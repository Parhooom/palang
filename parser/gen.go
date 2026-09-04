package parser

import (
	"fmt"
	"os"

	"palang/token"
)

func (p *Parser) genProgramStart() {
	fmt.Printf("format ELF64\n")
	fmt.Printf("section \".text\" executable\n")
}

func (p *Parser) genProgramEnd() {
	fmt.Printf("\n")
	fmt.Printf("section \".note.GNU-stack\"\n")
}

func (p *Parser) genFuncLabel(name string) {
	fmt.Printf("\n")
	fmt.Printf("%s:\n", name)
	if name == "main" {
		fmt.Printf("    public main\n")
	}
}

func (p *Parser) genFuncStart() {
	fmt.Println("    push rbp")
	fmt.Println("    mov rbp, rsp")
	fmt.Println()
}

func (p *Parser) genFuncEnd(returnType Type) {
	fmt.Println()
	switch returnType {
	case TYPE_VOID:
	case TYPE_INT:
		fmt.Println("    pop rax")
	}
	// fmt.Printf("    add rsp, %d\n", p.varsOffsets)
	fmt.Println("    mov rsp, rbp")
	fmt.Println("    pop rbp")
	fmt.Println("    ret")
}

func (p *Parser) genIntLiteral(num string) {
	fmt.Printf("    push %s\n", num)
}

func (p *Parser) genIntSymbol(symbol Symbol) {
	fmt.Printf("    push qword [rbp-%d]\n", symbol.Offset)
}

func (p *Parser) genBinaryExpression(leftType Type, rightType Type, operator token.Token) {
	if leftType != rightType {
		fmt.Fprintln(os.Stderr, "ERROR: binary expression operands must have same types")
		os.Exit(1)
	}

	if leftType == TYPE_INT {
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
	}
}

func (p *Parser) genVarDecl(offset uint) {
	fmt.Printf("    sub rsp, %d\n", offset)
}

func (p *Parser) genVarAssign(symbol Symbol) {
	if symbol.Kind == TYPE_INT {
		fmt.Printf("    pop rax\n")
		fmt.Printf("    mov [rbp-%d], rax\n", symbol.Offset)
	}
}
