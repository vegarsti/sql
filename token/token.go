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
	INT_LITERAL         = "INT_LITERAL"
	FLOAT_LITERAL       = "FLOAT_LITERAL"
	STRING_LITERAL      = "STRING_LITERAL"
	IDENTIFIER          = "IDENTIFIER"
	QUALIFIEDIDENTIFIER = "QUALIFIEDIDENTIFIER"
	BOOL_LITERAL        = "BOOL_LITERAL"
	NULL                = "NULL"

	// Operators
	PLUS                = "+"
	MINUS               = "-"
	ASTERISK            = "*"
	SLASH               = "/"
	EQUALS              = "="
	NOTEQUALS           = "!="
	LESSTHAN            = "<"
	LESSTHANOREQUALS    = "<="
	GREATERTHAN         = ">"
	GREATERTHANOREQUALS = ">="
	DOUBLEBAR           = "||"

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
	JOIN   = "JOIN"
	ON     = "ON"
	IS     = "IS"
	NOT    = "NOT"

	// Types
	STRING_TYPE  = "STRING"
	FLOAT_TYPE   = "FLOAT"
	INTEGER_TYPE = "INTEGER"
	BOOLEAN_TYPE = "BOOLEAN"

	// Delimiters
	LPAREN    = "("
	RPAREN    = ")"
	DOT       = "."
	QUOTE     = "'"
	COMMA     = ","
	SEMICOLON = ";"
)
