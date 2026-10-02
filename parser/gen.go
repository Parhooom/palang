package parser

import (
	"fmt"
	"os"

	"palang/token"
)

var argRegs = []string{
	"rdi",
	"rsi",
	"rdx",
	"rcx",
	"r8",
	"r9",
}

func (p *Parser) genProgramStart() {
	fmt.Println("format ELF64")
	fmt.Println()
	fmt.Println("section \".text\" executable")
	fmt.Println()
	fmt.Println("extrn \"printf\" as _printf")
	fmt.Println("printf = PLT _printf")
}

func (p *Parser) genProgramEnd() {
	fmt.Println()
	fmt.Println("section \".rodata\"")
	fmt.Println("    fmt_int db \"%d\", 10, 0")
	fmt.Println("    fmt_str db \"%s\", 10, 0")

	for _, str := range p.stringTable.strings {
		fmt.Printf("    %s db \"%s\", 0\n", str.Label, str.Value)
	}

	fmt.Println()
	fmt.Println("section \".note.GNU-stack\"")
}

func (p *Parser) genFuncLabel(name string) {
	fmt.Printf("\n")
	fmt.Printf("%s:\n", name)
	if name == "main" {
		fmt.Printf("    public main\n")
	}
}

func (p *Parser) genLocalLabel(label string) {
	fmt.Printf("%s:\n", label)
}

func (p *Parser) genFuncStart() {
	fmt.Println("    push rbp")
	fmt.Println("    mov rbp, rsp")
	fmt.Println()
}

func (p *Parser) genFuncEnd() {
	fmt.Println()
	fmt.Println(".end:")
	fmt.Println("    mov rsp, rbp")
	fmt.Println("    pop rbp")
	fmt.Println("    ret")
}

func (p *Parser) genFuncParam(offsetSize uint, offset uint, reg string) {
	fmt.Printf("    sub rsp, %d\n", offsetSize)
	fmt.Printf("    mov [rbp-%d], %s\n", offset, reg)
}

func (p *Parser) genPrintCall(argType Type) {
	aligned := (p.varsOffsets % 16) == 0

	fmt.Println("    pop rsi")
	if !aligned {
		fmt.Println("    sub rsp, 8")
	}

	switch argType {
	case TYPE_INT, TYPE_BOOL:
		fmt.Println("    lea rdi, [fmt_int]")
	case TYPE_STRING:
		fmt.Println("    lea rdi, [fmt_str]")
	}

	fmt.Println("    mov rax, 0")
	fmt.Println("    call printf")
	if !aligned {
		fmt.Println("    add rsp, 8")
	}
}

func (p *Parser) genFuncCall(name string, argsCount int) {
	for i := argsCount - 1; i >= 0; i-- {
		fmt.Printf("    pop %s\n", argRegs[i])
	}
	fmt.Printf("    call %s\n", name)
}

func (p *Parser) genFuncCallReturn() {
	fmt.Println("    push rax")
}

func (p *Parser) genReturnStatement(retType Type) {
	fmt.Println()

	if retType == TYPE_INT || retType == TYPE_BOOL || retType == TYPE_STRING {
		fmt.Println("    pop rax")
	}

	fmt.Println("    jmp .end")
}

func (p *Parser) genIntLiteral(num string) {
	fmt.Printf("    push %s\n", num)
}

func (p *Parser) genBoolLiteral(val bool) {
	if val {
		fmt.Println("    push 1")
	} else {
		fmt.Println("    push 0")
	}
}

func (p *Parser) genStringLiteral(label string) {
	fmt.Printf("    lea rax, [%s]\n", label)
	fmt.Printf("    push rax\n")
}

func (p *Parser) genSymbol(symbol Symbol) {
	fmt.Printf("    push qword [rbp-%d]\n", symbol.Offset)
}

func (p *Parser) checkBinaryExpressionTypes(leftType, rightType Type, operator token.Token) Type {
	if leftType != rightType {
		fmt.Fprintln(os.Stderr, "ERROR: binary expression operands must have same types")
		os.Exit(1)
	}

	switch operator.Kind {
	case token.TK_PLUS, token.TK_MINUS, token.TK_MULT, token.TK_DIV:
		if leftType != TYPE_INT {
			fmt.Fprintln(os.Stderr, "ERROR: arithmetic operands must be integers")
			os.Exit(1)
		}

		return TYPE_INT

	case token.TK_LT, token.TK_LTE, token.TK_GT, token.TK_GTE:
		if leftType != TYPE_INT {
			fmt.Fprintln(os.Stderr, "ERROR: comparison operands must be integers")
			os.Exit(1)
		}

		return TYPE_BOOL

	case token.TK_EQ, token.TK_NEQ:
		if !(leftType == TYPE_INT || leftType == TYPE_BOOL || leftType == TYPE_STRING) {
			fmt.Fprintln(os.Stderr, "ERROR: unsupported equality operands")
			os.Exit(1)
		}

		return TYPE_BOOL

	default:
		fmt.Fprintln(os.Stderr, "ERROR: unsupported binary operator")
		os.Exit(1)
		return TYPE_VOID
	}
}

func (p *Parser) genBinaryExpression(leftType, rightType Type, operator token.Token) Type {
	resultTypes := p.checkBinaryExpressionTypes(leftType, rightType, operator)

	fmt.Println("    pop rbx")
	fmt.Println("    pop rax")

	switch operator.Kind {
	case token.TK_PLUS:
		fmt.Println("    add rax, rbx")
	case token.TK_MINUS:
		fmt.Println("    sub rax, rbx")
	case token.TK_MULT:
		fmt.Println("    imul rbx")
	case token.TK_DIV:
		fmt.Println("    cqo")
		fmt.Println("    idiv rbx")
	case token.TK_EQ:
		fmt.Println("    cmp rax, rbx")
		fmt.Println("    sete al")
		fmt.Println("    movzx rax, al")
	case token.TK_NEQ:
		fmt.Println("    cmp rax, rbx")
		fmt.Println("    setne al")
		fmt.Println("    movzx rax, al")
	case token.TK_LT:
		fmt.Println("    cmp rax, rbx")
		fmt.Println("    setl al")
		fmt.Println("    movzx rax, al")
	case token.TK_LTE:
		fmt.Println("    cmp rax, rbx")
		fmt.Println("    setle al")
		fmt.Println("    movzx rax, al")
	case token.TK_GT:
		fmt.Println("    cmp rax, rbx")
		fmt.Println("    setg al")
		fmt.Println("    movzx rax, al")
	case token.TK_GTE:
		fmt.Println("    cmp rax, rbx")
		fmt.Println("    setge al")
		fmt.Println("    movzx rax, al")
	}

	fmt.Println("    push rax")
	return resultTypes
}

func (p *Parser) genVarDecl(offsetSize uint) {
	fmt.Printf("    sub rsp, %d\n", offsetSize)
}

func (p *Parser) genVarAssign(symbol Symbol) {
	fmt.Printf("    pop rax\n")
	fmt.Printf("    mov [rbp-%d], rax\n", symbol.Offset)
}

func (p *Parser) genJz(label string) {
	fmt.Printf("    pop rax\n")
	fmt.Printf("    cmp rax, 0\n")
	fmt.Printf("    jz %s\n", label)
}

func (p *Parser) genJmp(label string) {
	fmt.Printf("    jmp %s\n", label)
}
