package parser_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/vegarsti/sql/ast"
	"github.com/vegarsti/sql/lexer"
	"github.com/vegarsti/sql/parser"
	"github.com/vegarsti/sql/token"
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
		t.Fatalf("program.Statements does not contain 1 statement. got=%d", len(program.Statements))
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

func TestNull(t *testing.T) {
	input := "select null"
	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.SelectStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.SelectStatement. got=%T", program.Statements[0])
	}

	if len(stmt.Expressions) != 1 {
		t.Fatalf("stmt does not contain %d expressions. got=%d", 1, len(stmt.Expressions))
	}

	literal, ok := stmt.Expressions[0].(*ast.Null)
	if !ok {
		t.Fatalf("exp not *ast.Null. got=%T", stmt.Expressions[0])
	}
	expectedLiteral := "NULL"
	if literal.TokenLiteral() != expectedLiteral {
		t.Errorf("literal.TokenLiteral not %s. got=%s", expectedLiteral, literal.TokenLiteral())
	}
	checkParserErrors(t, p)
}

func TestBooleanLiteralExpression(t *testing.T) {
	input := "select true, false"
	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.SelectStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.SelectStatement. got=%T", program.Statements[0])
	}

	expectedLen := 2
	if len(stmt.Expressions) != expectedLen {
		t.Fatalf("stmt does not contain %d expressions. got=%d", 2, len(stmt.Expressions))
	}

	literal1, ok := stmt.Expressions[0].(*ast.BooleanLiteral)
	if !ok {
		t.Fatalf("exp not *ast.BooleanLiteral. got=%T", stmt.Expressions[0])
	}
	expectedTokenLiteral1 := "TRUE"
	if literal1.TokenLiteral() != expectedTokenLiteral1 {
		t.Errorf("literal.TokenLiteral not %s. got=%s", expectedTokenLiteral1, literal1.TokenLiteral())
	}

	literal2, ok := stmt.Expressions[1].(*ast.BooleanLiteral)
	if !ok {
		t.Fatalf("exp not *ast.BooleanLiteral. got=%T", stmt.Expressions[0])
	}
	expectedTokenLiteral2 := "FALSE"
	if literal2.TokenLiteral() != expectedTokenLiteral2 {
		t.Errorf("literal.TokenLiteral not %s. got=%s", expectedTokenLiteral2, literal2.TokenLiteral())
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

func TestNullPrefixExpression(t *testing.T) {
	prefixTest := []struct {
		input    string
		operator string
	}{
		{"select null is null", "IS NULL"},
		{"select null is not null", "IS NOT NULL"},
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

		exp, ok := stmt.Expressions[0].(*ast.PostfixExpression)
		if !ok {
			t.Fatalf("stmt is not ast.PostfixExpression. got=%T", stmt.Expressions[0])
		}
		if exp.Operator != tt.operator {
			t.Fatalf("exp.Operator is not '%s'. got=%s", tt.operator, exp.Operator)
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

func testBooleanLiteral(t *testing.T, bl ast.Expression, value bool) bool {
	b, ok := bl.(*ast.BooleanLiteral)
	if !ok {
		t.Errorf("bl not *ast.BooleanLiteral. got=%T", bl)
		return false
	}

	if b.Value != value {
		t.Errorf("b.Value not %t. got=%t", value, b.Value)
		return false
	}

	if b.TokenLiteral() != strings.ToUpper(fmt.Sprintf("%t", value)) {
		t.Errorf("b.TokenLiteral not %t. got=%s", value, b.TokenLiteral())
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

func TestParseBooleanPrefixExpression(t *testing.T) {
	prefixTest := []struct {
		input    string
		operator string
		value    bool
	}{
		{"select !true", "!", true},
		{"select !false", "!", false},
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
		if !testBooleanLiteral(t, exp.Right, tt.value) {
			return
		}
	}
}

func TestParseBooleanInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  bool
		operator   string
		rightValue bool
	}{
		{"select true = true", true, "=", true},
		{"select true != false", true, "!=", false},
		{"select true and false", true, "AND", false},
		{"select true or false", true, "OR", false},
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
		if !testBooleanLiteral(t, exp.Left, tt.leftValue) {
			return
		}
		if exp.Operator != tt.operator {
			t.Fatalf("exp.Operator is not '%s'. got=%s", tt.operator, exp.Operator)
		}
		if !testBooleanLiteral(t, exp.Right, tt.rightValue) {
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

func TestQualifiedIdentifier(t *testing.T) {
	input := "select foo.bar"
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
	expectedLiteral := "foo.bar"
	if literal.TokenLiteral() != expectedLiteral {
		t.Errorf("literal.TokenLiteral not %s. got=%s", expectedLiteral, literal.TokenLiteral())
	}
	expectedTable := "foo"
	if literal.Table != expectedTable {
		t.Errorf("literal.Table not %s. got=%s", expectedTable, literal.Table)
	}
	expectedValue := "bar"
	if literal.Value != expectedValue {
		t.Errorf("literal.Value not %s. got=%s", expectedValue, literal.Value)
	}
	checkParserErrors(t, p)
}

func TestSelectCartesianJoin(t *testing.T) {
	tt := []string{
		"select f.a, b.b from foo f, bar b",
	}
	for _, input := range tt {
		l := lexer.New(input)
		p := parser.New(l)

		program := p.ParseProgram()
		if program == nil {
			t.Fatalf("ParseProgram() returned nil")
		}

		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.SelectStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.SelectStatement. got=%T", program.Statements[0])
		}

		if len(stmt.Expressions) != 2 {
			t.Fatalf("stmt does not contain %d expressions. got=%d", 1, len(stmt.Expressions))
		}

		literal1, ok := stmt.Expressions[0].(*ast.Identifier)
		if !ok {
			t.Fatalf("exp not *ast.Identifier. got=%T", stmt.Expressions[0])
		}
		expectedLiteral1 := "f.a"
		if literal1.TokenLiteral() != expectedLiteral1 {
			t.Errorf("literal.TokenLiteral not %s. got=%s", expectedLiteral1, literal1.TokenLiteral())
		}
		expectedTable1 := "f"
		if literal1.Table != expectedTable1 {
			t.Errorf("literal.Table not %s. got=%s", expectedTable1, literal1.Table)
		}
		expectedValue1 := "a"
		if literal1.Value != expectedValue1 {
			t.Errorf("literal.Value not %s. got=%s", expectedValue1, literal1.Value)
		}

		literal2, ok := stmt.Expressions[1].(*ast.Identifier)
		if !ok {
			t.Fatalf("exp not *ast.Identifier. got=%T", stmt.Expressions[1])
		}
		expectedLiteral2 := "b.b"
		if literal2.TokenLiteral() != expectedLiteral2 {
			t.Errorf("literal.TokenLiteral not %s. got=%s", expectedLiteral2, literal2.TokenLiteral())
		}
		expectedTable2 := "b"
		if literal2.Table != expectedTable2 {
			t.Errorf("literal.Table not %s. got=%s", expectedTable2, literal2.Table)
		}
		expectedValue2 := "b"
		if literal2.Value != expectedValue2 {
			t.Errorf("literal.Value not %s. got=%s", expectedValue2, literal2.Value)
		}

		expectedFromLen := 2
		if len(stmt.From) != expectedFromLen {
			t.Fatalf("stmt.From not length %d. got=%d", expectedFromLen, len(stmt.From))
		}

		expectedFrom1 := "foo"
		if stmt.From[0].Table != expectedFrom1 {
			t.Fatalf("stmt.From[0] not %s. got=%s", expectedFrom1, stmt.From[0].Table)
		}
		expectedFromAlias1 := "f"
		if stmt.From[0].TableAlias != expectedFromAlias1 {
			t.Fatalf("stmt.FromAlias[0] not %s. got=%s", expectedFromAlias1, stmt.From[0].Table)
		}

		expectedFrom2 := "bar"
		if stmt.From[1].Table != expectedFrom2 {
			t.Fatalf("stmt.From[1] not %s. got=%s", expectedFrom2, stmt.From[1].Table)
		}
		expectedFromAlias2 := "b"
		if stmt.From[1].TableAlias != expectedFromAlias2 {
			t.Fatalf("stmt.FromAlias[1] not %s. got=%s", expectedFromAlias2, stmt.From[1].Table)
		}
	}
}

func TestSelectMultiple(t *testing.T) {
	input := "select 5, 'abc',0"
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

func TestSelectWithAs(t *testing.T) {
	input := "select 1, 5 as n, 'abc' as str, 2"
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

	if len(stmt.Expressions) != 4 {
		t.Fatalf("stmt does not contain %d expressions. got=%d", 4, len(stmt.Expressions))
	}

	if len(stmt.Aliases) != 4 {
		t.Fatalf("stmt does not contain %d names. got=%d", 4, len(stmt.Aliases))
	}

	literal1, ok := stmt.Expressions[0].(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("exp not *ast.IntegerLiteral. got=%T", stmt.Expressions[0])
	}
	expectedLiteral1 := "1"
	if literal1.TokenLiteral() != expectedLiteral1 {
		t.Errorf("literal.TokenLiteral not %s. got=%s", expectedLiteral1, literal1.TokenLiteral())
	}
	expectedName1 := ""
	if stmt.Aliases[0] != expectedName1 {
		t.Errorf("name not %s. got=%s", expectedName1, stmt.Aliases[0])
	}

	literal2, ok := stmt.Expressions[1].(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("exp not *ast.IntegerLiteral. got=%T", stmt.Expressions[1])
	}
	expectedLiteral2 := "5"
	if literal2.TokenLiteral() != expectedLiteral2 {
		t.Errorf("literal.TokenLiteral not %s. got=%s", expectedLiteral2, literal2.TokenLiteral())
	}
	expectedName2 := "n"
	if stmt.Aliases[1] != expectedName2 {
		t.Errorf("name not %s. got=%s", expectedName2, stmt.Aliases[1])
	}

	literal3, ok := stmt.Expressions[2].(*ast.StringLiteral)
	if !ok {
		t.Fatalf("exp not *ast.StringLiteral. got=%T", stmt.Expressions[2])
	}
	expectedLiteral3 := "abc"
	if literal3.TokenLiteral() != expectedLiteral3 {
		t.Errorf("literal.TokenLiteral not %s. got=%s", expectedLiteral3, literal3.TokenLiteral())
	}
	expectedName3 := "str"
	if stmt.Aliases[2] != expectedName3 {
		t.Errorf("name not %s. got=%s", expectedName3, stmt.Aliases[2])
	}

	literal4, ok := stmt.Expressions[3].(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("exp not *ast.IntegerLiteral. got=%T", stmt.Expressions[3])
	}
	expectedLiteral4 := "2"
	if literal4.TokenLiteral() != expectedLiteral4 {
		t.Errorf("literal.TokenLiteral not %s. got=%s", expectedLiteral4, literal4.TokenLiteral())
	}
	expectedName4 := ""
	if stmt.Aliases[3] != expectedName1 {
		t.Errorf("name not %s. got=%s", expectedName4, stmt.Aliases[3])
	}

	checkParserErrors(t, p)
}

func TestSelectOrderBy(t *testing.T) {
	input := "select a from foo order by a desc, b + 1 asc"
	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	checkParserErrors(t, p)

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
	expectedLiteral := "a"
	if literal.TokenLiteral() != expectedLiteral {
		t.Fatalf("literal.TokenLiteral not %s. got=%s", expectedLiteral, literal.TokenLiteral())
	}

	expectedFrom := "foo"
	expectedFromLen := 1
	if len(stmt.From) != expectedFromLen {
		t.Fatalf("stmt.From not length %d. got=%d", expectedFromLen, len(stmt.From))
	}
	if stmt.From[0].Table != expectedFrom {
		t.Fatalf("stmt.From not %s. got=%s", expectedFrom, stmt.From[0].Table)
	}

	expectedOrderBy := []string{"a", "(b + 1)"}
	if len(stmt.OrderBy) != len(expectedOrderBy) {
		t.Fatalf("stmt.OrderBy has length %d, expected %d", len(stmt.OrderBy), len(expectedOrderBy))
	}
	expectedOrderByString := strings.Join(expectedOrderBy, ", ")
	gotOrderByStrings := make([]string, len(stmt.OrderBy))
	for i, orderBy := range stmt.OrderBy {
		gotOrderByStrings[i] = orderBy.Expression.String()
	}
	if !stmt.OrderBy[0].Descending {
		t.Fatalf("expected stmt.OrderBy[0].Expression.Descending to be true. got false")
	}
	if stmt.OrderBy[1].Descending {
		t.Fatalf("expected stmt.OrderBy[1].Expression.Descending to be false. got true")
	}
	gotOrderByString := strings.Join(gotOrderByStrings, ", ")
	if gotOrderByString != expectedOrderByString {
		t.Fatalf("expected stmt.OrderBy to be %s. got=%s", expectedOrderByString, gotOrderByString)
	}
}

func TestSelectLimit(t *testing.T) {
	input := "select a from foo limit 1 offset 1"
	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	checkParserErrors(t, p)

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
	expectedLiteral := "a"
	if literal.TokenLiteral() != expectedLiteral {
		t.Fatalf("literal.TokenLiteral not %s. got=%s", expectedLiteral, literal.TokenLiteral())
	}

	expectedFrom := "foo"
	expectedFromLen := 1
	if len(stmt.From) != expectedFromLen {
		t.Fatalf("stmt.From not length %d. got=%d", expectedFromLen, len(stmt.From))
	}
	if stmt.From[0].Table != expectedFrom {
		t.Fatalf("stmt.From not %s. got=%s", expectedFrom, stmt.From[0].Table)
	}

	expectedLimit := 1
	if stmt.Limit == nil {
		t.Fatalf("stmt.Limit is nil")
	}
	if *stmt.Limit != expectedLimit {
		t.Fatalf("stmt.Limit not %d. got=%d", expectedLimit, *stmt.Limit)
	}

	expectedOffset := 1
	if stmt.Offset == nil {
		t.Fatalf("stmt.Offset is nil")
	}
	if *stmt.Offset != expectedOffset {
		t.Fatalf("stmt.Offset not %d. got=%d", expectedOffset, *stmt.Offset)
	}
}

func TestSelectWhere(t *testing.T) {
	input := "select a from foo where a = b and a"
	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	checkParserErrors(t, p)

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
	expectedLiteral := "a"
	if literal.TokenLiteral() != expectedLiteral {
		t.Fatalf("literal.TokenLiteral not %s. got=%s", expectedLiteral, literal.TokenLiteral())
	}

	expectedFrom := "foo"
	expectedFromLen := 1
	if len(stmt.From) != expectedFromLen {
		t.Fatalf("stmt.From not length %d. got=%d", expectedFromLen, len(stmt.From))
	}
	if stmt.From[0].Table != expectedFrom {
		t.Fatalf("stmt.From not %s. got=%s", expectedFrom, stmt.From[0].Table)
	}

	if stmt.Where == nil {
		t.Fatalf("stmt.Where is nil")
	}
	expectedWhere := "((a = b) AND a)"
	gotWhere := stmt.Where.String()
	if gotWhere != expectedWhere {
		t.Fatalf("expected stmt.Where to be %s. got=%s", expectedWhere, gotWhere)
	}
}

func TestCreateTable(t *testing.T) {
	input := "create table foo (a text, b integer, c float, d bool, e boolean, f int)"
	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.CreateTableStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.CreateTableStatement. got=%T", program.Statements[0])
	}

	if stmt == nil {
		t.Fatalf("ast.CreateTableStatement is nil")
	}

	expectedColumnNames := []string{
		"a",
		"b",
		"c",
		"d",
		"e",
		"f",
	}

	expectedColumnTypes := []token.Token{
		{Type: token.STRING_TYPE, Literal: token.STRING_TYPE},
		{Type: token.INTEGER_TYPE, Literal: token.INTEGER_TYPE},
		{Type: token.FLOAT_TYPE, Literal: token.FLOAT_TYPE},
		{Type: token.BOOLEAN_TYPE, Literal: token.BOOLEAN_TYPE},
		{Type: token.BOOLEAN_TYPE, Literal: token.BOOLEAN_TYPE},
		{Type: token.INTEGER_TYPE, Literal: token.INTEGER_TYPE},
	}

	if len(stmt.ColumnNames) != len(expectedColumnNames) {
		t.Fatalf("stmt does not contain %d column names. got=%d", len(stmt.ColumnNames), len(expectedColumnNames))
	}
	if len(stmt.ColumnTypes) != len(expectedColumnTypes) {
		t.Fatalf("stmt does not contain %d column types. got=%d", len(stmt.ColumnTypes), len(expectedColumnTypes))
	}

	for i := range expectedColumnNames {
		if stmt.ColumnNames[i] != expectedColumnNames[i] {
			t.Fatalf("expected column name %s. got=%s", expectedColumnNames[i], stmt.ColumnNames[i])
		}
		name := stmt.ColumnNames[i]
		if stmt.ColumnTypes[i].Literal != expectedColumnTypes[i].Literal {
			t.Fatalf("expected token literal %s for column %s. got=%s", expectedColumnTypes[i].Literal, name, stmt.ColumnTypes[i].Literal)
		}
		if stmt.ColumnTypes[i].Type != expectedColumnTypes[i].Type {
			t.Fatalf("expected token type %T for column %s. got=%T", expectedColumnTypes[i].Type, name, stmt.ColumnTypes[i].Type)
		}
	}
}

func TestInsert(t *testing.T) {
	input := "insert into foo values ('abc', 1, 3.14)"
	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.InsertStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.InsertStatement. got=%T", program.Statements[0])
	}

	if stmt == nil {
		t.Fatalf("ast.InsertStatement is nil")
	}

	expectedExpressions := []ast.Expression{
		&ast.StringLiteral{Token: token.Token{Literal: "abc", Type: token.STRING_LITERAL}},
		&ast.IntegerLiteral{Token: token.Token{Literal: "1", Type: token.INT_LITERAL}},
		&ast.FloatLiteral{Token: token.Token{Literal: "3.14", Type: token.FLOAT_LITERAL}},
	}

	if len(stmt.Expressions) != len(expectedExpressions) {
		t.Fatalf("stmt does not contain %d expressions. got=%d", len(stmt.Expressions), len(expectedExpressions))
	}

	for i, expectedExpr := range expectedExpressions {
		if stmt.Expressions[i].TokenLiteral() != expectedExpr.TokenLiteral() {
			t.Fatalf("expected stmt.Expressions[%d].TokenLiteral() to be %s. got=%s", i, expectedExpr.TokenLiteral(), stmt.Expressions[i].TokenLiteral())
		}
		if stmt.Expressions[i].String() != expectedExpr.String() {
			t.Fatalf("expected stmt.Expressions[%d].String() to be %s. got=%s", i, expectedExpr.String(), stmt.Expressions[i].String())
		}
	}
}

func TestSelectFrom(t *testing.T) {
	input := "select a from foo"
	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	checkParserErrors(t, p)

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
	expectedLiteral := "a"
	if literal.TokenLiteral() != expectedLiteral {
		t.Fatalf("literal.TokenLiteral not %s. got=%s", expectedLiteral, literal.TokenLiteral())
	}

	expectedFrom := "foo"
	expectedFromLen := 1
	if len(stmt.From) != expectedFromLen {
		t.Fatalf("stmt.From not length %d. got=%d", expectedFromLen, len(stmt.From))
	}
	if stmt.From[0].Table != expectedFrom {
		t.Fatalf("stmt.From not %s. got=%s", expectedFrom, stmt.From[0].Table)
	}
}

func TestParseMultipleStatementsOK(t *testing.T) {
	prefixTest := []struct {
		input                 string
		expectedStatementsLen int
	}{
		{"select 1;", 1},
		{"select 1; select 2 as n; select 3", 3},
		{"insert into foo values ('a', 'b', 'c')", 1},
		{"insert into foo values ('a', 'b', 'c');", 1},
		{"insert into foo values ('a', 'b', 'c'); select 1", 2},
		{"create table foo (a text, b integer, c float);", 1},
		{"create table foo (a text, b integer, c float); select 1", 2},
		{"create table foo (a text, b integer, c float); select 1; insert into foo values ('a', 'b', 'c')", 3},
	}
	for _, tt := range prefixTest {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)
		if len(program.Statements) != tt.expectedStatementsLen {
			t.Fatalf("program.Statements does not contain %d statements. got=%d", tt.expectedStatementsLen, len(program.Statements))
		}
	}
}

func TestParseMultipleStatementsError(t *testing.T) {
	prefixTest := []struct {
		input string
	}{
		{"select 1 select 2"},
		{"create table foo (a text, b integer, c double) create table bar (a text, b integer, c double)"},
		{"insert into foo values ('a', 'b', 'c') insert into foo values ('a', 'b', 'c')"},
	}
	for _, tt := range prefixTest {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		errors := p.Errors()
		if len(errors) != 1 {
			t.Fatalf("expected 1 parser error, got %d", len(errors))
		}
		if len(program.Statements) != 0 {
			t.Fatalf("program.Statements is not empty. got=%d", len(program.Statements))
		}
	}
}

func TestSelectJoin(t *testing.T) {
	input := "select a from foo join bar on foo.a = bar.b, baz join qux on x = y"
	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	checkParserErrors(t, p)

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
	expectedLiteral := "a"
	if literal.TokenLiteral() != expectedLiteral {
		t.Fatalf("literal.TokenLiteral not %s. got=%s", expectedLiteral, literal.TokenLiteral())
	}

	expectedFromLength := 2
	if len(stmt.From) != expectedFromLength {
		t.Fatalf("stmt.From not length %d. got=%d", expectedFromLength, len(stmt.From))
	}

	expectedFrom1 := "foo"
	if stmt.From[0].Table != expectedFrom1 {
		t.Fatalf("stmt.From[0].Table not %s. got=%s", expectedFrom1, stmt.From[0].Table)
	}
	if stmt.From[0].Join == nil {
		t.Fatalf("stmt.From[0].Join is nil")
	}
	expectedJoinTable1 := "bar"
	if stmt.From[0].Join.With.Table != expectedJoinTable1 {
		t.Fatalf("stmt.From[0].Join.Table is not %s. got=%s", expectedJoinTable1, stmt.From[0].Join.With.Table)
	}
	expectedJoinPred1 := "(a = b)"
	if stmt.From[0].Join.Predicate.String() != expectedJoinPred1 {
		t.Fatalf("stmt.From[0].Join.Predicate is not %s. got=%s", expectedJoinPred1, stmt.From[0].Join.Predicate)
	}

	expectedFrom2 := "baz"
	if stmt.From[1].Table != expectedFrom2 {
		t.Fatalf("stmt.From[1].Table not %s. got=%s", expectedFrom2, stmt.From[1].Table)
	}
	if stmt.From[1].Join == nil {
		t.Fatalf("stmt.From[1].Join is nil")
	}
	expectedJoinTable2 := "qux"
	if stmt.From[1].Join.With.Table != expectedJoinTable2 {
		t.Fatalf("stmt.From[1].Join.Table is not %s. got=%s", expectedJoinTable2, stmt.From[1].Join.With.Table)
	}
	expectedJoinPred2 := "(x = y)"
	if stmt.From[1].Join.Predicate.String() != expectedJoinPred2 {
		t.Fatalf("stmt.From[1].Join.Predicate is not %s. got=%s", expectedJoinPred2, stmt.From[1].Join.Predicate)
	}
}
