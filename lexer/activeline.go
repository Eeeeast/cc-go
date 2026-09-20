package activeline

import (
	"iter"
	"strings"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

// ActiveLine wraps a string slice and advances through tokens.
type ActiveLine struct {
	s string
}

// New creates a new ActiveLine.
func New(s string) *ActiveLine {
	return &ActiveLine{s: s}
}

// Str returns the remaining unprocessed string.
func (a *ActiveLine) Str() string {
	return a.s
}

// TrimStart removes leading whitespace from the remaining string.
func (a *ActiveLine) TrimStart() {
	a.s = strings.TrimLeftFunc(a.s, unicode.IsSpace)
}

// CharOffset calculates the rune offset of the current position relative to
// base, plus shift (for 1-based offset provide shift 1 by default).
// It returns (offset, true), or (0, false) if a.s does not originate from base.
// Using standard int prevents any integer overflow.
func (a *ActiveLine) CharOffset(shift int, base string) (int, bool) {
	if len(base) == 0 && len(a.s) == 0 {
		return shift, true
	}

	basePtr := uintptr(unsafe.Pointer(unsafe.StringData(base)))
	currPtr := uintptr(unsafe.Pointer(unsafe.StringData(a.s)))

	// Bounds check: verify that a.s points within base's byte buffer
	if currPtr < basePtr || currPtr > basePtr+uintptr(len(base)) {
		return 0, false
	}

	byteOffset := int(currPtr - basePtr)

	// Count runes up to the current offset, then apply shift.
	charCount := utf8.RuneCountInString(base[:byteOffset])
	return charCount + shift, true
}

// isBracket reports whether r is one of the designated delimiters.
func isBracket(r rune) bool {
	return strings.ContainsRune("()[]{}<>", r)
}

// Next extracts and returns the next token, advancing the internal position.
// Returns ("", false) when no tokens remain.
func (a *ActiveLine) Next() (string, bool) {
	a.TrimStart()
	if len(a.s) == 0 {
		return "", false
	}

	first, firstLen := utf8.DecodeRuneInString(a.s)
	isAlnum := unicode.IsLetter(first) || unicode.IsDigit(first)

	matchLen := -1
	byteOffset := 0

	for byteOffset < len(a.s) {
		r, size := utf8.DecodeRuneInString(a.s[byteOffset:])
		currAlnum := unicode.IsLetter(r) || unicode.IsDigit(r)

		if unicode.IsSpace(r) || isBracket(r) || currAlnum != isAlnum {
			matchLen = byteOffset
			break
		}
		byteOffset += size
	}

	// If no breaking boundary was found, consume up to the end of the string.
	if matchLen == -1 {
		matchLen = len(a.s)
	}

	// Guarantee at least the first rune is consumed (matching Rust's .max(first.len_utf8())).
	if matchLen < firstLen {
		matchLen = firstLen
	}

	item := a.s[:matchLen]
	a.s = a.s[matchLen:]
	return item, true
}

// All returns a push iterator compatible with `for tok := range line.All()`.
func (a *ActiveLine) All() iter.Seq[string] {
	return func(yield func(string) bool) {
		for {
			tok, ok := a.Next()
			if !ok || !yield(tok) {
				return
			}
		}
	}
}
