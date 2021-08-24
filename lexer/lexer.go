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

			// string type
			if strings.ToUpper(tok.Literal) == token.STRING_TYPE {
				return token.Token{Type: token.STRING_TYPE, Literal: token.STRING_TYPE}
			}
			if strings.ToUpper(tok.Literal) == "CHAR" {
				return token.Token{Type: token.STRING_TYPE, Literal: token.STRING_TYPE}
			}
			if strings.ToUpper(tok.Literal) == "TEXT" {
				return token.Token{Type: token.STRING_TYPE, Literal: token.STRING_TYPE}
			}
			if strings.ToUpper(tok.Literal) == "VARCHAR" {
				return token.Token{Type: token.STRING_TYPE, Literal: token.STRING_TYPE}
			}

			// float type
			if strings.ToUpper(tok.Literal) == token.FLOAT_TYPE {
				return token.Token{Type: token.FLOAT_TYPE, Literal: token.FLOAT_TYPE}
			}
			if strings.ToUpper(tok.Literal) == "DOUBLE" {
				return token.Token{Type: token.FLOAT_TYPE, Literal: token.FLOAT_TYPE}
			}

			// int type
			if strings.ToUpper(tok.Literal) == token.INTEGER_TYPE {
				return token.Token{Type: token.INTEGER_TYPE, Literal: token.INTEGER_TYPE}
			}
			if strings.ToUpper(tok.Literal) == "INT" {
				return token.Token{Type: token.INTEGER_TYPE, Literal: token.INTEGER_TYPE}
			}

			// bool type
			if strings.ToUpper(tok.Literal) == token.BOOLEAN_TYPE {
				return token.Token{Type: token.BOOLEAN_TYPE, Literal: token.BOOLEAN_TYPE}
			}
			if strings.ToUpper(tok.Literal) == "BOOL" {
				return token.Token{Type: token.BOOLEAN_TYPE, Literal: token.BOOLEAN_TYPE}
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
				return token.Token{Type: token.BOOL_LITERAL, Literal: token.TRUE}
			}
			if strings.ToUpper(tok.Literal) == token.FALSE {
				return token.Token{Type: token.BOOL_LITERAL, Literal: token.FALSE}
			}
			if strings.ToUpper(tok.Literal) == token.AND {
				return token.Token{Type: token.AND, Literal: token.AND}
			}
			if strings.ToUpper(tok.Literal) == token.OR {
				return token.Token{Type: token.OR, Literal: token.OR}
			}
			if strings.ToUpper(tok.Literal) == token.LIMIT {
				return token.Token{Type: token.LIMIT, Literal: token.LIMIT}
			}
			if strings.ToUpper(tok.Literal) == token.OFFSET {
				return token.Token{Type: token.OFFSET, Literal: token.OFFSET}
			}
			if strings.ToUpper(tok.Literal) == token.WHERE {
				return token.Token{Type: token.WHERE, Literal: token.WHERE}
			}
			if strings.ToUpper(tok.Literal) == token.JOIN {
				return token.Token{Type: token.JOIN, Literal: token.JOIN}
			}
			if strings.ToUpper(tok.Literal) == token.ON {
				return token.Token{Type: token.ON, Literal: token.ON}
			}
			if strings.ToUpper(tok.Literal) == token.NULL {
				return token.Token{Type: token.NULL, Literal: token.NULL}
			}
			if strings.ToUpper(tok.Literal) == token.IS {
				return token.Token{Type: token.IS, Literal: token.IS}
			}
			if strings.ToUpper(tok.Literal) == token.NOT {
				return token.Token{Type: token.NOT, Literal: token.NOT}
			}
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
