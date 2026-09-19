package token

type TokenKind uint8

type Token struct {
	Kind TokenKind
	Text string
}

const (
	TK_EOF TokenKind = iota
	TK_ILLEGAL

	TK_IDENT
	TK_INT
	TK_BASIC

	TK_LPAREN
	TK_RPAREN
	TK_LCURLY
	TK_RCURLY
	TK_COMMA
	TK_SEMICOLON
	TK_ASSIGN
	TK_BANG
	TK_EQ
	TK_NEQ
	TK_LT
	TK_LTE
	TK_GT
	TK_GTE

	TK_PLUS
	TK_MINUS
	TK_MULT
	TK_DIV

	TK_FUNC
	TK_RETURN
	TK_IF
	TK_WHILE
	TK_ELSE
	TK_TRUE
	TK_FALSE
)

var keywords = map[string]TokenKind{
	"func":   TK_FUNC,
	"return": TK_RETURN,
	"int":    TK_BASIC,
	"bool":   TK_BASIC,
	"if":     TK_IF,
	"while":  TK_WHILE,
	"else":   TK_ELSE,
	"true":   TK_TRUE,
	"false":  TK_FALSE,
}

func LookupIdent(literal string) TokenKind {
	tk, exists := keywords[literal]
	if exists {
		return tk
	}

	return TK_IDENT
}
