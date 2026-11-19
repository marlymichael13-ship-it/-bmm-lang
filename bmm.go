package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

type TokenType int

const (
	TOK_EOF TokenType = iota
	TOK_INT
	TOK_FLOAT
	TOK_STRING
	TOK_BOOL
	TOK_IDENT

	TOK_LET
	TOK_FN
	TOK_RETURN
	TOK_IF
	TOK_ELSE
	TOK_WHILE
	TOK_FOR
	TOK_BREAK
	TOK_CONTINUE
	TOK_PRINT
	TOK_INPUT
	TOK_LEN
	TOK_TYPEOF
	TOK_APPEND
	TOK_INT_CAST
	TOK_FLOAT_CAST
	TOK_STR_CAST

	TOK_PLUS
	TOK_MINUS
	TOK_STAR
	TOK_SLASH
	TOK_PERCENT
	TOK_PLUS_ASSIGN
	TOK_MINUS_ASSIGN
	TOK_STAR_ASSIGN
	TOK_SLASH_ASSIGN
	TOK_ASSIGN
	TOK_EQ
	TOK_NEQ
	TOK_LT
	TOK_LTE
	TOK_GT
	TOK_GTE
	TOK_AND
	TOK_OR
	TOK_BANG

	TOK_LPAREN
	TOK_RPAREN
	TOK_LBRACE
	TOK_RBRACE
	TOK_LBRACKET
	TOK_RBRACKET
	TOK_COMMA
	TOK_COLON
	TOK_DOT
)

func (t TokenType) String() string {
	switch t {
	case TOK_EOF:
		return "EOF"
	case TOK_INT:
		return "INT"
	case TOK_FLOAT:
		return "FLOAT"
	case TOK_STRING:
		return "STRING"
	case TOK_BOOL:
		return "BOOL"
	case TOK_IDENT:
		return "IDENT"
	case TOK_LET:
		return "LET"
	case TOK_FN:
		return "FN"
	case TOK_RETURN:
		return "RETURN"
	case TOK_IF:
		return "IF"
	case TOK_ELSE:
		return "ELSE"
	case TOK_WHILE:
		return "WHILE"
	case TOK_FOR:
		return "FOR"
	case TOK_BREAK:
		return "BREAK"
	case TOK_CONTINUE:
		return "CONTINUE"
	case TOK_PRINT:
		return "PRINT"
	case TOK_INPUT:
		return "INPUT"
	case TOK_LEN:
		return "LEN"
	case TOK_TYPEOF:
		return "TYPEOF"
	case TOK_APPEND:
		return "APPEND"
	case TOK_INT_CAST:
		return "INT_CAST"
	case TOK_FLOAT_CAST:
		return "FLOAT_CAST"
	case TOK_STR_CAST:
		return "STR_CAST"
	case TOK_PLUS:
		return "+"
	case TOK_MINUS:
		return "-"
	case TOK_STAR:
		return "*"
	case TOK_SLASH:
		return "/"
	case TOK_PERCENT:
		return "%"
	case TOK_PLUS_ASSIGN:
		return "+="
	case TOK_MINUS_ASSIGN:
		return "-="
	case TOK_STAR_ASSIGN:
		return "*="
	case TOK_SLASH_ASSIGN:
		return "/="
	case TOK_ASSIGN:
		return "="
	case TOK_EQ:
		return "=="
	case TOK_NEQ:
		return "!="
	case TOK_LT:
		return "<"
	case TOK_LTE:
		return "<="
	case TOK_GT:
		return ">"
	case TOK_GTE:
		return ">="
	case TOK_AND:
		return "&&"
	case TOK_OR:
		return "||"
	case TOK_BANG:
		return "!"
	case TOK_LPAREN:
		return "("
	case TOK_RPAREN:
		return ")"
	case TOK_LBRACE:
		return "{"
	case TOK_RBRACE:
		return "}"
	case TOK_LBRACKET:
		return "["
	case TOK_RBRACKET:
		return "]"
	case TOK_COMMA:
		return ","
	case TOK_COLON:
		return ":"
	case TOK_DOT:
		return "."
	default:
		return "UNKNOWN"
	}
}

type Token struct {
	Type TokenType
	Lex  string
	Line int
	Col  int
}

type Lexer struct {
	src  []rune
	pos  int
	line int
	col  int
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		src:  []rune(input),
		pos:  0,
		line: 1,
		col:  1,
	}
}

func (l *Lexer) peek() rune {
	if l.pos >= len(l.src) {
		return 0
	}
	return l.src[l.pos]
}

func (l *Lexer) peekNext() rune {
	if l.pos+1 >= len(l.src) {
		return 0
	}
	return l.src[l.pos+1]
}

func (l *Lexer) advance() rune {
	ch := l.peek()
	if ch == 0 {
		return 0
	}
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	l.pos++
	return ch
}

func (l *Lexer) skipWhitespaceAndComments() {
	for {
		ch := l.peek()
		if ch == 0 {
			return
		}
		if unicode.IsSpace(ch) {
			l.advance()
			continue
		}
		if ch == '/' && l.peekNext() == '/' {
			for l.peek() != 0 && l.peek() != '\n' {
				l.advance()
			}
			continue
		}
		if ch == '/' && l.peekNext() == '*' {
			l.advance()
			l.advance()
			for l.peek() != 0 {
				if l.peek() == '*' && l.peekNext() == '/' {
					l.advance()
					l.advance()
					break
				}
				l.advance()
			}
			continue
		}
		break
	}
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespaceAndComments()
	line := l.line
	col := l.col
	ch := l.peek()
	if ch == 0 {
		return Token{Type: TOK_EOF, Lex: "", Line: line, Col: col}
	}

	if ch == '"' || ch == '\'' {
		quote := l.advance()
		var sb strings.Builder
		for l.peek() != 0 && l.peek() != quote {
			c := l.advance()
			if c == '\\' {
				esc := l.advance()
				switch esc {
				case 'n':
					sb.WriteRune('\n')
				case 'r':
					sb.WriteRune('\r')
				case 't':
					sb.WriteRune('\t')
				case '\\':
					sb.WriteRune('\\')
				case '"':
					sb.WriteRune('"')
				case '\'':
					sb.WriteRune('\'')
				default:
					sb.WriteRune(esc)
				}
			} else {
				sb.WriteRune(c)
			}
		}
		if l.peek() == quote {
			l.advance()
		}
		return Token{Type: TOK_STRING, Lex: sb.String(), Line: line, Col: col}
	}

	if unicode.IsDigit(ch) {
		var sb strings.Builder
		isFloat := false
		for unicode.IsDigit(l.peek()) || (l.peek() == '.' && !isFloat && unicode.IsDigit(l.peekNext())) {
			c := l.advance()
			if c == '.' {
				isFloat = true
			}
			sb.WriteRune(c)
		}
		if isFloat {
			return Token{Type: TOK_FLOAT, Lex: sb.String(), Line: line, Col: col}
		}
		return Token{Type: TOK_INT, Lex: sb.String(), Line: line, Col: col}
	}

	if unicode.IsLetter(ch) || ch == '_' {
		var sb strings.Builder
		for unicode.IsLetter(l.peek()) || unicode.IsDigit(l.peek()) || l.peek() == '_' {
			sb.WriteRune(l.advance())
		}
		word := sb.String()
		switch word {
		case "let":
			return Token{Type: TOK_LET, Lex: word, Line: line, Col: col}
		case "fn":
			return Token{Type: TOK_FN, Lex: word, Line: line, Col: col}
		case "return":
			return Token{Type: TOK_RETURN, Lex: word, Line: line, Col: col}
		case "if":
			return Token{Type: TOK_IF, Lex: word, Line: line, Col: col}
		case "else":
			return Token{Type: TOK_ELSE, Lex: word, Line: line, Col: col}
		case "while":
			return Token{Type: TOK_WHILE, Lex: word, Line: line, Col: col}
		case "for":
			return Token{Type: TOK_FOR, Lex: word, Line: line, Col: col}
		case "break":
			return Token{Type: TOK_BREAK, Lex: word, Line: line, Col: col}
		case "continue":
			return Token{Type: TOK_CONTINUE, Lex: word, Line: line, Col: col}
		case "print":
			return Token{Type: TOK_PRINT, Lex: word, Line: line, Col: col}
		case "input":
			return Token{Type: TOK_INPUT, Lex: word, Line: line, Col: col}
		case "len":
			return Token{Type: TOK_LEN, Lex: word, Line: line, Col: col}
		case "typeof":
			return Token{Type: TOK_TYPEOF, Lex: word, Line: line, Col: col}
		case "append":
			return Token{Type: TOK_APPEND, Lex: word, Line: line, Col: col}
		case "int":
			return Token{Type: TOK_INT_CAST, Lex: word, Line: line, Col: col}
		case "float":
			return Token{Type: TOK_FLOAT_CAST, Lex: word, Line: line, Col: col}
		case "str":
			return Token{Type: TOK_STR_CAST, Lex: word, Line: line, Col: col}
		case "true", "false":
			return Token{Type: TOK_BOOL, Lex: word, Line: line, Col: col}
		default:
			return Token{Type: TOK_IDENT, Lex: word, Line: line, Col: col}
		}
	}

	two := string([]rune{ch, l.peekNext()})
	switch two {
	case "==":
		l.advance()
		l.advance()
		return Token{Type: TOK_EQ, Lex: "==", Line: line, Col: col}
	case "!=":
		l.advance()
		l.advance()
		return Token{Type: TOK_NEQ, Lex: "!=", Line: line, Col: col}
	case "<=":
		l.advance()
		l.advance()
		return Token{Type: TOK_LTE, Lex: "<=", Line: line, Col: col}
	case ">=":
		l.advance()
		l.advance()
		return Token{Type: TOK_GTE, Lex: ">=", Line: line, Col: col}
	case "&&":
		l.advance()
		l.advance()
		return Token{Type: TOK_AND, Lex: "&&", Line: line, Col: col}
	case "||":
		l.advance()
		l.advance()
		return Token{Type: TOK_OR, Lex: "||", Line: line, Col: col}
	case "+=":
		l.advance()
		l.advance()
		return Token{Type: TOK_PLUS_ASSIGN, Lex: "+=", Line: line, Col: col}
	case "-=":
		l.advance()
		l.advance()
		return Token{Type: TOK_MINUS_ASSIGN, Lex: "-=", Line: line, Col: col}
	case "*=":
		l.advance()
		l.advance()
		return Token{Type: TOK_STAR_ASSIGN, Lex: "*=", Line: line, Col: col}
	case "/=":
		l.advance()
		l.advance()
		return Token{Type: TOK_SLASH_ASSIGN, Lex: "/=", Line: line, Col: col}
	}

	l.advance()
	switch ch {
	case '+':
		return Token{Type: TOK_PLUS, Lex: "+", Line: line, Col: col}
	case '-':
		return Token{Type: TOK_MINUS, Lex: "-", Line: line, Col: col}
	case '*':
		return Token{Type: TOK_STAR, Lex: "*", Line: line, Col: col}
	case '/':
		return Token{Type: TOK_SLASH, Lex: "/", Line: line, Col: col}
	case '%':
		return Token{Type: TOK_PERCENT, Lex: "%", Line: line, Col: col}
	case '=':
		return Token{Type: TOK_ASSIGN, Lex: "=", Line: line, Col: col}
	case '<':
		return Token{Type: TOK_LT, Lex: "<", Line: line, Col: col}
	case '>':
		return Token{Type: TOK_GT, Lex: ">", Line: line, Col: col}
	case '!':
		return Token{Type: TOK_BANG, Lex: "!", Line: line, Col: col}
	case '(':
		return Token{Type: TOK_LPAREN, Lex: "(", Line: line, Col: col}
	case ')':
		return Token{Type: TOK_RPAREN, Lex: ")", Line: line, Col: col}
	case '{':
		return Token{Type: TOK_LBRACE, Lex: "{", Line: line, Col: col}
	case '}':
		return Token{Type: TOK_RBRACE, Lex: "}", Line: line, Col: col}
	case '[':
		return Token{Type: TOK_LBRACKET, Lex: "[", Line: line, Col: col}
	case ']':
		return Token{Type: TOK_RBRACKET, Lex: "]", Line: line, Col: col}
	case ',':
		return Token{Type: TOK_COMMA, Lex: ",", Line: line, Col: col}
	case ':':
		return Token{Type: TOK_COLON, Lex: ":", Line: line, Col: col}
	case '.':
		return Token{Type: TOK_DOT, Lex: ".", Line: line, Col: col}
	default:
		return Token{Type: TOK_EOF, Lex: string(ch), Line: line, Col: col}
	}
}

type Expr interface {
	exprNode()
}

type Stmt interface {
	stmtNode()
}

type IntExpr struct {
	Val int64
}

func (*IntExpr) exprNode() {}

type FloatExpr struct {
	Val float64
}

func (*FloatExpr) exprNode() {}

type StrExpr struct {
	Val string
}

func (*StrExpr) exprNode() {}

type BoolExpr struct {
	Val bool
}

func (*BoolExpr) exprNode() {}

type VarExpr struct {
	Name string
}

func (*VarExpr) exprNode() {}

type ArrayExpr struct {
	Elements []Expr
}

func (*ArrayExpr) exprNode() {}

type IndexExpr struct {
	Target Expr
	Index  Expr
}

func (*IndexExpr) exprNode() {}

type UnaryExpr struct {
	Op    string
	Right Expr
}

func (*UnaryExpr) exprNode() {}

type BinaryExpr struct {
	Left  Expr
	Op    string
	Right Expr
}

func (*BinaryExpr) exprNode() {}

type CallExpr struct {
	Callee string
	Args   []Expr
}

func (*CallExpr) exprNode() {}

type BuiltinCallExpr struct {
	Name string
	Args []Expr
}

func (*BuiltinCallExpr) exprNode() {}

type LetStmt struct {
	Name  string
	Value Expr
}

func (*LetStmt) stmtNode() {}

type AssignStmt struct {
	Name  string
	Op    string
	Value Expr
}

func (*AssignStmt) stmtNode() {}

type IndexAssignStmt struct {
	Target Expr
	Index  Expr
	Op     string
	Value  Expr
}

func (*IndexAssignStmt) stmtNode() {}

type PrintStmt struct {
	Args []Expr
}

func (*PrintStmt) stmtNode() {}

type ReturnStmt struct {
	Value Expr
}

func (*ReturnStmt) stmtNode() {}

type BreakStmt struct{}

func (*BreakStmt) stmtNode() {}

type ContinueStmt struct{}

func (*ContinueStmt) stmtNode() {}

type BlockStmt struct {
	Stmts []Stmt
}

func (*BlockStmt) stmtNode() {}

type IfStmt struct {
	Cond Expr
	Then *BlockStmt
	Else *BlockStmt
}

func (*IfStmt) stmtNode() {}

type WhileStmt struct {
	Cond Expr
	Body *BlockStmt
}

func (*WhileStmt) stmtNode() {}

type ForStmt struct {
	Init Stmt
	Cond Expr
	Post Stmt
	Body *BlockStmt
}

func (*ForStmt) stmtNode() {}

type FnDeclStmt struct {
	Name   string
	Params []string
	Body   *BlockStmt
}

func (*FnDeclStmt) stmtNode() {}

type ExprStmt struct {
	E Expr
}

func (*ExprStmt) stmtNode() {}

type Parser struct {
	lex *Lexer
	cur Token
}

func NewParser(input string) *Parser {
	p := &Parser{lex: NewLexer(input)}
	p.next()
	return p
}

func (p *Parser) next() {
	p.cur = p.lex.NextToken()
}

func (p *Parser) expect(t TokenType) error {
	if p.cur.Type != t {
		return fmt.Errorf("line %d, col %d: expected %s, got %q", p.cur.Line, p.cur.Col, t.String(), p.cur.Lex)
	}
	p.next()
	return nil
}

func (p *Parser) ParseProgram() ([]Stmt, error) {
	var stmts []Stmt
	for p.cur.Type != TOK_EOF {
		s, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, s)
	}
	return stmts, nil
}

func (p *Parser) parseBlock() (*BlockStmt, error) {
	if err := p.expect(TOK_LBRACE); err != nil {
		return nil, err
	}
	var stmts []Stmt
	for p.cur.Type != TOK_RBRACE && p.cur.Type != TOK_EOF {
		s, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, s)
	}
	if err := p.expect(TOK_RBRACE); err != nil {
		return nil, err
	}
	return &BlockStmt{Stmts: stmts}, nil
}

func (p *Parser) parseArgList() ([]Expr, error) {
	if err := p.expect(TOK_LPAREN); err != nil {
		return nil, err
	}
	var args []Expr
	for p.cur.Type != TOK_RPAREN && p.cur.Type != TOK_EOF {
		e, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		args = append(args, e)
		if p.cur.Type == TOK_COMMA {
			p.next()
		} else {
			break
		}
	}
	if err := p.expect(TOK_RPAREN); err != nil {
		return nil, err
	}
	return args, nil
}

func (p *Parser) parseStmt() (Stmt, error) {
	switch p.cur.Type {
	case TOK_LET:
		p.next()
		if p.cur.Type != TOK_IDENT {
			return nil, fmt.Errorf("line %d, col %d: expected variable identifier", p.cur.Line, p.cur.Col)
		}
		name := p.cur.Lex
		p.next()
		if err := p.expect(TOK_ASSIGN); err != nil {
			return nil, err
		}
		val, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		return &LetStmt{Name: name, Value: val}, nil

	case TOK_FN:
		p.next()
		if p.cur.Type != TOK_IDENT {
			return nil, fmt.Errorf("line %d, col %d: expected function identifier", p.cur.Line, p.cur.Col)
		}
		name := p.cur.Lex
		p.next()
		if err := p.expect(TOK_LPAREN); err != nil {
			return nil, err
		}
		var params []string
		for p.cur.Type != TOK_RPAREN && p.cur.Type != TOK_EOF {
			if p.cur.Type != TOK_IDENT {
				return nil, fmt.Errorf("line %d, col %d: expected parameter identifier", p.cur.Line, p.cur.Col)
			}
			params = append(params, p.cur.Lex)
			p.next()
			if p.cur.Type == TOK_COMMA {
				p.next()
			} else {
				break
			}
		}
		if err := p.expect(TOK_RPAREN); err != nil {
			return nil, err
		}
		body, err := p.parseBlock()
		if err != nil {
			return nil, err
		}
		return &FnDeclStmt{Name: name, Params: params, Body: body}, nil

	case TOK_RETURN:
		p.next()
		var val Expr
		if p.cur.Type != TOK_RBRACE && p.cur.Type != TOK_EOF {
			var err error
			val, err = p.parseExpr()
			if err != nil {
				return nil, err
			}
		}
		return &ReturnStmt{Value: val}, nil

	case TOK_BREAK:
		p.next()
		return &BreakStmt{}, nil

	case TOK_CONTINUE:
		p.next()
		return &ContinueStmt{}, nil

	case TOK_IF:
		p.next()
		cond, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		thenBlk, err := p.parseBlock()
		if err != nil {
			return nil, err
		}
		var elseBlk *BlockStmt
		if p.cur.Type == TOK_ELSE {
			p.next()
			if p.cur.Type == TOK_IF {
				nestedIf, err := p.parseStmt()
				if err != nil {
					return nil, err
				}
				elseBlk = &BlockStmt{Stmts: []Stmt{nestedIf}}
			} else {
				elseBlk, err = p.parseBlock()
				if err != nil {
					return nil, err
				}
			}
		}
		return &IfStmt{Cond: cond, Then: thenBlk, Else: elseBlk}, nil

	case TOK_WHILE:
		p.next()
		cond, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		body, err := p.parseBlock()
		if err != nil {
			return nil, err
		}
		return &WhileStmt{Cond: cond, Body: body}, nil

	case TOK_FOR:
		p.next()
		var initStmt Stmt
		if p.cur.Type != TOK_COLON {
			s, err := p.parseStmt()
			if err != nil {
				return nil, err
			}
			initStmt = s
		}
		if err := p.expect(TOK_COLON); err != nil {
			return nil, err
		}
		var condExpr Expr
		if p.cur.Type != TOK_COLON {
			c, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			condExpr = c
		}
		if err := p.expect(TOK_COLON); err != nil {
			return nil, err
		}
		var postStmt Stmt
		if p.cur.Type != TOK_LBRACE {
			ps, err := p.parseStmt()
			if err != nil {
				return nil, err
			}
			postStmt = ps
		}
		body, err := p.parseBlock()
		if err != nil {
			return nil, err
		}
		return &ForStmt{Init: initStmt, Cond: condExpr, Post: postStmt, Body: body}, nil

	case TOK_PRINT:
		p.next()
		args, err := p.parseArgList()
		if err != nil {
			return nil, err
		}
		return &PrintStmt{Args: args}, nil

	case TOK_IDENT:
		ident := p.cur.Lex
		p.next()
		switch p.cur.Type {
		case TOK_ASSIGN, TOK_PLUS_ASSIGN, TOK_MINUS_ASSIGN, TOK_STAR_ASSIGN, TOK_SLASH_ASSIGN:
			op := p.cur.Lex
			p.next()
			val, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			return &AssignStmt{Name: ident, Op: op, Value: val}, nil
		case TOK_LBRACKET:
			p.next()
			idx, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			if err := p.expect(TOK_RBRACKET); err != nil {
				return nil, err
			}
			switch p.cur.Type {
			case TOK_ASSIGN, TOK_PLUS_ASSIGN, TOK_MINUS_ASSIGN, TOK_STAR_ASSIGN, TOK_SLASH_ASSIGN:
				op := p.cur.Lex
				p.next()
				val, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				return &IndexAssignStmt{Target: &VarExpr{Name: ident}, Index: idx, Op: op, Value: val}, nil
			default:
				e, err := p.finishPostfix(&IndexExpr{Target: &VarExpr{Name: ident}, Index: idx})
				if err != nil {
					return nil, err
				}
				return &ExprStmt{E: e}, nil
			}
		case TOK_LPAREN:
			call, err := p.finishCall(ident)
			if err != nil {
				return nil, err
			}
			e, err := p.finishPostfix(call)
			if err != nil {
				return nil, err
			}
			return &ExprStmt{E: e}, nil
		default:
			e, err := p.finishBinaryFrom(&VarExpr{Name: ident})
			if err != nil {
				return nil, err
			}
			return &ExprStmt{E: e}, nil
		}

	default:
		e, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		return &ExprStmt{E: e}, nil
	}
}

func (p *Parser) parseExpr() (Expr, error) {
	return p.parseOr()
}

func (p *Parser) finishBinaryFrom(left Expr) (Expr, error) {
	return p.finishOr(left)
}

func (p *Parser) finishOr(left Expr) (Expr, error) {
	for p.cur.Type == TOK_OR {
		op := p.cur.Lex
		p.next()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left, nil
}

func (p *Parser) parseOr() (Expr, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	return p.finishOr(left)
}

func (p *Parser) parseAnd() (Expr, error) {
	left, err := p.parseEq()
	if err != nil {
		return nil, err
	}
	for p.cur.Type == TOK_AND {
		op := p.cur.Lex
		p.next()
		right, err := p.parseEq()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left, nil
}

func (p *Parser) parseEq() (Expr, error) {
	left, err := p.parseRel()
	if err != nil {
		return nil, err
	}
	for p.cur.Type == TOK_EQ || p.cur.Type == TOK_NEQ {
		op := p.cur.Lex
		p.next()
		right, err := p.parseRel()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left, nil
}

func (p *Parser) parseRel() (Expr, error) {
	left, err := p.parseAdd()
	if err != nil {
		return nil, err
	}
	for p.cur.Type == TOK_LT || p.cur.Type == TOK_LTE || p.cur.Type == TOK_GT || p.cur.Type == TOK_GTE {
		op := p.cur.Lex
		p.next()
		right, err := p.parseAdd()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left, nil
}

func (p *Parser) parseAdd() (Expr, error) {
	left, err := p.parseMul()
	if err != nil {
		return nil, err
	}
	for p.cur.Type == TOK_PLUS || p.cur.Type == TOK_MINUS {
		op := p.cur.Lex
		p.next()
		right, err := p.parseMul()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left, nil
}

func (p *Parser) parseMul() (Expr, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for p.cur.Type == TOK_STAR || p.cur.Type == TOK_SLASH || p.cur.Type == TOK_PERCENT {
		op := p.cur.Lex
		p.next()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left, nil
}

func (p *Parser) parseUnary() (Expr, error) {
	if p.cur.Type == TOK_BANG || p.cur.Type == TOK_MINUS {
		op := p.cur.Lex
		p.next()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Op: op, Right: right}, nil
	}
	return p.parsePrimary()
}

func (p *Parser) finishCall(name string) (Expr, error) {
	args, err := p.parseArgList()
	if err != nil {
		return nil, err
	}
	return &CallExpr{Callee: name, Args: args}, nil
}

func (p *Parser) finishPostfix(e Expr) (Expr, error) {
	for p.cur.Type == TOK_LBRACKET {
		p.next()
		idx, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if err := p.expect(TOK_RBRACKET); err != nil {
			return nil, err
		}
		e = &IndexExpr{Target: e, Index: idx}
	}
	return e, nil
}

func (p *Parser) parsePrimary() (Expr, error) {
	switch p.cur.Type {
	case TOK_INT:
		v, err := strconv.ParseInt(p.cur.Lex, 10, 64)
		if err != nil {
			return nil, err
		}
		p.next()
		return &IntExpr{Val: v}, nil

	case TOK_FLOAT:
		v, err := strconv.ParseFloat(p.cur.Lex, 64)
		if err != nil {
			return nil, err
		}
		p.next()
		return &FloatExpr{Val: v}, nil

	case TOK_STRING:
		s := p.cur.Lex
		p.next()
		return &StrExpr{Val: s}, nil

	case TOK_BOOL:
		b := p.cur.Lex == "true"
		p.next()
		return &BoolExpr{Val: b}, nil

	case TOK_LEN, TOK_TYPEOF, TOK_APPEND, TOK_INPUT, TOK_INT_CAST, TOK_FLOAT_CAST, TOK_STR_CAST:
		bName := p.cur.Lex
		p.next()
		args, err := p.parseArgList()
		if err != nil {
			return nil, err
		}
		return p.finishPostfix(&BuiltinCallExpr{Name: bName, Args: args})

	case TOK_IDENT:
		name := p.cur.Lex
		p.next()
		if p.cur.Type == TOK_LPAREN {
			call, err := p.finishCall(name)
			if err != nil {
				return nil, err
			}
			return p.finishPostfix(call)
		}
		return p.finishPostfix(&VarExpr{Name: name})

	case TOK_LBRACKET:
		p.next()

		var elems []Expr
		for p.cur.Type != TOK_RBRACKET && p.cur.Type != TOK_EOF {
			el, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			elems = append(elems, el)
			if p.cur.Type == TOK_COMMA {
				p.next()
			} else {
				break
			}
		}
		if err := p.expect(TOK_RBRACKET); err != nil {
			return nil, err
		}
		return p.finishPostfix(&ArrayExpr{Elements: elems})

	case TOK_LPAREN:
		p.next()
		inner, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if err := p.expect(TOK_RPAREN); err != nil {
			return nil, err
		}
		return p.finishPostfix(inner)

	default:
		return nil, fmt.Errorf("line %d, col %d: unexpected token %q", p.cur.Line, p.cur.Col, p.cur.Lex)
	}
}

type Value interface {
	String() string
	TypeName() string
}

type IntVal struct {
	V int64
}

func (i *IntVal) String() string   { return strconv.FormatInt(i.V, 10) }
func (i *IntVal) TypeName() string { return "int" }

type FloatVal struct {
	V float64
}

func (f *FloatVal) String() string   { return strconv.FormatFloat(f.V, 'g', -1, 64) }
func (f *FloatVal) TypeName() string { return "float" }

type StrVal struct {
	V string
}

func (s *StrVal) String() string   { return s.V }
func (s *StrVal) TypeName() string { return "string" }

type BoolVal struct {
	V bool
}

func (b *BoolVal) String() string {
	if b.V {
		return "true"
	}
	return "false"
}
func (b *BoolVal) TypeName() string { return "bool" }

type ArrayVal struct {
	Elems []Value
}

func (a *ArrayVal) String() string {
	parts := make([]string, len(a.Elems))
	for i, e := range a.Elems {
		if s, ok := e.(*StrVal); ok {
			parts[i] = strconv.Quote(s.V)
		} else {
			parts[i] = e.String()
		}
	}
	return "[" + strings.Join(parts, ", ") + "]"
}
func (a *ArrayVal) TypeName() string { return "array" }

type UserFn struct {
	Params []string
	Body   *BlockStmt
	Parent *Env
}

func (u *UserFn) String() string   { return "<fn>" }
func (u *UserFn) TypeName() string { return "function" }

type returnSignal struct {
	Val Value
}

func (returnSignal) Error() string { return "return" }

type breakSignal struct{}

func (breakSignal) Error() string { return "break" }

type continueSignal struct{}

func (continueSignal) Error() string { return "continue" }

type Env struct {
	vars   map[string]Value
	parent *Env
}

func NewEnv(parent *Env) *Env {
	return &Env{
		vars:   make(map[string]Value),
		parent: parent,
	}
}

func (e *Env) SetLocal(name string, v Value) {
	e.vars[name] = v
}

func (e *Env) Assign(name string, v Value) error {
	if _, ok := e.vars[name]; ok {
		e.vars[name] = v
		return nil
	}
	if e.parent != nil {
		return e.parent.Assign(name, v)
	}
	return fmt.Errorf("undefined variable %q", name)
}

func (e *Env) Get(name string) (Value, error) {
	if v, ok := e.vars[name]; ok {
		return v, nil
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, fmt.Errorf("undefined variable %q", name)
}

type Interpreter struct {
	reader *bufio.Reader
}

func NewInterpreter() *Interpreter {
	return &Interpreter{
		reader: bufio.NewReader(os.Stdin),
	}
}

func isTruthy(v Value) bool {
	switch x := v.(type) {
	case *BoolVal:
		return x.V
	case *IntVal:
		return x.V != 0
	case *FloatVal:
		return x.V != 0
	case *StrVal:
		return x.V != ""
	case *ArrayVal:
		return len(x.Elems) > 0
	default:
		return true
	}
}

func toFloat(v Value) (float64, bool) {
	switch x := v.(type) {
	case *IntVal:
		return float64(x.V), true
	case *FloatVal:
		return x.V, true
	default:
		return 0, false
	}
}

func applyCompound(op string, oldVal Value, rhs Value) (Value, error) {
	if op == "=" {
		return rhs, nil
	}
	baseOp := strings.TrimSuffix(op, "=")
	return evalBinaryOp(baseOp, oldVal, rhs)
}

func (it *Interpreter) EvalBlock(blk *BlockStmt, env *Env) (Value, error) {
	var last Value = &BoolVal{V: false}
	for _, s := range blk.Stmts {
		val, err := it.ExecStmt(s, env)
		if err != nil {
			return nil, err
		}
		if val != nil {
			last = val
		}
	}
	return last, nil
}

func (it *Interpreter) ExecStmt(s Stmt, env *Env) (Value, error) {
	switch st := s.(type) {
	case *LetStmt:
		v, err := it.EvalExpr(st.Value, env)
		if err != nil {
			return nil, err
		}
		env.SetLocal(st.Name, v)
		return v, nil

	case *AssignStmt:
		curVal, err := env.Get(st.Name)
		if err != nil {
			return nil, err
		}
		rhsVal, err := it.EvalExpr(st.Value, env)
		if err != nil {
			return nil, err
		}
		finalVal, err := applyCompound(st.Op, curVal, rhsVal)
		if err != nil {
			return nil, err
		}
		if err := env.Assign(st.Name, finalVal); err != nil {
			return nil, err
		}
		return finalVal, nil

	case *IndexAssignStmt:
		tVal, err := it.EvalExpr(st.Target, env)
		if err != nil {
			return nil, err
		}
		arr, ok := tVal.(*ArrayVal)
		if !ok {
			return nil, fmt.Errorf("index assignment requires an array")
		}
		idxVal, err := it.EvalExpr(st.Index, env)
		if err != nil {
			return nil, err
		}
		idxInt, ok := idxVal.(*IntVal)
		if !ok {
			return nil, fmt.Errorf("array index must be an integer")
		}
		if idxInt.V < 0 || int(idxInt.V) >= len(arr.Elems) {
			return nil, fmt.Errorf("array index %d out of bounds", idxInt.V)
		}
		rhsVal, err := it.EvalExpr(st.Value, env)
		if err != nil {
			return nil, err
		}
		finalVal, err := applyCompound(st.Op, arr.Elems[int(idxInt.V)], rhsVal)
		if err != nil {
			return nil, err
		}
		arr.Elems[int(idxInt.V)] = finalVal
		return finalVal, nil

	case *PrintStmt:
		parts := make([]string, len(st.Args))
		for i, a := range st.Args {
			v, err := it.EvalExpr(a, env)
			if err != nil {
				return nil, err
			}
			parts[i] = v.String()
		}
		fmt.Println(strings.Join(parts, " "))
		return &BoolVal{V: true}, nil

	case *ReturnStmt:
		var v Value = &BoolVal{V: false}
		if st.Value != nil {
			var err error
			v, err = it.EvalExpr(st.Value, env)
			if err != nil {
				return nil, err
			}
		}
		return nil, returnSignal{Val: v}

	case *BreakStmt:
		return nil, breakSignal{}

	case *ContinueStmt:
		return nil, continueSignal{}

	case *BlockStmt:
		return it.EvalBlock(st, NewEnv(env))

	case *IfStmt:
		cond, err := it.EvalExpr(st.Cond, env)
		if err != nil {
			return nil, err
		}
		if isTruthy(cond) {
			return it.EvalBlock(st.Then, NewEnv(env))
		}
		if st.Else != nil {
			return it.EvalBlock(st.Else, NewEnv(env))
		}
		return &BoolVal{V: false}, nil

	case *WhileStmt:
		var last Value = &BoolVal{V: false}
		for {
			cond, err := it.EvalExpr(st.Cond, env)
			if err != nil {
				return nil, err
			}
			if !isTruthy(cond) {
				break
			}
			val, err := it.EvalBlock(st.Body, NewEnv(env))
			if err != nil {
				if _, ok := err.(breakSignal); ok {
					break
				}
				if _, ok := err.(continueSignal); ok {
					continue
				}
				return nil, err
			}
			if val != nil {
				last = val
			}
		}
		return last, nil

	case *ForStmt:
		loopEnv := NewEnv(env)
		if st.Init != nil {
			if _, err := it.ExecStmt(st.Init, loopEnv); err != nil {
				return nil, err
			}
		}
		var last Value = &BoolVal{V: false}
		for {
			if st.Cond != nil {
				condVal, err := it.EvalExpr(st.Cond, loopEnv)
				if err != nil {
					return nil, err
				}
				if !isTruthy(condVal) {
					break
				}
			}
			val, err := it.EvalBlock(st.Body, NewEnv(loopEnv))
			if err != nil {
				if _, ok := err.(breakSignal); ok {
					break
				}
				if _, ok := err.(continueSignal); ok {
					if st.Post != nil {
						if _, pErr := it.ExecStmt(st.Post, loopEnv); pErr != nil {
							return nil, pErr
						}
					}
					continue
				}
				return nil, err
			}
			if val != nil {
				last = val
			}
			if st.Post != nil {
				if _, err := it.ExecStmt(st.Post, loopEnv); err != nil {
					return nil, err
				}
			}
		}
		return last, nil

	case *FnDeclStmt:
		fn := &UserFn{
			Params: st.Params,
			Body:   st.Body,
			Parent: env,
		}
		env.SetLocal(st.Name, fn)
		return fn, nil

	case *ExprStmt:
		return it.EvalExpr(st.E, env)

	default:
		return nil, fmt.Errorf("unknown statement node")
	}
}

func evalBinaryOp(op string, lv Value, rv Value) (Value, error) {
	if op == "+" {
		if _, ok := lv.(*StrVal); ok {
			return &StrVal{V: lv.String() + rv.String()}, nil
		}
		if _, ok := rv.(*StrVal); ok {
			return &StrVal{V: lv.String() + rv.String()}, nil
		}
	}

	if li, okL := lv.(*IntVal); okL {
		if ri, okR := rv.(*IntVal); okR {
			switch op {
			case "+":
				return &IntVal{V: li.V + ri.V}, nil
			case "-":
				return &IntVal{V: li.V - ri.V}, nil
			case "*":
				return &IntVal{V: li.V * ri.V}, nil
			case "/":
				if ri.V == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				return &IntVal{V: li.V / ri.V}, nil
			case "%":
				if ri.V == 0 {
					return nil, fmt.Errorf("modulo by zero")
				}
				return &IntVal{V: li.V % ri.V}, nil
			case "==":
				return &BoolVal{V: li.V == ri.V}, nil
			case "!=":
				return &BoolVal{V: li.V != ri.V}, nil
			case "<":
				return &BoolVal{V: li.V < ri.V}, nil
			case "<=":
				return &BoolVal{V: li.V <= ri.V}, nil
			case ">":
				return &BoolVal{V: li.V > ri.V}, nil
			case ">=":
				return &BoolVal{V: li.V >= ri.V}, nil
			}
		}
	}

	if lf, okL := toFloat(lv); okL {
		if rf, okR := toFloat(rv); okR {
			switch op {
			case "+":
				return &FloatVal{V: lf + rf}, nil
			case "-":
				return &FloatVal{V: lf - rf}, nil
			case "*":
				return &FloatVal{V: lf * rf}, nil
			case "/":
				if rf == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				return &FloatVal{V: lf / rf}, nil
			case "==":
				return &BoolVal{V: lf == rf}, nil
			case "!=":
				return &BoolVal{V: lf != rf}, nil
			case "<":
				return &BoolVal{V: lf < rf}, nil
			case "<=":
				return &BoolVal{V: lf <= rf}, nil
			case ">":
				return &BoolVal{V: lf > rf}, nil
			case ">=":
				return &BoolVal{V: lf >= rf}, nil
			}
		}
	}

	switch op {
	case "==":
		return &BoolVal{V: lv.String() == rv.String()}, nil
	case "!=":
		return &BoolVal{V: lv.String() != rv.String()}, nil
	}

	return nil, fmt.Errorf("unsupported binary operation %q between %s and %s", op, lv.TypeName(), rv.TypeName())
}

func (it *Interpreter) EvalExpr(e Expr, env *Env) (Value, error) {
	switch ex := e.(type) {
	case *IntExpr:
		return &IntVal{V: ex.Val}, nil

	case *FloatExpr:
		return &FloatVal{V: ex.Val}, nil

	case *StrExpr:
		return &StrVal{V: ex.Val}, nil

	case *BoolExpr:
		return &BoolVal{V: ex.Val}, nil

	case *VarExpr:
		return env.Get(ex.Name)

	case *ArrayExpr:
		elems := make([]Value, len(ex.Elements))
		for i, el := range ex.Elements {
			v, err := it.EvalExpr(el, env)
			if err != nil {
				return nil, err
			}
			elems[i] = v
		}
		return &ArrayVal{Elems: elems}, nil

	case *IndexExpr:
		tVal, err := it.EvalExpr(ex.Target, env)
		if err != nil {
			return nil, err
		}
		idxVal, err := it.EvalExpr(ex.Index, env)
		if err != nil {
			return nil, err
		}
		idxInt, ok := idxVal.(*IntVal)
		if !ok {
			return nil, fmt.Errorf("index must be an integer")
		}
		switch t := tVal.(type) {
		case *ArrayVal:
			if idxInt.V < 0 || int(idxInt.V) >= len(t.Elems) {
				return nil, fmt.Errorf("array index %d out of bounds", idxInt.V)
			}
			return t.Elems[int(idxInt.V)], nil
		case *StrVal:
			runes := []rune(t.V)
			if idxInt.V < 0 || int(idxInt.V) >= len(runes) {
				return nil, fmt.Errorf("string index %d out of bounds", idxInt.V)
			}
			return &StrVal{V: string(runes[int(idxInt.V)])}, nil
		default:
			return nil, fmt.Errorf("cannot index value of type %s", tVal.TypeName())
		}

	case *BuiltinCallExpr:
		switch ex.Name {
		case "len":
			if len(ex.Args) != 1 {
				return nil, fmt.Errorf("len expects 1 argument, got %d", len(ex.Args))
			}
			v, err := it.EvalExpr(ex.Args[0], env)
			if err != nil {
				return nil, err
			}
			switch x := v.(type) {
			case *ArrayVal:
				return &IntVal{V: int64(len(x.Elems))}, nil
			case *StrVal:
				return &IntVal{V: int64(len([]rune(x.V)))}, nil
			default:
				return nil, fmt.Errorf("len expects array or string")
			}

		case "typeof":
			if len(ex.Args) != 1 {
				return nil, fmt.Errorf("typeof expects 1 argument, got %d", len(ex.Args))
			}
			v, err := it.EvalExpr(ex.Args[0], env)
			if err != nil {
				return nil, err
			}
			return &StrVal{V: v.TypeName()}, nil

		case "append":
			if len(ex.Args) != 2 {
				return nil, fmt.Errorf("append expects 2 arguments, got %d", len(ex.Args))
			}
			arrVal, err := it.EvalExpr(ex.Args[0], env)
			if err != nil {
				return nil, err
			}
			arr, ok := arrVal.(*ArrayVal)
			if !ok {
				return nil, fmt.Errorf("append first argument must be an array")
			}
			elemVal, err := it.EvalExpr(ex.Args[1], env)
			if err != nil {
				return nil, err
			}
			newElems := make([]Value, len(arr.Elems)+1)
			copy(newElems, arr.Elems)
			newElems[len(arr.Elems)] = elemVal
			return &ArrayVal{Elems: newElems}, nil

		case "input":
			if len(ex.Args) > 1 {
				return nil, fmt.Errorf("input expects at most 1 argument, got %d", len(ex.Args))
			}
			if len(ex.Args) == 1 {
				pv, err := it.EvalExpr(ex.Args[0], env)
				if err != nil {
					return nil, err
				}
				fmt.Print(pv.String())
			}
			line, err := it.reader.ReadString('\n')
			if err != nil && len(line) == 0 {
				return &StrVal{V: ""}, nil
			}
			line = strings.TrimRight(line, "\r\n")
			trimmed := strings.TrimSpace(line)
			if iv, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
				return &IntVal{V: iv}, nil
			}
			if fv, err := strconv.ParseFloat(trimmed, 64); err == nil {
				return &FloatVal{V: fv}, nil
			}
			return &StrVal{V: line}, nil

		case "int":
			if len(ex.Args) != 1 {
				return nil, fmt.Errorf("int expects 1 argument, got %d", len(ex.Args))
			}
			v, err := it.EvalExpr(ex.Args[0], env)
			if err != nil {
				return nil, err
			}
			switch x := v.(type) {
			case *IntVal:
				return x, nil
			case *FloatVal:
				return &IntVal{V: int64(x.V)}, nil
			case *BoolVal:
				if x.V {
					return &IntVal{V: 1}, nil
				}
				return &IntVal{V: 0}, nil
			case *StrVal:
				iv, err := strconv.ParseInt(strings.TrimSpace(x.V), 10, 64)
				if err != nil {
					return nil, fmt.Errorf("cannot convert %q to int", x.V)
				}
				return &IntVal{V: iv}, nil
			default:
				return nil, fmt.Errorf("cannot convert %s to int", v.TypeName())
			}

		case "float":
			if len(ex.Args) != 1 {
				return nil, fmt.Errorf("float expects 1 argument, got %d", len(ex.Args))
			}
			v, err := it.EvalExpr(ex.Args[0], env)
			if err != nil {
				return nil, err
			}
			switch x := v.(type) {
			case *FloatVal:
				return x, nil
			case *IntVal:
				return &FloatVal{V: float64(x.V)}, nil
			case *StrVal:
				fv, err := strconv.ParseFloat(strings.TrimSpace(x.V), 64)
				if err != nil {
					return nil, fmt.Errorf("cannot convert %q to float", x.V)
				}
				return &FloatVal{V: fv}, nil
			default:
				return nil, fmt.Errorf("cannot convert %s to float", v.TypeName())
			}

		case "str":
			if len(ex.Args) != 1 {
				return nil, fmt.Errorf("str expects 1 argument, got %d", len(ex.Args))
			}
			v, err := it.EvalExpr(ex.Args[0], env)
			if err != nil {
				return nil, err
			}
			return &StrVal{V: v.String()}, nil
		}

	case *UnaryExpr:
		rv, err := it.EvalExpr(ex.Right, env)
		if err != nil {
			return nil, err
		}
		switch ex.Op {
		case "!":
			return &BoolVal{V: !isTruthy(rv)}, nil
		case "-":
			if iv, ok := rv.(*IntVal); ok {
				return &IntVal{V: -iv.V}, nil
			}
			if fv, ok := rv.(*FloatVal); ok {
				return &FloatVal{V: -fv.V}, nil
			}
			return nil, fmt.Errorf("unary minus requires number")
		}

	case *BinaryExpr:
		if ex.Op == "&&" {
			lv, err := it.EvalExpr(ex.Left, env)
			if err != nil {
				return nil, err
			}
			if !isTruthy(lv) {
				return &BoolVal{V: false}, nil
			}
			rv, err := it.EvalExpr(ex.Right, env)
			if err != nil {
				return nil, err
			}
			return &BoolVal{V: isTruthy(rv)}, nil
		}
		if ex.Op == "||" {
			lv, err := it.EvalExpr(ex.Left, env)
			if err != nil {
				return nil, err
			}
			if isTruthy(lv) {
				return &BoolVal{V: true}, nil
			}
			rv, err := it.EvalExpr(ex.Right, env)
			if err != nil {
				return nil, err
			}
			return &BoolVal{V: isTruthy(rv)}, nil
		}

		lv, err := it.EvalExpr(ex.Left, env)
		if err != nil {
			return nil, err
		}
		rv, err := it.EvalExpr(ex.Right, env)
		if err != nil {
			return nil, err
		}
		return evalBinaryOp(ex.Op, lv, rv)

	case *CallExpr:
		fnVal, err := env.Get(ex.Callee)
		if err != nil {
			return nil, err
		}
		fn, ok := fnVal.(*UserFn)
		if !ok {
			return nil, fmt.Errorf("%q is not a function", ex.Callee)
		}
		if len(ex.Args) != len(fn.Params) {
			return nil, fmt.Errorf("function %q expects %d args, got %d", ex.Callee, len(fn.Params), len(ex.Args))
		}
		argVals := make([]Value, len(ex.Args))
		for i, a := range ex.Args {
			v, err := it.EvalExpr(a, env)
			if err != nil {
				return nil, err
			}
			argVals[i] = v
		}
		callEnv := NewEnv(fn.Parent)
		for i, pName := range fn.Params {
			callEnv.SetLocal(pName, argVals[i])
		}
		res, err := it.EvalBlock(fn.Body, callEnv)
		if err != nil {
			if ret, ok := err.(returnSignal); ok {
				return ret.Val, nil
			}
			return nil, err
		}
		return res, nil
	}

	return nil, fmt.Errorf("unknown expression node")
}

func (it *Interpreter) RunSource(code string, env *Env) error {
	p := NewParser(code)
	stmts, err := p.ParseProgram()
	if err != nil {
		return err
	}
	for _, s := range stmts {
		_, err := it.ExecStmt(s, env)
		if err != nil {
			if _, ok := err.(returnSignal); ok {
				continue
			}
			return err
		}
	}
	return nil
}

func main() {
	it := NewInterpreter()
	globalEnv := NewEnv(nil)

	if len(os.Args) >= 2 {
		data, err := os.ReadFile(os.Args[1])
		if err != nil {
			fmt.Println("Error reading file:", err)
			os.Exit(1)
		}
		if err := it.RunSource(string(data), globalEnv); err != nil {
			fmt.Println("Runtime Error:", err)
			os.Exit(1)
		}
		return
	}

	fmt.Println("BMM Language Interactive Shell (type 'exit' to quit)")
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("bmm> ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "exit" {
			break
		}
		if line == "" {
			continue
		}
		if err := it.RunSource(line, globalEnv); err != nil {
			fmt.Println("Error:", err)
		}
	}
}

