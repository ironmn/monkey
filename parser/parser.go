package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
)

type Parser struct {
	l *lexer.Lexer

	curToken  token.Token
	peerToken token.Token
	errors    []string
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}
	//读取两个token
	p.nextToken()
	p.nextToken()
	return p
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

func (p *Parser) ParserProgram() *ast.Program {
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
		return p.parserLetStatement()
	default:
		return nil
	}
}

func (p *Parser) parserLetStatement() *ast.LetStatement { //关于这里为什么要加指针返回值，是因为LetStatement接口实现时使用的是指针接收者，这个指针接收者继承了Statement的方法集。但是其值接收者并没有（因为指针接收者和值接收者二者的方法集是不同的）
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
