package parser

import (
	"fmt"
	"strconv"

	"compiler/internal/lex"
	"compiler/internal/vm"
)

type builtinSpec struct {
	op    vm.OpCode
	arity int
}

var builtins = map[string]builtinSpec{
	"fmt.Println": {op: vm.OpCallPrintln, arity: 1},
	"fmt.ReadInt": {op: vm.OpCallReadInt, arity: 0},
	"abs":         {op: vm.OpCallAbs, arity: 1},
	"min":         {op: vm.OpCallMin, arity: 2},
	"max":         {op: vm.OpCallMax, arity: 2},
}

type Parser struct {
	lexer *lex.Lexer
	curr  lex.Token
	code  []vm.Instruction
}

func New(lexer *lex.Lexer) (*Parser, error) {
	p := &Parser{lexer: lexer}
	if err := p.advance(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Parser) advance() error {
	tok, err := p.lexer.NextToken()
	if err != nil {
		return err
	}
	p.curr = tok
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

func (p *Parser) emit(op vm.OpCode, arg int64, pos lex.SrcPos) {
	p.code = append(p.code, vm.Instruction{
		Op:  op,
		Arg: arg,
		Pos: pos,
	})
}

func (p *Parser) syntaxError(pos lex.SrcPos, format string, args ...any) error {
	return lex.NewParseError(pos, fmt.Sprintf(format, args...))
}

// program = "func" , "main" , "(" , ")" , "{" , { statement } , "}" , ? конец файла ? ;
func (p *Parser) ParseProgram() ([]vm.Instruction, error) {
	if _, err := p.expect(lex.TokenKwFunc); err != nil {
		return nil, err
	}

	nameTok, err := p.expect(lex.TokenIdent)
	if err != nil {
		return nil, err
	}
	if nameTok.Lexeme != "main" {
		return nil, p.syntaxError(nameTok.Pos, "expected entry point 'main', got %q", nameTok.Lexeme)
	}

	if _, err := p.expect(lex.TokenLParen); err != nil {
		return nil, err
	}
	if _, err := p.expect(lex.TokenRParen); err != nil {
		return nil, err
	}
	if _, err := p.expect(lex.TokenLBrace); err != nil {
		return nil, err
	}

	for p.curr.Type != lex.TokenRBrace && p.curr.Type != lex.TokenEOF {
		if err := p.ParseStatement(); err != nil {
			return nil, err
		}
	}

	if _, err := p.expect(lex.TokenRBrace); err != nil {
		return nil, err
	}
	if _, err := p.expect(lex.TokenEOF); err != nil {
		return nil, err
	}

	return p.code, nil
}

// statement = expression , ";" ;
func (p *Parser) ParseStatement() error {
	if err := p.ParseExpression(); err != nil {
		return err
	}
	semiTok, err := p.expect(lex.TokenSemicolon)
	if err != nil {
		return err
	}
	p.emit(vm.OpPop, 0, semiTok.Pos)
	return nil
}

// expression = additive_expression ;
func (p *Parser) ParseExpression() error {
	return p.ParseAdditiveExpression()
}

// additive_expression = multiplicative_expression , { ( "+" | "-" ) , multiplicative_expression } ;
func (p *Parser) ParseAdditiveExpression() error {
	if err := p.ParseMultiplicativeExpression(); err != nil {
		return err
	}

	for p.curr.Type == lex.TokenPlus || p.curr.Type == lex.TokenMinus {
		opTok := p.curr
		if err := p.advance(); err != nil {
			return err
		}
		if err := p.ParseMultiplicativeExpression(); err != nil {
			return err
		}
		if opTok.Type == lex.TokenPlus {
			p.emit(vm.OpAdd, 0, opTok.Pos)
		} else {
			p.emit(vm.OpSub, 0, opTok.Pos)
		}
	}

	return nil
}

// multiplicative_expression = unary_expression , { ( "*" | "/" | "%" ) , unary_expression } ;
func (p *Parser) ParseMultiplicativeExpression() error {
	if err := p.ParseUnaryExpression(); err != nil {
		return err
	}

	for p.curr.Type == lex.TokenStar || p.curr.Type == lex.TokenSlash || p.curr.Type == lex.TokenPercent {
		opTok := p.curr
		if err := p.advance(); err != nil {
			return err
		}
		if err := p.ParseUnaryExpression(); err != nil {
			return err
		}
		switch opTok.Type {
		case lex.TokenStar:
			p.emit(vm.OpMul, 0, opTok.Pos)
		case lex.TokenSlash:
			p.emit(vm.OpDiv, 0, opTok.Pos)
		case lex.TokenPercent:
			p.emit(vm.OpMod, 0, opTok.Pos)
		}
	}

	return nil
}

// unary_expression = "-" , unary_expression | primary_expression ;
func (p *Parser) ParseUnaryExpression() error {
	if p.curr.Type == lex.TokenMinus {
		opTok := p.curr
		if err := p.advance(); err != nil {
			return err
		}
		if err := p.ParseUnaryExpression(); err != nil {
			return err
		}
		p.emit(vm.OpNeg, 0, opTok.Pos)
		return nil
	}
	return p.ParsePrimaryExpression()
}

// primary_expression = integer_literal | call_expression | "(" , expression , ")" ;
func (p *Parser) ParsePrimaryExpression() error {
	switch p.curr.Type {
	case lex.TokenInt:
		tok := p.curr
		val, err := strconv.ParseInt(tok.Lexeme, 10, 64)
		if err != nil {
			return p.syntaxError(tok.Pos, "integer literal out of int64 range: %s", tok.Lexeme)
		}
		if err := p.advance(); err != nil {
			return err
		}
		p.emit(vm.OpPush, val, tok.Pos)
		return nil

	case lex.TokenIdent:
		return p.ParseCallExpression()

	case lex.TokenLParen:
		if err := p.advance(); err != nil {
			return err
		}
		if err := p.ParseExpression(); err != nil {
			return err
		}
		if _, err := p.expect(lex.TokenRParen); err != nil {
			return err
		}
		return nil

	default:
		return p.syntaxError(p.curr.Pos, "unexpected token %s in expression", formatToken(p.curr))
	}
}

// call_expression = identifier , [ "." , identifier ] , "(" , [ argument_list ] , ")" ;
func (p *Parser) ParseCallExpression() error {
	first, err := p.expect(lex.TokenIdent)
	if err != nil {
		return err
	}

	funcName := first.Lexeme
	if p.curr.Type == lex.TokenDot {
		if err := p.advance(); err != nil {
			return err
		}
		second, err := p.expect(lex.TokenIdent)
		if err != nil {
			return err
		}
		funcName += "." + second.Lexeme
	}

	if _, err := p.expect(lex.TokenLParen); err != nil {
		return err
	}

	argCount := 0
	if p.curr.Type != lex.TokenRParen {
		argCount, err = p.ParseArgumentList()
		if err != nil {
			return err
		}
	}

	if _, err := p.expect(lex.TokenRParen); err != nil {
		return err
	}

	spec, ok := builtins[funcName]
	if !ok {
		return p.syntaxError(first.Pos, "unknown function %q", funcName)
	}
	if argCount != spec.arity {
		return p.syntaxError(
			first.Pos,
			"function %q expects %d argument(s), got %d",
			funcName,
			spec.arity,
			argCount,
		)
	}

	p.emit(spec.op, 0, first.Pos)
	return nil
}

// argument_list = expression , { "," , expression } ;
func (p *Parser) ParseArgumentList() (int, error) {
	if err := p.ParseExpression(); err != nil {
		return 0, err
	}

	count := 1
	for p.curr.Type == lex.TokenComma {
		if err := p.advance(); err != nil {
			return 0, err
		}
		if err := p.ParseExpression(); err != nil {
			return 0, err
		}
		count++
	}

	return count, nil
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
