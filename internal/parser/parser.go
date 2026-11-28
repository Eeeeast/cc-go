package parser

import (
	"fmt"
	"strconv"

	"compiler/internal/ast"
	"compiler/internal/lex"
)

type Parser struct {
	lexer *lex.Lexer
	curr  lex.Token
	peek  lex.Token
}

func New(lexer *lex.Lexer) (*Parser, error) {
	p := &Parser{lexer: lexer}
	tok, err := p.lexer.NextToken()
	if err != nil {
		return nil, err
	}
	p.curr = tok
	if p.curr.Type != lex.TokenEOF {
		next, err := p.lexer.NextToken()
		if err != nil {
			return nil, err
		}
		p.peek = next
	} else {
		p.peek = p.curr
	}
	return p, nil
}

func (p *Parser) advance() error {
	p.curr = p.peek
	if p.curr.Type == lex.TokenEOF {
		return nil
	}
	next, err := p.lexer.NextToken()
	if err != nil {
		return err
	}
	p.peek = next
	return nil
}

func (p *Parser) expect(tt lex.TokenType) (lex.Token, error) {
	if p.curr.Type != tt {
		return lex.Token{}, p.syntaxError(
			p.curr.Pos,
			"expected %s, got %s",
			tokenTypeName(tt),
			formatToken(p.curr),
		)
	}
	tok := p.curr
	if err := p.advance(); err != nil {
		return lex.Token{}, err
	}
	return tok, nil
}

func (p *Parser) expectIntLiteral() (*ast.IntLiteral, error) {
	tok, err := p.expect(lex.TokenInt)
	if err != nil {
		return nil, err
	}
	val, err := strconv.ParseInt(tok.Lexeme, 10, 64)
	if err != nil {
		return nil, p.syntaxError(tok.Pos, "integer literal out of int64 range: %s", tok.Lexeme)
	}
	return &ast.IntLiteral{Value: val, SrcPos: tok.Pos}, nil
}

func (p *Parser) syntaxError(pos lex.SrcPos, format string, args ...any) error {
	return lex.NewParseError(pos, fmt.Sprintf(format, args...))
}

// program = { top_level_decl } , ? конец файла ? ;
func (p *Parser) ParseProgram() (*ast.Program, error) {
	startPos := p.curr.Pos
	var decls []ast.Decl

	for p.curr.Type != lex.TokenEOF {
		decl, err := p.ParseTopLevelDecl()
		if err != nil {
			return nil, err
		}
		decls = append(decls, decl)
	}

	if _, err := p.expect(lex.TokenEOF); err != nil {
		return nil, err
	}

	return &ast.Program{
		Decls:  decls,
		SrcPos: startPos,
	}, nil
}

// top_level_decl = global_var_decl | func_decl ;
func (p *Parser) ParseTopLevelDecl() (ast.Decl, error) {
	switch p.curr.Type {
	case lex.TokenKwVar:
		return p.ParseGlobalVarDecl()
	case lex.TokenKwFunc:
		return p.ParseFuncDecl()
	default:
		return nil, p.syntaxError(p.curr.Pos, "expected 'var' or 'func', got %s", formatToken(p.curr))
	}
}

// global_var_decl = "var" , identifier , type , [ "=" , global_initializer ] , ";" ;
func (p *Parser) ParseGlobalVarDecl() (*ast.VarDecl, error) {
	varTok, err := p.expect(lex.TokenKwVar)
	if err != nil {
		return nil, err
	}

	nameTok, err := p.expect(lex.TokenIdent)
	if err != nil {
		return nil, err
	}

	typeName, err := p.ParseType()
	if err != nil {
		return nil, err
	}

	var initExpr ast.Expr
	if p.curr.Type == lex.TokenAssign {
		if err := p.advance(); err != nil {
			return nil, err
		}
		initExpr, err = p.ParseGlobalInitializer()
		if err != nil {
			return nil, err
		}
	}

	if _, err := p.expect(lex.TokenSemicolon); err != nil {
		return nil, err
	}

	return &ast.VarDecl{
		Name:   nameTok.Lexeme,
		Type:   typeName,
		Init:   initExpr,
		SrcPos: varTok.Pos,
	}, nil
}

// global_initializer = global_additive_expression ;
func (p *Parser) ParseGlobalInitializer() (ast.Expr, error) {
	return p.ParseGlobalAdditiveExpression()
}

// global_additive_expression = global_multiplicative_expression , { ( "+" | "-" ) , global_multiplicative_expression } ;
func (p *Parser) ParseGlobalAdditiveExpression() (ast.Expr, error) {
	left, err := p.ParseGlobalMultiplicativeExpression()
	if err != nil {
		return nil, err
	}

	for p.curr.Type == lex.TokenPlus || p.curr.Type == lex.TokenMinus {
		opTok := p.curr
		if err := p.advance(); err != nil {
			return nil, err
		}
		right, err := p.ParseGlobalMultiplicativeExpression()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Op: opTok.Type, Left: left, Right: right, SrcPos: opTok.Pos}
	}

	return left, nil
}

// global_multiplicative_expression = global_unary_expression , { ( "*" | "/" | "%" ) , global_unary_expression } ;
func (p *Parser) ParseGlobalMultiplicativeExpression() (ast.Expr, error) {
	left, err := p.ParseGlobalUnaryExpression()
	if err != nil {
		return nil, err
	}

	for p.curr.Type == lex.TokenStar || p.curr.Type == lex.TokenSlash || p.curr.Type == lex.TokenPercent {
		opTok := p.curr
		if err := p.advance(); err != nil {
			return nil, err
		}
		right, err := p.ParseGlobalUnaryExpression()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Op: opTok.Type, Left: left, Right: right, SrcPos: opTok.Pos}
	}

	return left, nil
}

// global_unary_expression = "-" , global_unary_expression | global_primary_expression ;
func (p *Parser) ParseGlobalUnaryExpression() (ast.Expr, error) {
	if p.curr.Type == lex.TokenMinus {
		opTok := p.curr
		if err := p.advance(); err != nil {
			return nil, err
		}
		operand, err := p.ParseGlobalUnaryExpression()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpr{Op: opTok.Type, Operand: operand, SrcPos: opTok.Pos}, nil
	}
	return p.ParseGlobalPrimaryExpression()
}

// global_primary_expression = integer_literal | "(" , global_initializer , ")" ;
func (p *Parser) ParseGlobalPrimaryExpression() (ast.Expr, error) {
	switch p.curr.Type {
	case lex.TokenInt:
		return p.expectIntLiteral()

	case lex.TokenLParen:
		if err := p.advance(); err != nil {
			return nil, err
		}
		expr, err := p.ParseGlobalInitializer()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(lex.TokenRParen); err != nil {
			return nil, err
		}
		return expr, nil

	default:
		return nil, p.syntaxError(p.curr.Pos, "unexpected token %s in global initializer", formatToken(p.curr))
	}
}

// func_decl = "func" , identifier , "(" , [ param_list ] , ")" , [ type ] , block ;
func (p *Parser) ParseFuncDecl() (*ast.FuncDecl, error) {
	funcTok, err := p.expect(lex.TokenKwFunc)
	if err != nil {
		return nil, err
	}

	nameTok, err := p.expect(lex.TokenIdent)
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(lex.TokenLParen); err != nil {
		return nil, err
	}

	var params []ast.Param
	if p.curr.Type != lex.TokenRParen {
		params, err = p.ParseParamList()
		if err != nil {
			return nil, err
		}
	}

	if _, err := p.expect(lex.TokenRParen); err != nil {
		return nil, err
	}

	retType := ""
	if p.curr.Type != lex.TokenLBrace {
		retType, err = p.ParseType()
		if err != nil {
			return nil, err
		}
	}

	body, err := p.ParseBlock()
	if err != nil {
		return nil, err
	}

	return &ast.FuncDecl{
		Name:       nameTok.Lexeme,
		Params:     params,
		ReturnType: retType,
		Body:       body,
		SrcPos:     funcTok.Pos,
	}, nil
}

// param_list = param , { "," , param } ;
func (p *Parser) ParseParamList() ([]ast.Param, error) {
	first, err := p.ParseParam()
	if err != nil {
		return nil, err
	}
	params := []ast.Param{first}

	for p.curr.Type == lex.TokenComma {
		if err := p.advance(); err != nil {
			return nil, err
		}
		param, err := p.ParseParam()
		if err != nil {
			return nil, err
		}
		params = append(params, param)
	}
	return params, nil
}

// param = identifier , type ;
func (p *Parser) ParseParam() (ast.Param, error) {
	nameTok, err := p.expect(lex.TokenIdent)
	if err != nil {
		return ast.Param{}, err
	}
	typeName, err := p.ParseType()
	if err != nil {
		return ast.Param{}, err
	}
	return ast.Param{
		Name:   nameTok.Lexeme,
		Type:   typeName,
		SrcPos: nameTok.Pos,
	}, nil
}

// type = "int" ;
func (p *Parser) ParseType() (string, error) {
	tok, err := p.expect(lex.TokenIdent)
	if err != nil {
		return "", err
	}
	if tok.Lexeme != "int" {
		return "", p.syntaxError(tok.Pos, "expected type 'int', got %q", tok.Lexeme)
	}
	return tok.Lexeme, nil
}

// block = "{" , { statement } , "}" ;
func (p *Parser) ParseBlock() (*ast.BlockStmt, error) {
	lbraceTok, err := p.expect(lex.TokenLBrace)
	if err != nil {
		return nil, err
	}

	var stmts []ast.Stmt
	for p.curr.Type != lex.TokenRBrace && p.curr.Type != lex.TokenEOF {
		stmt, err := p.ParseStatement()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)
	}

	if _, err := p.expect(lex.TokenRBrace); err != nil {
		return nil, err
	}

	return &ast.BlockStmt{
		Stmts:  stmts,
		SrcPos: lbraceTok.Pos,
	}, nil
}

// statement = local_var_decl | return_statement | block | assignment_statement | expression_statement ;
func (p *Parser) ParseStatement() (ast.Stmt, error) {
	switch p.curr.Type {
	case lex.TokenKwVar:
		return p.ParseLocalVarDecl()
	case lex.TokenKwReturn:
		return p.ParseReturnStatement()
	case lex.TokenLBrace:
		return p.ParseBlock()
	case lex.TokenIdent:
		if p.peek.Type == lex.TokenAssign {
			return p.ParseAssignmentStatement()
		}
	}
	return p.ParseExpressionStatement()
}

// local_var_decl = "var" , identifier , type , [ "=" , expression ] , ";" ;
func (p *Parser) ParseLocalVarDecl() (*ast.VarDecl, error) {
	varTok, err := p.expect(lex.TokenKwVar)
	if err != nil {
		return nil, err
	}

	nameTok, err := p.expect(lex.TokenIdent)
	if err != nil {
		return nil, err
	}

	typeName, err := p.ParseType()
	if err != nil {
		return nil, err
	}

	var initExpr ast.Expr
	if p.curr.Type == lex.TokenAssign {
		if err := p.advance(); err != nil {
			return nil, err
		}
		initExpr, err = p.ParseExpression()
		if err != nil {
			return nil, err
		}
	}

	if _, err := p.expect(lex.TokenSemicolon); err != nil {
		return nil, err
	}

	return &ast.VarDecl{
		Name:   nameTok.Lexeme,
		Type:   typeName,
		Init:   initExpr,
		SrcPos: varTok.Pos,
	}, nil
}

// return_statement = "return" , [ expression ] , ";" ;
func (p *Parser) ParseReturnStatement() (*ast.ReturnStmt, error) {
	retTok, err := p.expect(lex.TokenKwReturn)
	if err != nil {
		return nil, err
	}

	var val ast.Expr
	if p.curr.Type != lex.TokenSemicolon {
		val, err = p.ParseExpression()
		if err != nil {
			return nil, err
		}
	}

	if _, err := p.expect(lex.TokenSemicolon); err != nil {
		return nil, err
	}

	return &ast.ReturnStmt{
		Value:  val,
		SrcPos: retTok.Pos,
	}, nil
}

// assignment_statement = identifier , "=" , expression , ";" ;
func (p *Parser) ParseAssignmentStatement() (*ast.AssignStmt, error) {
	nameTok, err := p.expect(lex.TokenIdent)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lex.TokenAssign); err != nil {
		return nil, err
	}
	val, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lex.TokenSemicolon); err != nil {
		return nil, err
	}
	return &ast.AssignStmt{
		Name:   nameTok.Lexeme,
		Value:  val,
		SrcPos: nameTok.Pos,
	}, nil
}

// expression_statement = expression , ";" ;
func (p *Parser) ParseExpressionStatement() (*ast.ExprStmt, error) {
	expr, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}
	semiTok, err := p.expect(lex.TokenSemicolon)
	if err != nil {
		return nil, err
	}
	return &ast.ExprStmt{
		Expr:    expr,
		SrcPos:  expr.Pos(),
		SemiPos: semiTok.Pos,
	}, nil
}

// expression = additive_expression ;
func (p *Parser) ParseExpression() (ast.Expr, error) {
	return p.ParseAdditiveExpression()
}

// additive_expression = multiplicative_expression , { ( "+" | "-" ) , multiplicative_expression } ;
func (p *Parser) ParseAdditiveExpression() (ast.Expr, error) {
	left, err := p.ParseMultiplicativeExpression()
	if err != nil {
		return nil, err
	}

	for p.curr.Type == lex.TokenPlus || p.curr.Type == lex.TokenMinus {
		opTok := p.curr
		if err := p.advance(); err != nil {
			return nil, err
		}
		right, err := p.ParseMultiplicativeExpression()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Op: opTok.Type, Left: left, Right: right, SrcPos: opTok.Pos}
	}

	return left, nil
}

// multiplicative_expression = unary_expression , { ( "*" | "/" | "%" ) , unary_expression } ;
func (p *Parser) ParseMultiplicativeExpression() (ast.Expr, error) {
	left, err := p.ParseUnaryExpression()
	if err != nil {
		return nil, err
	}

	for p.curr.Type == lex.TokenStar || p.curr.Type == lex.TokenSlash || p.curr.Type == lex.TokenPercent {
		opTok := p.curr
		if err := p.advance(); err != nil {
			return nil, err
		}
		right, err := p.ParseUnaryExpression()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Op: opTok.Type, Left: left, Right: right, SrcPos: opTok.Pos}
	}

	return left, nil
}

// unary_expression = "-" , unary_expression | primary_expression ;
func (p *Parser) ParseUnaryExpression() (ast.Expr, error) {
	if p.curr.Type == lex.TokenMinus {
		opTok := p.curr
		if err := p.advance(); err != nil {
			return nil, err
		}
		operand, err := p.ParseUnaryExpression()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpr{Op: opTok.Type, Operand: operand, SrcPos: opTok.Pos}, nil
	}
	return p.ParsePrimaryExpression()
}

// primary_expression = integer_literal | identifier , [ call_suffix ] | "(" , expression , ")" ;
func (p *Parser) ParsePrimaryExpression() (ast.Expr, error) {
	switch p.curr.Type {
	case lex.TokenInt:
		return p.expectIntLiteral()

	case lex.TokenIdent:
		identTok, err := p.expect(lex.TokenIdent)
		if err != nil {
			return nil, err
		}
		if p.curr.Type == lex.TokenDot || p.curr.Type == lex.TokenLParen {
			return p.ParseCallSuffix(identTok)
		}
		return &ast.IdentExpr{Name: identTok.Lexeme, SrcPos: identTok.Pos}, nil

	case lex.TokenLParen:
		if err := p.advance(); err != nil {
			return nil, err
		}
		expr, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(lex.TokenRParen); err != nil {
			return nil, err
		}
		return expr, nil

	default:
		return nil, p.syntaxError(p.curr.Pos, "unexpected token %s in expression", formatToken(p.curr))
	}
}

// call_suffix = [ "." , identifier ] , "(" , [ argument_list ] , ")" ;
func (p *Parser) ParseCallSuffix(first lex.Token) (*ast.CallExpr, error) {
	funcName := first.Lexeme
	if p.curr.Type == lex.TokenDot {
		if err := p.advance(); err != nil {
			return nil, err
		}
		second, err := p.expect(lex.TokenIdent)
		if err != nil {
			return nil, err
		}
		funcName += "." + second.Lexeme
	}

	if _, err := p.expect(lex.TokenLParen); err != nil {
		return nil, err
	}

	var args []ast.Expr
	if p.curr.Type != lex.TokenRParen {
		var err error
		args, err = p.ParseArgumentList()
		if err != nil {
			return nil, err
		}
	}

	if _, err := p.expect(lex.TokenRParen); err != nil {
		return nil, err
	}

	return &ast.CallExpr{
		Callee: funcName,
		Args:   args,
		SrcPos: first.Pos,
	}, nil
}

// argument_list = expression , { "," , expression } ;
func (p *Parser) ParseArgumentList() ([]ast.Expr, error) {
	first, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}

	args := []ast.Expr{first}
	for p.curr.Type == lex.TokenComma {
		if err := p.advance(); err != nil {
			return nil, err
		}
		arg, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}

	return args, nil
}

func formatToken(tok lex.Token) string {
	if tok.Type == lex.TokenEOF {
		return "EOF"
	}
	return fmt.Sprintf("%q", tok.Lexeme)
}

func tokenTypeName(tt lex.TokenType) string {
	switch tt {
	case lex.TokenEOF:
		return "EOF"
	case lex.TokenIdent:
		return "identifier"
	case lex.TokenInt:
		return "integer literal"
	case lex.TokenKwFunc:
		return "'func'"
	case lex.TokenKwVar:
		return "'var'"
	case lex.TokenKwReturn:
		return "'return'"
	case lex.TokenAssign:
		return "'='"
	case lex.TokenLParen:
		return "'('"
	case lex.TokenRParen:
		return "')'"
	case lex.TokenLBrace:
		return "'{'"
	case lex.TokenRBrace:
		return "'}'"
	case lex.TokenSemicolon:
		return "';'"
	case lex.TokenComma:
		return "','"
	case lex.TokenDot:
		return "'.'"
	default:
		return fmt.Sprintf("token(%d)", tt)
	}
}
