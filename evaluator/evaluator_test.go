package evaluator_test

import (
	"fmt"
	"math"
	"sort"
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
		{"select 7 % 2", 1},
		{"select 5 ^ 2", 25},
		{"select 2*5^2+1", 51},
		{"select 1 / 2", 0},
		{"select 1 / 1", 1},
	}
	for _, tt := range tests {
		evaluated := testEval(newTestBackend(), tt.input)
		result, ok := evaluated.(*object.Result)
		if !ok {
			if errorEvaluated, errorOK := evaluated.(*object.Error); errorOK {
				t.Fatalf("object is Error: %s", errorEvaluated.Inspect())
			}
			t.Fatalf("object is not Result. got=%T", evaluated)
		}
		if len(result.Rows) != 1 {
			t.Fatalf("expected result to contain 1 row. got=%d", len(result.Rows))
		}
		row := result.Rows[0]
		if len(row.Values) != 1 {
			t.Fatalf("expected row to contain 1 element. got=%d", len(row.Values))
		}
		testIntegerObject(t, row.Values[0], tt.expected)
	}
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

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"select true", true},
		{"select false", false},
		{"select true = true", true},
		{"select true != true", false},
		{"select not true", false},
		{"select not false", true},
		{"select false and true", false},
		{"select false or true", true},
		{"select false or not true", false},
		{"select 1 < 1", false},
		{"select 1 <= 1", true},
		{"select 1 >= 1", true},
		{"select 1 > 1", false},
		{"select null is null", true},
		{"select null is not null", false},
		{"select 1 is not null", true},
	}
	for _, tt := range tests {
		evaluated := testEval(newTestBackend(), tt.input)
		result, ok := evaluated.(*object.Result)
		if !ok {
			if errorEvaluated, errorOK := evaluated.(*object.Error); errorOK {
				t.Fatalf("object is Error: %s", errorEvaluated.Inspect())
			}
			t.Fatalf("%s: object is not Result. got=%T", tt.input, evaluated)
		}
		if len(result.Rows) != 1 {
			t.Fatalf("expected result to contain 1 row. got=%d", len(result.Rows))
		}
		row := result.Rows[0]
		if len(row.Values) != 1 {
			t.Fatalf("expected row to contain 1 element. got=%d", len(row.Values))
		}
		testBooleanObject(t, row.Values[0], tt.expected)
	}
}

func TestEvalNullExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected object.Object
	}{
		{"select null", object.NULL},
	}
	for _, tt := range tests {
		evaluated := testEval(newTestBackend(), tt.input)
		result, ok := evaluated.(*object.Result)
		if !ok {
			if errorEvaluated, errorOK := evaluated.(*object.Error); errorOK {
				t.Fatalf("object is Error: %s", errorEvaluated.Inspect())
			}
			t.Fatalf("object is not Result. got=%T", evaluated)
		}
		if len(result.Rows) != 1 {
			t.Fatalf("expected result to contain 1 row. got=%d", len(result.Rows))
		}
		row := result.Rows[0]
		if len(row.Values) != 1 {
			t.Fatalf("expected row to contain 1 element. got=%d", len(row.Values))
		}
		testNull(t, row.Values[0])
	}
}

func testNull(t *testing.T, obj object.Object) bool {
	if _, ok := obj.(*object.Null); !ok {
		t.Errorf("object is not Null. got=%T (%+v)", obj, obj)
		return false
	}
	return true
}

func testEval(backend *testBackend, input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	return evaluator.Eval(backend, program)
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%t, want=%t", result.Value, expected)
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
		{"select 5.0 ^ 2.0", 25.0},
		{"select 4.8 % 2", 0.8},
	}
	for _, tt := range tests {
		evaluated := testEval(newTestBackend(), tt.input)
		result, ok := evaluated.(*object.Result)
		if !ok {
			t.Fatalf("object is not Result. got=%T", evaluated)
		}
		if len(result.Rows) != 1 {
			t.Fatalf("expected result to contain 1 row. got=%d", len(result.Rows))
		}
		row := result.Rows[0]
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
	if math.Abs(result.Value-expected) > 0.01 {
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
		{"select 'hello' || 'world'", "helloworld"},
	}
	for _, tt := range tests {
		evaluated := testEval(newTestBackend(), tt.input)
		result, ok := evaluated.(*object.Result)
		if !ok {
			t.Fatalf("object is not Result. got=%T", evaluated)
		}
		if len(result.Rows) != 1 {
			t.Fatalf("expected result to contain 1 row. got=%d", len(result.Rows))
		}
		row := result.Rows[0]
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

func TestErrors(t *testing.T) {
	tests := []struct {
		input                string
		expectedErrorMessage string
	}{
		{"select foo", `column "foo" does not exist`},
		{"select bar.foo", `missing FROM-clause entry for table "bar"`},
		{"select a from foo, bar", `column reference "a" is ambiguous`},
		{"select a from foo order by d", `column "d" does not exist`},
		{"select a from foo where 1", `argument of WHERE must be type boolean, not type integer: 1`},
		{"select foo.a from foo f where 1", `invalid reference to FROM-clause entry for table "foo". Perhaps you meant to reference the table alias "f"`},
		{"select 1 from foo, foo", `table name "foo" specified more than once`},
		{"insert into foo values (1)", `table "foo" has 2 columns but 1 value were supplied`},
		{"insert into foo values (1)", `table "foo" has 2 columns but 1 value were supplied`},
		{"insert into foo values ('hello', 'world')", `cannot insert STRING with value 'world' in INTEGER column in table "foo"`},
		{"insert into foo values (1, 2)", `cannot insert INTEGER with value 1 in STRING column in table "foo"`},
	}
	for _, tt := range tests {
		backend := newTestBackend()

		// table `foo`
		backend.tables["foo"] = []object.Column{
			{Name: "a", Type: object.STRING},
			{Name: "c", Type: object.INTEGER},
		}
		backend.rows["foo"] = []object.Row{
			{
				Values: []object.Object{
					&object.String{Value: "abc"},
					&object.Integer{Value: 1},
				},
				Aliases:   []string{"a", "c"},
				TableName: []string{"foo", "foo"},
			},
			{
				Values: []object.Object{
					&object.String{Value: "bcd"},
					&object.Integer{Value: 2},
				},
				Aliases:   []string{"a", "c"},
				TableName: []string{"foo", "foo"},
			},
		}

		// table `bar`
		backend.tables["bar"] = []object.Column{{Name: "a", Type: object.STRING}}

		evaluated := testEval(backend, tt.input)
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
			[]string{"'abc'", "n", "pi", "(-1)"},
		},
	}
	for _, tt := range tests {
		evaluated := testEval(newTestBackend(), tt.input)
		result, ok := evaluated.(*object.Result)
		if !ok {
			t.Fatalf("object is not Result. got=%T", evaluated)
		}
		if len(result.Rows) != 1 {
			t.Fatalf("expected result to contain 1 row. got=%d", len(result.Rows))
		}
		row := result.Rows[0]
		if len(row.Values) != 4 {
			t.Fatalf("expected row.Values to contain 4 elements. got=%d", len(row.Values))
		}
		if len(row.Aliases) != 4 {
			t.Fatalf("expected row.Names to have 4 elements. got=%d", len(row.Aliases))
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
		rowNames := strings.Join(row.Aliases, ", ")
		if rowNames != expectedRowNames {
			t.Fatalf("expected row names to be [%s], but was [%s]", expectedRowNames, rowNames)
		}
	}
}

type testBackend struct {
	tables map[string][]object.Column
	rows   map[string][]object.Row
}

func (tb *testBackend) CreateTable(name string, columns []object.Column) error {
	if _, ok := tb.tables[name]; ok {
		return fmt.Errorf(`relation "%s" already exists`, name)
	}
	tb.tables[name] = columns
	tb.rows[name] = make([]object.Row, 0)
	return nil
}

func (tb *testBackend) Insert(name string, row object.Row) error {
	if _, ok := tb.tables[name]; !ok {
		return fmt.Errorf(`relation "%s" does not exist`, name)
	}
	tb.rows[name] = append(tb.rows[name], row)
	// Populate aliases
	for i := range tb.rows[name] {
		tb.rows[name][i].Aliases = make([]string, len(tb.tables[name]))
		for j, column := range tb.tables[name] {
			tb.rows[name][i].Aliases[j] = column.Name
		}
	}
	return nil
}

func (tb *testBackend) Rows(name string) ([]object.Row, error) {
	rows, ok := tb.rows[name]
	if !ok {
		return nil, fmt.Errorf(`relation "%s" does not exist`, name)
	}
	return rows, nil
}

func (tb *testBackend) Columns(name string) ([]object.Column, error) {
	return tb.tables[name], nil
}

func newTestBackend() *testBackend {
	return &testBackend{
		tables: make(map[string][]object.Column),
		rows:   make(map[string][]object.Row),
	}
}

func TestEvalCreateTable(t *testing.T) {
	tests := []struct {
		input         string
		tableName     string
		expectedTable object.Table
	}{
		{
			"create table foo (a text, b integer, c float)",
			"foo",
			object.Table{
				Name: "foo",
				Columns: []object.Column{
					{Name: "a", Type: object.STRING},
					{Name: "b", Type: object.INTEGER},
					{Name: "c", Type: object.FLOAT},
				},
			},
		},
	}
	for _, tt := range tests {
		backend := newTestBackend()
		evaluated := testEval(backend, tt.input)
		if _, ok := evaluated.(*object.OK); !ok {
			t.Fatalf("object is not OK. got=%T", evaluated)
		}
		columns := backend.tables[tt.tableName]
		expectedColumns := tt.expectedTable.Columns
		if len(columns) != len(expectedColumns) {
			t.Fatalf("expected %d columns. got=%d", len(expectedColumns), len(columns))
		}

		// Assert column lists are equal
		expectedColumnsMap := make(map[string]object.DataType)
		expectedColumnsMapKeys := make([]string, len(expectedColumns))
		for i, s := range expectedColumns {
			expectedColumnsMap[s.Name] = s.Type
			expectedColumnsMapKeys[i] = s.Name
		}
		sort.Strings(expectedColumnsMapKeys)
		columnsMap := make(map[string]object.DataType)
		columnsMapKeys := make([]string, len(columns))
		for i, s := range columns {
			columnsMap[s.Name] = s.Type
			columnsMapKeys[i] = s.Name
		}
		sort.Strings(columnsMapKeys)
		for i := range columns {
			if columnsMapKeys[i] != expectedColumnsMapKeys[i] {
				t.Fatalf("expected column %s. got=%s", expectedColumnsMapKeys[i], columnsMapKeys[i])
			}
			if columnsMap[columnsMapKeys[i]] != expectedColumnsMap[expectedColumnsMapKeys[i]] {
				t.Fatalf("expected column type %s. got=%s", expectedColumnsMap[expectedColumnsMapKeys[i]], columnsMap[columnsMapKeys[i]])
			}
		}
	}
}

func TestEvalInsert(t *testing.T) {
	tests := []struct {
		input        string
		expectedRows [][]object.Object
	}{
		{
			"insert into foo values ('abc', 1, 3.14), ('def', 2, 6.28)",
			[][]object.Object{
				{
					&object.String{Value: "abc"},
					&object.Integer{Value: 1},
					&object.Float{Value: 3.14},
				},
				{
					&object.String{Value: "def"},
					&object.Integer{Value: 2},
					&object.Float{Value: 6.28},
				},
			},
		},
	}
	for _, tt := range tests {
		backend := newTestBackend()
		backend.tables["foo"] = []object.Column{
			{Name: "a", Type: object.DataType("STRING")},
			{Name: "b", Type: object.DataType("INTEGER")},
			{Name: "c", Type: object.DataType("FLOAT")},
		}

		evaluated := testEval(backend, tt.input)
		if _, ok := evaluated.(*object.OK); !ok {
			if errorEvaluated, errorOK := evaluated.(*object.Error); errorOK {
				t.Fatalf("%s: %s", tt.input, errorEvaluated.Inspect())
			}
			t.Fatalf("object is not OK. got=%T", evaluated)
		}
		rows, ok := backend.rows["foo"]
		if !ok {
			t.Fatalf("row doesn't exist")
		}
		if len(rows) != len(tt.expectedRows) {
			t.Fatalf("expected table to have %d rows. got=%d", len(tt.expectedRows), len(rows))
		}
		for i, expectedValues := range tt.expectedRows {
			for j := range expectedValues {
				if rows[i].Values[j].Type() != expectedValues[j].Type() {
					t.Fatalf("expected rows[%d][%d] to have %v value. got=%v", i, j, expectedValues[j].Type(), rows[i].Values[j].Type())
				}
				if rows[i].Values[j].Inspect() != expectedValues[j].Inspect() {
					t.Fatalf("expected rows[%d][%d] to have %v value. got=%v", i, j, expectedValues[j].Inspect(), rows[i].Values[j].Inspect())
				}
			}
			expectedAliases := "a, b, c"
			gotAliases := strings.Join(rows[i].Aliases, ", ")
			if gotAliases != expectedAliases {
				t.Fatalf("expected aliases to be %s. got=%s", expectedAliases, gotAliases)
			}
		}
	}
}

func TestEvalSelectFrom(t *testing.T) {
	tests := []struct {
		input           string
		expected        []object.Row
		expectedAliases string
	}{
		{
			"select a, b from foo",
			[]object.Row{
				{
					Values: []object.Object{
						&object.String{Value: "abc"},
						&object.String{Value: "efg"},
					},
				},
				{
					Values: []object.Object{
						&object.String{Value: "bcd"},
						&object.String{Value: "def"},
					},
				},
			},
			"a, b",
		},
		{
			"select a from foo",
			[]object.Row{
				{
					Values: []object.Object{
						&object.String{Value: "abc"},
					},
				},
				{
					Values: []object.Object{
						&object.String{Value: "bcd"},
					},
				},
			},
			"a",
		},
		{
			"select foo.a from foo",
			[]object.Row{
				{
					Values: []object.Object{
						&object.String{Value: "abc"},
					},
				},
				{
					Values: []object.Object{
						&object.String{Value: "bcd"},
					},
				},
			},
			"a",
		},
		{
			"select b from foo",
			[]object.Row{
				{
					Values: []object.Object{
						&object.String{Value: "efg"},
					},
				},
				{
					Values: []object.Object{
						&object.String{Value: "def"},
					},
				},
			},
			"b",
		},
		{
			"select foo.a, bar.a from foo, bar",
			[]object.Row{
				{
					Values: []object.Object{
						&object.String{Value: "abc"},
						&object.String{Value: "m"},
					},
				},
				{
					Values: []object.Object{
						&object.String{Value: "abc"},
						&object.String{Value: "n"},
					},
				},
				{
					Values: []object.Object{
						&object.String{Value: "bcd"},
						&object.String{Value: "m"},
					},
				},
				{
					Values: []object.Object{
						&object.String{Value: "bcd"},
						&object.String{Value: "n"},
					},
				},
			},
			"a, a",
		},
		{
			"select foo.a, bar.a from foo join bar on true",
			[]object.Row{
				{
					Values: []object.Object{
						&object.String{Value: "abc"},
						&object.String{Value: "m"},
					},
				},
				{
					Values: []object.Object{
						&object.String{Value: "abc"},
						&object.String{Value: "n"},
					},
				},
				{
					Values: []object.Object{
						&object.String{Value: "bcd"},
						&object.String{Value: "m"},
					},
				},
				{
					Values: []object.Object{
						&object.String{Value: "bcd"},
						&object.String{Value: "n"},
					},
				},
			},
			"a, a",
		},
		{
			"select f.a, b.a from foo f join bar b on true",
			[]object.Row{
				{
					Values: []object.Object{
						&object.String{Value: "abc"},
						&object.String{Value: "m"},
					},
				},
				{
					Values: []object.Object{
						&object.String{Value: "abc"},
						&object.String{Value: "n"},
					},
				},
				{
					Values: []object.Object{
						&object.String{Value: "bcd"},
						&object.String{Value: "m"},
					},
				},
				{
					Values: []object.Object{
						&object.String{Value: "bcd"},
						&object.String{Value: "n"},
					},
				},
			},
			"a, a",
		},
		{
			"select f.a, b.a, c from foo f join bar b on true",
			[]object.Row{
				{
					Values: []object.Object{
						&object.String{Value: "abc"},
						&object.String{Value: "m"},
						&object.String{Value: "1"},
					},
				},
				{
					Values: []object.Object{
						&object.String{Value: "abc"},
						&object.String{Value: "n"},
						&object.String{Value: "1"},
					},
				},
				{
					Values: []object.Object{
						&object.String{Value: "bcd"},
						&object.String{Value: "m"},
						&object.String{Value: "2"},
					},
				},
				{
					Values: []object.Object{
						&object.String{Value: "bcd"},
						&object.String{Value: "n"},
						&object.String{Value: "2"},
					},
				},
			},
			"a, a, c",
		},
		{
			"select f.a, b.a, c from bar b join foo f on true",
			[]object.Row{
				{
					Values: []object.Object{
						&object.String{Value: "abc"},
						&object.String{Value: "m"},
						&object.String{Value: "1"},
					},
				},
				{
					Values: []object.Object{
						&object.String{Value: "bcd"},
						&object.String{Value: "m"},
						&object.String{Value: "2"},
					},
				},
				{
					Values: []object.Object{
						&object.String{Value: "abc"},
						&object.String{Value: "n"},
						&object.String{Value: "1"},
					},
				},
				{
					Values: []object.Object{
						&object.String{Value: "bcd"},
						&object.String{Value: "n"},
						&object.String{Value: "2"},
					},
				},
			},
			"a, a, c",
		},
		{
			"select f.a, b.a, c, x from bar b join foo f on true join baz on true",
			[]object.Row{
				{
					Values: []object.Object{
						&object.String{Value: "abc"},
						&object.String{Value: "m"},
						&object.String{Value: "1"},
						&object.String{Value: "x"},
					},
				},
				{
					Values: []object.Object{
						&object.String{Value: "bcd"},
						&object.String{Value: "m"},
						&object.String{Value: "2"},
						&object.String{Value: "x"},
					},
				},
				{
					Values: []object.Object{
						&object.String{Value: "abc"},
						&object.String{Value: "n"},
						&object.String{Value: "1"},
						&object.String{Value: "x"},
					},
				},
				{
					Values: []object.Object{
						&object.String{Value: "bcd"},
						&object.String{Value: "n"},
						&object.String{Value: "2"},
						&object.String{Value: "x"},
					},
				},
			},
			"a, a, c, x",
		},
		{
			"select f.a, b.a, c from bar b, foo f",
			[]object.Row{
				{
					Values: []object.Object{
						&object.String{Value: "abc"},
						&object.String{Value: "m"},
						&object.String{Value: "1"},
					},
				},
				{
					Values: []object.Object{
						&object.String{Value: "bcd"},
						&object.String{Value: "m"},
						&object.String{Value: "2"},
					},
				},
				{
					Values: []object.Object{
						&object.String{Value: "abc"},
						&object.String{Value: "n"},
						&object.String{Value: "1"},
					},
				},
				{
					Values: []object.Object{
						&object.String{Value: "bcd"},
						&object.String{Value: "n"},
						&object.String{Value: "2"},
					},
				},
			},
			"a, a, c",
		},
		{
			"select b from foo order by b",
			[]object.Row{
				{
					Values: []object.Object{
						&object.String{Value: "def"},
					},
				},
				{
					Values: []object.Object{
						&object.String{Value: "efg"},
					},
				},
			},
			"b",
		},
		{
			"select b from foo order by b desc",
			[]object.Row{
				{
					Values: []object.Object{
						&object.String{Value: "efg"},
					},
				},
				{
					Values: []object.Object{
						&object.String{Value: "def"},
					},
				},
			},
			"b",
		},
		{
			"select b from foo order by b desc limit 100",
			[]object.Row{
				{
					Values: []object.Object{
						&object.String{Value: "efg"},
					},
				},
				{
					Values: []object.Object{
						&object.String{Value: "def"},
					},
				},
			},
			"b",
		},
		{
			"select b from foo order by b desc limit 100 offset 1",
			[]object.Row{
				{
					Values: []object.Object{
						&object.String{Value: "def"},
					},
				},
			},
			"b",
		},
		{
			"select b from foo order by b desc limit 100 offset 10",
			[]object.Row{},
			"b",
		},
		{
			"select b from foo limit 1",
			[]object.Row{
				{
					Values: []object.Object{
						&object.String{Value: "efg"},
					},
				},
			},
			"b",
		},
		{
			"select b from foo where b = 'def' limit 1",
			[]object.Row{
				{
					Values: []object.Object{
						&object.String{Value: "def"},
					},
				},
			},
			"b",
		},
		{
			"select b from foo where b = 'def' and false limit 1",
			[]object.Row{},
			"b",
		},
		{
			"select b from foo limit 0",
			[]object.Row{},
			"b",
		},
	}
	for _, tt := range tests {
		backend := newTestBackend()

		// table `foo`
		backend.tables["foo"] = []object.Column{
			{Name: "a", Type: object.STRING},
			{Name: "b", Type: object.STRING},
			{Name: "c", Type: object.STRING},
		}
		backend.rows["foo"] = []object.Row{
			{
				Values: []object.Object{
					&object.String{Value: "abc"},
					&object.String{Value: "efg"},
					&object.String{Value: "1"},
				},
				Aliases:   []string{"a", "b", "c"},
				TableName: []string{"foo", "foo", "foo"},
			},
			{
				Values: []object.Object{
					&object.String{Value: "bcd"},
					&object.String{Value: "def"},
					&object.String{Value: "2"},
				},
				Aliases:   []string{"a", "b", "c"},
				TableName: []string{"foo", "foo", "foo"},
			},
		}

		// table `bar`
		backend.tables["bar"] = []object.Column{
			{Name: "a", Type: object.STRING},
		}
		backend.rows["bar"] = []object.Row{
			{
				Values: []object.Object{
					&object.String{Value: "m"},
				},
				Aliases:   []string{"a"},
				TableName: []string{"bar"},
			},
			{
				Values: []object.Object{
					&object.String{Value: "n"},
				},
				Aliases:   []string{"a"},
				TableName: []string{"bar"},
			},
		}

		// table `baz`
		backend.tables["baz"] = []object.Column{
			{Name: "x", Type: object.STRING},
		}
		backend.rows["baz"] = []object.Row{
			{
				Values: []object.Object{
					&object.String{Value: "x"},
				},
				Aliases:   []string{"x"},
				TableName: []string{"baz"},
			},
		}

		evaluated := testEval(backend, tt.input)
		result, ok := evaluated.(*object.Result)
		if !ok {
			if errorEvaluated, errorOK := evaluated.(*object.Error); errorOK {
				t.Fatalf("%s: %s", tt.input, errorEvaluated.Inspect())
			}
			t.Fatalf("object is not Result. got=%T", evaluated)
		}
		if len(result.Rows) != len(tt.expected) {
			t.Fatalf("%s: expected result to contain %d rows. got=%d", tt.input, len(tt.expected), len(result.Rows))
		}
		if len(result.Rows) == 0 {
			continue
		}
		for i, gotRow := range result.Rows {
			if gotRow.Inspect() != tt.expected[i].Inspect() {
				t.Fatalf("%s: expected row %d to be %s. got=%s", tt.input, i, tt.expected[i].Inspect(), gotRow.Inspect())
			}
		}
		gotAliases := strings.Join(result.Aliases, ", ")
		if gotAliases != tt.expectedAliases {
			t.Fatalf("expected aliases '%s'. got='%s'", tt.expectedAliases, gotAliases)
		}
	}
}
