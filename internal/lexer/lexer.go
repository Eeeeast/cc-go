package lexer

import (
	"fmt"
	"strconv"
)

type Lexer struct {
	inner *Tokenizer
	buf   []Token
}

func NewLexer(inner *Tokenizer) *Lexer {
	return &Lexer{
		inner: inner,
		buf:   make([]Token, 0, 4),
	}
}

func (p *Lexer) Interner() *Interner {
	return p.inner.Interner()
}

func (p *Lexer) fill(count int) {
	for len(p.buf) < count {
		tok, ok := p.inner.Next()
		if !ok {
			break
		}
		p.buf = append(p.buf, tok)
	}
}

func (p *Lexer) Peek() (Token, bool) {
	return p.PeekN(0)
}

func (p *Lexer) PeekN(n int) (Token, bool) {
	if n < 0 {
		return Token{}, false
	}
	p.fill(n + 1)
	if n < len(p.buf) {
		return p.buf[n], true
	}
	return Token{}, false
}

func (p *Lexer) Consume(n int) {
	for n > 0 {
		if len(p.buf) > 0 {
			p.buf = p.buf[1:]
			n--
		} else {
			if _, ok := p.inner.Next(); !ok {
				break
			}
			n--
		}
	}
}

// Next concatenates composite tokens (e.g. Float) and returns the final Token
func (p *Lexer) Next() (Token, bool) {
	tok0, ok0 := p.PeekN(0)
	if !ok0 {
		return Token{}, false
	}

	// Check the Float pattern: KindInt + SymDot + KindInt
	if tok0.Kind == KindInt {
		tok1, ok1 := p.PeekN(1)
		tok2, ok2 := p.PeekN(2)

		if ok1 && ok2 && tok1.Kind == KindPunctuation && Symbol(tok1.data) == SymDot && tok2.Kind == KindInt {
			p.Consume(3)

			intPart, _ := tok0.AsInt()
			fracPart, _ := tok2.AsInt()

			fVal, _ := strconv.ParseFloat(fmt.Sprintf("%d.%d", intPart, fracPart), 64)
			return NewFloatToken(tok0.Span, fVal), true
		}
	}

	p.Consume(1)
	return tok0, true
}

func (p *Lexer) Comment() string {
	return p.inner.Comment()
}

func (p *Lexer) TakeComment() string {
	return p.inner.TakeComment()
}
