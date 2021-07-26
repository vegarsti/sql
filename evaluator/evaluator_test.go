package evaluator_test

import (
	"strings"
	"testing"

	"github.com/vegarsti/sql/evaluator"
	"github.com/vegarsti/sql/lexer"
	"github.com/vegarsti/sql/object"
	"github.com/vegarsti/sql/parser"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"select 5", 5},
		{"select 10", 10},
		{"select 6497869", 6497869},
		{"select -5", -5},
		{"select -10", -10},
		{"select 5 + 5 + 5 + 5 - 10", 10},
		{"select 2 * 2 * 2 * 2 * 2", 32},
		{"select -50 + 100 + -50", 0},
		{"select 5 * 2 + 10", 20},
		{"select 5 + 2 * 10", 25},
		{"select 20 + 2 * -10", 0},
		{"select 50 + 2 * 2 + 10", 64},
		{"select 2 * (5 + 10)", 30},
		{"select 3 * 3 * 3 + 10", 37},
		{"select 3 * (3 * 3) + 10", 37},
		{"select (5 + 10 * 2 + 15 * 3) * 2 + -10", 130},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		row, ok := evaluated.(*object.Row)
		if !ok {
			t.Fatalf("object is not Row. got=%T", row)
		}
		if len(row.Values) != 1 {
			t.Fatalf("expected row to contain 1 element. got=%d", len(row.Values))
		}
		testIntegerObject(t, row.Values[0], tt.expected)
	}
}

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	return evaluator.Eval(program)
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%d, want=%d", result.Value, expected)
		return false
	}
	return true
}

func TestEvalFloatExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"select 5.1", 5.1},
		{"select 3.14", 3.14},
		{"select 0.", 0},
		{"select 1 * 3.14", 3.14},
		{"select 3.14 * 1", 3.14},
		{"select 1 / 2", 0.5},
		{"select 1 / 1", 1},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		row, ok := evaluated.(*object.Row)
		if !ok {
			t.Fatalf("object is not Row. got=%T", row)
		}
		if len(row.Values) != 1 {
			t.Fatalf("expected row to contain 1 element. got=%d", len(row.Values))
		}
		testFloatObject(t, row.Values[0], tt.expected)
	}
}

func testFloatObject(t *testing.T, obj object.Object, expected float64) bool {
	result, ok := obj.(*object.Float)
	if !ok {
		t.Errorf("object is not Float. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%f, want=%f", result.Value, expected)
		return false
	}
	return true
}

func TestEvalStringExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"select 'abc'", "abc"},
		{"select 'def'", "def"},
		{`select 'a string with spaces and "quotes"'`, `a string with spaces and "quotes"`},
		{"select '🤩'", "🤩"},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		row, ok := evaluated.(*object.Row)
		if !ok {
			t.Fatalf("object is not Row. got=%T", row)
		}
		if len(row.Values) != 1 {
			t.Fatalf("expected row to contain 1 element. got=%d", len(row.Values))
		}
		testStringObject(t, row.Values[0], tt.expected)
	}
}

func testStringObject(t *testing.T, obj object.Object, expected string) bool {
	result, ok := obj.(*object.String)
	if !ok {
		t.Errorf("object is not String. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%s, want=%s", result.Value, expected)
		return false
	}
	return true
}

func TestEvalIdentifierExpression(t *testing.T) {
	tests := []struct {
		input                string
		expectedErrorMessage string
	}{
		{"select foo", "no such column: foo"},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testError(t, evaluated, tt.expectedErrorMessage)
	}
}

func testError(t *testing.T, obj object.Object, expectedMessage string) bool {
	result, ok := obj.(*object.Error)
	if !ok {
		t.Errorf("object is not Error. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Message != expectedMessage {
		t.Errorf("object has wrong value. got=%s, want=%s", result.Message, expectedMessage)
		return false
	}
	return true
}

func TestEvalSelectMultiple(t *testing.T) {
	tests := []struct {
		input          string
		expectedValues []interface{}
		expectedNames  []string
	}{
		{
			"select 'abc', 1 as n, 3.14 as pi, -1",
			[]interface{}{"abc", int64(1), float64(3.14), int64(-1)},
			[]string{"?column?", "n", "pi", "?column?"},
		},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		row, ok := evaluated.(*object.Row)
		if !ok {
			t.Fatalf("object is not Row. got=%T", row)
		}
		if len(row.Values) != 4 {
			t.Fatalf("expected row.Values to contain 4 elements. got=%d", len(row.Values))
		}
		if len(row.Names) != 4 {
			t.Fatalf("expected row.Names to have 4 elements. got=%d", len(row.Names))
		}

		// assert values
		s := tt.expectedValues[0].(string)
		testStringObject(t, row.Values[0], s)
		n := tt.expectedValues[1].(int64)
		testIntegerObject(t, row.Values[1], n)
		f := tt.expectedValues[2].(float64)
		testFloatObject(t, row.Values[2], f)
		m := tt.expectedValues[3].(int64)
		testIntegerObject(t, row.Values[3], m)

		// assert names
		expectedRowNames := strings.Join(tt.expectedNames, ", ")
		rowNames := strings.Join(row.Names, ", ")
		if rowNames != expectedRowNames {
			t.Fatalf("expected row names to be [%s], but was [%s]", expectedRowNames, rowNames)
		}
	}
}
