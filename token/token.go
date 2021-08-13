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
	CREATE = "CREATE"
	TABLE  = "TABLE"
	INSERT = "INSERT"
	INTO   = "INTO"
	VALUES = "VALUES"
	FROM   = "FROM"
	ORDER  = "ORDER"
	BY     = "BY"
	DESC   = "DESC"

	// Types
	TEXT    = "TEXT"
	DOUBLE  = "DOUBLE"
	INTEGER = "INTEGER"

	// Delimiters
	LPAREN    = "("
	RPAREN    = ")"
	DOT       = "."
	QUOTE     = "'"
	COMMA     = ","
	SEMICOLON = ";"
)
