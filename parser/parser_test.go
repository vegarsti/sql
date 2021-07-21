package parser_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/vegarsti/sql/ast"
	"github.com/vegarsti/sql/lexer"
	"github.com/vegarsti/sql/parser"
)

func TestIntegerLiteralExpression(t *testing.T) {
	input := "select 5"
	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.SelectStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.SelectStatement. got=%T", program.Statements[0])
	}

	if len(stmt.Expressions) != 1 {
		t.Fatalf("stmt does not contain %d expressions. got=%d", 1, len(stmt.Expressions))
	}

	literal, ok := stmt.Expressions[0].(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("exp not *ast.IntegerLiteral. got=%T", stmt.Expressions[0])
	}
	if literal.TokenLiteral() != "5" {
		t.Errorf("literal.TokenLiteral not %s. got=%s", "5", literal.TokenLiteral())
	}
	checkParserErrors(t, p)
}

func TestFloatLiteralExpression(t *testing.T) {
	input := "select 3.14"
	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.SelectStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.SelectStatement. got=%T", program.Statements[0])
	}

	if len(stmt.Expressions) != 1 {
		t.Fatalf("stmt does not contain %d expressions. got=%d", 1, len(stmt.Expressions))
	}

	literal, ok := stmt.Expressions[0].(*ast.FloatLiteral)
	if !ok {
		t.Fatalf("exp not *ast.FloatLiteral. got=%T", stmt.Expressions[0])
	}
	if literal.TokenLiteral() != "3.14" {
		t.Errorf("literal.TokenLiteral not %s. got=%s", "3.14", literal.TokenLiteral())
	}
	checkParserErrors(t, p)
}

func TestParsingPrefixExpression(t *testing.T) {
	prefixTest := []struct {
		input        string
		operator     string
		integerValue int64
	}{
		{"select -15", "-", 15},
	}
	for _, tt := range prefixTest {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d", 1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.SelectStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.SelectStatement. got=%T", program.Statements[0])
		}

		if len(stmt.Expressions) != 1 {
			t.Fatalf("stmt does not contain %d expressions. got=%d", 1, len(stmt.Expressions))
		}

		exp, ok := stmt.Expressions[0].(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("stmt is not ast.PrefixExpression. got=%T", stmt.Expressions[0])
		}
		if exp.Operator != tt.operator {
			t.Fatalf("exp.Operator is not '%s'. got=%s", tt.operator, exp.Operator)
		}
		if !testIntegerLiteral(t, exp.Right, tt.integerValue) {
			return
		}
	}
}

func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) bool {
	integ, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("il not *ast.IntegerLiteral. got=%T", il)
		return false
	}

	if integ.Value != value {
		t.Errorf("integ.Value not %d. got=%d", value, integ.Value)
		return false
	}

	if integ.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf("integ.TokenLiteral not %d. got=%s", value, integ.TokenLiteral())
		return false
	}

	return true
}

func TestParsingInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  int64
		operator   string
		rightValue int64
	}{
		{"select 5 + 5", 5, "+", 5},
		{"select 5 - 5", 5, "-", 5},
		{"select 5 * 5", 5, "*", 5},
		{"select 5 / 5", 5, "/", 5},
	}
	for _, tt := range infixTests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d", 1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.SelectStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.SelectStatement. got=%T", program.Statements[0])
		}

		if len(stmt.Expressions) != 1 {
			t.Fatalf("stmt does not contain %d expressions. got=%d", 1, len(stmt.Expressions))
		}

		exp, ok := stmt.Expressions[0].(*ast.InfixExpression)
		if !ok {
			t.Fatalf("stmt is not ast.InfixExpression. got=%T", stmt.Expressions[0])
		}
		if !testIntegerLiteral(t, exp.Left, tt.leftValue) {
			return
		}
		if exp.Operator != tt.operator {
			t.Fatalf("exp.Operator is not '%s'. got=%s", tt.operator, exp.Operator)
		}
		if !testIntegerLiteral(t, exp.Right, tt.rightValue) {
			return
		}
	}
}

func checkParserErrors(t *testing.T, p *parser.Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"select -1 * 2",
			"SELECT ((-1) * 2)",
		},
		{
			"select 1 + 2 + 3",
			"SELECT ((1 + 2) + 3)",
		},
		{
			"select 1 + 2 - 3",
			"SELECT ((1 + 2) - 3)",
		},
		{
			"select 1 * 2 * 3",
			"SELECT ((1 * 2) * 3)",
		},
		{
			"select 1 * 2 / 3",
			"SELECT ((1 * 2) / 3)",
		},
		{
			"select 1 + 2 / 3",
			"SELECT (1 + (2 / 3))",
		},
		{
			"select 1 + (2 + 3) + 4",
			"SELECT ((1 + (2 + 3)) + 4)",
		},
		{
			"select (5 + 5) * 2",
			"SELECT ((5 + 5) * 2)",
		},
		{
			"select 2 / (5 + 5)",
			"SELECT (2 / (5 + 5))",
		},
		{
			"select -(5 + 5)",
			"SELECT (-(5 + 5))",
		},
	}
	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		actual := program.String()
		if actual != tt.expected {
			t.Errorf("expected=%q, got=%q", tt.expected, actual)
		}
	}
}

func TestStringLiteralExpression(t *testing.T) {
	input := "select 'abc'"
	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.SelectStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.SelectStatement. got=%T", program.Statements[0])
	}

	if len(stmt.Expressions) != 1 {
		t.Fatalf("stmt does not contain %d expressions. got=%d", 1, len(stmt.Expressions))
	}

	literal, ok := stmt.Expressions[0].(*ast.StringLiteral)
	if !ok {
		t.Fatalf("exp not *ast.StringLiteral. got=%T", stmt.Expressions[0])
	}
	expectedLiteral := "abc"
	if literal.TokenLiteral() != expectedLiteral {
		t.Errorf("literal.TokenLiteral not %s. got=%s", expectedLiteral, literal.TokenLiteral())
	}
	checkParserErrors(t, p)
}

func TestIdentifierExpression(t *testing.T) {
	input := "select foo"
	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.SelectStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.SelectStatement. got=%T", program.Statements[0])
	}

	if len(stmt.Expressions) != 1 {
		t.Fatalf("stmt does not contain %d expressions. got=%d", 1, len(stmt.Expressions))
	}

	literal, ok := stmt.Expressions[0].(*ast.Identifier)
	if !ok {
		t.Fatalf("exp not *ast.Identifier. got=%T", stmt.Expressions[0])
	}
	expectedLiteral := "foo"
	if literal.TokenLiteral() != expectedLiteral {
		t.Errorf("literal.TokenLiteral not %s. got=%s", expectedLiteral, literal.TokenLiteral())
	}
	checkParserErrors(t, p)
}

func TestSelectMultiple(t *testing.T) {
	input := "select 5, 'abc',0"
	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	log.Println(program)
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.SelectStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.SelectStatement. got=%T", program.Statements[0])
	}

	if len(stmt.Expressions) != 3 {
		t.Fatalf("stmt does not contain %d expressions. got=%d", 3, len(stmt.Expressions))
	}

	literal1, ok := stmt.Expressions[0].(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("exp not *ast.IntegerLiteral. got=%T", stmt.Expressions[0])
	}
	expectedLiteral1 := "5"
	if literal1.TokenLiteral() != expectedLiteral1 {
		t.Errorf("literal.TokenLiteral not %s. got=%s", expectedLiteral1, literal1.TokenLiteral())
	}

	literal2, ok := stmt.Expressions[1].(*ast.StringLiteral)
	if !ok {
		t.Fatalf("exp not *ast.StringLiteral. got=%T", stmt.Expressions[1])
	}
	expectedLiteral2 := "abc"
	if literal2.TokenLiteral() != expectedLiteral2 {
		t.Errorf("literal.TokenLiteral not %s. got=%s", expectedLiteral2, literal2.TokenLiteral())
	}

	literal3, ok := stmt.Expressions[2].(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("exp not *ast.IntegerLiteral. got=%T", stmt.Expressions[2])
	}
	expectedLiteral3 := "0"
	if literal3.TokenLiteral() != expectedLiteral3 {
		t.Errorf("literal.TokenLiteral not %s. got=%s", expectedLiteral3, literal3.TokenLiteral())
	}

	checkParserErrors(t, p)
}
