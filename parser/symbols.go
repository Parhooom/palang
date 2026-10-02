package parser

import (
	"fmt"
	"os"
)

type Type uint8

const (
	_ Type = iota
	TYPE_VOID
	TYPE_INT
	TYPE_BOOL
	TYPE_STRING
)

func (t Type) Size() uint {
	switch t {
	case TYPE_INT, TYPE_BOOL, TYPE_STRING:
		return 8
	default:
		return 0
	}
}

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
		fmt.Fprintf(os.Stderr, "ERROR: '%s' is already declared\n", name)
		os.Exit(1)
	}

	(*ft)[name] = funcSymbol
}

func newFuncSymbol(returnType Type, paramsTypes []Type) FuncSymbol {
	return FuncSymbol{ReturnType: returnType, ParamsTypes: paramsTypes}
}

type StringConstant struct {
	Label string
	Value string
}

type StringTable struct {
	table   map[string]string
	strings []StringConstant
	count   uint
}

func (st *StringTable) getOrInsert(str string) string {
	label, exists := st.table[str]
	if exists {
		return label
	}

	st.count += 1
	label = fmt.Sprintf("str%d", st.count)
	st.table[str] = label
	st.strings = append(st.strings, StringConstant{Label: label, Value: str})
	return label
}
