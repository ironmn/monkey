package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	EQUALS      //==
	LESSGREATER //< or >
	SUM         //+
	PRODUCT     //*
	PREFIX      //--,++,-,!...
	CALL        //add(x+y)
)

//添加优先级表
var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
}

//定义前缀函数和中缀函数，并设置这两种之间的关联（通过参数传递）
type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(expression ast.Expression) ast.Expression
)

type Parser struct {
	l *lexer.Lexer

	curToken  token.Token
	peerToken token.Token
	errors    []string

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}
	//读取两个token
	p.nextToken()
	p.nextToken()

	//初始化关联函数的过程，这个过程主要是用来模拟递归下降预测分析表的构建过程
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)

	//关联infix函数
	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	return p
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) nextToken() {
	p.curToken = p.peerToken
	p.peerToken = p.l.NextToken()
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peerToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) peekPrecedence() int {
	if ans, ok := precedences[p.peerToken.Type]; ok {
		return ans
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if ans, ok := precedences[p.curToken.Type]; ok {
		return ans
	}
	return LOWEST
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}              //首先构建一个空的Program对象
	program.Statements = []ast.Statement{} //构建Program对象的Statement成员变量
	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement { //关于这里为什么要加指针返回值，是因为LetStatement接口实现时使用的是指针接收者，这个指针接收者继承了Statement的方法集。但是其值接收者并没有（因为指针接收者和值接收者二者的方法集是不同的）
	//当这里用了值接收者作为返回对象时，由于值接收者并没有实现Statement接口的所有方法，因此在作为泛型时它就不能作为Statement的返回对象
	stmt := &ast.LetStatement{Token: p.curToken}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
	//@todo: 跳过对表达式的解析
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()
	//@todo: 跳过对于表达式的解析
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peerTokenIs(t token.TokenType) bool {
	return p.peerToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peerTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t) //添加错误
		return false
	}
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)

	if p.peerTokenIs(token.SEMICOLON) { //由于';'不影响表达式的解析，因此这里可以选择跳过
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peerTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peerToken.Type]
		if infix == nil { //如果发现没有对应的infix函数，就直接返回左值
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}
	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
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

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{Token: p.curToken, Operator: p.curToken.Literal}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)
	return expression
}

func (p *Parser) parseInfixExpression(leftExp ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{Token: p.curToken, Left: leftExp, Operator: p.curToken.Literal}
	p.nextToken()                                                            //由于已经完成了左值的读取，那么就需要继续获取下一个词法单元
	expression.Right = p.parseExpression(precedences[expression.Token.Type]) //使用expression的表达式优先级作为获取右值的参数
	return expression
}
