package parser

import (
	"monkey/ast"
	"monkey/lexer"
	"testing"
)

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{"let x = 5;", "x", 5},
		{"let y = true;", "y", true},
		{"let foobar = y;", "foobar", "y"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParserProgram()
		//checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d",
				len(program.Statements))
		}

		stmt := program.Statements[0]
		if !testLetStatement(t, stmt, tt.expectedIdentifier) {
			return
		}

		//val := stmt.(*ast.LetStatement).Value
		//if !testLiteralExpression(t, val, tt.expectedValue) {
		//	return
		//}
	}
}

func testLetStatement(t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != "let" {
		t.Errorf("s.TokenLiteral not 'let'. got=%q", s.TokenLiteral())
		return false
	}

	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		t.Errorf("s not *ast.LetStatement. got=%T", s)
		return false
	}

	if letStmt.Name.Value != name {
		t.Errorf("letStmt.Name.Value not '%s'. got=%s", name, letStmt.Name.Value)
		return false
	}

	if letStmt.Name.TokenLiteral() != name {
		t.Errorf("letStmt.Name.TokenLiteral() not '%s'. got=%s",
			name, letStmt.Name.TokenLiteral())
		return false
	}

	return true
}

//func testLiteralExpression(
//	t *testing.T,
//	exp ast.Expression,
//	expected interface{},
//) bool {
//	switch v := expected.(type) {
//	case int:
//		return testIntegerLiteral(t, exp, int64(v))
//	case int64:
//		return testIntegerLiteral(t, exp, v)
//	case string:
//		return testIdentifier(t, exp, v)
//	case bool:
//		return testBooleanLiteral(t, exp, v)
//	}
//	t.Errorf("type of exp not handled. got=%T", exp)
//	return false
//}
