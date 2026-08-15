package parser

import "fmt"

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
