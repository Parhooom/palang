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

	case token.TK_STRING:
		label := p.stringTable.getOrInsert(p.currToken.Text)
		p.genStringLiteral(label)
		p.nextToken()
		return TYPE_STRING

	case token.TK_IDENT:
		if p.peekToken.Kind == token.TK_LPAREN {
			funcName, returnType := p.FuncCall()
			if returnType == TYPE_VOID {
				fmt.Fprintf(os.Stderr, "ERROR: void function '%s' cannot be used in expression\n",
					funcName.Text)
				os.Exit(1)
			}

			p.genFuncCallReturn()
			return returnType
		} else {
			name := p.currToken.Text

			symbol, exists := p.symbolTable.get(name)
			if !exists {
				fmt.Fprintf(os.Stderr, "ERROR: variable '%s' not defined\n", name)
				os.Exit(1)
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
		fmt.Fprintf(os.Stderr, "ERROR: expected primary expression, got '%s'\n", p.currToken.Text)
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
	case "string":
		return TYPE_STRING

	default:
		fmt.Fprintf(os.Stderr, "ERROR: unknown type '%s'\n", typeToken.Text)
		os.Exit(1)
		return TYPE_VOID
	}
}

func (p *Parser) printFunc(argsTypes []Type) {
	if len(argsTypes) != 1 {
		fmt.Fprintf(os.Stderr, "ERROR: function 'print' expected 1 argument, got %d\n",
			len(argsTypes))
		os.Exit(1)
	}

	argType := argsTypes[0]
	if argType != TYPE_INT && argType != TYPE_BOOL && argType != TYPE_STRING {
		fmt.Fprintf(os.Stderr, "ERROR: function 'print' cannot print this type of expression\n")
		os.Exit(1)
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
		p.printFunc(argsTypes)
		return name, TYPE_VOID
	}

	fn, exists := p.functionTable[name.Text]
	if !exists {
		fmt.Fprintf(os.Stderr, "ERROR: function '%s' is not defined\n", name.Text)
		os.Exit(1)
	}

	if len(argsTypes) != len(fn.ParamsTypes) {
		fmt.Fprintf(os.Stderr, "ERROR: function '%s' expected %d arguments, got %d\n",
			name.Text, len(fn.ParamsTypes), len(argsTypes))
		os.Exit(1)
	}

	for i, expectedType := range fn.ParamsTypes {
		if argsTypes[i] != expectedType {
			fmt.Fprintf(os.Stderr, "ERROR: argument %d of '%s' type mismatch\n", i+1, name.Text)
			os.Exit(1)
		}
	}

	p.genFuncCall(name.Text, len(argsTypes))
	return name, fn.ReturnType
}

func (p *Parser) FuncParam(paramIndex int) Type {
	if paramIndex >= len(argRegs) {
		fmt.Fprintf(os.Stderr, "ERROR: only %d parameters are supported", len(argRegs))
		os.Exit(1)
	}

	paramType := p.Type()
	paramName := p.expect(token.TK_IDENT, "identifier")

	paramOffset := paramType.Size()
	p.varsOffsets += paramOffset
	p.symbolTable.put(paramName.Text, newSymbol(paramType, p.varsOffsets))
	p.genFuncParam(paramOffset, p.varsOffsets, argRegs[paramIndex])

	return paramType
}

func (p *Parser) FuncParams() []Type {
	p.expect(token.TK_LPAREN, "(")

	var paramsTypes []Type
	paramIndex := 0

	if p.currToken.Kind != token.TK_RPAREN {
		paramsTypes = append(paramsTypes, p.FuncParam(paramIndex))
		paramIndex += 1

		for p.currToken.Kind == token.TK_COMMA {
			p.nextToken()
			paramsTypes = append(paramsTypes, p.FuncParam(paramIndex))
			paramIndex += 1
		}
	}

	p.expect(token.TK_RPAREN, ")")
	return paramsTypes
}

func (p *Parser) ReturnStatement() {
	p.expect(token.TK_RETURN, "return")

	if p.currToken.Kind != token.TK_SEMICOLON {
		if p.currFuncRetType == TYPE_VOID {
			fmt.Fprintln(os.Stderr, "ERROR: void function cannot return a value")
			os.Exit(1)
		}

		returnType := p.Expression()
		if returnType != p.currFuncRetType {
			fmt.Fprintln(os.Stderr, "ERROR: return type mismatch")
			os.Exit(1)
		}

		p.genReturnStatement(returnType)
	} else {
		if p.currFuncRetType != TYPE_VOID {
			fmt.Fprintln(os.Stderr, "ERROR: function must return a value")
			os.Exit(1)
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

	paramsTypes := p.FuncParams()

	returnType := TYPE_VOID
	if p.currToken.Kind != token.TK_LCURLY {
		returnType = p.Type()
	}
	p.currFuncRetType = returnType
	p.functionTable.put(funcName.Text, newFuncSymbol(returnType, paramsTypes))

	p.Block()
	p.genFuncEnd()
}
