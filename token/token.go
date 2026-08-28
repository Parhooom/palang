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

	TK_PLUS
	TK_MINUS
	TK_MULT
	TK_DIV

	TK_FUNC
	TK_RETURN
)

var keywords = map[string]TokenKind{
	"func":   TK_FUNC,
	"return": TK_RETURN,
	"int":    TK_BASIC,
}

func LookupIdent(literal string) TokenKind {
	tk, exists := keywords[literal]
	if exists {
		return tk
	}

	return TK_IDENT
}
