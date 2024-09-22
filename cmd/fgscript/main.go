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
