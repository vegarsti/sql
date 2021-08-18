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
	ColumnsInTable(string) []string
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
	case *ast.BooleanLiteral:
		if node.Value {
			return &object.True
		}
		return &object.False
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
		panic(fmt.Sprintf("no such column: %s", node.Value))
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

// identifiersInExpression walks the node and returns a slice of all identifiers as strings
func identifiersInExpression(node ast.Expression) ([]ast.Identifier, error) {
	switch node := node.(type) {
	case *ast.PrefixExpression:
		right, err := identifiersInExpression(node.Right)
		if err != nil {
			return nil, err
		}
		return right, nil
	case *ast.InfixExpression:
		left, err := identifiersInExpression(node.Left)
		if err != nil {
			return nil, err
		}
		right, err := identifiersInExpression(node.Right)
		if err != nil {
			return nil, err
		}
		var identifiers []ast.Identifier
		identifiers = append(identifiers, left...)
		identifiers = append(identifiers, right...)
		return identifiers, nil
	case *ast.IntegerLiteral, *ast.BooleanLiteral, *ast.FloatLiteral, *ast.StringLiteral:
		return nil, nil
	case *ast.Identifier:
		return []ast.Identifier{*node}, nil
	}
	return nil, fmt.Errorf("unknown expression type %T", node)
}

func evalSelectStatement(backend Backend, ss *ast.SelectStatement) object.Object {
	// 1) Traverse tree looking for any column identifiers
	// 2) Get rows from backend if applicable
	// 3) For each row, evaluate expression

	columns := make(map[string]bool)
	if ss.From != "" {
		for _, c := range backend.ColumnsInTable(ss.From) {
			columns[c] = true
		}
	}

	identifiers := make(map[ast.Identifier]bool)
	for _, expr := range ss.Expressions {
		ids, err := identifiersInExpression(expr)
		if err != nil {
			return newError(err.Error())
		}
		for _, id := range ids {
			// identifier is on the form `table_name.column_name`,
			// but not selecting from `table_name`
			if id.Table != "" && id.Table != ss.From {
				return newError(`missing FROM-clause entry for table "%s"`, id.Table)
			}
			identifiers[id] = true
		}
	}

	for identifier := range identifiers {
		if !columns[identifier.Value] {
			return newError("no such column: %s", identifier.Value)
		}
	}

	// fetch rows if selecting from a table
	var rows []object.Row
	var err error
	if ss.From != "" {
		rows, err = backend.Rows(ss.From)
		if err != nil {
			return newError(err.Error())
		}
	} else {
		rows = []object.Row{{}}
	}

	// iterate over rows and evaluate expressions for each row
	rowsToReturn := make([]*object.Row, 0)
	aliases := ss.Aliases
	for _, backendRow := range rows {
		row := &object.Row{
			Aliases:      ss.Aliases,
			Values:       make([]object.Object, len(ss.Expressions)),
			SortByValues: make([]object.SortBy, len(ss.OrderBy)),
		}
		for i, e := range ss.Expressions {
			row.Values[i] = evalExpression(backendRow, e)
			if isError(row.Values[i]) {
				return row.Values[i]
			}
		}
		if ss.Where != nil {
			v := evalExpression(backendRow, ss.Where)
			if isError(v) {
				return v
			}
			include, ok := v.(*object.Boolean)
			if !ok {
				return newError("argument of WHERE must be type boolean, not type integer: %s", v.Inspect())
			}
			if !include.Value {
				continue
			}
		}
		for i, e := range ss.OrderBy {
			v := evalExpression(backendRow, e.Expression)
			if isError(v) {
				return v
			}
			row.SortByValues[i].Value = v
			row.SortByValues[i].Descending = e.Descending
		}
		rowsToReturn = append(rowsToReturn, row)
	}
	// Populate aliases
	for i, alias := range ss.Aliases {
		if alias == "" {
			aliases[i] = ss.Expressions[i].String()
		}
	}
	// Sort
	if len(ss.OrderBy) != 0 {
		sortRows(rowsToReturn)
	}
	// Limit
	if ss.Limit != nil {
		offset := 0
		if ss.Offset != nil {
			offset = *ss.Offset
		}
		end := len(rowsToReturn)
		if len(rowsToReturn) > offset+*ss.Limit {
			end = offset + *ss.Limit
		}
		if end < offset {
			offset = 0
			end = 0
		}
		rowsToReturn = rowsToReturn[offset:end]
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
			sign := float64(1)
			if rows[j].SortByValues[k].Descending {
				sign = -1
			}
			iValue := sign * rows[i].SortByValues[k].Value.SortValue()
			jValue := sign * rows[j].SortByValues[k].Value.SortValue()
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
	case "!":
		return evalBangPrefixOperatorExpression(right)
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

func evalBangPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() == object.BOOLEAN_OBJ {
		value := right.(*object.Boolean).Value
		return &object.Boolean{Value: !value}
	}
	return newError("unknown operator: !%s", right.Type())
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
	// bool
	case left.Type() == object.BOOLEAN_OBJ && right.Type() == object.BOOLEAN_OBJ:
		return evalBooleanInfixExpression(operator, left, right)
	// string
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
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
	case "=":
		return &object.Boolean{Value: leftVal == rightVal}
	case "!=":
		return &object.Boolean{Value: leftVal != rightVal}
	case ">":
		return &object.Boolean{Value: leftVal > rightVal}
	case ">=":
		return &object.Boolean{Value: leftVal >= rightVal}
	case "<":
		return &object.Boolean{Value: leftVal < rightVal}
	case "<=":
		return &object.Boolean{Value: leftVal <= rightVal}
	default:
		return newError("unknown integer operator: %s %s %s", left.Type(), operator, right.Type())
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
	case "=":
		return &object.Boolean{Value: leftVal == rightVal}
	case "!=":
		return &object.Boolean{Value: leftVal != rightVal}
	case ">":
		return &object.Boolean{Value: leftVal > rightVal}
	case ">=":
		return &object.Boolean{Value: leftVal >= rightVal}
	case "<":
		return &object.Boolean{Value: leftVal < rightVal}
	case "<=":
		return &object.Boolean{Value: leftVal <= rightVal}
	default:
		return newError("unknown float operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalBooleanInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	leftVal := left.(*object.Boolean).Value
	rightVal := right.(*object.Boolean).Value

	switch operator {
	case "=":
		return &object.Boolean{Value: leftVal == rightVal}
	case "!=":
		return &object.Boolean{Value: leftVal != rightVal}
	case "AND":
		return &object.Boolean{Value: leftVal && rightVal}
	case "OR":
		return &object.Boolean{Value: leftVal || rightVal}
	default:
		return newError("unknown boolean operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	switch operator {
	case "=":
		return &object.Boolean{Value: leftVal == rightVal}
	case "!=":
		return &object.Boolean{Value: leftVal != rightVal}
	case ">":
		return &object.Boolean{Value: leftVal > rightVal}
	case ">=":
		return &object.Boolean{Value: leftVal >= rightVal}
	case "<":
		return &object.Boolean{Value: leftVal < rightVal}
	case "<=":
		return &object.Boolean{Value: leftVal <= rightVal}
	default:
		return newError("unknown string operator: %s %s %s", left.Type(), operator, right.Type())
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
