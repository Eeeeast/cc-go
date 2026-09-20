package lexer

import (
	"iter"
	"strings"
	"unicode"
)

// Span represents a source code location.
type Span struct {
	Line   int
	Column int
}

// First represents line 1, column 1.
var First = Span{
	Line:   1,
	Column: 1,
}

// NewSpanWithLine creates a Span for the given line, defaulting to column 1.
func NewSpanWithLine(line int) Span {
	return Span{
		Line:   line,
		Column: 1,
	}
}

// NewSpan creates a Span with explicit line and column coordinates.
func NewSpan(line, column int) Span {
	return Span{
		Line:   line,
		Column: column,
	}
}

// Token represents a lexical token with its location span and string value.
type Token struct {
	Span Span
	Val  string
}

// NewToken constructs a Token with its location span and text.
func NewToken(span Span, val string) Token {
	return Token{
		Span: span,
		Val:  val,
	}
}

// activeLineState holds the current line coordinates, accumulated comment,
// the original raw string (base for offset calculations), and the ActiveLine scanner.
type activeLineState struct {
	span    Span
	comment string
	base    string
	rem     ActiveLine
}

// Tokenizer produces tokens from source lines and accumulates adjacent comments.
type Tokenizer struct {
	rawLines    []string
	lineIdx     int
	activeLine  *activeLineState
	lastComment string
}

// NewTokenizer creates a new Tokenizer from input source text.
func NewTokenizer(src string) *Tokenizer {
	return &Tokenizer{
		rawLines: strings.Split(src, "\n"),
		lineIdx:  0,
	}
}

// Next implements the iterator step, equivalent to:
// self.drain_active_line().or_else(|| self.fetch_next_line())
func (l *Tokenizer) Next() (Token, bool) {
	if tok, ok := l.drainActiveLine(); ok {
		return tok, true
	}
	return l.fetchNextLine()
}

// drainActiveLine extracts the next token from the currently active line if available.
func (l *Tokenizer) drainActiveLine() (Token, bool) {
	if l.activeLine == nil {
		return Token{}, false
	}

	// 1. Skip leading spaces
	l.activeLine.rem.TrimStart()
	if l.activeLine.rem.Str() == "" {
		l.activeLine = nil
		return Token{}, false
	}

	// 2. Check for an inline trailing comment on this line
	remStr := l.activeLine.rem.Str()
	if strings.HasPrefix(remStr, "//") {
		commentText := strings.TrimRightFunc(remStr, unicode.IsSpace)
		if l.activeLine.comment == "" {
			l.activeLine.comment = commentText
		} else {
			l.activeLine.comment += "\n" + commentText
		}
		l.lastComment = l.activeLine.comment
		l.activeLine = nil
		return Token{}, false
	}

	// 3. Compute the token's column relative to the original raw line
	col, _ := l.activeLine.rem.CharOffset(1, l.activeLine.base)
	span := NewSpan(l.activeLine.span.Line, col)

	// 4. Extract token value
	val, ok := l.activeLine.rem.Next()
	if !ok {
		l.activeLine = nil
		return Token{}, false
	}

	// Store any comment that preceded this token on or before this line
	l.lastComment = l.activeLine.comment

	return NewToken(span, val), true
}

// fetchNextLine loops through raw lines, aggregating adjacent comment lines
// until it finds a line containing actual code or reaches EOF.
func (l *Tokenizer) fetchNextLine() (Token, bool) {
	var accumulatedComments []string

	for l.lineIdx < len(l.rawLines) {
		raw := l.rawLines[l.lineIdx]
		lineNum := l.lineIdx + 1
		l.lineIdx++

		trimmed := strings.TrimSpace(raw)

		// Empty lines break comment continuity
		if trimmed == "" {
			accumulatedComments = nil
			continue
		}

		// Comment line: accumulate and continue checking adjacent lines
		if strings.HasPrefix(trimmed, "//") {
			accumulatedComments = append(accumulatedComments, trimmed)
			continue
		}

		// Found a code line: join all preceding contiguous comments
		joinedComment := strings.Join(accumulatedComments, "\n")

		l.activeLine = &activeLineState{
			span:    NewSpanWithLine(lineNum),
			comment: joinedComment,
			base:    raw,
			rem:     *New(raw),
		}

		// Try to drain the first token from this newly activated line
		if tok, ok := l.drainActiveLine(); ok {
			return tok, true
		}
	}

	// Trailing comments at the end of the file with no subsequent tokens
	if len(accumulatedComments) > 0 {
		l.lastComment = strings.Join(accumulatedComments, "\n")
	}

	return Token{}, false
}

// Comment returns the last associated comment block without clearing it.
func (l *Tokenizer) Comment() string {
	return l.lastComment
}

// TakeComment returns the accumulated comment block and resets it.
func (l *Tokenizer) TakeComment() string {
	c := l.lastComment
	l.lastComment = ""
	return c
}

// All provides standard `for tok := range lex.All()` support.
func (l *Tokenizer) All() iter.Seq[Token] {
	return func(yield func(Token) bool) {
		for {
			tok, ok := l.Next()
			if !ok || !yield(tok) {
				return
			}
		}
	}
}
