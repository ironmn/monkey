package lexer

import (
	"monkey/token"
)

type Lexer struct {
	input        string
	position     int //这个是当前的字符
	readPosition int //这个代表向后看字符
	ch           byte
}

func New(input string) *Lexer { //返回的是一个指针类型
	l := &Lexer{input: input} //直接获取初始化的指针变量地址
	l.readChar()              //在新建的时候直接对其初始化
	return l
}

/**
辅助函数，用于将移动指针的这种原子操作抽象出来。
*/
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 //表明已经到了字符串结尾
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1 //结束，保证readPosition始终在position前面一个字符
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token
	switch l.ch {
	case '=':
		tok = newToken(token.ASSIGN, l.ch)
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '-':
		tok = newToken(token.MINUS, l.ch)
	case 0: //0表示字符串的末尾
		tok.Literal = ""
		tok.Type = token.EOF
	}
	l.readChar()
	return tok
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}
