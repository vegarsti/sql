package object

import (
	"fmt"
)

type ObjectType string

const (
	INTEGER_OBJ    = "INTEGER"
	FLOAT_OBJ      = "FLOAT"
	STRING_OBJ     = "STRING"
	IDENTIFIER_OBJ = "IDENTIFIER"
	ERROR_OBJ      = "ERROR"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

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

type Identifier struct {
	Value string
}

func (i *Identifier) Inspect() string  { return i.Value }
func (i *Identifier) Type() ObjectType { return IDENTIFIER_OBJ }

type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }
