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
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
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

	//for循环的作用在于，能够结合连续的同等优先级的，并且大于当前precedence的项（如a + b * c * d）
	//使用递归能够让每一个高优先级的项先于for循环进行结合，然后再通过for循环将回溯的结果进行结合
	for !p.peerTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() { //这个比较判断语句的意思是，如果当前表达式的运算符优先级低于下一个运算符的优先级，那么就需要将当前的右值进行右结合，而非左结合
		//如果precedence>p.peekPrecedence，那么说明当前运算符的左结合能力大于下一个运算符的右结合能力，所以不能将当前右值和下一个左值进行结合。
		//这个过程实质上是根据优先级的大小进行对语法树进行中序遍历，将遍历的结果存放在expression对象中
		infix := p.infixParseFns[p.peerToken.Type]
		if infix == nil { //如果发现没有对应的infix函数，就直接返回左值
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp) //递归函数的作用主要是用来结合更高优先级的项
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
	precedence := p.curPrecedence()
	p.nextToken()                                    //由于已经完成了左值的读取，那么就需要继续获取下一个词法单元
	expression.Right = p.parseExpression(precedence) //使用expression的表达式优先级作为获取右值的参数
	return expression
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST) //由于parseExpression这个函数内部也有递归函数，同时所有的递归函数中，传递优先级的参数只会增加不会减少，能够控制
	//这个优先级参数参数的方式就是在外部调用的时候传递一个优先级参数。当遇到左括号的时候，就强制调用一个新的parseExpression递归函数，由于递归程序的执行顺序是永远先于for循环的，因此这里能够顺利地得到结果。

	if !p.expectPeek(token.RPAREN) { //如果没有遇到右括号，则说明出现了语法错误
		return nil //返回一个空的值
	}
	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}
	if !p.expectPeek(token.LPAREN) { //如果不是左括号，则说明语法分析出现了错误，暂时返回nil
		return nil
	}
	p.nextToken()                                    //将词法指针移动到表达式的开头
	expression.Condition = p.parseExpression(LOWEST) //去解析IF语句里面的表达式
	if !p.expectPeek(token.RPAREN) {                 //解析完条件表达式之后，如果遇到的不是右括号，则说明语法分析器出现了问题，返回nil
		return nil
	}

	if !p.expectPeek(token.LBRACE) { //如果不是左大括号
		return nil
	}

	expression.Consequence = p.parseBlockStatement()

	if p.peerTokenIs(token.ELSE) {
		p.nextToken()

		if !p.expectPeek(token.LBRACE) {
			return nil
		}
		expression.Alternative = p.parseBlockStatement() //解析else后面的语句
	}
	return expression
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	blockStatements := &ast.BlockStatement{Token: p.curToken} //设置左大括号为该语法单元的词法标记
	p.nextToken()
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		statement := p.parseStatement()
		blockStatements.Statements = append(blockStatements.Statements, statement)
		p.nextToken()
	}
	return blockStatements
}
