package lexer

import (
	"strings"

	"github.com/vegarsti/sql/token"
)

type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
}

var keywords = []token.TokenType{
	token.SELECT,
	token.AS,
	token.CREATE,
	token.TABLE,
	token.INSERT,
	token.INTO,
	token.VALUES,
	token.FROM,
	token.ORDER,
	token.BY,
	token.DESC,
	token.ASC,
	token.TRUE,
	token.FALSE,
	token.AND,
	token.OR,
	token.LIMIT,
	token.OFFSET,
	token.WHERE,
	token.JOIN,
	token.ON,
	token.NULL,
	token.IS,
	token.NOT,
	token.TRUE,
	token.FALSE,
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
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case '=':
		tok = newToken(token.EQUALS, l.ch)
	case '>':
		if l.input[l.position+1] == '=' {
			tok = token.Token{Type: token.GREATERTHANOREQUALS, Literal: ">="}
			l.readChar()
			l.readChar()
			return tok
		}
		tok = newToken(token.GREATERTHAN, l.ch)
	case '<':
		if l.input[l.position+1] == '=' {
			tok = token.Token{Type: token.LESSTHANOREQUALS, Literal: "<="}
			l.readChar()
			l.readChar()
			return tok
		}
		tok = newToken(token.LESSTHAN, l.ch)
	case '!':
		if l.input[l.position+1] == '=' {
			tok = token.Token{Type: token.NOTEQUALS, Literal: "!="}
			l.readChar()
			l.readChar()
			return tok
		}
		tok = newToken(token.ILLEGAL, l.ch)
	case '|':
		if l.input[l.position+1] == '|' {
			tok = token.Token{Type: token.DOUBLEBAR, Literal: "||"}
			l.readChar()
			l.readChar()
			return tok
		}
		tok = newToken(token.ILLEGAL, l.ch)
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
		if isLetter(l.ch) {
			tok = l.readIdentifier()
			literal := strings.ToUpper(tok.Literal)
			// string type
			if map[string]bool{token.STRING_TYPE: true, "CHAR": true, "TEXT": true, "VARCHAR": true}[literal] {
				return token.Token{Type: token.STRING_TYPE, Literal: token.STRING_TYPE}
			}

			// float type
			if map[string]bool{token.FLOAT_TYPE: true, "DOUBLE": true}[literal] {
				return token.Token{Type: token.FLOAT_TYPE, Literal: token.FLOAT_TYPE}
			}

			// int type
			if map[string]bool{token.INTEGER_TYPE: true, "INT": true}[literal] {
				return token.Token{Type: token.INTEGER_TYPE, Literal: token.INTEGER_TYPE}
			}

			// bool type
			if map[string]bool{token.BOOLEAN_TYPE: true, "BOOL": true}[literal] {
				return token.Token{Type: token.BOOLEAN_TYPE, Literal: token.BOOLEAN_TYPE}
			}

			for _, keyword := range keywords {
				if literal == string(keyword) {
					return token.Token{Type: keyword, Literal: literal}
				}
			}
			// don't allow uppercase identifiers
			for _, ch := range tok.Literal {
				if isUppercase(byte(ch)) {
					return token.Token{Type: token.ILLEGAL, Literal: tok.Literal}
				}
			}
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
			Type:    token.FLOAT_LITERAL,
			Literal: l.input[position:l.position],
		}
	}

	return token.Token{
		Type:    token.INT_LITERAL,
		Literal: l.input[position:l.position],
	}
}

func (l *Lexer) readIdentifier() token.Token {
	position := l.position
	for isLetter(l.ch) || l.ch == '_' || l.ch == '.' {
		l.readChar()
	}
	literal := l.input[position:l.position]
	nDots := strings.Count(literal, ".")
	if nDots > 1 {
		return token.Token{
			Type:    token.ILLEGAL,
			Literal: literal,
		}
	}
	if nDots == 1 {
		return token.Token{
			Type:    token.QUALIFIEDIDENTIFIER,
			Literal: literal,
		}
	}
	return token.Token{
		Type:    token.IDENTIFIER,
		Literal: literal,
	}
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isLowercase(ch byte) bool {
	return 'a' <= ch && ch <= 'z'
}

func isUppercase(ch byte) bool {
	return 'A' <= ch && ch <= 'Z'
}

func isLetter(ch byte) bool {
	return isLowercase(ch) || isUppercase(ch)
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
		Type:    token.STRING_LITERAL,
		Literal: literal,
	}
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func (l *Lexer) skipWhitespace() {
	for isWhitespace(l.ch) {
		l.readChar()
	}
}
