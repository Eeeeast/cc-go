package lex

import (
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// `LineNo` and `ColNo` 1-based, so 0 mean None
type SrcPos struct {
	SrcName string
	LineNo  int
	ColNo   int
}

func NewSrcPos(srcName string, lineNo int, colNo int) SrcPos {
	return SrcPos{
		SrcName: srcName,
		LineNo:  lineNo,
		ColNo:   colNo,
	}
}

type ParseError struct {
	Pos SrcPos
	Msg string
}

func (e *ParseError) Error() string {
	if e.Pos.LineNo == 0 {
		return fmt.Sprintf("%s: %s", e.Pos.SrcName, e.Msg)
	} else if e.Pos.ColNo == 0 {
		return fmt.Sprintf("%s:%d: %s", e.Pos.SrcName, e.Pos.LineNo, e.Msg)
	} else {
		return fmt.Sprintf("%s:%d:%d: %s", e.Pos.SrcName, e.Pos.LineNo, e.Pos.ColNo, e.Msg)
	}
}

func NewParseError(l *Lexer, msg string) *ParseError {
	return &ParseError{
		Pos: l.GetPos(),
		Msg: msg,
	}
}

// Check if a character can be the start of an identifier
func IsIdentStart(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

// Check if a character can be part of an identifier
func IsIdentCh(ch rune) bool {
	return IsIdentStart(ch) || (ch >= '0' && ch <= '9')
}

// Check if an identifier is a reserved keyword
func IsReserved(ident string) bool {
	switch ident {
	case
		// Currently in use and potentially unusable
		// or confusing if chosen as a name
		"break", "func", "else", "if",
		"continue", "return",

		// Reserved for future use
		"default", "interface", "select", "case",
		"defer", "go", "map", "struct",
		"chan", "goto", "package", "switch",
		"const", "fallthrough", "range", "type",
		"for", "import", "var":
		return true
	default:
		return false
	}
}

type Lexer struct {
	input   string
	idx     int
	SrcName string
	LineNo  int
	ColNo   int
}

func NewLexer(inputStr string, srcName string) *Lexer {
	return &Lexer{
		input:   inputStr,
		idx:     0,
		SrcName: srcName,
		LineNo:  1,
		ColNo:   1,
	}
}

func FromFile(fileName string) (*Lexer, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, &ParseError{
			Pos: SrcPos{},
			Msg: fmt.Sprintf("unreadable input file %q", fileName),
		}
	}
	return NewLexer(string(data), fileName), nil
}

func (l *Lexer) GetSrcName() string {
	return l.SrcName
}

func (l *Lexer) GetPos() SrcPos {
	return NewSrcPos(l.SrcName, l.LineNo, l.ColNo)
}

// Test if the end of the input has been reached
func (l *Lexer) EOF() bool {
	return l.idx >= len(l.input)
}

// Peek at a character from the input
func (l *Lexer) PeekCh() rune {
	if l.idx >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.idx:])
	return r
}

// Peek at a character from the input
func (l *Lexer) PeekChAt(charOffset int) rune {
	bytePos := l.idx
	for i := 0; i < charOffset; i++ {
		if bytePos >= len(l.input) {
			return 0
		}
		_, width := utf8.DecodeRuneInString(l.input[bytePos:])
		bytePos += width
	}
	if bytePos >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[bytePos:])
	return r
}

// Consume a character from the input
func (l *Lexer) EatCh() rune {
	if l.idx >= len(l.input) {
		l.ColNo++
		return 0
	}
	ch, width := utf8.DecodeRuneInString(l.input[l.idx:])
	l.idx += width

	if ch == '\n' {
		l.LineNo++
		l.ColNo = 1
	} else {
		l.ColNo++
	}
	return ch
}

// Match a single character in the input
func (l *Lexer) MatchChar(ch rune) bool {
	if l.PeekCh() == ch {
		l.EatCh()
		return true
	}
	return false
}

// Peek for a sequence of characters
func (l *Lexer) PeekChars(chars []rune) bool {
	bytePos := l.idx
	for _, expected := range chars {
		if bytePos >= len(l.input) {
			return false
		}
		r, width := utf8.DecodeRuneInString(l.input[bytePos:])
		if r != expected {
			return false
		}
		bytePos += width
	}
	return true
}

// Match characters in the input
func (l *Lexer) MatchChars(chars []rune) bool {
	if !l.PeekChars(chars) {
		return false
	}
	for range chars {
		l.EatCh()
	}
	return true
}

// Peek for a string
func (l *Lexer) PeekStr(s string) bool {
	return strings.HasPrefix(l.input[l.idx:], s)
}

// Consume characters until the end of a single-line comment
func (l *Lexer) EatComment() {
	for !l.EOF() {
		if l.EatCh() == '\n' {
			break
		}
	}
}

// Consume characters until the end of a multi-line comment
func (l *Lexer) EatMultiComment() error {
	depth := 1
	for {
		if l.EOF() {
			return l.ParseError("unexpected end of input inside multi-line comment")
		} else if l.MatchChars([]rune{'/', '*'}) {
			depth++
		} else if l.MatchChars([]rune{'*', '/'}) {
			depth--
			if depth == 0 {
				break
			}
		} else {
			l.EatCh()
		}
	}
	return nil
}

// Consume whitespace
func (l *Lexer) EatWS() error {
	for !l.EOF() {
		ch := l.PeekCh()

		if ch == '/' {
			if l.MatchChars([]rune{'/', '/'}) {
				l.EatComment()
				continue
			}
			if l.MatchChars([]rune{'/', '*'}) {
				if err := l.EatMultiComment(); err != nil {
					return err
				}
				continue
			}
		}

		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == '\x0c' {
			l.EatCh()
			continue
		}

		break
	}
	return nil
}

// Match a string in the input, ignoring preceding whitespace
func (l *Lexer) MatchToken(token string) (bool, error) {
	if err := l.EatWS(); err != nil {
		return false, err
	}
	if !l.PeekStr(token) {
		return false, nil
	}
	for range token {
		l.EatCh()
	}
	return true, nil
}

// Match a keyword in the input, ignoring preceding whitespace
// It's different from match_token because following identifier
// chars unacceptable
func (l *Lexer) MatchKeyword(keyword string) (bool, error) {
	if err := l.EatWS(); err != nil {
		return false, err
	}
	if !l.PeekStr(keyword) {
		return false, nil
	}

	endBytePos := l.idx + len(keyword)
	if endBytePos < len(l.input) {
		nextRune, _ := utf8.DecodeRuneInString(l.input[endBytePos:])
		// Can't match as a keyword if the next is identifier chars
		if IsIdentCh(nextRune) {
			return false, nil
		}
	}

	for range keyword {
		l.EatCh()
	}
	return true, nil
}

func (l *Lexer) ParseError(msg string) error {
	return NewParseError(l, msg)
}

func (l *Lexer) ExpectToken(token string) error {
	matched, err := l.MatchToken(token)
	if err != nil {
		return err
	}
	if matched {
		return nil
	}
	return l.ParseError(fmt.Sprintf("expected token %q", token))
}

// Parse an integer value
func (l *Lexer) ParseInt(radix int) (*big.Int, error) {
	if l.EOF() || runeToDigit(l.PeekCh(), radix) < 0 {
		return nil, l.ParseError("expected digit")
	}

	intVal := new(big.Int)
	radixBig := big.NewInt(int64(radix))

	for !l.EOF() {
		ch := l.PeekCh()

		if ch == '_' {
			if runeToDigit(l.PeekChAt(1), radix) < 0 {
				return nil, l.ParseError("only one numeric separator can occur between digits")
			}
			l.EatCh()
			continue
		}

		digit := runeToDigit(ch, radix)
		if digit < 0 {
			break
		}

		intVal.Mul(intVal, radixBig)
		intVal.Add(intVal, big.NewInt(int64(digit)))
		l.EatCh()
	}

	return intVal, nil
}

// Read the characters of a numeric value into a string
func (l *Lexer) ReadNumeric() (string, error) {
	readDigits := func() (bool, error) {
		ch := l.PeekCh()
		if ch < '0' || ch > '9' {
			return false, nil
		}

		for {
			c := l.PeekCh()
			if c == '_' {
				next := l.PeekChAt(1)
				if next < '0' || next > '9' {
					return false, l.ParseError("only one numeric separator can occur between digits")
				}
				l.EatCh()
				continue
			}

			if c < '0' || c > '9' {
				break
			}
			l.EatCh()
		}
		return true, nil
	}

	readSign := func() {
		_ = l.MatchChar('+') || l.MatchChar('-')
	}

	startByteIdx := l.idx

	readSign()

	if _, err := readDigits(); err != nil {
		return "", err
	}

	if l.PeekCh() == '.' && l.PeekChAt(1) == '_' {
		return "", l.ParseError("only one numeric separator can occur between digits")
	}

	if l.PeekCh() == '.' && l.PeekChAt(1) >= '0' && l.PeekChAt(1) <= '9' {
		l.EatCh()
		if _, err := readDigits(); err != nil {
			return "", err
		}
	}

	if l.MatchChar('e') || l.MatchChar('E') {
		readSign()
		hasDigits, err := readDigits()
		if err != nil {
			return "", err
		}
		if !hasDigits {
			return "", l.ParseError("expected digits in floating-point exponent")
		}
	}

	endByteIdx := l.idx
	numStr := l.input[startByteIdx:endByteIdx]
	return strings.ReplaceAll(numStr, "_", ""), nil
}

// Parse a string literal
func (l *Lexer) ParseStr(endCh rune) (string, error) {
	l.EatCh()

	var out strings.Builder

	for {
		if l.EOF() {
			return "", l.ParseError("unexpected end of input while parsing string literal")
		}

		ch := l.EatCh()
		if ch == endCh {
			break
		}

		if ch == '\\' {
			switch l.EatCh() {
			case '\\':
				out.WriteRune('\\')
			case '\'':
				out.WriteRune('\'')
			case '"':
				out.WriteRune('"')
			case 't':
				out.WriteRune('\t')
			case 'r':
				out.WriteRune('\r')
			case 'n':
				out.WriteRune('\n')
			case '0':
				out.WriteRune(0)

			// Hexadecimal escape sequence, limited to ASCII since
			// strings are UTF-8 and a single byte above 0x7F is invalid
			case 'x':
				d0 := runeToDigit(l.EatCh(), 16)
				d1 := runeToDigit(l.EatCh(), 16)
				if d0 < 0 || d1 < 0 {
					return "", l.ParseError("invalid hexadecimal escape sequence")
				}
				byteVal := (d0 << 4) + d1
				if byteVal > 0x7F {
					return "", l.ParseError("hexadecimal escape sequence must be in the ASCII range \\x00-\\x7F")
				}
				out.WriteByte(byte(byteVal))

			// Unicode escape sequence with 1 to 6 hex digits, e.g. \u1F600
			case 'u':
				var codepoint int64
				numDigits := 0

				for {
					d := runeToDigit(l.PeekCh(), 16)
					if d < 0 {
						break
					}
					if numDigits == 6 {
						return "", l.ParseError("unicode escape sequence has more than 6 hex digits")
					}
					l.EatCh()
					codepoint = (codepoint << 4) + int64(d)
					numDigits++
				}

				if numDigits == 0 {
					return "", l.ParseError("expected hex digits in unicode escape sequence")
				}

				if !utf8.ValidRune(rune(codepoint)) || codepoint > unicode.MaxRune {
					return "", l.ParseError("invalid unicode codepoint in escape sequence")
				}
				out.WriteRune(rune(codepoint))
			default:
				return "", l.ParseError("unknown escape sequence")
			}
			continue
		}

		out.WriteRune(ch)
	}

	return out.String(), nil
}

// Parse a C-style alphanumeric identifier
func (l *Lexer) ParseIdent() (string, error) {
	if l.EOF() || !IsIdentStart(l.PeekCh()) {
		return "", l.ParseError("expected identifier")
	}

	startByteIdx := l.idx
	for !l.EOF() {
		if !IsIdentCh(l.PeekCh()) {
			break
		}
		l.EatCh()
	}

	return l.input[startByteIdx:l.idx], nil
}

func runeToDigit(ch rune, radix int) int {
	d, err := strconv.ParseInt(string(ch), radix, 64)
	if err != nil {
		return -1
	}
	return int(d)
}
