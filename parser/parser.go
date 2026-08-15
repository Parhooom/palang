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

func (p *Parser) Program() {
	p.genProgramStart()

	for p.currToken.Kind != token.TK_EOF {
		p.Func()
	}

	p.genProgramEnd()
}

func (p *Parser) FuncParam() {
	p.expect(token.TK_IDENT, "identifier")
	p.expect(token.TK_BASIC, "int type")
}

func (p *Parser) FuncParams() {
	p.expect(token.TK_LPAREN, "(")
	if p.currToken.Kind != token.TK_RPAREN {
		p.FuncParam()
		for p.currToken.Kind == token.TK_COMMA {
			p.nextToken()
			p.FuncParam()
		}
	}
	p.expect(token.TK_RPAREN, ")")
}

func (p *Parser) Type() uint8 {
	retType := p.expect(token.TK_BASIC, "type")

	if retType.Text == "int" {
		return TYPE_INT
	}

	return TYPE_VOID
}

func (p *Parser) ReturnStatement() {
	p.expect(token.TK_RETURN, "return")
	if p.currToken.Kind != token.TK_SEMICOLON {
		statusCode := p.expect(token.TK_INT, "integer")
		p.genFuncEnd(statusCode.Text)
	}
	p.expect(token.TK_SEMICOLON, ";")
}

func (p *Parser) Statement() {
	if p.currToken.Kind == token.TK_RETURN {
		p.ReturnStatement()
	} else {
		fmt.Fprintln(os.Stderr, "ERROR: not implemented")
		os.Exit(1)
	}
}

func (p *Parser) Statements() {
	for p.currToken.Kind != token.TK_RCURLY && p.currToken.Kind != token.TK_EOF {
		p.Statement()
	}
}

func (p *Parser) Block() {
	p.expect(token.TK_LCURLY, "{")
	p.Statements()
	p.expect(token.TK_RCURLY, "}")
}

func (p *Parser) Func() {
	p.expect(token.TK_FUNC, "func")

	funcName := p.expect(token.TK_IDENT, "function name")
	p.genFuncLabel(funcName.Text)

	p.FuncParams()
	if p.currToken.Kind != token.TK_LCURLY {
		p.Type()
	}

	p.Block()
}
