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
	INT        = "INT"
	FLOAT      = "FLOAT"
	STRING     = "STRING"
	IDENTIFIER = "IDENTIFIER"

	// Operators
	PLUS     = "+"
	MINUS    = "-"
	ASTERISK = "*"
	SLASH    = "/"

	// Keywords
	SELECT = "SELECT"
	AS     = "AS"

	// Delimiters
	LPAREN = "("
	RPAREN = ")"
	DOT    = "."
	QUOTE  = "'"
	COMMA  = ","
)
