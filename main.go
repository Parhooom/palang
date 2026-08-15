package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"palang/lexer"
	"palang/parser"
)

func main() {
	args := os.Args

	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "ERROR: Input is not provided")
		usage(os.Stderr, args[0])
		os.Exit(1)
	}

	f, err := os.Open(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())
		usage(os.Stderr, args[0])
		os.Exit(1)
	}
	defer f.Close()

	l := lexer.New(bufio.NewReader(f))

	// for tok := l.NextToken(); tok.Kind != token.TK_EOF; tok = l.NextToken() {
	// 	fmt.Println(tok.Text)
	// }

	p := parser.New(l)
	p.Program()
}

func usage(w io.Writer, program string) {
	fmt.Fprintf(w, "Usage: %s <input.palang>\n", program)
}
