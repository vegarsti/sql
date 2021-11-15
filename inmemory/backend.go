package inmemory

import (
	"fmt"

	"github.com/vegarsti/sql/object"
)

type Backend struct {
	Tables map[string][]object.Column
	Tuples map[string][]object.Row
}

func (b *Backend) Open() error {
	return nil
}

func (b *Backend) Close() error {
	return nil
}

func (b *Backend) CreateTable(name string, columns []object.Column) error {
	if _, ok := b.Tables[name]; ok {
		return fmt.Errorf(`relation "%s" already exists`, name)
	}
	b.Tables[name] = columns
	b.Tuples[name] = make([]object.Row, 0)
	return nil
}

func (b *Backend) Insert(name string, row object.Row) error {
	if _, ok := b.Tables[name]; !ok {
		return fmt.Errorf(`relation "%s" does not exist`, name)
	}
	b.Tuples[name] = append(b.Tuples[name], row)
	// Populate aliases
	for i := range b.Tuples[name] {
		b.Tuples[name][i].Aliases = make([]string, len(b.Tables[name]))
		for j, column := range b.Tables[name] {
			b.Tuples[name][i].Aliases[j] = column.Name
		}
	}
	return nil
}

func (b *Backend) Rows(name string) ([]object.Row, error) {
	rows, ok := b.Tuples[name]
	if !ok {
		return nil, fmt.Errorf(`relation "%s" does not exist`, name)
	}
	return rows, nil
}

func (b *Backend) Columns(name string) ([]object.Column, error) {
	return b.Tables[name], nil
}

func NewBackend() *Backend {
	return &Backend{
		Tables: make(map[string][]object.Column),
		Tuples: make(map[string][]object.Row),
	}
}
