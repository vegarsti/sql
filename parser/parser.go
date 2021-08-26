package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/vegarsti/sql/ast"
	"github.com/vegarsti/sql/lexer"
	"github.com/vegarsti/sql/token"
)

type (
	prefixParseFn  func() ast.Expression
	postfixParseFn func(ast.Expression) ast.Expression // the argument is the expression on the left side
	infixParseFn   func(ast.Expression) ast.Expression // ditto
)

const (
	_ int = iota
	LOWEST
	SUM     // +
	PRODUCT // *
	PREFIX  // -X
)

type Parser struct {
	l *lexer.Lexer

	curToken  token.Token
	peekToken token.Token

	errors []string

	prefixParseFns  map[token.TokenType]prefixParseFn
	postfixParseFns map[token.TokenType]postfixParseFn
	infixParseFns   map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.INT_LITERAL, p.parseIntegerLiteral)
	p.registerPrefix(token.TRUE, p.parseBooleanLiteral)
	p.registerPrefix(token.FALSE, p.parseBooleanLiteral)
	p.registerPrefix(token.FLOAT_LITERAL, p.parseFloatLiteral)
	p.registerPrefix(token.STRING_LITERAL, p.parseStringLiteral)
	p.registerPrefix(token.IDENTIFIER, p.parseIdentifier)
	p.registerPrefix(token.QUALIFIEDIDENTIFIER, p.parseQualifiedIdentifier)
	p.registerPrefix(token.NULL, p.parseNull)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.NOT, p.parsePrefixExpression)

	p.postfixParseFns = make(map[token.TokenType]postfixParseFn)
	p.registerPostfix(token.IS, p.parseNullIsPostfixExpression)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQUALS, p.parseInfixExpression)
	p.registerInfix(token.NOTEQUALS, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.LESSTHAN, p.parseInfixExpression)
	p.registerInfix(token.LESSTHANOREQUALS, p.parseInfixExpression)
	p.registerInfix(token.GREATERTHAN, p.parseInfixExpression)
	p.registerInfix(token.GREATERTHANOREQUALS, p.parseInfixExpression)
	p.registerInfix(token.DOUBLEBAR, p.parseInfixExpression)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)
	return expression
}

func (p *Parser) parseNullIsPostfixExpression(left ast.Expression) ast.Expression {
	op := "IS NULL"
	p.nextToken()
	if p.peekToken.Type == token.NOT {
		op = "IS NOT NULL"
		p.nextToken()
	}
	if !p.expectPeek(token.NULL) {
		return nil
	}
	expression := &ast.PostfixExpression{
		Token:    p.curToken,
		Operator: op,
		Left:     left,
	}
	p.nextToken()
	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	if p.curTokenIs(token.EOF) {
		p.errors = append(p.errors, "expected operand")
		return nil
	}
	expression.Right = p.parseExpression(precedence)
	return expression
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		// got errors, abort parsing
		if stmt == nil {
			return program
		}
		program.Statements = append(program.Statements, stmt)
		p.nextToken()
	}

	return program
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	lit := &ast.BooleanLiteral{Token: p.curToken}
	value, err := strconv.ParseBool(p.curToken.Literal)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as boolean", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	lit := &ast.StringLiteral{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
	return lit
}

func (p *Parser) parseIdentifier() ast.Expression {
	lit := &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
	return lit
}

func (p *Parser) parseQualifiedIdentifier() ast.Expression {
	split := strings.Split(p.curToken.Literal, ".")
	lit := &ast.Identifier{
		Token: p.curToken,
		Value: split[1],
		Table: split[0],
	}
	return lit
}

func (p *Parser) parseNull() ast.Expression {
	return ast.NULL
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.SELECT:
		return p.parseSelectStatement()
	case token.CREATE:
		return p.parseCreateTableStatement()
	case token.INSERT:
		return p.parseInsertStatement()
	default:
		p.errors = append(p.errors, fmt.Sprintf("expected start of statement, got %s token with literal %s", p.curToken.Type, p.curToken.Literal))
		return nil
	}
}

func (p *Parser) parseElementInSelect() (ast.Expression, string) {
	p.nextToken()
	expr := p.parseExpression(LOWEST)
	// check for AS
	if p.peekToken.Type == token.AS {
		p.nextToken()
		if !p.expectPeek(token.IDENTIFIER) {
			return nil, ""
		}
		alias := p.curToken.Literal
		return expr, alias
	}
	return expr, ""
}

func (p *Parser) parseJoin() *ast.Join {
	p.nextToken()
	if !p.expectPeek(token.IDENTIFIER) {
		return nil
	}
	joinWith := &ast.From{
		Table: p.curToken.Literal,
	}
	if p.peekToken.Type == token.IDENTIFIER {
		p.nextToken()
		joinWith.TableAlias = p.curToken.Literal
	}
	if !p.expectPeek(token.ON) {
		return nil
	}
	p.nextToken()
	joinExpr := p.parseExpression(LOWEST)
	join := &ast.Join{
		With:      joinWith,
		Predicate: joinExpr,
		JoinType:  ast.INNERJOIN,
	}
	// table alias
	if p.peekToken.Type == token.IDENTIFIER {
		p.nextToken()
		join.With.TableAlias = p.curToken.Literal
	}
	return join
}

func (p *Parser) parseFrom() *ast.From {
	p.nextToken()
	if !p.expectPeek(token.IDENTIFIER) {
		return nil
	}
	from := &ast.From{Table: p.curToken.Literal}

	if p.peekToken.Type == token.IDENTIFIER {
		p.nextToken()
		from.TableAlias = p.curToken.Literal
	}

	if p.peekToken.Type == token.JOIN {
		join := p.parseJoin()
		if join == nil {
			return nil
		}
		from.Join = join
	}
	return from
}

func (p *Parser) parseOrderBy() *ast.OrderByExpression {
	p.nextToken()
	sortExpr := p.parseExpression(LOWEST)
	if sortExpr == nil {
		return nil
	}
	orderBy := &ast.OrderByExpression{Expression: sortExpr}
	if p.peekToken.Type == token.DESC {
		orderBy.Descending = true
		p.nextToken()
	} else if p.peekToken.Type == token.ASC {
		p.nextToken()
	}
	return orderBy
}

func (p *Parser) parseLimit() *int {
	if !p.expectPeek(token.INT_LITERAL) {
		return nil
	}
	n, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	if n < 0 {
		p.errors = append(p.errors, fmt.Sprintf("limit must be non-negative, got %d", n))
		return nil
	}
	x := int(n)
	return &x
}

func (p *Parser) parseOffset() *int {
	if !p.expectPeek(token.INT_LITERAL) {
		return nil
	}
	n, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	if n < 0 {
		p.errors = append(p.errors, fmt.Sprintf("offset must be non-negative, got %d", n))
		return nil
	}
	x := int(n)
	return &x
}

func (p *Parser) expectPeekIsEndOfStatement() bool {
	if !(p.peekToken.Type == token.SEMICOLON || p.peekToken.Type == token.EOF) {
		msg := fmt.Sprintf("expected next token to be %s or %s, got %s '%s' instead", token.SEMICOLON, token.EOF, p.peekToken.Type, p.peekToken.Literal)
		p.errors = append(p.errors, msg)
		return false
	}
	p.nextToken()
	return true
}

func (p *Parser) parseSelectStatement() ast.Statement {
	stmt := &ast.SelectStatement{
		Expressions: make([]ast.Expression, 0),
		Aliases:     make([]string, 0),
		From:        make([]*ast.From, 0),
		OrderBy:     make([]*ast.OrderByExpression, 0),
		Limit:       nil,
		Where:       nil,
	}
	expression, alias := p.parseElementInSelect()
	if expression == nil {
		return nil
	}
	stmt.Expressions = append(stmt.Expressions, expression)
	stmt.Aliases = append(stmt.Aliases, alias)

	for p.peekToken.Type == token.COMMA {
		p.nextToken()
		expression, alias := p.parseElementInSelect()
		if expression == nil {
			return nil
		}
		stmt.Expressions = append(stmt.Expressions, expression)
		stmt.Aliases = append(stmt.Aliases, alias)
	}

	if p.peekToken.Type == token.FROM {
		from := p.parseFrom()
		if from == nil {
			return nil
		}
		stmt.From = append(stmt.From, from)
	}

	for p.peekToken.Type == token.COMMA {
		from := p.parseFrom()
		if from == nil {
			return nil
		}
		stmt.From = append(stmt.From, from)
	}

	if p.peekToken.Type == token.WHERE {
		p.nextToken()
		p.nextToken()
		stmt.Where = p.parseExpression(LOWEST)
	}

	if p.peekToken.Type == token.ORDER {
		p.nextToken()
		if !p.expectPeek(token.BY) {
			return nil
		}
		orderBy := p.parseOrderBy()
		if orderBy == nil {
			return nil
		}
		stmt.OrderBy = append(stmt.OrderBy, orderBy)
		for p.peekToken.Type == token.COMMA {
			p.nextToken()
			orderBy := p.parseOrderBy()
			if orderBy == nil {
				return nil
			}
			stmt.OrderBy = append(stmt.OrderBy, orderBy)
		}
	}

	if p.peekToken.Type == token.LIMIT {
		p.nextToken()
		limit := p.parseLimit()
		if limit == nil {
			return nil
		}
		stmt.Limit = limit
	}

	if p.peekToken.Type == token.OFFSET {
		p.nextToken()
		offset := p.parseOffset()
		if offset == nil {
			return nil
		}
		stmt.Offset = offset
	}

	if !p.expectPeekIsEndOfStatement() {
		return nil
	}

	return stmt
}

func (p *Parser) expectPeekType() bool {
	if !(p.peekToken.Type == token.STRING_TYPE || p.peekToken.Type == token.FLOAT_TYPE || p.peekToken.Type == token.INTEGER_TYPE || p.peekToken.Type == token.BOOLEAN_TYPE) {
		p.errors = append(p.errors, fmt.Sprintf("expected type, got %s token with literal %s", p.peekToken.Type, p.peekToken.Literal))
		return false
	}
	p.nextToken()
	return true
}

func (p *Parser) parseCreateTableStatement() ast.Statement {
	stmt := &ast.CreateTableStatement{
		ColumnNames: make([]string, 0),
		ColumnTypes: make([]token.Token, 0),
	}

	if !p.expectPeek(token.TABLE) {
		return nil
	}

	if !p.expectPeek(token.IDENTIFIER) {
		return nil
	}
	stmt.Name = p.curToken.Literal

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	// parse column pairs
	if !p.expectPeek(token.IDENTIFIER) {
		return nil
	}
	stmt.ColumnNames = append(stmt.ColumnNames, p.curToken.Literal)
	if !p.expectPeekType() {
		return nil
	}
	stmt.ColumnTypes = append(stmt.ColumnTypes, p.curToken)

	for p.peekToken.Type == token.COMMA {
		p.nextToken()
		if !p.expectPeek(token.IDENTIFIER) {
			return nil
		}
		stmt.ColumnNames = append(stmt.ColumnNames, p.curToken.Literal)

		if !p.expectPeekType() {
			return nil
		}
		stmt.ColumnTypes = append(stmt.ColumnTypes, p.curToken)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeekIsEndOfStatement() {
		return nil
	}

	return stmt
}

func (p *Parser) parseInsertStatement() ast.Statement {
	stmt := &ast.InsertStatement{
		Expressions: make([]ast.Expression, 0),
	}

	if !p.expectPeek(token.INTO) {
		return nil
	}

	if !p.expectPeek(token.IDENTIFIER) {
		return nil
	}
	stmt.TableName = p.curToken.Literal

	if !p.expectPeek(token.VALUES) {
		return nil
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()

	expr := p.parseExpression(LOWEST)
	if expr == nil {
		return nil
	}
	stmt.Expressions = append(stmt.Expressions, expr)

	for p.peekToken.Type == token.COMMA {
		p.nextToken()
		p.nextToken()

		expr := p.parseExpression(LOWEST)
		if expr == nil {
			return nil
		}
		stmt.Expressions = append(stmt.Expressions, expr)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeekIsEndOfStatement() {
		return nil
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type, p.curToken.Literal)
		return nil
	}
	leftExp := prefix()

	for precedence < p.peekPrecedence() {
		if postfix := p.postfixParseFns[p.peekToken.Type]; postfix != nil {
			return postfix(leftExp)
		}
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s '%s' instead",
		t, p.peekToken.Type, p.peekToken.Literal)
	p.errors = append(p.errors, msg)
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerPostfix(tokenType token.TokenType, fn postfixParseFn) {
	p.postfixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) noPrefixParseFnError(t token.TokenType, literal string) {
	msg := fmt.Sprintf("no prefix parse function for %s token with literal '%s' found", t, literal)
	p.errors = append(p.errors, msg)
}

var precedences = map[token.TokenType]int{
	token.PLUS:                SUM,
	token.MINUS:               SUM,
	token.SLASH:               PRODUCT,
	token.ASTERISK:            PRODUCT,
	token.EQUALS:              PRODUCT,
	token.NOTEQUALS:           PRODUCT,
	token.AND:                 SUM,
	token.OR:                  SUM,
	token.LESSTHAN:            SUM,
	token.LESSTHANOREQUALS:    SUM,
	token.GREATERTHAN:         SUM,
	token.GREATERTHANOREQUALS: SUM,
	token.IS:                  PRODUCT,
	token.DOUBLEBAR:           SUM,
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return exp
}
