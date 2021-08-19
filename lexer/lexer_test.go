package lexer_test

import (
	"testing"

	"github.com/vegarsti/sql/lexer"
	"github.com/vegarsti/sql/token"
)

func TestExpressionValue(t *testing.T) {
	input := `
1 + 2 * (30 / 5) - 1 + 3.14 + 'abc' 1.0 'def' select SELECT SeLeCT aWord , AS as aS As create table text double integer insert into values from identifier_with_underscore;
order by desc asc false true = != !2 and or limit offset where < <= > >= table_name.column_name bool boolean int
`
	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.INT_LITERAL, "1"},
		{token.PLUS, "+"},
		{token.INT_LITERAL, "2"},
		{token.ASTERISK, "*"},
		{token.LPAREN, "("},
		{token.INT_LITERAL, "30"},
		{token.SLASH, "/"},
		{token.INT_LITERAL, "5"},
		{token.RPAREN, ")"},
		{token.MINUS, "-"},
		{token.INT_LITERAL, "1"},
		{token.PLUS, "+"},
		{token.FLOAT_LITERAL, "3.14"},
		{token.PLUS, "+"},
		{token.STRING_LITERAL, "abc"},
		{token.FLOAT_LITERAL, "1.0"},
		{token.STRING_LITERAL, "def"},
		{token.SELECT, "SELECT"},
		{token.SELECT, "SELECT"},
		{token.SELECT, "SELECT"},
		{token.IDENTIFIER, "aWord"},
		{token.COMMA, ","},
		{token.AS, "AS"},
		{token.AS, "AS"},
		{token.AS, "AS"},
		{token.AS, "AS"},
		{token.CREATE, "CREATE"},
		{token.TABLE, "TABLE"},
		{token.TEXT_TYPE, "TEXT"},
		{token.FLOAT_TYPE, "DOUBLE"},
		{token.INTEGER_TYPE, "INTEGER"},
		{token.INSERT, "INSERT"},
		{token.INTO, "INTO"},
		{token.VALUES, "VALUES"},
		{token.FROM, "FROM"},
		{token.IDENTIFIER, "identifier_with_underscore"},
		{token.SEMICOLON, ";"},
		{token.ORDER, "ORDER"},
		{token.BY, "BY"},
		{token.DESC, "DESC"},
		{token.ASC, "ASC"},
		{token.BOOL_LITERAL, "FALSE"},
		{token.BOOL_LITERAL, "TRUE"},
		{token.EQUALS, "="},
		{token.NOTEQUALS, "!="},
		{token.BANG, "!"},
		{token.INT_LITERAL, "2"},
		{token.AND, "AND"},
		{token.OR, "OR"},
		{token.LIMIT, "LIMIT"},
		{token.OFFSET, "OFFSET"},
		{token.WHERE, "WHERE"},
		{token.LESSTHAN, "<"},
		{token.LESSTHANOREQUALS, "<="},
		{token.GREATERTHAN, ">"},
		{token.GREATERTHANOREQUALS, ">="},
		{token.QUALIFIEDIDENTIFIER, "table_name.column_name"},
		{token.BOOLEAN_TYPE, "BOOLEAN"},
		{token.BOOLEAN_TYPE, "BOOLEAN"},
		{token.INTEGER_TYPE, "INTEGER"},
	}
	l := lexer.New(input)
	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - token type wrong. expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - token literal wrong. expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}
