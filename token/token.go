package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Literals
	INT   = "INT"
	FLOAT = "FLOAT"

	// Operators
	PLUS     = "+"
	MINUS    = "-"
	ASTERISK = "*"
	SLASH    = "/"

	// Delimiters
	LPAREN = "("
	RPAREN = ")"
	DOT    = "."
)
