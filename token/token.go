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
	INT                       = "INT"
	FLOAT                     = "FLOAT"
	STRING                    = "STRING"
	IDENTIFIER                = "IDENTIFIER"
	IDENTIFIER_WITH_TABLENAME = "IDENTIFIER_WITH_TABLENAME"
	BOOL                      = "BOOL"

	// Operators
	PLUS                = "+"
	MINUS               = "-"
	ASTERISK            = "*"
	SLASH               = "/"
	EQUALS              = "="
	BANG                = "!"
	NOTEQUALS           = "!="
	LESSTHAN            = "<"
	LESSTHANOREQUALS    = "<="
	GREATERTHAN         = ">"
	GREATERTHANOREQUALS = ">="

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
	ASC    = "ASC"
	FALSE  = "FALSE"
	TRUE   = "TRUE"
	AND    = "AND"
	OR     = "OR"
	LIMIT  = "LIMIT"
	OFFSET = "OFFSET"
	WHERE  = "WHERE"

	// Types
	TEXT    = "TEXT"
	DOUBLE  = "DOUBLE"
	INTEGER = "INTEGER"
	BOOLEAN = "BOOLEAN"

	// Delimiters
	LPAREN    = "("
	RPAREN    = ")"
	DOT       = "."
	QUOTE     = "'"
	COMMA     = ","
	SEMICOLON = ";"
)
