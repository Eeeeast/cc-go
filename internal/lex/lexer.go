package lex

import (
	"fmt"
	"math/big"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)

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

func IsIdentStart(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func IsIdentCh(ch rune) bool {
	return IsIdentStart(ch) || (ch >= '0' && ch <= '9')
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

func (l *Lexer) EOF() bool {
	return l.idx >= len(l.input)
}

func (l *Lexer) PeekCh() rune {
	if l.idx >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.idx:])
	return r
}

func (l *Lexer) PeekChAt(charOffset int) rune {
	bytePos := l.idx
	for range charOffset {
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

func (l *Lexer) EatCh() rune {
	if l.idx >= len(l.input) {
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

func (l *Lexer) MatchChar(ch rune) bool {
	if l.PeekCh() == ch {
		l.EatCh()
		return true
	}
	return false
}

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

func (l *Lexer) MatchChars(chars []rune) bool {
	if !l.PeekChars(chars) {
		return false
	}
	for range chars {
		l.EatCh()
	}
	return true
}

func (l *Lexer) PeekStr(s string) bool {
	return strings.HasPrefix(l.input[l.idx:], s)
}

func (l *Lexer) EatComment() {
	for !l.EOF() {
		if l.EatCh() == '\n' {
			break
		}
	}
}

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

func (l *Lexer) NextToken() (Token, error) {
	for {
		if err := l.EatWS(); err != nil {
			return Token{}, err
		}

		pos := l.GetPos()
		if l.EOF() {
			return Token{
				Type:   TokenEOF,
				Lexeme: "",
				Pos:    pos,
			}, nil
		}

		ch := l.PeekCh()

		// Ключевые слова (>>, >|), логические литералы (><) и идентификаторы (><)
		if IsIdentStart(ch) {
			startIdx := l.idx
			state := daRoot
			inTrie := true

			for !l.EOF() && IsIdentCh(l.PeekCh()) {
				r := l.EatCh()
				if inTrie {
					if next, ok := defaultTrie.Step(state, r); ok {
						state = next
					} else {
						inTrie = false
					}
				}
			}

			lexeme := l.input[startIdx:l.idx]
			if inTrie && defaultTrie.output[state] != TokenNone {
				return Token{
					Type:   defaultTrie.output[state],
					Lexeme: lexeme,
					Pos:    pos,
				}, nil
			}
			return Token{
				Type:   TokenIdent,
				Lexeme: lexeme,
				Pos:    pos,
			}, nil
		}

		// Числовые литералы (><)
		if ch >= '0' && ch <= '9' {
			lexeme, err := l.ReadNumeric()
			if err != nil {
				return Token{}, err
			}
			return Token{
				Type:   TokenInt,
				Lexeme: lexeme,
				Pos:    pos,
			}, nil
		}

		// Строковый литерал (><)
		if ch == '"' || ch == '\'' {
			strVal, err := l.ParseStr(ch)
			if err != nil {
				return Token{}, err
			}
			return Token{
				Type:   TokenString,
				Lexeme: strVal,
				Pos:    pos,
			}, nil
		}

		// Операторы (<>), разделители (>|) и комментарии.
		startIdx := l.idx
		state := daRoot
		for !l.EOF() {
			next, ok := defaultTrie.Step(state, l.PeekCh())
			if !ok {
				break
			}
			l.EatCh()
			state = next
		}

		if state == daRoot || defaultTrie.output[state] == TokenNone {
			return Token{}, l.ParseError(fmt.Sprintf("unexpected character %q", ch))
		}

		tt := defaultTrie.output[state]
		if tt == TokenCommentLine {
			l.EatComment()
			continue
		}
		if tt == TokenCommentBlock {
			if err := l.EatMultiComment(); err != nil {
				return Token{}, err
			}
			continue
		}

		return Token{
			Type:   tt,
			Lexeme: l.input[startIdx:l.idx],
			Pos:    pos,
		}, nil
	}
}

// Функция TokenizeAll считывает все токены, включая TokenEOF
func (l *Lexer) TokenizeAll() ([]Token, error) {
	var tokens []Token
	for {
		tok, err := l.NextToken()
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, tok)
		if tok.Type == TokenEOF {
			return tokens, nil
		}
	}
}

// Сопоставляет строку во входных данных, игнорируя предшествующие пробелы
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

// Считываеь ключевое слово во входных данных, игнорируя предшествующие пробелы
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

// Парсит целочисленное число
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
				return nil, l.ParseError("'_' must separate successive digits")
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

// Парсит строковый литерал
func (l *Lexer) ParseStr(endCh rune) (string, error) {
	if !l.MatchChar(endCh) {
		return "", l.ParseError(fmt.Sprintf("expected opening quote %q", endCh))
	}

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

// Парсит целочисленный литерал в системах счисления 2 (0b), 8 (0o), 16 (0x) или 10
func (l *Lexer) ReadNumeric() (string, error) {
	radix := 10
	if l.PeekCh() == '0' {
		p1 := l.PeekChAt(1)
		switch p1 {
		case 'b', 'B':
			radix = 2
		case 'o', 'O':
			radix = 8
		case 'x', 'X':
			radix = 16
		}
	}

	if radix != 10 {
		l.EatCh()
		l.EatCh()
	}

	val, err := l.ParseInt(radix)
	if err != nil {
		return "", err
	}
	return val.String(), nil
}

// Парсит буквенно-цифровой идентификатор
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
	var d int
	switch {
	case ch >= '0' && ch <= '9':
		d = int(ch - '0')
	case ch >= 'a' && ch <= 'z':
		d = int(ch-'a') + 10
	case ch >= 'A' && ch <= 'Z':
		d = int(ch-'A') + 10
	default:
		return -1
	}
	if d >= radix {
		return -1
	}
	return d
}
