package lexer

import (
	"fmt"
	"math"
)

// Unified token
type TokenKind uint8

const (
	KindError TokenKind = iota
	KindInt
	KindFloat
	KindBool
	KindString
	KindIdent
	KindDelimiter
	KindPunctuation
	KindKeyword
)

type Token struct {
	Kind TokenKind
	Span Span
	data uint64 // i64, f64, bool or ID
}

func NewIntToken(span Span, val int64) Token {
	return Token{Kind: KindInt, Span: span, data: uint64(val)}
}

func NewFloatToken(span Span, val float64) Token {
	return Token{Kind: KindFloat, Span: span, data: math.Float64bits(val)}
}

func NewBoolToken(span Span, val bool) Token {
	var d uint64
	if val {
		d = 1
	}
	return Token{Kind: KindBool, Span: span, data: d}
}

func NewSymbolToken(kind TokenKind, span Span, sym Symbol) Token {
	return Token{Kind: kind, Span: span, data: uint64(sym)}
}

func (t Token) AsInt() (int64, error) {
	if t.Kind != KindInt {
		return 0, ErrTypeMismatch
	}
	return int64(t.data), nil
}

func (t Token) AsFloat() (float64, error) {
	if t.Kind != KindFloat {
		return 0, ErrTypeMismatch
	}
	return math.Float64frombits(t.data), nil
}

func (t Token) AsBool() (bool, error) {
	if t.Kind != KindBool {
		return false, ErrTypeMismatch
	}
	return t.data == 1, nil
}

func (t Token) AsID() (Symbol, error) {
	switch t.Kind {
	case KindIdent, KindString, KindKeyword, KindDelimiter, KindPunctuation:
		return Symbol(t.data), nil
	default:
		return SymError, ErrTypeMismatch
	}
}

// Format returns a textual representation of the token with names from the interning pool
func (t Token) Format(in *Interner) string {
	switch t.Kind {
	case KindInt:
		val, _ := t.AsInt()
		return fmt.Sprintf("(Int, %d, %d:%d)", val, t.Span.Line, t.Span.Column)
	case KindFloat:
		val, _ := t.AsFloat()
		return fmt.Sprintf("(Float, %f, %d:%d)", val, t.Span.Line, t.Span.Column)
	case KindBool:
		val, _ := t.AsBool()
		return fmt.Sprintf("(Bool, %t, %d:%d)", val, t.Span.Line, t.Span.Column)
	case KindString:
		sym, _ := t.AsID()
		return fmt.Sprintf("(String, %q, %d:%d)", in.Resolve(sym), t.Span.Line, t.Span.Column)
	case KindIdent:
		sym, _ := t.AsID()
		return fmt.Sprintf("(Ident, %q, %d:%d)", in.Resolve(sym), t.Span.Line, t.Span.Column)
	case KindKeyword:
		sym, _ := t.AsID()
		return fmt.Sprintf("(Keyword, %s, %d:%d)", in.Resolve(sym), t.Span.Line, t.Span.Column)
	case KindDelimiter:
		sym, _ := t.AsID()
		return fmt.Sprintf("(Delimiter, %s, %d:%d)", in.Resolve(sym), t.Span.Line, t.Span.Column)
	case KindPunctuation:
		sym, _ := t.AsID()
		return fmt.Sprintf("(Punctuation, %s, %d:%d)", in.Resolve(sym), t.Span.Line, t.Span.Column)
	default:
		return fmt.Sprintf("(Error, %d:%d)", t.Span.Line, t.Span.Column)
	}
}
