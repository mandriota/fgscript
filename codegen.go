package fgscript

import (
	"fmt"
	"strings"
)

func Tokenize(src string) []string {
	tokens := []string{}
	currentToken := strings.Builder{}

	inQuotes := false
	inParens := false

	for i := 0; i < len(src); i++ {
		char := src[i]
		switch char {
		case '"':
			currentToken.WriteByte(char)

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
				currentToken.WriteByte(char)
			} else {
				currentToken.WriteByte(char)
			}
		case ')':
			if inParens && !inQuotes {
				currentToken.WriteByte(char)
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
				inParens = false
			} else {
				currentToken.WriteByte(char)
			}
		case ' ', '\t':
			if inQuotes || inParens {
				currentToken.WriteByte(char)
			} else if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
		default:
			currentToken.WriteByte(char)
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

func (g *Generator) WriteStFunction(tokens []string) error {
	if len(tokens) != 5 {
		return fmt.Errorf("expected 5 words (found %d)", len(tokens))
	}

	fnName := tokens[1]
	fnArgs := tokens[2]
	retVarName := tokens[3]
	retVarType := tokens[4]

	if retVarName == "_" {
		retVarName = ""
	}

	g.sb.WriteString(fmt.Sprintf("<function name=\"%s\" type=\"%s\" variable=\"%s\">", fnName, retVarType, retVarName))

	if len(fnArgs) < 2 || fnArgs[0] != '(' || fnArgs[len(fnArgs)-1] != ')' {
		return fmt.Errorf("wrong arguments format for function \"%s\"", fnName)
	}

	g.stStack = append(g.stStack, "</body></function>")

	fnArgsTokens := Tokenize(fnArgs[1 : len(fnArgs)-1])

	if len(fnArgsTokens) == 0 {
		g.sb.WriteString("<parameters/><body>")
		return nil
	}

	if len(fnArgsTokens)%2 != 0 {
		return fmt.Errorf("wrong arguments format for function \"%s\"", fnName)
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
		return fmt.Errorf("wrong if statement format: expected expression")
	}

	g.sb.WriteString(fmt.Sprintf("<if expression=\"%s\"><then>", strings.Join(tokens[1:], " ")))
	g.stStack = append(g.stStack, "</then><else/></if>")
	return nil
}

func (g *Generator) WriteStElse(tokens []string) error {
	if len(g.stStack) == 0 || !strings.HasPrefix(g.stStack[len(g.stStack)-1], "</then><else/></if>") {
		return fmt.Errorf("else block found but if block not found")
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
		g.stStack[len(g.stStack)-1] = "</else></if>" + strings.TrimPrefix(g.stStack[len(g.stStack)-1], "</then><else/></if>")
	}
	return nil
}

func (g *Generator) WriteStWhile(tokens []string) error {
	if len(tokens) < 2 {
		return fmt.Errorf("wrong while statement format: expected expression")
	}

	g.sb.WriteString(fmt.Sprintf("<while expression=\"%s\">", strings.Join(tokens[1:], " ")))
	g.stStack = append(g.stStack, "</while>")
	return nil
}

func (g *Generator) WriteStDo(tokens []string) error {
	if len(tokens) < 2 {
		return fmt.Errorf("wrong do statement format: expected expression")
	}

	g.sb.WriteString(fmt.Sprintf("<do expression=\"%s\">", strings.Join(tokens[1:], " ")))
	g.stStack = append(g.stStack, "</do>")
	return nil
}

func (g *Generator) WriteStFor(tokens []string) error {
	if len(tokens) < 8 {
		return fmt.Errorf("wrong do statement format: expected expression")
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
		return fmt.Errorf("wrong for statement format: expected expression from")
	}
	tokenPos++

	fromExprBegPos := tokenPos
	for tokenPos < len(tokens) && tokens[tokenPos] != "to" {
		tokenPos++
	}
	fromExprEndPos := tokenPos

	if tokenPos >= len(tokens)-1 || tokens[tokenPos] != "to" {
		return fmt.Errorf("wrong for statement format: expected expression to")
	}

	toExprBegPos := tokenPos + 1
	for tokenPos < len(tokens) && tokens[tokenPos] != "step" {
		tokenPos++
	}
	toExprEndPos := tokenPos

	if tokenPos >= len(tokens)-1 || tokens[tokenPos] != "step" {
		return fmt.Errorf("wrong for statement format: expected expression step")
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
	if len(g.stStack) == 0 {
		return fmt.Errorf("unexpected end of statement")
	}

	if len(tokens) > 1 {
		return fmt.Errorf("expected 0 arguments (found %d)", len(tokens)-1)
	}

	g.sb.WriteString(g.stStack[len(g.stStack)-1])
	g.stStack = g.stStack[:len(g.stStack)-1]
	return nil
}

func (g *Generator) WriteLnStDeclaration(tokens []string) error {
	if len(tokens) < 3 {
		return fmt.Errorf("expected at least 3 words (found %d)", len(tokens))
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
		return fmt.Errorf("expected at least 3 words")
	}

	g.sb.WriteString(fmt.Sprintf(
		"<assign variable=\"%s\" expression=\"%s\"/>",
		tokens[1],
		strings.Join(tokens[2:], " "),
	))
	return nil
}

func (g *Generator) WriteLnStPrint(tokens []string, newline bool) error {
	if len(tokens) < 2 {
		return fmt.Errorf("expected at least 2 words")
	}

	g.sb.WriteString(fmt.Sprintf(
		"<output expression=\"%s\" newline=\"%s\"/>",
		strings.ReplaceAll(strings.Join(tokens[1:], " &amp; "), "\"", "&quot;"),
		map[bool]string{
			true:  "True",
			false: "False",
		}[newline],
	))
	return nil
}

func (g *Generator) WriteLnStScan(tokens []string) error {
	if len(tokens) < 2 {
		return fmt.Errorf("expected at least 2 words")
	}

	g.sb.WriteString(fmt.Sprintf("<input variable=\"%s\"/>", tokens[1]))
	return nil
}

func (g *Generator) WriteComment(tokens []string) error {
	if len(tokens) < 2 {
		return fmt.Errorf("expected at least 2 words")
	}

	g.sb.WriteString(fmt.Sprintf(
		"<comment text=\"%s\"/>",
		strings.Join(tokens[1:], " "),
	))
	return nil
}
