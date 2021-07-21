package object

import (
	"fmt"
	"strings"
)

type ObjectType string

const (
	ROW_OBJ     = "ROW"
	INTEGER_OBJ = "INTEGER"
	FLOAT_OBJ   = "FLOAT"
	STRING_OBJ  = "STRING"
	ERROR_OBJ   = "ERROR"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Row struct {
	Values []Object
}

func (r *Row) Inspect() string {
	values := make([]string, len(r.Values))
	for i, v := range r.Values {
		values[i] = v.Inspect()
	}
	return strings.Join(values, ", ")
}
func (r *Row) Type() ObjectType { return ROW_OBJ }

type Integer struct {
	Value int64
}

func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) Type() ObjectType { return INTEGER_OBJ }

type Float struct {
	Value float64
}

func (f *Float) Inspect() string  { return fmt.Sprintf("%f", f.Value) }
func (f *Float) Type() ObjectType { return FLOAT_OBJ }

type String struct {
	Value string
}

func (s *String) Inspect() string  { return s.Value }
func (s *String) Type() ObjectType { return STRING_OBJ }

type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }
