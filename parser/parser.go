package parser

import (
	"fmt"
	"os"

	"palang/lexer"
	"palang/token"
)

type Parser struct {
	l         *lexer.Lexer
	currToken token.Token
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l: l,
	}

	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.currToken = p.l.NextToken()
}

func (p *Parser) expect(tk token.TokenKind, msg string) token.Token {
	if p.currToken.Kind != tk {
		fmt.Fprintf(os.Stderr, "ERROR: Expected '%s', got '%s'\n", msg, p.currToken.Text)
		os.Exit(1)
	}

	currToken := p.currToken
	p.nextToken()
	return currToken
}

func (p *Parser) ParseProgram() {
	fmt.Printf("format ELF64\n")
	fmt.Printf("section \".text\" executable\n")

	for p.currToken.Kind != token.TK_EOF {
		p.expect(token.TK_FUNC, "func")
		ident := p.expect(token.TK_IDENT, "identifier")
		fmt.Printf("%s:\n", ident.Text)
		if ident.Text == "main" {
			fmt.Printf("    public main\n")
		}

		p.expect(token.TK_LPAREN, "(")
		p.expect(token.TK_RPAREN, ")")
		p.expect(token.TK_BASIC, "basic type")
		p.expect(token.TK_LCURLY, "{")
		p.expect(token.TK_RETURN, "return")
		num := p.expect(token.TK_INT, "integer")
		fmt.Printf("    mov rax, %s\n", num.Text)
		fmt.Printf("    ret\n")

		p.expect(token.TK_SEMICOLON, ";")
		p.expect(token.TK_RCURLY, "}")
	}

	fmt.Printf("section \".note.GNU-stack\"\n")
}
