package lexer

// Scaner wraps Lexer to support arbitrary lookahead (peeking N tokens ahead).
type Scaner struct {
	inner *Tokenizer
	buf   []Token
}

// NewScaner creates a new Scaner wrapping the given Tokenizer.
func NewScaner(inner *Tokenizer) *Scaner {
	return &Scaner{
		inner: inner,
		buf:   make([]Token, 0, 4),
	}
}

// fill ensures at least count tokens are buffered, if available.
func (p *Scaner) fill(count int) {
	for len(p.buf) < count {
		tok, ok := p.inner.Next()
		if !ok {
			break
		}
		p.buf = append(p.buf, tok)
	}
}

// Peek returns the very next token without advancing the cursor.
// Returns (Token{}, false) if EOF is reached.
func (p *Scaner) Peek() (Token, bool) {
	return p.PeekN(0)
}

// PeekN returns the token at lookahead offset n (0 is the next token, 1 is the one after, etc.)
// without consuming any tokens.
// Returns (Token{}, false) if offset n is at or past EOF.
func (p *Scaner) PeekN(n int) (Token, bool) {
	if n < 0 {
		return Token{}, false
	}
	p.fill(n + 1)
	if n < len(p.buf) {
		return p.buf[n], true
	}
	return Token{}, false
}

// Next consumes and returns the next token.
// Returns (Token{}, false) if EOF is reached.
func (p *Scaner) Next() (Token, bool) {
	if len(p.buf) > 0 {
		tok := p.buf[0]
		p.buf = p.buf[1:]
		return tok, true
	}
	return p.inner.Next()
}

// Consume advances the cursor by n tokens, discarding them.
func (p *Scaner) Consume(n int) {
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

// Take tokens if match occurs and returns true.
func (p *Scaner) TakeIf(expected ...string) bool {
	for i, exp := range expected {
		tok, ok := p.PeekN(i)
		if !ok || tok.Val != exp {
			return false
		}
	}
	p.Consume(len(expected))
	return true
}

// Comment delegates to the underlying Lexer to retrieve comments.
func (p *Scaner) Comment() string {
	return p.inner.Comment()
}

// TakeComment delegates to the underlying Lexer.
func (p *Scaner) TakeComment() string {
	return p.inner.TakeComment()
}
