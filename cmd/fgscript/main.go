package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/mandriota/fgscript"
)

func main() {
	l := log.New(os.Stderr, "", 0)

	if len(os.Args) != 2 {
		l.Fatalln("file name expected")
	}

	fs, err := os.Open(os.Args[1])
	if err != nil {
		l.Fatalln("error while openning file")
	}
	defer fs.Close()

	gen := fgscript.Generator{}
	gen.WriteHeader()

	s := bufio.NewScanner(fs)

	lineCount := 0

	for s.Scan() {
		line := s.Text()
		lineCount++

		tokens := fgscript.Tokenize(line)
		if len(tokens) == 0 {
			continue
		}

		err := (error)(nil)

		switch action := tokens[0]; action {
		case "fn":
			err = gen.WriteStFunction(tokens)
		case "if":
			err = gen.WriteStIf(tokens)
		case "else":
			err = gen.WriteStElse(tokens)
		case "while":
			err = gen.WriteStWhile(tokens)
		case "do":
			err = gen.WriteStDo(tokens)
		case "for":
			err = gen.WriteStFor(tokens)
		case "end":
			err = gen.WriteEndSt(tokens)
		case "var":
			err = gen.WriteLnStDeclaration(tokens)
		case "set":
			err = gen.WriteLnStAssignment(tokens)
		case "print":
			err = gen.WriteLnStPrint(tokens, false)
		case "println":
			err = gen.WriteLnStPrint(tokens, true)
		case "scan":
			err = gen.WriteLnStScan(tokens)
		case "#":
			err = gen.WriteComment(tokens)
		default:
			err = fmt.Errorf("unknown command \"%s\"", action)
		}

		if err != nil {
			l.Fatalf("line %d: %v\n", lineCount, err)
		}
		l.Println(len(tokens), tokens)
	}

	if err := gen.WriteFooter(); err != nil {
		l.Fatalf("line %d: %v\n", lineCount, err)
	}
	os.Stdout.WriteString(gen.String())
}
