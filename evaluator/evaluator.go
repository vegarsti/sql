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
	case *ast.PostfixExpression:
		left := evalExpression(row, node.Left)
		if isError(left) {
			return left
		}
		return evalPostfixExpression(left, node.Operator)
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
	case *ast.Null:
		return object.NULL
	case *ast.Identifier:
		if row.Values != nil && row.Aliases != nil {
			for i := range row.Values {
				if row.Aliases[i] == node.Value && row.TableName[i] == node.Table {
					return row.Values[i]
				}
			}
		}
		if node.Table != "" {
			return newError("column %s.%s does not exist", node.Table, node.Value)
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

// concatenateRows by simply splicing the tuples together.
// Used when joining tables
func concatenateRows(row1 object.Row, row2 object.Row) object.Row {
	aliases := append(row1.Aliases, row2.Aliases...)
	values := append(row1.Values, row2.Values...)
	tableNames := append(row1.TableName, row2.TableName...)
	return object.Row{
		Aliases:   aliases,
		Values:    values,
		TableName: tableNames,
	}
}

// identifiersInExpression walks the node and returns a slice of all identifiers as strings
func identifiersInExpression(node ast.Expression) ([]*ast.Identifier, error) {
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
		var identifiers []*ast.Identifier
		identifiers = append(identifiers, left...)
		identifiers = append(identifiers, right...)
		return identifiers, nil
	case *ast.IntegerLiteral, *ast.BooleanLiteral, *ast.FloatLiteral, *ast.StringLiteral, *ast.PostfixExpression, *ast.Null:
		return nil, nil
	case *ast.Identifier:
		return []*ast.Identifier{node}, nil
	}
	return nil, fmt.Errorf("unknown expression type %T", node)
}

func evalSelectStatement(backend Backend, stmt *ast.SelectStatement) object.Object {
	// 1) Traverse tree looking for any column identifiers
	// 2) Get rows from backend if applicable
	// 3) For each row, evaluate expression

	columns := make(map[string]map[string]bool)
	tableToAlias := make(map[string]string)
	tableReferences := make(map[string]int)
	for _, from := range stmt.From {
		tableReferences[from.Table]++
		for _, c := range backend.ColumnsInTable(from.Table) {
			if _, ok := columns[c]; !ok {
				columns[c] = make(map[string]bool)
			}
			table := from.Table
			if from.TableAlias != "" {
				table = from.TableAlias
				tableToAlias[from.Table] = from.TableAlias
			}
			columns[c][table] = true
		}
		if from.Join != nil {
			tableReferences[from.Join.With.Table]++
			for _, c := range backend.ColumnsInTable(from.Join.With.Table) {
				if _, ok := columns[c]; !ok {
					columns[c] = make(map[string]bool)
				}
				table := from.Table
				if from.Join.With.TableAlias != "" {
					table = from.Join.With.TableAlias
					tableToAlias[from.TableAlias] = from.Table
				}
				columns[c][table] = true
			}
		}
	}

	for table, count := range tableReferences {
		if count > 1 {
			return newError(`table name "%s" specified more than once`, table)
		}
	}

	// gather pointers to all identifiers used in statement
	var allIdentifiers []*ast.Identifier
	for _, expr := range stmt.Expressions {
		ids, err := identifiersInExpression(expr)
		if err != nil {
			return newError(err.Error())
		}
		allIdentifiers = append(allIdentifiers, ids...)
	}
	for _, from := range stmt.From {
		if from.Join != nil {
			ids, err := identifiersInExpression(from.Join.Predicate)
			if err != nil {
				return newError(err.Error())
			}
			allIdentifiers = append(allIdentifiers, ids...)
		}
	}
	if stmt.Where != nil {
		ids, err := identifiersInExpression(stmt.Where)
		if err != nil {
			return newError(err.Error())
		}
		allIdentifiers = append(allIdentifiers, ids...)
	}
	for _, orderBy := range stmt.OrderBy {
		ids, err := identifiersInExpression(orderBy.Expression)
		if err != nil {
			return newError(err.Error())
		}
		allIdentifiers = append(allIdentifiers, ids...)
	}

	// check for missing FROM clause
	for _, id := range allIdentifiers {
		// identifier is on the form `table_name.column_name`,
		// but not selecting from `table_name`
		if id.Table != "" {
			missingFrom := true
			for _, from := range stmt.From {
				// alias not provided
				if from.TableAlias == "" && id.Table == from.Table {
					missingFrom = false
				}

				// alias is provided
				if from.TableAlias != "" && id.Table == from.TableAlias {
					missingFrom = false
					id.Table = from.Table
				}
				if from.Join != nil {
					// alias not provided
					if from.Join.With.TableAlias == "" && id.Table == from.Join.With.Table {
						missingFrom = false
					}

					// alias is provided
					if from.Join.With.TableAlias != "" && id.Table == from.Join.With.TableAlias {
						missingFrom = false
						id.Table = from.Join.With.Table
					}
				}
			}
			if alias, ok := tableToAlias[id.Table]; missingFrom && ok {
				return newError(`invalid reference to FROM-clause entry for table "%s". Perhaps you meant to reference the table alias "%s"`, id.Table, alias)
			}
			if missingFrom {
				return newError(`missing FROM-clause entry for table "%s"`, id.Table)
			}
		}
	}

	// populate column identifiers with which table the column belongs to, and fail if ambiguous or non-existent
	for _, identifier := range allIdentifiers {
		tables, ok := columns[identifier.Value]
		if !ok {
			return newError(`column "%s" does not exist`, identifier.Value)
		}
		if identifier.Table != "" {
			continue
		}
		if len(tables) > 1 {
			return newError(`column reference "%s" is ambiguous`, identifier.Value)
		}
		var table string
		for t := range tables {
			table = t
		}
		identifier.Table = table
	}

	// fetch rows
	rows := []object.Row{{}}
	for _, from := range stmt.From {
		r, err := backend.Rows(from.Table)
		if err != nil {
			return newError(err.Error())
		}
		var newRows []object.Row
		for _, row1 := range rows {
			for _, row2 := range r {
				newRow := concatenateRows(row1, row2)
				newRows = append(newRows, newRow)
			}
		}
		rows = newRows
		if from.Join != nil {
			newRows = []object.Row{}
			r, err = backend.Rows(from.Join.With.Table)
			if err != nil {
				return newError(err.Error())
			}
			for _, row1 := range rows {
				for _, row2 := range r {
					newRow := concatenateRows(row1, row2)
					v := evalExpression(newRow, from.Join.Predicate)
					if isError(v) {
						return v
					}
					include, ok := v.(*object.Boolean)
					if !ok {
						return newError("join condition must be of type boolean, not %s: %s", v.Type(), v.Inspect())
					}
					if !include.Value {
						continue
					}
					newRows = append(newRows, newRow)
				}
			}
			rows = newRows
		}
	}

	// iterate over rows and evaluate expressions for each row
	rowsToReturn := make([]*object.Row, 0)
	aliases := stmt.Aliases
	for _, backendRow := range rows {
		row := &object.Row{
			Aliases:      stmt.Aliases,
			Values:       make([]object.Object, len(stmt.Expressions)),
			SortByValues: make([]object.SortBy, len(stmt.OrderBy)),
		}
		for i, e := range stmt.Expressions {
			row.Values[i] = evalExpression(backendRow, e)
			if isError(row.Values[i]) {
				return row.Values[i]
			}
		}
		if stmt.Where != nil {
			v := evalExpression(backendRow, stmt.Where)
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
		for i, e := range stmt.OrderBy {
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
	for i, alias := range stmt.Aliases {
		if alias == "" {
			aliases[i] = stmt.Expressions[i].String()
		}
	}
	// Sort
	if len(stmt.OrderBy) != 0 {
		sortRows(rowsToReturn)
	}
	// Limit
	if stmt.Limit != nil {
		offset := 0
		if stmt.Offset != nil {
			offset = *stmt.Offset
		}
		end := len(rowsToReturn)
		if len(rowsToReturn) > offset+*stmt.Limit {
			end = offset + *stmt.Limit
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
	if len(rows) == 0 {
		return
	}
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
	columns := make([]object.Column, len(cst.ColumnNames))
	for i := range cst.ColumnNames {
		columns[i] = object.Column{
			Name: cst.ColumnNames[i],
			Type: object.DataType(cst.ColumnTypes[i].Literal), // This should always be valid since it has parsed successfully
		}
	}
	if err := backend.CreateTable(cst.Name, columns); err != nil {
		return newError(err.Error())
	}
	return &object.OK{}
}

func evalInsertStatement(backend Backend, is *ast.InsertStatement) object.Object {
	columns := backend.ColumnsInTable(is.TableName)
	row := object.Row{
		Values:    make([]object.Object, len(is.Expressions)),
		TableName: make([]string, len(is.Expressions)),
		Aliases:   columns,
	}
	for i, es := range is.Expressions {
		obj := Eval(backend, es)
		row.Values[i] = obj
		row.TableName[i] = is.TableName
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

func evalPostfixExpression(left object.Object, operator string) object.Object {
	switch operator {
	case "IS NULL":
		switch left.(type) {
		case *object.Null:
			return &object.True
		default:
			return &object.False
		}
	case "IS NOT NULL":
		switch left.(type) {
		case *object.Null:
			return &object.False
		default:
			return &object.True
		}
	default:
		return newError("unknown operator: %s %s", left.Type(), operator)
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
