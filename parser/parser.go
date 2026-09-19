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

	symbolTable   SymbolTable
	varsOffsets   uint
	localLabelNum uint
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

func (p *Parser) newLocalLabel() string {
	p.localLabelNum += 1
	return fmt.Sprintf(".label%d", p.localLabelNum)
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

	case token.TK_TRUE:
		p.genBoolLiteral(true)
		p.nextToken()
		return TYPE_BOOL

	case token.TK_FALSE:
		p.genBoolLiteral(false)
		p.nextToken()
		return TYPE_BOOL

	case token.TK_IDENT:
		name := p.currToken.Text

		symbol, exists := p.symbolTable.get(name)
		if !exists {
			fmt.Fprintf(os.Stderr, "ERROR: variable %s not defined\n", name)
			os.Exit(1)
		}
		p.genSymbol(symbol)

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
		leftType = p.genBinaryExpression(leftType, rightType, operator)
	}

	return leftType
}

func (p *Parser) Term() Type {
	leftType := p.Factor()

	for p.currToken.Kind == token.TK_PLUS || p.currToken.Kind == token.TK_MINUS {
		operator := p.currToken
		p.nextToken()
		rightType := p.Factor()
		leftType = p.genBinaryExpression(leftType, rightType, operator)
	}

	return leftType
}

func (p *Parser) Comparison() Type {
	leftType := p.Term()

	for p.currToken.Kind == token.TK_LT || p.currToken.Kind == token.TK_LTE || p.currToken.Kind == token.TK_GT || p.currToken.Kind == token.TK_GTE {
		operator := p.currToken
		p.nextToken()
		rightType := p.Term()
		leftType = p.genBinaryExpression(leftType, rightType, operator)
	}

	return leftType
}

func (p *Parser) Equality() Type {
	leftType := p.Comparison()

	for p.currToken.Kind == token.TK_EQ || p.currToken.Kind == token.TK_NEQ {
		operator := p.currToken
		p.nextToken()
		rightType := p.Comparison()
		leftType = p.genBinaryExpression(leftType, rightType, operator)
	}

	return leftType
}

func (p *Parser) Expression() Type {
	return p.Equality()
}

func (p *Parser) Type() Type {
	typeToken := p.expect(token.TK_BASIC, "type")

	switch typeToken.Text {
	case "int":
		return TYPE_INT
	case "bool":
		return TYPE_BOOL

	default:
		fmt.Fprintf(os.Stderr, "ERROR: unknown type %s\n", typeToken.Text)
		os.Exit(1)
		return TYPE_VOID
	}
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
		p.genReturnStatement(returnType)
	} else {
		p.genReturnStatement(TYPE_VOID)
	}

	p.expect(token.TK_SEMICOLON, ";")
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

func (p *Parser) IfStatement() {
	p.expect(token.TK_IF, "if")

	p.expect(token.TK_LPAREN, "(")
	p.Expression()
	p.expect(token.TK_RPAREN, ")")

	elseLabel := p.newLocalLabel()
	p.genJz(elseLabel)

	p.Block()

	if p.currToken.Kind == token.TK_ELSE {
		p.nextToken()

		endLabel := p.newLocalLabel()
		p.genJmp(endLabel)
		p.genLocalLabel(elseLabel)

		if p.currToken.Kind == token.TK_IF {
			p.IfStatement()
		} else {
			p.Block()
		}

		p.genLocalLabel(endLabel)
	} else {
		p.genLocalLabel(elseLabel)
	}
}

func (p *Parser) WhileStatement() {
	p.expect(token.TK_WHILE, "while")

	startLable := p.newLocalLabel()
	endLabel := p.newLocalLabel()

	p.genLocalLabel(startLable)

	p.expect(token.TK_LPAREN, "(")
	p.Expression()
	p.expect(token.TK_RPAREN, ")")

	p.genJz(endLabel)

	p.Block()

	p.genJmp(startLable)
	p.genLocalLabel(endLabel)
}

func (p *Parser) Statement() {
	switch p.currToken.Kind {
	case token.TK_RETURN:
		p.ReturnStatement()
	case token.TK_BASIC:
		p.VarDecl()
	case token.TK_IDENT:
		p.VarAssign()
	case token.TK_IF:
		p.IfStatement()
	case token.TK_WHILE:
		p.WhileStatement()

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
	p.varsOffsets = 0
	p.localLabelNum = 0

	p.expect(token.TK_FUNC, "func")

	funcName := p.expect(token.TK_IDENT, "function name")
	p.genFuncLabel(funcName.Text)
	p.genFuncStart()

	p.FuncParams()
	if p.currToken.Kind != token.TK_LCURLY {
		p.Type()
	}

	p.Block()
	p.genFuncEnd()
}
