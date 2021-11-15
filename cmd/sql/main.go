package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/chzyer/readline"
	"github.com/vegarsti/sql/bolt"
	"github.com/vegarsti/sql/evaluator"
	"github.com/vegarsti/sql/inmemory"
	"github.com/vegarsti/sql/lexer"
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
	if len(os.Args) > 2 {
		fmt.Fprint(os.Stderr, "usage: sql [database file]")
		os.Exit(1)
	}
	var backend evaluator.Backend
	if len(os.Args) == 1 {
		backend = inmemory.NewBackend()
	}
	if len(os.Args) == 2 {
		backend = bolt.NewBackend(os.Args[1])
	}
	if err := backend.Open(); err != nil {
		fmt.Fprintf(os.Stderr, "backend open: %v", err)
		os.Exit(1)
	}
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
	} else {
		RunInteractive(backend)
	}
	if err := backend.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "backend close: %v", err)
		os.Exit(1)
	}
}
