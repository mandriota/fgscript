package main

import (
	"bufio"
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

		if err := gen.MatchAndWrite(tokens); err != nil {
			l.Fatalf("line %d: %v\n", lineCount, err)
		}
		l.Println(len(tokens), tokens)
	}

	if err := gen.WriteFooter(); err != nil {
		l.Fatalf("line %d: %v\n", lineCount, err)
	}
	os.Stdout.WriteString(gen.String())
}
