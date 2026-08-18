package lexer

import (
	"bufio"
	"bytes"

	"palang/token"
)

type Lexer struct {
	input *bufio.Reader
	char  byte
}

func New(input *bufio.Reader) *Lexer {
	return &Lexer{
		input: input,
	}
}

func (l *Lexer) nextChar() {
	char, err := l.input.ReadByte()
	if err != nil {
		l.char = 0
	}

	l.char = char
}

func (l *Lexer) peekChar() byte {
	peekChar, err := l.input.Peek(1)
	if err != nil {
		return 0
	}

	return peekChar[0]
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token
	l.nextChar()

	for l.char == ' ' || l.char == '\n' || l.char == '\t' {
		l.nextChar()
	}

	switch l.char {
	case '(':
		tok = token.Token{Kind: token.TK_LPAREN, Text: "("}
	case ')':
		tok = token.Token{Kind: token.TK_RPAREN, Text: ")"}
	case '{':
		tok = token.Token{Kind: token.TK_LCURLY, Text: "{"}
	case '}':
		tok = token.Token{Kind: token.TK_RCURLY, Text: "}"}
	case ',':
		tok = token.Token{Kind: token.TK_COMMA, Text: ","}
	case ';':
		tok = token.Token{Kind: token.TK_SEMICOLON, Text: ";"}
	case '+':
		tok = token.Token{Kind: token.TK_PLUS, Text: "+"}
	case '-':
		tok = token.Token{Kind: token.TK_MINUS, Text: "-"}
	case '*':
		tok = token.Token{Kind: token.TK_MULT, Text: "*"}
	case '/':
		tok = token.Token{Kind: token.TK_DIV, Text: "/"}
	case 0:
		tok = token.Token{Kind: token.TK_EOF, Text: ""}

	default:
		if isLetter(l.char) {
			literal := l.readIdentifier()
			tok = token.Token{Kind: token.LookupIdent(literal), Text: literal}
		} else if isDigit(l.char) {
			literal := l.readInteger()
			tok = token.Token{Kind: token.TK_INT, Text: literal}
		} else {
			tok = token.Token{Kind: token.TK_ILLEGAL, Text: "illegal"}
		}
	}

	return tok
}

func (l *Lexer) readIdentifier() string {
	var buffer bytes.Buffer
	buffer.WriteByte(l.char)

	for peek := l.peekChar(); isLetter(peek) || isDigit(peek); peek = l.peekChar() {
		l.nextChar()
		buffer.WriteByte(l.char)
	}

	return buffer.String()
}

func (l *Lexer) readInteger() string {
	var buffer bytes.Buffer
	buffer.WriteByte(l.char)

	for peek := l.peekChar(); isDigit(peek); peek = l.peekChar() {
		l.nextChar()
		buffer.WriteByte(l.char)
	}

	return buffer.String()
}

func isLetter(char byte) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || char == '_'
}

func isDigit(char byte) bool {
	return '0' <= char && char <= '9'
}
