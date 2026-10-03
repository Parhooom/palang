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
	peekToken token.Token

	symbolTable     SymbolTable
	functionTable   FuncTable
	stringTable     StringTable
	currFuncRetType Type
	varsOffsets     uint
	localLabelNum   uint
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:             l,
		functionTable: make(FuncTable),
		stringTable:   StringTable{table: make(map[string]string)},
	}

	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.currToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func report(tok token.Token, msg string) {
	fmt.Fprintf(os.Stderr, "%s: error: %s\n", tok.Loc(), msg)
	os.Exit(1)
}

func (p *Parser) expect(tk token.TokenKind, msg string) token.Token {
	if p.currToken.Kind != tk {
		report(p.currToken, fmt.Sprintf("expected '%s', got '%s'", msg, p.currToken.Text))
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

	case token.TK_STRING:
		label := p.stringTable.getOrInsert(p.currToken.Text)
		p.genStringLiteral(label)
		p.nextToken()
		return TYPE_STRING

	case token.TK_IDENT:
		if p.peekToken.Kind == token.TK_LPAREN {
			funcName, returnType := p.FuncCall()
			if returnType == TYPE_VOID {
				report(
					funcName,
					fmt.Sprintf("void function '%s' cannot be used in expression", funcName.Text),
				)
			}

			p.genFuncCallReturn()
			return returnType
		} else {
			name := p.currToken.Text

			symbol, exists := p.symbolTable.get(name)
			if !exists {
				report(p.currToken, fmt.Sprintf("variable '%s' not defined", name))
			}
			p.genSymbol(symbol)

			p.nextToken()
			return symbol.Kind
		}

	case token.TK_LPAREN:
		p.nextToken()
		exprType := p.Expression()
		p.expect(token.TK_RPAREN, ")")
		return exprType

	default:
		report(p.currToken, fmt.Sprintf("expected primary expression, got '%s'", p.currToken.Text))
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
	case "string":
		return TYPE_STRING

	default:
		report(typeToken, fmt.Sprintf("unknown type '%s'", typeToken.Text))
		return TYPE_VOID
	}
}

func (p *Parser) printFunc(argsTypes []Type, name token.Token) {
	if len(argsTypes) != 1 {
		report(name, fmt.Sprintf("function 'print' expected 1 argument, got %d", len(argsTypes)))
	}

	argType := argsTypes[0]
	if argType != TYPE_INT && argType != TYPE_BOOL && argType != TYPE_STRING {
		report(name, "function 'print' cannot print this type of expression")
	}

	p.genPrintCall(argType)
}

func (p *Parser) FuncCall() (token.Token, Type) {
	name := p.expect(token.TK_IDENT, "function name")

	p.expect(token.TK_LPAREN, "(")

	var argsTypes []Type
	if p.currToken.Kind != token.TK_RPAREN {
		argsTypes = append(argsTypes, p.Expression())

		for p.currToken.Kind == token.TK_COMMA {
			p.nextToken()
			argsTypes = append(argsTypes, p.Expression())
		}
	}

	p.expect(token.TK_RPAREN, ")")

	if name.Text == "print" {
		p.printFunc(argsTypes, name)
		return name, TYPE_VOID
	}

	fn, exists := p.functionTable[name.Text]
	if !exists {
		report(name, fmt.Sprintf("function '%s' is not defined", name.Text))
	}

	if len(argsTypes) != len(fn.ParamsTypes) {
		report(
			name,
			fmt.Sprintf("function '%s' expected %d arguments, got %d",
				name.Text, len(fn.ParamsTypes), len(argsTypes)),
		)
	}

	for i, expectedType := range fn.ParamsTypes {
		if argsTypes[i] != expectedType {
			report(name, fmt.Sprintf("argument %d of '%s' type mismatch", i+1, name.Text))
		}
	}

	p.genFuncCall(name.Text, len(argsTypes))
	return name, fn.ReturnType
}

func (p *Parser) FuncParam(paramIndex int, funcName token.Token) Type {
	if paramIndex >= len(argRegs) {
		report(funcName, fmt.Sprintf("only %d parameters are supported", len(argRegs)))
	}

	paramType := p.Type()
	paramName := p.expect(token.TK_IDENT, "identifier")

	paramOffset := paramType.Size()
	p.varsOffsets += paramOffset
	p.symbolTable.put(paramName, newSymbol(paramType, p.varsOffsets))
	p.genFuncParam(paramOffset, p.varsOffsets, argRegs[paramIndex])

	return paramType
}

func (p *Parser) FuncParams(funcName token.Token) []Type {
	p.expect(token.TK_LPAREN, "(")

	var paramsTypes []Type
	paramIndex := 0

	if p.currToken.Kind != token.TK_RPAREN {
		paramsTypes = append(paramsTypes, p.FuncParam(paramIndex, funcName))
		paramIndex += 1

		for p.currToken.Kind == token.TK_COMMA {
			p.nextToken()
			paramsTypes = append(paramsTypes, p.FuncParam(paramIndex, funcName))
			paramIndex += 1
		}
	}

	p.expect(token.TK_RPAREN, ")")
	return paramsTypes
}

func (p *Parser) ReturnStatement() {
	returnToken := p.expect(token.TK_RETURN, "return")

	if p.currToken.Kind != token.TK_SEMICOLON {
		if p.currFuncRetType == TYPE_VOID {
			report(returnToken, "void function cannot return a value")
		}

		returnType := p.Expression()
		if returnType != p.currFuncRetType {
			report(returnToken, "return type mismatch")
		}

		p.genReturnStatement(returnType)
	} else {
		if p.currFuncRetType != TYPE_VOID {
			report(returnToken, "function must return a value")
		}

		p.genReturnStatement(TYPE_VOID)
	}

	p.expect(token.TK_SEMICOLON, ";")
}

func (p *Parser) VarDecl() {
	varType := p.Type()
	varToken := p.expect(token.TK_IDENT, "identifier")
	p.expect(token.TK_SEMICOLON, ";")

	varTypeOffset := varType.Size()
	p.varsOffsets += varTypeOffset
	p.symbolTable.put(varToken, newSymbol(varType, p.varsOffsets))
	p.genVarDecl(varTypeOffset)
}

func (p *Parser) VarAssign() {
	varToken := p.expect(token.TK_IDENT, "identifier")
	p.expect(token.TK_ASSIGN, "=")
	exprType := p.Expression()
	p.expect(token.TK_SEMICOLON, ";")

	symbol, exists := p.symbolTable.get(varToken.Text)
	if !exists {
		report(varToken, fmt.Sprintf("variable '%s' not defined", varToken.Text))
	}
	if symbol.Kind != exprType {
		report(varToken, "type mismatch")
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
		if p.peekToken.Kind == token.TK_LPAREN {
			p.FuncCall()
			p.expect(token.TK_SEMICOLON, ";")
		} else {
			p.VarAssign()
		}
	case token.TK_IF:
		p.IfStatement()
	case token.TK_WHILE:
		p.WhileStatement()

	default:
		report(p.currToken, "statement not supported")
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

	paramsTypes := p.FuncParams(funcName)

	returnType := TYPE_VOID
	if p.currToken.Kind != token.TK_LCURLY {
		returnType = p.Type()
	}
	p.currFuncRetType = returnType
	p.functionTable.put(funcName, newFuncSymbol(returnType, paramsTypes))

	p.Block()
	p.genFuncEnd()
}
