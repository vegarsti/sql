package lexer_test

import (
	"testing"

	"github.com/vegarsti/sql/lexer"
	"github.com/vegarsti/sql/token"
)

func TestExpressionValue(t *testing.T) {
	input := "1 + 2 * (30 / 5) - 1 + 3.14 + 1.0"
	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.INT, "1"},
		{token.PLUS, "+"},
		{token.INT, "2"},
		{token.ASTERISK, "*"},
		{token.LPAREN, "("},
		{token.INT, "30"},
		{token.SLASH, "/"},
		{token.INT, "5"},
		{token.RPAREN, ")"},
		{token.MINUS, "-"},
		{token.INT, "1"},
		{token.PLUS, "+"},
		{token.FLOAT, "3.14"},
		{token.PLUS, "+"},
		{token.FLOAT, "1.0"},
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
