package object

import (
	"fmt"
	"math"
	"strings"
)

type ObjectType string

const (
	ROW_OBJ     = "ROW"
	RESULT_OBJ  = "RESULT"
	INTEGER_OBJ = "INTEGER"
	BOOLEAN_OBJ = "BOOLEAN"
	FLOAT_OBJ   = "FLOAT"
	STRING_OBJ  = "STRING"
	ERROR_OBJ   = "ERROR"
	OK_OBJ      = "OK"
)

type Object interface {
	Type() ObjectType
	Inspect() string
	SortValue() float64
}

type SortBy struct {
	Value      Object
	Descending bool
}

type Row struct {
	Aliases      []string
	Values       []Object
	SortByValues []SortBy
	TableName    []string
}

func (r *Row) Inspect() string {
	values := make([]string, len(r.Values))
	for i, v := range r.Values {
		values[i] = v.Inspect()
	}
	return strings.Join(values, "\t")
}
func (r *Row) Type() ObjectType   { return ROW_OBJ }
func (r *Row) SortValue() float64 { panic("a row doesn't have a sort value") }

type Result struct {
	Aliases []string
	Rows    []*Row
}

func (r *Result) Inspect() string {
	rows := make([]string, len(r.Rows))
	for i, v := range r.Rows {
		rows[i] = v.Inspect()
	}
	header := strings.Join(r.Aliases, "\t")
	if len(rows) == 0 {
		return header
	}
	allRowsString := strings.Join(rows, "\n")
	return strings.Join([]string{
		header,
		allRowsString,
	}, "\n")
}
func (r *Result) Type() ObjectType   { return RESULT_OBJ }
func (r *Result) SortValue() float64 { panic("a result doesn't have a sort value") }

type DataType string

const (
	TEXT    = "TEXT"
	INTEGER = "INTEGER"
	FLOAT   = "FLOAT"
)

type Table struct {
	Name    string
	Columns []Column
}

type Column struct {
	Name string
	Type DataType
}

type Integer struct {
	Value int64
}

func (i *Integer) Inspect() string    { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) Type() ObjectType   { return INTEGER_OBJ }
func (i *Integer) SortValue() float64 { return float64(i.Value) }

type Boolean struct {
	Value bool
}

func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) SortValue() float64 {
	if b.Value {
		return 1
	}
	return 0
}

var True = Boolean{Value: true}
var False = Boolean{Value: false}

type Float struct {
	Value float64
}

func (f *Float) Inspect() string    { return fmt.Sprintf("%f", f.Value) }
func (f *Float) Type() ObjectType   { return FLOAT_OBJ }
func (f *Float) SortValue() float64 { return f.Value }

type String struct {
	Value string
}

func (s *String) Inspect() string  { return "'" + s.Value + "'" }
func (s *String) Type() ObjectType { return STRING_OBJ }

// SortValue of a string is the sum of all individual u * 10^-i,
// where u is the unicode value of each rune,
// and i is the index of that rune in the string.
func (s *String) SortValue() float64 {
	sortValue := float64(0)
	for i, r := range s.Value {
		unicodeValue := int(r)
		sortValue += float64(unicodeValue) * math.Pow10(-i)
	}
	return sortValue
}

type Error struct {
	Message string
}

func (e *Error) Type() ObjectType   { return ERROR_OBJ }
func (e *Error) Inspect() string    { return "ERROR: " + e.Message }
func (e *Error) SortValue() float64 { panic("an error doesn't have a sort value") }

type OK struct {
}

func (ok *OK) Type() ObjectType   { return OK_OBJ }
func (ok *OK) Inspect() string    { return "OK" }
func (ok *OK) SortValue() float64 { panic("an OK doesn't have a sort value") }
