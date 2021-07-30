package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/vegarsti/sql/evaluator"
	"github.com/vegarsti/sql/lexer"
	"github.com/vegarsti/sql/object"
	"github.com/vegarsti/sql/parser"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	w := tabwriter.NewWriter(out, 0, 0, 1, ' ', 0)
	scanner := bufio.NewScanner(in)

	backend := newTestBackend()

	for {
		fmt.Print(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		evaluated := evaluator.Eval(backend, program)
		if evaluated != nil {
			w.Write([]byte(evaluated.Inspect()))
			w.Write([]byte("\n"))
		}
		w.Flush()
	}
}

func printParserErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		io.WriteString(out, "ERROR: "+msg+"\n")
	}
}

func main() {
	Start(os.Stdin, os.Stdout)
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

func newTestBackend() *testBackend {
	return &testBackend{
		tables: make(map[string][]object.Column),
		rows:   make(map[string][]object.Row),
	}
}
