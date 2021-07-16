package lexer

import (
	"github.com/vegarsti/sql/token"
)

type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	switch l.ch {
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '-':
		tok = newToken(token.MINUS, l.ch)
	case '*':
		tok = newToken(token.ASTERISK, l.ch)
	case '/':
		tok = newToken(token.SLASH, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case []byte("'")[0]:
		tok := l.readString()
		return tok
	case 0:
		tok = token.Token{Type: token.EOF, Literal: ""}
	default:
		if isDigit(l.ch) {
			tok = l.readNumber()
			return tok
		}
		tok = newToken(token.ILLEGAL, l.ch)
	}

	l.readChar()
	return tok
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func (l *Lexer) readNumber() token.Token {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}

	// float
	if isDot(l.ch) {
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
		return token.Token{
			Type:    token.FLOAT,
			Literal: l.input[position:l.position],
		}
	}

	return token.Token{
		Type:    token.INT,
		Literal: l.input[position:l.position],
	}
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isDot(ch byte) bool {
	return ch == '.'
}

func isQuote(ch byte) bool {
	return string(ch) == "'"
}

func isEOF(ch byte) bool {
	return ch == 0
}

func (l *Lexer) readString() token.Token {
	l.readChar() // read first quote character
	// read until string is terminated or EOF is reached (which is an error)
	position := l.position
	for {
		// reached second quote character; break
		if isQuote(l.ch) {
			break
		}
		// if an EOF is reached before, it's an error
		if isEOF(l.ch) {
			return token.Token{Type: token.EOF, Literal: ""}
		}
		l.readChar()
	}
	literal := l.input[position:l.position]
	l.readChar() // read second quote character
	return token.Token{
		Type:    token.STRING,
		Literal: literal,
	}
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}
