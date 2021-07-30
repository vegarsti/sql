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

	backend := NewBackend()

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

type Backend struct {
}

func (tb *Backend) CreateTable(name string, columns []object.Column) error {
	return fmt.Errorf("no support for tables yet")
}

func (tb *Backend) InsertInto(name string, row object.Row) error {
	return fmt.Errorf("no support for tables yet")
}

func (tb *Backend) Rows(name string) ([]object.Row, error) {
	return nil, fmt.Errorf("no support for tables yet")
}

func NewBackend() *Backend {
	return &Backend{}
}
