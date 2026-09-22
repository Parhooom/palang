package parser

import (
	"fmt"
	"os"
)

type Symbol struct {
	Kind   Type
	Offset uint
}

type SymbolTable struct {
	Prev  *SymbolTable
	Table map[string]Symbol
}

func (st *SymbolTable) put(name string, symbol Symbol) {
	_, exists := st.Table[name]
	if exists {
		fmt.Fprintf(os.Stderr, "ERROR: '%s' exists in this scope\n", name)
		os.Exit(1)
	}

	st.Table[name] = symbol
}

func (st *SymbolTable) get(name string) (Symbol, bool) {
	for symbolTable := st; symbolTable != nil; symbolTable = symbolTable.Prev {
		symbol, exists := symbolTable.Table[name]
		if exists {
			return symbol, true
		}
	}

	return Symbol{}, false
}

func newSymbol(kind Type, offset uint) Symbol {
	return Symbol{Kind: kind, Offset: offset}
}

type FuncSymbol struct {
	ReturnType  Type
	ParamsTypes []Type
}

type FuncTable map[string]FuncSymbol

func (ft *FuncTable) put(name string, funcSymbol FuncSymbol) {
	_, exists := (*ft)[name]
	if exists {
		fmt.Fprintf(os.Stderr, "ERROR: '%s' function already declared\n", name)
		os.Exit(1)
	}

	(*ft)[name] = funcSymbol
}

func newFuncSymbol(returnType Type, paramsTypes []Type) FuncSymbol {
	return FuncSymbol{ReturnType: returnType, ParamsTypes: paramsTypes}
}
