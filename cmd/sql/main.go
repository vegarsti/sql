package main

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/chzyer/readline"
	"github.com/vegarsti/sql/evaluator"
	"github.com/vegarsti/sql/lexer"
	"github.com/vegarsti/sql/object"
	"github.com/vegarsti/sql/parser"
)

const PROMPT = ">> "

func Start(r *readline.Instance) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	backend := newTestBackend()

	for {
		line, err := r.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}
		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(os.Stdout, p.Errors())
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
	r, err := readline.New(">> ")
	if err != nil {
		panic(err)
	}
	defer r.Close()
	Start(r)
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

func (tb *testBackend) Columns(name string) []string {
	return tb.columns[name]
}

func newTestBackend() *testBackend {
	return &testBackend{
		tables:  make(map[string][]object.Column),
		rows:    make(map[string][]object.Row),
		columns: make(map[string][]string),
	}
}
