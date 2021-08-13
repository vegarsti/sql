package evaluator

import (
	"fmt"
	"sort"

	"github.com/vegarsti/sql/ast"
	"github.com/vegarsti/sql/object"
)

type Backend interface {
	CreateTable(string, []object.Column) error
	InsertInto(string, object.Row) error
	Rows(string) ([]object.Row, error)
}

func Eval(backend Backend, node ast.Node) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalStatements(backend, node.Statements)
	case *ast.SelectStatement:
		return evalSelectStatement(backend, node)
	case *ast.CreateTableStatement:
		return evalCreateTableStatement(backend, node)
	case *ast.InsertStatement:
		return evalInsertStatement(backend, node)
	default:
		if expression, ok := node.(ast.Expression); ok {
			return evalExpression(object.Row{}, expression)
		}
		return newError("unknown node type %T", node)
	}
}

func evalExpression(row object.Row, node ast.Expression) object.Object {
	switch node := node.(type) {
	case *ast.PrefixExpression:
		right := evalExpression(row, node.Right)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := evalExpression(row, node.Left)
		if isError(left) {
			return left
		}
		right := evalExpression(row, node.Right)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value}
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.Identifier:
		if row.Values != nil && row.Aliases != nil {
			for i := range row.Values {
				if row.Aliases[i] == node.Value {
					return row.Values[i]
				}
			}
		}
		return newError("no such column: %s", node.Value)
	default:
		return newError("unknown expression type %T", node)
	}
}

func evalStatements(backend Backend, stmts []ast.Statement) object.Object {
	var result object.Object
	for _, statement := range stmts {
		result = Eval(backend, statement)
		if isError(result) {
			return result
		}
	}
	return result
}

func evalSelectStatement(backend Backend, ss *ast.SelectStatement) object.Object {
	// fetch rows if selecting from a table
	var rows []object.Row
	var err error
	if ss.From != "" {
		rows, err = backend.Rows(ss.From)
		if err != nil {
			return newError(err.Error())
		}
	}

	// iterate over rows from backend (or a dummy row if not selecting from a table)
	// and evaluate the expression for each row
	if len(rows) == 0 {
		rows = []object.Row{{}}
	}
	rowsToReturn := make([]*object.Row, 0)
	aliases := ss.Aliases
	for _, backendRow := range rows {
		row := &object.Row{
			Aliases:      ss.Aliases,
			Values:       make([]object.Object, len(ss.Expressions)),
			SortByValues: make([]object.Object, len(ss.OrderBy)),
		}
		for i, e := range ss.Expressions {
			row.Values[i] = evalExpression(backendRow, e)
			if isError(row.Values[i]) {
				return row.Values[i]
			}
		}
		for i, e := range ss.OrderBy {
			row.SortByValues[i] = evalExpression(backendRow, e.Expression)
			if isError(row.SortByValues[i]) {
				return row.SortByValues[i]
			}
		}
		rowsToReturn = append(rowsToReturn, row)
	}
	// Populate aliases
	for i, alias := range ss.Aliases {
		if alias == "" {
			aliases[i] = ss.Expressions[i].String()
		}
	}
	if len(ss.OrderBy) != 0 {
		sortRows(rowsToReturn)
	}
	result := &object.Result{
		Aliases: aliases,
		Rows:    rowsToReturn,
	}
	return result
}

func sortRows(rows []*object.Row) {
	n := len(rows[0].SortByValues) // must be same length for all rows
	sort.Slice(rows, func(i, j int) bool {
		for k := 0; k < n; k++ {
			iValue := rows[i].SortByValues[k].SortValue()
			jValue := rows[j].SortByValues[k].SortValue()
			if iValue != jValue {
				return iValue < jValue
			}
		}
		return false
	})
}

func evalCreateTableStatement(backend Backend, cst *ast.CreateTableStatement) object.Object {
	columns := make([]object.Column, len(cst.Columns))
	i := 0
	for name, typeToken := range cst.Columns {
		columns[i] = object.Column{
			Name: name,
			Type: object.DataType(typeToken.Literal), // This should always be valid since it has parsed successfully
		}
		i++
	}
	if err := backend.CreateTable(cst.Name, columns); err != nil {
		return newError(err.Error())
	}
	return &object.OK{}
}

func evalInsertStatement(backend Backend, is *ast.InsertStatement) object.Object {
	row := object.Row{
		Values: make([]object.Object, len(is.Expressions)),
	}
	for i, es := range is.Expressions {
		obj := Eval(backend, es)
		row.Values[i] = obj
	}
	if err := backend.InsertInto(is.TableName, row); err != nil {
		return newError(err.Error())
	}
	return &object.OK{}
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() == object.INTEGER_OBJ {
		value := right.(*object.Integer).Value
		return &object.Integer{Value: -value}
	}
	if right.Type() == object.FLOAT_OBJ {
		value := right.(*object.Float).Value
		return &object.Float{Value: -value}
	}
	return newError("unknown operator: -%s", right.Type())
}

func evalInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	switch {
	// int, int
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	// float, float
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.FLOAT_OBJ:
		return evalFloatInfixExpression(operator, left, right)
	// float, int
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.INTEGER_OBJ:
		rightInteger := right.(*object.Integer)
		right = &object.Float{Value: float64(rightInteger.Value)}
		return evalFloatInfixExpression(operator, left, right)
	// int, float
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.FLOAT_OBJ:
		leftInteger := left.(*object.Integer)
		left = &object.Float{Value: float64(leftInteger.Value)}
		return evalFloatInfixExpression(operator, left, right)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIntegerInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Float{Value: float64(leftVal) / float64(rightVal)}
	default:
		panic("unreachable; should have been caught by check in outer function")
	}
}

func evalFloatInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	leftVal := left.(*object.Float).Value
	rightVal := right.(*object.Float).Value

	switch operator {
	case "+":
		return &object.Float{Value: leftVal + rightVal}
	case "-":
		return &object.Float{Value: leftVal - rightVal}
	case "*":
		return &object.Float{Value: leftVal * rightVal}
	case "/":
		return &object.Float{Value: leftVal / rightVal}
	default:
		panic("unreachable; should have been caught by check in outer function")
	}
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}
