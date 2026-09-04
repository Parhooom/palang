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

	symbolTable    SymbolTable
	isFuncReturned bool
	varsOffsets    uint
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

func (p *Parser) Primary() Type {
	switch p.currToken.Kind {
	case token.TK_INT:
		p.genIntLiteral(p.currToken.Text)
		p.nextToken()
		return TYPE_INT

	case token.TK_IDENT:
		name := p.currToken.Text

		symbol, exists := p.symbolTable.get(name)
		if !exists {
			fmt.Fprintf(os.Stderr, "ERROR: variable %s not defined\n", name)
			os.Exit(1)
		}

		switch symbol.Kind {
		case TYPE_INT:
			p.genIntSymbol(symbol)
		default:
			fmt.Fprintf(os.Stderr, "ERROR: symbol type is not supported\n")
			os.Exit(1)
		}

		p.nextToken()
		return symbol.Kind

	case token.TK_LPAREN:
		p.nextToken()
		exprType := p.Expression()
		p.expect(token.TK_RPAREN, ")")
		return exprType

	default:
		fmt.Fprintf(os.Stderr, "ERROR: expected primary expression, got %s\n", p.currToken.Text)
		os.Exit(1)
		return TYPE_VOID
	}
}

func (p *Parser) Factor() Type {
	leftType := p.Primary()

	for p.currToken.Kind == token.TK_MULT || p.currToken.Kind == token.TK_DIV {
		operator := p.currToken
		p.nextToken()
		rightType := p.Primary()
		p.genBinaryExpression(leftType, rightType, operator)
	}

	return leftType
}

func (p *Parser) Term() Type {
	leftType := p.Factor()

	for p.currToken.Kind == token.TK_PLUS || p.currToken.Kind == token.TK_MINUS {
		operator := p.currToken
		p.nextToken()
		rightType := p.Factor()
		p.genBinaryExpression(leftType, rightType, operator)
	}

	return leftType
}

func (p *Parser) Comparison() Type {
	leftType := p.Term()

	for p.currToken.Kind == token.TK_LT || p.currToken.Kind == token.TK_LTE || p.currToken.Kind == token.TK_GT || p.currToken.Kind == token.TK_GTE {
		operator := p.currToken
		p.nextToken()
		rightType := p.Term()
		p.genBinaryExpression(leftType, rightType, operator)
	}

	return leftType
}

func (p *Parser) Equality() Type {
	leftType := p.Comparison()

	for p.currToken.Kind == token.TK_EQ || p.currToken.Kind == token.TK_NEQ {
		operator := p.currToken
		p.nextToken()
		rightType := p.Comparison()
		p.genBinaryExpression(leftType, rightType, operator)
	}

	return leftType
}

func (p *Parser) Expression() Type {
	return p.Equality()
}

func (p *Parser) Type() Type {
	typeToken := p.expect(token.TK_BASIC, "type")

	if typeToken.Text == "int" {
		return TYPE_INT
	}

	return TYPE_VOID
}

// TODO: function params generation
func (p *Parser) FuncParam() {
	p.Type()
	p.expect(token.TK_IDENT, "identifier")
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

func (p *Parser) ReturnStatement() {
	p.expect(token.TK_RETURN, "return")

	if p.currToken.Kind != token.TK_SEMICOLON {
		returnType := p.Expression()
		p.genFuncEnd(returnType)
	} else {
		p.genFuncEnd(TYPE_VOID)
	}

	p.expect(token.TK_SEMICOLON, ";")
	p.isFuncReturned = true
}

func (p *Parser) VarDecl() {
	varType := p.Type()
	varToken := p.expect(token.TK_IDENT, "identifier")
	p.expect(token.TK_SEMICOLON, ";")

	varTypeOffset := offsetOf[varType]
	p.varsOffsets += varTypeOffset
	p.symbolTable.put(varToken.Text, newSymbol(varType, p.varsOffsets))
	p.genVarDecl(varTypeOffset)
}

func (p *Parser) VarAssign() {
	varToken := p.expect(token.TK_IDENT, "identifier")
	p.expect(token.TK_ASSIGN, "=")
	exprType := p.Expression()
	p.expect(token.TK_SEMICOLON, ";")

	symbol, exists := p.symbolTable.get(varToken.Text)
	if !exists {
		fmt.Fprintln(os.Stderr, "ERROR: variable not defined")
		os.Exit(1)
	}
	if symbol.Kind != exprType {
		fmt.Fprintln(os.Stderr, "ERROR: type mismatch")
		os.Exit(1)
	}

	p.genVarAssign(symbol)
}

func (p *Parser) Statement() {
	switch p.currToken.Kind {
	case token.TK_RETURN:
		p.ReturnStatement()
	case token.TK_BASIC:
		p.VarDecl()
	case token.TK_IDENT:
		p.VarAssign()
	default:
		fmt.Fprintln(os.Stderr, "ERROR: statement not implemented")
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
	p.symbolTable = SymbolTable{Table: make(map[string]Symbol)}
	p.isFuncReturned = false
	p.varsOffsets = 0

	p.expect(token.TK_FUNC, "func")

	funcName := p.expect(token.TK_IDENT, "function name")
	p.genFuncLabel(funcName.Text)
	p.genFuncStart()

	p.FuncParams()
	if p.currToken.Kind != token.TK_LCURLY {
		p.Type()
	}

	p.Block()

	if !p.isFuncReturned {
		p.genFuncEnd(TYPE_VOID)
	}

	// fmt.Printf("%+v\n", p.symbolTable)
}
