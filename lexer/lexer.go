package lexer

import (
	"bufio"
	"bytes"

	"palang/token"
)

type Lexer struct {
	input *bufio.Reader
	char  byte

	row      uint
	col      uint
	fileName string
}

func New(input *bufio.Reader, fileName string) *Lexer {
	return &Lexer{
		input:    input,
		fileName: fileName,
		row:      1,
	}
}

func (l *Lexer) nextChar() {
	char, err := l.input.ReadByte()
	if err != nil {
		l.char = 0
	}

	l.char = char

	switch l.char {
	case '\n':
		l.row += 1
		l.col = 0
	case '\t':
		l.col += 4
	default:
		l.col += 1
	}
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

	location := token.Location{
		Row:      l.row,
		Col:      l.col,
		FileName: l.fileName,
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
	case '=':
		if l.peekChar() == '=' {
			tok = token.Token{Kind: token.TK_EQ, Text: "=="}
			l.nextChar()
		} else {
			tok = token.Token{Kind: token.TK_ASSIGN, Text: "="}
		}
	case '!':
		if l.peekChar() == '=' {
			tok = token.Token{Kind: token.TK_NEQ, Text: "!="}
			l.nextChar()
		} else {
			tok = token.Token{Kind: token.TK_BANG, Text: "!"}
		}
	case '<':
		if l.peekChar() == '=' {
			tok = token.Token{Kind: token.TK_LTE, Text: "<="}
			l.nextChar()
		} else {
			tok = token.Token{Kind: token.TK_LT, Text: "<"}
		}
	case '>':
		if l.peekChar() == '=' {
			tok = token.Token{Kind: token.TK_GTE, Text: ">="}
			l.nextChar()
		} else {
			tok = token.Token{Kind: token.TK_GT, Text: ">"}
		}
	case '"':
		literal := l.readString()
		tok = token.Token{Kind: token.TK_STRING, Text: literal}

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

	tok.Location = location
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

func (l *Lexer) readString() string {
	var buffer bytes.Buffer

	for peek := l.peekChar(); peek != '"' && peek != 0; peek = l.peekChar() {
		l.nextChar()
		buffer.WriteByte(l.char)
	}

	if l.peekChar() == '"' {
		l.nextChar()
	}

	return buffer.String()
}

func isLetter(char byte) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || char == '_'
}

func isDigit(char byte) bool {
	return '0' <= char && char <= '9'
}
