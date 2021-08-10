package parser

import (
	"fmt"
	"strconv"

	"github.com/vegarsti/sql/ast"
	"github.com/vegarsti/sql/lexer"
	"github.com/vegarsti/sql/token"
)

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression // the argument is the expression on the left side
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

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.IDENTIFIER, p.parseIdentifier)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)

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

func (p *Parser) parseSelectStatement() ast.Statement {
	stmt := &ast.SelectStatement{
		Token:       p.curToken,
		Expressions: make([]ast.Expression, 0),
		Aliases:     make([]string, 0),
	}
	p.nextToken() // read SELECT token

	stmt.Expressions = append(stmt.Expressions, p.parseExpression(LOWEST))
	p.nextToken()

	// check for AS
	if p.curToken.Type == token.AS {
		p.nextToken() // read AS

		// assert next token is an identifier
		if p.curToken.Type != token.IDENTIFIER {
			p.errors = append(p.errors, fmt.Sprintf("expected identifier, got %s token with literal %s", p.curToken.Type, p.curToken.Literal))
			return nil
		}
		stmt.Aliases = append(stmt.Aliases, p.curToken.Literal)

		p.nextToken()
	} else {
		stmt.Aliases = append(stmt.Aliases, "")
	}

	for p.curToken.Type == token.COMMA {
		p.nextToken() // read comma
		stmt.Expressions = append(stmt.Expressions, p.parseExpression(LOWEST))
		p.nextToken() // advance to next token

		// check for AS
		if p.curToken.Type != token.AS {
			stmt.Aliases = append(stmt.Aliases, "")
			continue
		}
		p.nextToken() // read AS

		// assert next token is an identifier
		if p.curToken.Type != token.IDENTIFIER {
			p.errors = append(p.errors, fmt.Sprintf("expected identifier, got %s token with literal %s", p.curToken.Type, p.curToken.Literal))
			return nil
		}
		stmt.Aliases = append(stmt.Aliases, p.curToken.Literal)

		p.nextToken() // advance to next token
	}

	if p.curToken.Type == token.FROM {
		p.nextToken()
		// assert next token is an identifier
		if p.curToken.Type != token.IDENTIFIER {
			p.errors = append(p.errors, fmt.Sprintf("expected identifier, got %s token with literal %s", p.curToken.Type, p.curToken.Literal))
			return nil
		}
		stmt.From = p.curToken.Literal
		p.nextToken()
	}

	if !(p.curToken.Type == token.SEMICOLON || p.curToken.Type == token.EOF) {
		msg := fmt.Sprintf("expected next token to be %s or %s, got %s '%s' instead", token.SEMICOLON, token.EOF, p.peekToken.Type, p.peekToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	return stmt
}

func (p *Parser) parseCreateTableStatement() ast.Statement {
	stmt := &ast.CreateTableStatement{
		Token:   p.curToken,
		Columns: make(map[string]token.Token),
	}

	if !p.expectPeek(token.TABLE) {
		return nil
	}

	// assert next token is an identifier
	if p.peekToken.Type != token.IDENTIFIER {
		p.errors = append(p.errors, fmt.Sprintf("expected identifier, got %T token with literal %s", p.peekToken.Type, p.peekToken.Literal))
		return nil
	}
	p.nextToken()
	stmt.Name = p.curToken.Literal

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	// parse column pairs
	// assert next token is a column identifier
	if p.peekToken.Type != token.IDENTIFIER {
		p.errors = append(p.errors, fmt.Sprintf("expected identifier, got %T token with literal %s", p.peekToken.Type, p.peekToken.Literal))
		return nil
	}
	p.nextToken()
	columnLiteral := p.curToken.Literal

	// assert next token is a column type
	if !(p.peekToken.Type == token.TEXT || p.peekToken.Type == token.DOUBLE || p.peekToken.Type == token.INTEGER) {
		p.errors = append(p.errors, fmt.Sprintf("expected type, got %s token with literal %s", p.peekToken.Type, p.peekToken.Literal))
		return nil
	}
	p.nextToken()
	stmt.Columns[columnLiteral] = p.curToken

	for p.peekToken.Type == token.COMMA {
		p.nextToken()
		// assert next token is a column identifier
		if p.peekToken.Type != token.IDENTIFIER {
			p.errors = append(p.errors, fmt.Sprintf("expected identifier, got %T token with literal %s", p.peekToken.Type, p.peekToken.Literal))
			return nil
		}
		p.nextToken()
		columnLiteral := p.curToken.Literal

		// assert next token is a column type
		if !(p.peekToken.Type == token.TEXT || p.peekToken.Type == token.DOUBLE || p.peekToken.Type == token.INTEGER) {
			p.errors = append(p.errors, fmt.Sprintf("expected type, got %s token with literal %s", p.peekToken.Type, p.peekToken.Literal))
			return nil
		}
		p.nextToken()
		stmt.Columns[columnLiteral] = p.curToken
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !(p.peekToken.Type == token.SEMICOLON || p.peekToken.Type == token.EOF) {
		msg := fmt.Sprintf("expected next token to be %s or %s, got %s '%s' instead", token.SEMICOLON, token.EOF, p.peekToken.Type, p.peekToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	p.nextToken()

	return stmt
}

func (p *Parser) parseInsertStatement() ast.Statement {
	stmt := &ast.InsertStatement{
		Token:       p.curToken,
		Expressions: make([]ast.Expression, 0),
	}

	if !p.expectPeek(token.INTO) {
		return nil
	}

	// assert next token is an identifier
	if p.peekToken.Type != token.IDENTIFIER {
		p.errors = append(p.errors, fmt.Sprintf("expected identifier, got %T token with literal %s", p.peekToken.Type, p.peekToken.Literal))
		return nil
	}
	p.nextToken()
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

	if !(p.peekToken.Type == token.SEMICOLON || p.peekToken.Type == token.EOF) {
		msg := fmt.Sprintf("expected next token to be %s or %s, got %s '%s' instead", token.SEMICOLON, token.EOF, p.peekToken.Type, p.peekToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	p.nextToken()

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

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) noPrefixParseFnError(t token.TokenType, literal string) {
	msg := fmt.Sprintf("no prefix parse function for %s token with literal '%s' found", t, literal)
	p.errors = append(p.errors, msg)
}

var precedences = map[token.TokenType]int{
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
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
