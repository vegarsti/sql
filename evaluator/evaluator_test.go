package evaluator_test

import (
	"fmt"
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
		{"select !true", false},
		{"select !false", true},
		{"select false and true", false},
		{"select false or true", true},
		{"select false or !true", false},
		{"select 1 < 1", false},
		{"select 1 <= 1", true},
		{"select 1 >= 1", true},
		{"select 1 > 1", false},
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
		testBooleanObject(t, row.Values[0], tt.expected)
	}
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
		{"select 1 / 2", 0.5},
		{"select 1 / 1", 1},
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

func TestEvalIdentifierExpression(t *testing.T) {
	tests := []struct {
		input                string
		expectedErrorMessage string
	}{
		{"select foo", "no such column: foo"},
	}
	for _, tt := range tests {
		evaluated := testEval(newTestBackend(), tt.input)
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
	tables  map[string][]object.Column
	rows    map[string][]object.Row
	columns map[string][]string
}

func (tb *testBackend) CreateTable(name string, columns []object.Column) error {
	if _, ok := tb.tables[name]; ok {
		return fmt.Errorf(`relation "%s" already exists`, name)
	}
	tb.tables[name] = columns
	tb.rows[name] = make([]object.Row, 0)
	tb.columns[name] = make([]string, len(columns))
	for i, c := range columns {
		tb.columns[name][i] = c.Name
	}
	return nil
}

func (tb *testBackend) InsertInto(name string, row object.Row) error {
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

func (tb *testBackend) ColumnsInTable(name string) []string {
	return tb.columns[name]
}

func newTestBackend() *testBackend {
	return &testBackend{
		tables:  make(map[string][]object.Column),
		rows:    make(map[string][]object.Row),
		columns: make(map[string][]string),
	}
}

func TestEvalCreateTable(t *testing.T) {
	tests := []struct {
		input         string
		tableName     string
		expectedTable object.Table
	}{
		{
			"create table foo (a text, b integer, c double)",
			"foo",
			object.Table{
				Name: "foo",
				Columns: []object.Column{
					{Name: "a", Type: object.TEXT},
					{Name: "b", Type: object.INTEGER},
					{Name: "c", Type: object.DOUBLE},
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
		input          string
		expectedValues []object.Object
	}{
		{
			"insert into foo values ('abc', 1, 3.14)",
			[]object.Object{
				&object.String{Value: "abc"},
				&object.Integer{Value: 1},
				&object.Float{Value: 3.14},
			},
		},
	}
	for _, tt := range tests {
		backend := newTestBackend()
		backend.tables["foo"] = []object.Column{
			{Name: "a", Type: object.DataType("TEXT")},
			{Name: "b", Type: object.DataType("INTEGER")},
			{Name: "c", Type: object.DataType("DOUBLE")},
		}
		evaluated := testEval(backend, tt.input)
		if _, ok := evaluated.(*object.OK); !ok {
			if errorEvaluated, errorOK := evaluated.(*object.Error); errorOK {
				t.Fatalf("object is Error: %s", errorEvaluated.Inspect())
			}
			t.Fatalf("object is not OK. got=%T", evaluated)
		}
		rows, ok := backend.rows["foo"]
		if !ok {
			t.Fatalf("row doesn't exist")
		}
		if len(rows) != 1 {
			t.Fatalf("expected table to have %d rows. got=%d", 1, len(rows))
		}
		for i := range tt.expectedValues {
			if rows[0].Values[i].Type() != tt.expectedValues[i].Type() {
				t.Fatalf("expected row[%d] to have %v value. got=%v", i, tt.expectedValues[i].Type(), rows[0].Values[i].Type())
			}
			if rows[0].Values[i].Inspect() != tt.expectedValues[i].Inspect() {
				t.Fatalf("expected row[%d] to have %v value. got=%v", i, tt.expectedValues[i].Inspect(), rows[0].Values[i].Inspect())
			}
			expectedAliases := "a, b, c"
			gotAliases := strings.Join(rows[0].Aliases, ", ")
			if gotAliases != expectedAliases {
				t.Fatalf("expected aliases to be %s. got=%s", expectedAliases, gotAliases)
			}
		}
	}
}

func TestEvalSelectFrom(t *testing.T) {
	tests := []struct {
		input           string
		expected        [][]string
		expectedAliases string
	}{
		{
			"select a, b from foo",
			[][]string{
				{"abc", "efg"},
				{"bcd", "def"},
			},
			"a, b",
		},
		{
			"select a from foo",
			[][]string{
				{"abc"},
				{"bcd"},
			},
			"a",
		},
		{
			"select foo.a from foo",
			[][]string{
				{"abc"},
				{"bcd"},
			},
			"foo.a",
		},
		{
			"select b from foo",
			[][]string{
				{"efg"},
				{"def"},
			},
			"b",
		},
		{
			"select b from foo order by b",
			[][]string{
				{"def"},
				{"efg"},
			},
			"b",
		},
		{
			"select b from foo order by b desc",
			[][]string{
				{"efg"},
				{"def"},
			},
			"b",
		},
		{
			"select b from foo order by b desc limit 100",
			[][]string{
				{"efg"},
				{"def"},
			},
			"b",
		},
		{
			"select b from foo order by b desc limit 100 offset 1",
			[][]string{
				{"def"},
			},
			"b",
		},
		{
			"select b from foo order by b desc limit 100 offset 10",
			[][]string{},
			"b",
		},
		{
			"select b from foo limit 1",
			[][]string{
				{"efg"},
			},
			"b",
		},
		{
			"select b from foo where b = 'def' limit 1",
			[][]string{
				{"def"},
			},
			"b",
		},
		{
			"select b from foo where b = 'def' and false limit 1",
			[][]string{},
			"b",
		},
		{
			"select b from foo limit 0",
			[][]string{},
			"b",
		},
	}
	for _, tt := range tests {
		backend := newTestBackend()
		backend.tables["foo"] = []object.Column{
			{Name: "a", Type: object.DataType("TEXT")},
			{Name: "b", Type: object.DataType("TEXT")},
		}
		backend.rows["foo"] = []object.Row{
			{
				Values: []object.Object{
					&object.String{Value: "abc"},
					&object.String{Value: "efg"},
					&object.Integer{Value: 1},
				},
				Aliases: []string{"a", "b"},
			},
			{
				Values: []object.Object{
					&object.String{Value: "bcd"},
					&object.String{Value: "def"},
					&object.Integer{Value: 2},
				},
				Aliases: []string{"a", "b"},
			},
		}
		backend.columns["foo"] = []string{"a", "b"}
		evaluated := testEval(backend, tt.input)
		result, ok := evaluated.(*object.Result)
		if !ok {
			if errorEvaluated, errorOK := evaluated.(*object.Error); errorOK {
				t.Fatalf("object is Error: %s", errorEvaluated.Inspect())
			}
			t.Fatalf("object is not Result. got=%T", evaluated)
		}
		if len(result.Rows) != len(tt.expected) {
			t.Fatalf("expected result to contain %d rows. got=%d", len(tt.expected), len(result.Rows))
		}
		if len(result.Rows) == 0 {
			continue
		}
		row1 := result.Rows[0]
		if len(row1.Values) != len(tt.expected[0]) {
			t.Fatalf("expected row to contain %d element. got=%d", len(tt.expected[0]), len(row1.Values))
		}
		for i := range row1.Values {
			// log.Println(row1.Values[i])
			testStringObject(t, row1.Values[i], tt.expected[0][i])
		}
		if len(result.Rows) == 1 {
			continue
		}
		row2 := result.Rows[1]
		if len(row2.Values) != len(tt.expected[1]) {
			t.Fatalf("expected row to contain %d element. got=%d", len(tt.expected[1]), len(row1.Values))
		}
		for i := range row2.Values {
			testStringObject(t, row2.Values[i], tt.expected[1][i])
		}

		gotAliases := strings.Join(result.Aliases, ", ")
		if gotAliases != tt.expectedAliases {
			t.Fatalf("expected aliases %s. got=%s", tt.expectedAliases, gotAliases)
		}
	}
}
