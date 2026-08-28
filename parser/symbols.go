package parser

import (
	"fmt"
)

type Symbol struct {
	Kind   Type
	Offset uint
}

type SymbolTable struct {
	Prev  *SymbolTable
	Table map[string]Symbol
}

func (st *SymbolTable) put(name string, symbol Symbol) error {
	s, exists := st.Table[name]
	if exists && s.Kind != symbol.Kind {
		return fmt.Errorf("ERROR: %s symbol was defined with different signiture", name)
	}

	st.Table[name] = symbol
	return nil
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
