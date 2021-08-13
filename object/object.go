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

type Row struct {
	Aliases      []string
	Values       []Object
	SortByValues []Object
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
	allRowsString := strings.Join(rows, "\n")
	return strings.Join([]string{
		strings.Join(r.Aliases, "\t"),
		allRowsString,
	}, "\n")
}
func (r *Result) Type() ObjectType   { return RESULT_OBJ }
func (r *Result) SortValue() float64 { panic("a result doesn't have a sort value") }

type DataType string

const (
	TEXT    = "TEXT"
	INTEGER = "INTEGER"
	DOUBLE  = "DOUBLE"
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

// SortValue of a string is u * 10^-i,
// where u is the unicode value of the rune,
// and i is the index of the rune in the string
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
