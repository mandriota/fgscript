// Copyright (C) 2024  Mark Mandriota
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

// Package fgscript implements FGScript macro system execution and generation of Flowgorithm file.
package fgscript

import (
	"fmt"
	"slices"
	"strings"
)

func Tokenize(src string) []string {
	tokens := []string{}
	currentToken := strings.Builder{}

	inQuotes := false
	inParens := false

	for _, r := range src {
		switch r {
		case '"':
			currentToken.WriteRune(r)

			if inQuotes {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
				inQuotes = false
			} else {
				inQuotes = true
			}
		case '(':
			if !inQuotes {
				inParens = true
				currentToken.WriteRune(r)
			} else {
				currentToken.WriteRune(r)
			}
		case ')':
			if inParens && !inQuotes {
				currentToken.WriteRune(r)
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
				inParens = false
			} else {
				currentToken.WriteRune(r)
			}
		case ' ', '\t':
			if inQuotes || inParens {
				currentToken.WriteRune(r)
			} else if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
		default:
			currentToken.WriteRune(r)
		}
	}

	if currentToken.Len() > 0 {
		tokens = append(tokens, currentToken.String())
	}
	return tokens
}

type Generator struct {
	sb strings.Builder

	stStack []string
}

func (g *Generator) String() string {
	return g.sb.String()
}

func (g *Generator) TopLevel() bool {
	return len(g.stStack) == 0
}

func (g *Generator) WriteHeader() {
	g.sb.WriteString("<?xml version=\"1.0\"?><flowgorithm fileversion=\"4.2\">")
}

func (g *Generator) WriteFooter() error {
	if len(g.stStack) != 0 {
		return fmt.Errorf("not closed statement")
	}

	g.sb.WriteString("</flowgorithm>\n")
	return nil
}

func (g *Generator) MatchAndWrite(tokens []string) error {
	if g.TopLevel() && !slices.Contains([]string{"fn", "#"}, tokens[0]) {
		return fmt.Errorf("statement \"%s\" is not allowed outside of function\n", tokens[0])
	}
	
	switch tokens[0] {
	case "fn":
		return g.WriteStFunction(tokens)
	case "if":
		return g.WriteStIf(tokens)
	case "else":
		return g.WriteStElse(tokens)
	case "while":
		return g.WriteStWhile(tokens)
	case "do":
		return g.WriteStDo(tokens)
	case "for":
		return g.WriteStFor(tokens)
	case "end":
		return g.WriteEndSt(tokens)
	case "var":
		return g.WriteLnStDeclaration(tokens)
	case "set":
		return g.WriteLnStAssignment(tokens)
	case "call":
		return g.WriteLnStCall(tokens)
	case "print":
		return g.WriteLnStPrint(tokens, false)
	case "println":
		return g.WriteLnStPrint(tokens, true)
	case "scan":
		return g.WriteLnStScan(tokens)
	case "#":
		return g.WriteComment(tokens)
	default:
		return fmt.Errorf("unknown command \"%s\"", tokens[0])
	}
}

func (g *Generator) WriteStFunction(tokens []string) error {
	if len(tokens) != 5 {
		return fmt.Errorf("\"%s\" expects 4 arguments (found %d)", tokens[0], len(tokens)-1)
	}

	fnName := tokens[1]
	fnArgs := tokens[2]
	retVarName := tokens[3]
	retVarType := tokens[4]

	if len(g.stStack) != 0 {
		return fmt.Errorf("function \"%s\" is nestedly declared", fnName)
	}

	if retVarName == "_" {
		retVarName = ""
	}

	g.sb.WriteString(fmt.Sprintf(
		"<function name=\"%s\" type=\"%s\" variable=\"%s\">",
		fnName,
		retVarType,
		retVarName,
	))

	if len(fnArgs) < 2 || fnArgs[0] != '(' || fnArgs[len(fnArgs)-1] != ')' {
		return fmt.Errorf("function \"%s\" has wrong arguments format", fnName)
	}

	g.stStack = append(g.stStack, "</body></function>")

	fnArgsTokens := Tokenize(fnArgs[1 : len(fnArgs)-1])

	if len(fnArgsTokens) == 0 {
		g.sb.WriteString("<parameters/><body>")
		return nil
	}

	if len(fnArgsTokens)%2 != 0 {
		return fmt.Errorf("function \"%s\" has wrong arguments format", fnName)
	}

	g.sb.WriteString("<parameters>")

	for i := 0; i < len(fnArgsTokens); i += 2 {
		g.sb.WriteString(fmt.Sprintf(
			"<parameter name=\"%s\" type=\"%s\" array=\"False\"/>",
			fnArgsTokens[i],
			fnArgsTokens[i+1],
		))
	}

	g.sb.WriteString("</parameters><body>")
	return nil
}

func (g *Generator) WriteStIf(tokens []string) error {
	if len(tokens) < 2 {
		return fmt.Errorf("\"%s\" expects expression", tokens[0])
	}

	g.sb.WriteString(fmt.Sprintf("<if expression=\"%s\"><then>", strings.Join(tokens[1:], " ")))
	g.stStack = append(g.stStack, "</then><else/></if>")
	return nil
}

func (g *Generator) WriteStElse(tokens []string) error {
	if !strings.HasPrefix(g.stStack[len(g.stStack)-1], "</then><else/></if>") {
		return fmt.Errorf("\"%s\" must be preceded by an if", tokens[0])
	}

	if len(tokens) > 1 && tokens[1] != "if" {
		return fmt.Errorf("expected 0 arguments (found %d)", len(tokens)-1)
	}

	g.sb.WriteString("</then><else>")

	if len(tokens) > 1 {
		g.WriteStIf(tokens[1:])
		g.stStack = g.stStack[:len(g.stStack)-1]
		g.stStack[len(g.stStack)-1] += "</else></if>"
	} else {
		g.stStack[len(g.stStack)-1] = "</else></if>" +
			strings.TrimPrefix(g.stStack[len(g.stStack)-1], "</then><else/></if>")
	}
	return nil
}

func (g *Generator) WriteStWhile(tokens []string) error {
	if len(tokens) < 2 {
		return fmt.Errorf("\"%s\" expects expression", tokens[0])
	}

	g.sb.WriteString(fmt.Sprintf("<while expression=\"%s\">", strings.Join(tokens[1:], " ")))
	g.stStack = append(g.stStack, "</while>")
	return nil
}

func (g *Generator) WriteStDo(tokens []string) error {
	if len(tokens) < 2 {
		return fmt.Errorf("\"%s\" expects expression", tokens[0])
	}

	g.sb.WriteString(fmt.Sprintf("<do expression=\"%s\">", strings.Join(tokens[1:], " ")))
	g.stStack = append(g.stStack, "</do>")
	return nil
}

func (g *Generator) WriteStFor(tokens []string) error {
	if len(tokens) < 8 {
		return fmt.Errorf("\"%s\" expects expression", tokens[0])
	}

	tokenPos := 1

	direction := "inc"
	if tokens[tokenPos] == "backward" {
		direction = "dec"
		tokenPos++
	}

	varName := tokens[tokenPos]
	tokenPos++

	if tokens[tokenPos] != "from" {
		return fmt.Errorf("\"%s\" expects expression \"from\"", tokens[0])
	}
	tokenPos++

	fromExprBegPos := tokenPos
	for tokenPos < len(tokens) && tokens[tokenPos] != "to" {
		tokenPos++
	}
	fromExprEndPos := tokenPos

	if tokenPos >= len(tokens)-1 || tokens[tokenPos] != "to" {
		return fmt.Errorf("\"%s\" expects expression \"to\"", tokens[0])
	}

	toExprBegPos := tokenPos + 1
	for tokenPos < len(tokens) && tokens[tokenPos] != "step" {
		tokenPos++
	}
	toExprEndPos := tokenPos

	if tokenPos >= len(tokens)-1 || tokens[tokenPos] != "step" {
		return fmt.Errorf("\"%s\" expects expression \"step\"", tokens[0])
	}

	stepExprBegPos := tokenPos + 1

	g.sb.WriteString(fmt.Sprintf(
		"<for variable=\"%s\" start=\"%s\" end=\"%s\" direction=\"%s\" step=\"%s\">",
		varName,
		strings.Join(tokens[fromExprBegPos:fromExprEndPos], " "),
		strings.Join(tokens[toExprBegPos:toExprEndPos], " "),
		direction,
		strings.Join(tokens[stepExprBegPos:], " "),
	))
	g.stStack = append(g.stStack, "</for>")
	return nil
}

func (g *Generator) WriteEndSt(tokens []string) error {
	if len(tokens) > 1 {
		return fmt.Errorf("\"%s\" expects 0 arguments (found %d)", tokens[0], len(tokens)-1)
	}

	g.sb.WriteString(g.stStack[len(g.stStack)-1])
	g.stStack = g.stStack[:len(g.stStack)-1]
	return nil
}

func (g *Generator) WriteLnStDeclaration(tokens []string) error {
	if len(tokens) < 3 {
		return fmt.Errorf("\"%s\" expects at least 2 arguments (found %d)", tokens[0], len(tokens)-1)
	}

	g.sb.WriteString(fmt.Sprintf(
		"<declare name=\"%s\" type=\"%s\" array=\"False\" size=\"\"/>",
		strings.Join(tokens[1:len(tokens)-1], ", "),
		tokens[len(tokens)-1],
	))
	return nil
}

func (g *Generator) WriteLnStAssignment(tokens []string) error {
	if len(tokens) < 3 {
		return fmt.Errorf("\"%s\" expects at least 2 arguments (found %d)", tokens[0], len(tokens)-1)
	}

	g.sb.WriteString(fmt.Sprintf(
		"<assign variable=\"%s\" expression=\"%s\"/>",
		tokens[1],
		strings.Join(tokens[2:], " "),
	))
	return nil
}

func (g *Generator) WriteLnStCall(tokens []string) error {
	if len(tokens) < 2 {
		return fmt.Errorf("\"%s\" expects at least 1 argument (found %d)", tokens[0], len(tokens)-1)
	}

	g.sb.WriteString(fmt.Sprintf(
		"<call expression=\"%s\"/>",
		strings.Join(tokens[1:], " "),
	))
	return nil
}

func (g *Generator) WriteLnStPrint(tokens []string, newline bool) error {
	if len(tokens) < 2 {
		return fmt.Errorf("\"%s\" expects at least 1 argument (found %d)", tokens[0], len(tokens)-1)
	}

	newlineStringify := "False"
	if newline {
		newlineStringify = "True"
	}

	g.sb.WriteString(fmt.Sprintf(
		"<output expression=\"%s\" newline=\"%s\"/>",
		strings.ReplaceAll(strings.Join(tokens[1:], " &amp; "), "\"", "&quot;"),
		newlineStringify,
	))
	return nil
}

func (g *Generator) WriteLnStScan(tokens []string) error {
	if len(tokens) < 2 {
		return fmt.Errorf("\"%s\" expects at least 1 argument (found %d)", tokens[0], len(tokens)-1)
	}

	g.sb.WriteString(fmt.Sprintf("<input variable=\"%s\"/>", tokens[1]))
	return nil
}

func (g *Generator) WriteComment(tokens []string) error {
	if len(tokens) < 2 {
		return fmt.Errorf("\"%s\" expects at least 1 argument (found %d)", tokens[0], len(tokens)-1)
	}

	if len(g.stStack) == 0 {
		return nil
	}

	g.sb.WriteString(fmt.Sprintf(
		"<comment text=\"%s\"/>",
		strings.Join(tokens[1:], " "),
	))
	return nil
}
