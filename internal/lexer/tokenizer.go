package lexer

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Span represents a source code location.
type Span struct {
	Line   int
	Column int
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

type activeLineState struct {
	span    Span
	comment string
	base    string
	rem     ActiveLine
}

type Tokenizer struct {
	rawLines    []string
	lineIdx     int
	activeLine  *activeLineState
	lastComment string
	interner    *Interner
}

func NewTokenizer(src string, in *Interner) *Tokenizer {
	if in == nil {
		in = NewInterner()
	}
	return &Tokenizer{
		rawLines: strings.Split(src, "\n"),
		lineIdx:  0,
		interner: in,
	}
}

func (l *Tokenizer) Interner() *Interner {
	return l.interner
}

func isAllDigits(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func isValidIdent(s string) bool {
	if len(s) == 0 {
		return false
	}
	first, _ := utf8.DecodeRuneInString(s)
	if unicode.IsDigit(first) {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
	}
	return true
}

func (l *Tokenizer) classifyGeneric(span Span, raw string) Token {
	// Integer
	if isAllDigits(raw) {
		val, _ := strconv.ParseInt(raw, 10, 64)
		return NewIntToken(span, val)
	}

	// Keywords, Delimiters, Punctuation
	if sym, exists := l.interner.lookup[raw]; exists {
		if sym == SymTrue {
			return NewBoolToken(span, true)
		}
		if sym == SymFalse {
			return NewBoolToken(span, false)
		}
		if sym >= SymOpenParen && sym <= SymCloseBracket {
			return NewSymbolToken(KindDelimiter, span, sym)
		}
		if sym >= SymEq && sym <= SymDot {
			return NewSymbolToken(KindPunctuation, span, sym)
		}
		if sym >= SymUnderscore && sym < firstDynamicSymbol {
			return NewSymbolToken(KindKeyword, span, sym)
		}
	}

	// Identificators
	if isValidIdent(raw) {
		sym := l.interner.Intern(raw)
		return NewSymbolToken(KindIdent, span, sym)
	}

	// Unrecognized
	return Token{Kind: KindError, Span: span}
}

func (l *Tokenizer) Next() (Token, bool) {
	if tok, ok := l.drainActiveLine(); ok {
		return tok, true
	}
	return l.fetchNextLine()
}

func (l *Tokenizer) drainActiveLine() (Token, bool) {
	if l.activeLine == nil {
		return Token{}, false
	}

	// Whitespace trimming
	l.activeLine.rem.TrimStart()
	if l.activeLine.rem.Str() == "" {
		l.activeLine = nil
		return Token{}, false
	}

	// Comment
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

	// Span
	col, _ := l.activeLine.rem.CharOffset(1, l.activeLine.base)
	span := NewSpan(l.activeLine.span.Line, col)

	// Next chunk
	chunk, ok := l.activeLine.rem.Next()
	if !ok {
		l.activeLine = nil
		return Token{}, false
	}

	l.lastComment = l.activeLine.comment

	// If already string
	if chunk.Kind == ChunkString {
		sym := l.interner.Intern(chunk.Val)
		return NewSymbolToken(KindString, span, sym), true
	}

	// Others (nums, keywords, punctuation, identificators)
	return l.classifyGeneric(span, chunk.Val), true
}

func (l *Tokenizer) fetchNextLine() (Token, bool) {
	var accumulatedComments []string

	for l.lineIdx < len(l.rawLines) {
		raw := l.rawLines[l.lineIdx]
		lineNum := l.lineIdx + 1
		l.lineIdx++

		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			accumulatedComments = nil
			continue
		}

		if strings.HasPrefix(trimmed, "//") {
			accumulatedComments = append(accumulatedComments, trimmed)
			continue
		}

		joinedComment := strings.Join(accumulatedComments, "\n")
		l.activeLine = &activeLineState{
			span:    NewSpanWithLine(lineNum),
			comment: joinedComment,
			base:    raw,
			rem:     *New(raw),
		}

		if tok, ok := l.drainActiveLine(); ok {
			return tok, true
		}
	}

	if len(accumulatedComments) > 0 {
		l.lastComment = strings.Join(accumulatedComments, "\n")
	}

	return Token{}, false
}

func (l *Tokenizer) Comment() string {
	return l.lastComment
}

func (l *Tokenizer) TakeComment() string {
	c := l.lastComment
	l.lastComment = ""
	return c
}
