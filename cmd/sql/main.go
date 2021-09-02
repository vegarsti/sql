package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/chzyer/readline"
	"github.com/vegarsti/sql/evaluator"
	"github.com/vegarsti/sql/lexer"
	"github.com/vegarsti/sql/object"
	"github.com/vegarsti/sql/parser"
)

func RunInteractive(backend evaluator.Backend) {
	r, err := readline.New(">> ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "readline.New(): %v", err)
		os.Exit(1)
	}
	defer r.Close()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)

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

func RunScript(backend evaluator.Backend, input string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		printParserErrors(os.Stdout, p.Errors())
		return
	}
	evaluated := evaluator.Eval(backend, program)
	if evaluated != nil {
		w.Write([]byte(evaluated.Inspect()))
		w.Write([]byte("\n"))
	}
	w.Flush()
}

func printParserErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		io.WriteString(out, "ERROR: "+msg+"\n")
	}
}

func main() {
	fi, err := os.Stdin.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "os.Stdin.Stat(): %v", err)
		os.Exit(1)
	}
	backend := newTestBackend()
	receivedInputFromStdin := (fi.Mode() & os.ModeCharDevice) == 0
	if receivedInputFromStdin {
		s := bufio.NewScanner(os.Stdin)
		var lines []string
		for s.Scan() {
			line := s.Text()
			lines = append(lines, line)
		}
		input := strings.Join(lines, " ")
		RunScript(backend, input)
		return
	}
	RunInteractive(backend)
}

type testBackend struct {
	tables      map[string][]object.Column
	rows        map[string][]object.Row
	columns     map[string][]string
	columnTypes map[string][]object.DataType
}

func (tb *testBackend) CreateTable(name string, columns []object.Column) error {
	if _, ok := tb.tables[name]; ok {
		return fmt.Errorf(`relation "%s" already exists`, name)
	}
	tb.tables[name] = columns
	tb.rows[name] = make([]object.Row, 0)
	tb.columns[name] = make([]string, len(columns))
	tb.columnTypes[name] = make([]object.DataType, len(columns))
	for i, c := range columns {
		tb.columns[name][i] = c.Name
		tb.columnTypes[name][i] = c.Type
	}
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

func (tb *testBackend) Columns(name string) []string {
	return tb.columns[name]
}

func (tb *testBackend) ColumnTypes(name string) []object.DataType {
	return tb.columnTypes[name]
}

func newTestBackend() *testBackend {
	return &testBackend{
		tables:      make(map[string][]object.Column),
		rows:        make(map[string][]object.Row),
		columns:     make(map[string][]string),
		columnTypes: make(map[string][]object.DataType),
	}
}
