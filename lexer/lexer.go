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
	case '!':
		if string(l.input[l.position:l.position+2]) == "!=" {
			tok = token.Token{Type: token.NOTEQUALS, Literal: "!="}
			l.readChar()
			l.readChar()
			return tok
		}
		tok = newToken(token.BANG, l.ch)
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
			if strings.ToUpper(tok.Literal) == token.SELECT {
				return token.Token{Type: token.SELECT, Literal: token.SELECT}
			}
			if strings.ToUpper(tok.Literal) == token.AS {
				return token.Token{Type: token.AS, Literal: token.AS}
			}
			if strings.ToUpper(tok.Literal) == token.CREATE {
				return token.Token{Type: token.CREATE, Literal: token.CREATE}
			}
			if strings.ToUpper(tok.Literal) == token.TABLE {
				return token.Token{Type: token.TABLE, Literal: token.TABLE}
			}
			if strings.ToUpper(tok.Literal) == token.TEXT {
				return token.Token{Type: token.TEXT, Literal: token.TEXT}
			}
			if strings.ToUpper(tok.Literal) == token.DOUBLE {
				return token.Token{Type: token.DOUBLE, Literal: token.DOUBLE}
			}
			if strings.ToUpper(tok.Literal) == token.INTEGER {
				return token.Token{Type: token.INTEGER, Literal: token.INTEGER}
			}
			if strings.ToUpper(tok.Literal) == token.INT {
				return token.Token{Type: token.INTEGER, Literal: token.INTEGER}
			}
			if strings.ToUpper(tok.Literal) == token.INSERT {
				return token.Token{Type: token.INSERT, Literal: token.INSERT}
			}
			if strings.ToUpper(tok.Literal) == token.INTO {
				return token.Token{Type: token.INTO, Literal: token.INTO}
			}
			if strings.ToUpper(tok.Literal) == token.VALUES {
				return token.Token{Type: token.VALUES, Literal: token.VALUES}
			}
			if strings.ToUpper(tok.Literal) == token.FROM {
				return token.Token{Type: token.FROM, Literal: token.FROM}
			}
			if strings.ToUpper(tok.Literal) == token.ORDER {
				return token.Token{Type: token.ORDER, Literal: token.ORDER}
			}
			if strings.ToUpper(tok.Literal) == token.BY {
				return token.Token{Type: token.BY, Literal: token.BY}
			}
			if strings.ToUpper(tok.Literal) == token.DESC {
				return token.Token{Type: token.DESC, Literal: token.DESC}
			}
			if strings.ToUpper(tok.Literal) == token.ASC {
				return token.Token{Type: token.ASC, Literal: token.ASC}
			}
			if strings.ToUpper(tok.Literal) == token.TRUE {
				return token.Token{Type: token.BOOL, Literal: token.TRUE}
			}
			if strings.ToUpper(tok.Literal) == token.FALSE {
				return token.Token{Type: token.BOOL, Literal: token.FALSE}
			}
			if strings.ToUpper(tok.Literal) == token.AND {
				return token.Token{Type: token.AND, Literal: token.AND}
			}
			if strings.ToUpper(tok.Literal) == token.OR {
				return token.Token{Type: token.OR, Literal: token.OR}
			}
			if strings.ToUpper(tok.Literal) == token.BOOL {
				return token.Token{Type: token.BOOLEAN, Literal: token.BOOLEAN}
			}
			if strings.ToUpper(tok.Literal) == token.BOOLEAN {
				return token.Token{Type: token.BOOLEAN, Literal: token.BOOLEAN}
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
			Type:    token.FLOAT,
			Literal: l.input[position:l.position],
		}
	}

	return token.Token{
		Type:    token.INT,
		Literal: l.input[position:l.position],
	}
}

func (l *Lexer) readIdentifier() token.Token {
	position := l.position
	for isLetter(l.ch) || l.ch == '_' {
		l.readChar()
	}
	return token.Token{
		Type:    token.IDENTIFIER,
		Literal: l.input[position:l.position],
	}
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isLetter(ch byte) bool {
	isLowercase := 'a' <= ch && ch <= 'z'
	isUppercase := 'A' <= ch && ch <= 'Z'
	return isLowercase || isUppercase
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

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func (l *Lexer) skipWhitespace() {
	for isWhitespace(l.ch) {
		l.readChar()
	}
}
