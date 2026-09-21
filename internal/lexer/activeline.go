package lexer

import (
	"strings"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

type ChunkKind uint8

const (
	ChunkGeneric ChunkKind = iota // any else char chunk
	ChunkString                   // string literal ("..." or r#"..."#)
)

type Chunk struct {
	Kind ChunkKind
	Val  string
}

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
	return strings.ContainsRune("()[]{}", r)
}

// isPunctuation reports whether r is one of operators, separators, and other parts.
func isPunctuation(r rune) bool {
	return strings.ContainsRune("#\"", r)
}

// consumeStringLiteral checks if s starts with a regular string literal:
//
//	"...", #"..."#, r##"..."##, etc.
//
// or a raw string literal:
//
//	r"...", r#"..."#, r##"..."##, etc.
//
// Returns the byte length of the complete literal, or 0 if it does not match
// and decoded content.
func consumeStringLiteral(s string) (int, string) {
	if len(s) == 0 {
		return 0, ""
	}

	idx := 0
	isRaw := false

	// Check optional leading 'r'
	if s[0] == 'r' {
		isRaw = true
		idx++
	}

	// Count leading '#' characters
	numHashes := 0
	for idx < len(s) && s[idx] == '#' {
		numHashes++
		idx++
	}

	// Must be followed immediately by '"'
	if idx >= len(s) || s[idx] != '"' {
		return 0, ""
	}

	// If it had a leading 'r', but no hashes and no quote (handled above),
	// ensure that an identifier like 'result' is not mistakenly matched.
	// A raw string MUST have quotes: r"..." or r#"..."#
	if isRaw && s[idx] != '"' {
		return 0, ""
	}

	// Regular string ("...") has no 'r' and no '#'
	if !isRaw && numHashes > 0 {
		return 0, ""
	}

	// Content after opening '"'
	contentStart := idx + 1
	idx++

	// For standard strings (non-raw), handle escape sequences like \"
	if !isRaw {
		var builder strings.Builder
		for idx < len(s) {
			if s[idx] == '\\' {
				if idx+1 >= len(s) {
					builder.WriteByte('\\')
					return len(s), builder.String()
				}

				nextChar := s[idx+1]
				switch nextChar {
				case 'n':
					builder.WriteByte('\n')
				case 't':
					builder.WriteByte('\t')
				case 'r':
					builder.WriteByte('\r')
				case '\\':
					builder.WriteByte('\\')
				case '"':
					builder.WriteByte('"')
				case '\'':
					builder.WriteByte('\'')
				case '0':
					builder.WriteByte(0)
				default:
					builder.WriteByte('\\')
					builder.WriteByte(nextChar)
				}
				idx += 2
				continue
			}

			if s[idx] == '"' {
				totalLen := idx + 1
				return totalLen, builder.String()
			}

			builder.WriteByte(s[idx])
			idx++
		}
		// Unterminated string: consume the remainder of the line/string
		return len(s), s[contentStart:]
	}

	// For raw strings (r"..." or r#"..."#), content is literal until the closing delimiter
	closingDelimiter := `"` + strings.Repeat("#", numHashes)
	closeRelIdx := strings.Index(s[contentStart:], closingDelimiter)
	if closeRelIdx == -1 {
		// Unterminated raw string: consume until end of line
		return len(s), s[contentStart:]
	}

	contentEnd := contentStart + closeRelIdx
	totalLen := contentEnd + len(closingDelimiter)
	return totalLen, s[contentStart:contentEnd]
}

// Next extracts and returns the next token, advancing the internal position.
// Returns ("", false) when no tokens remain.
func (a *ActiveLine) Next() (Chunk, bool) {
	a.TrimStart()
	if len(a.s) == 0 {
		return Chunk{}, false
	}

	// Attempt to match a string literal first (raw or standard)
	if totalLen, content := consumeStringLiteral(a.s); totalLen > 0 {
		a.s = a.s[totalLen:]
		return Chunk{Kind: ChunkString, Val: content}, true
	}

	// Standard tokenization fallback
	first, firstLen := utf8.DecodeRuneInString(a.s)
	isAlnum := unicode.IsLetter(first) || unicode.IsDigit(first) || first == '_'

	matchLen := -1
	byteOffset := 0

	for byteOffset < len(a.s) {
		r, size := utf8.DecodeRuneInString(a.s[byteOffset:])
		currAlnum := unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'

		if unicode.IsSpace(r) || isBracket(r) || isPunctuation(r) || currAlnum != isAlnum {
			matchLen = byteOffset
			break
		}
		byteOffset += size
	}

	// If no breaking boundary was found, consume up to the end of the string.
	if matchLen == -1 {
		matchLen = len(a.s)
	}

	// Guarantee at least the first rune is consumed.
	if matchLen < firstLen {
		matchLen = firstLen
	}

	item := a.s[:matchLen]
	a.s = a.s[matchLen:]
	return Chunk{Kind: ChunkGeneric, Val: item}, true
}
