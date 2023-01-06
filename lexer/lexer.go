package lexer

import (
	"fmt"
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

func (l *Lexer) unreadChar() { //回退向前看字符
	if l.position == 0 {
		l.ch = 0
		return
	}
	l.ch = l.input[l.readPosition]
	l.readPosition = l.position
	l.position -= 1
}

func (l *Lexer) NextToken() token.Token {
	for l.ch == ' ' || l.ch == '\n' || l.ch == '\b' || l.ch == '\t' || l.ch == '\r' { //吸收空字符
		l.readChar()
	}
	fmt.Print(l.ch)
	var tok token.Token
	switch l.ch {
	case '=':
		tok = newToken(token.ASSIGN, l.ch)
		if l.input[l.readPosition] == '=' { //如果下一个字符依然是=
			tok = token.Token{Type: token.EQ, Literal: "=="}
			l.readChar()
		}
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
	default:
		if isLetter(l.ch) { //如果是英文字符开头
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isNumber(l.ch) { //如果是number开头的，默认其为INT类型
			tok.Literal = l.readNumber()
			tok.Type = token.INT
			return tok
		} else if l.ch == ' ' || l.ch == '\n' || l.ch == '\t' {
			break
		}
	}
	l.readChar()
	return tok
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func isLetter(ch byte) bool {
	return ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch == '_'
}
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}
func isNumber(ch byte) bool {
	if ch >= '0' && ch <= '9' {
		return true
	}
	return false
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isNumber(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}
