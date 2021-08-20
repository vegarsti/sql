package ast

import (
	"bytes"
	"strings"

	"github.com/vegarsti/sql/token"
)

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type OrderByExpression struct {
	Expression Expression
	Descending bool
}

type Join struct {
	Table     string
	Predicate Expression
}

type From struct {
	Table string
	Join  *Join
}

type SelectStatement struct {
	Token       token.Token // the SELECT token
	Expressions []Expression
	Aliases     []string // SELECT value AS some_alias
	From        []*From
	OrderBy     []*OrderByExpression // can be any expression that would be valid in the query's select list
	Limit       *int
	Offset      *int
	Where       Expression
}

func (es *SelectStatement) statementNode()       {}
func (es *SelectStatement) TokenLiteral() string { return es.Token.Literal }
func (es *SelectStatement) String() string {
	if len(es.Expressions) == 0 {
		return ""
	}
	expressions := make([]string, len(es.Expressions))
	for i, expr := range es.Expressions {
		expressions[i] = expr.String()
	}
	return es.TokenLiteral() + " " + strings.Join(expressions, ", ")
}

type CreateTableStatement struct {
	Token   token.Token // the CREATE token
	Name    string
	Columns map[string]token.Token
}

func (cts *CreateTableStatement) statementNode()       {}
func (cts *CreateTableStatement) TokenLiteral() string { return cts.Token.Literal }
func (cts *CreateTableStatement) String() string {
	if len(cts.Columns) == 0 {
		return ""
	}
	columns := make([]string, len(cts.Columns))
	i := 0
	for name, tok := range cts.Columns {
		columns[i] = name + " " + tok.Literal
		i++
	}
	return "CREATE TABLE " + cts.Name + " " + "(" + strings.Join(columns, ", ") + ")"
}

type InsertStatement struct {
	Token       token.Token // the INSERT token
	TableName   string
	Expressions []Expression
}

func (is *InsertStatement) statementNode()       {}
func (is *InsertStatement) TokenLiteral() string { return is.Token.Literal }
func (is *InsertStatement) String() string {
	expressions := make([]string, len(is.Expressions))
	for i, e := range is.Expressions {
		expressions[i] = e.String()
	}
	return "INSERT INTO " + is.TableName + " VALUES " + "(" + strings.Join(expressions, ", ") + ")"
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }
func (bl *BooleanLiteral) String() string       { return bl.Token.Literal }

type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return "'" + sl.Token.Literal + "'" }

type Identifier struct {
	Token token.Token
	Value string
	Table string
}

func (il *Identifier) expressionNode()      {}
func (il *Identifier) TokenLiteral() string { return il.Token.Literal }
func (il *Identifier) String() string       { return il.Value }

type PrefixExpression struct {
	Token    token.Token // The prefix token, e.g. -
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

type InfixExpression struct {
	Token    token.Token // The infix token, e.g. !
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	return out.String()
}
