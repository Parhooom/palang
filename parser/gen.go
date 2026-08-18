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

func (p *Parser) genFuncEnd(statusCode string) {
	fmt.Printf("    mov rax, %s\n", statusCode)
	fmt.Printf("    ret\n")
}

func (p *Parser) genIntLiteral(num string) {
	fmt.Printf("    push %s\n", num)
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
		}

		fmt.Println("    push rax")
	}
}

func (p *Parser) genReturnStatement(returnType Type) {
	switch returnType {
	case TYPE_VOID:
		fmt.Println("    ret")
	case TYPE_INT:
		fmt.Println("    pop rax")
		fmt.Println("    ret")
	}
}
