package lex

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"
)

type ParseError struct {
	Pos SrcPos
	Msg string
}

func (e *ParseError) Error() string {
	if e.Pos.SrcName == "" {
		return fmt.Sprintf("%s", e.Msg)
	}
	if e.Pos.LineNo == 0 {
		return fmt.Sprintf("%s: %s", e.Pos.SrcName, e.Msg)
	}
	if e.Pos.ColNo == 0 {
		return fmt.Sprintf("%s:%d: %s", e.Pos.SrcName, e.Pos.LineNo, e.Msg)
	}
	return fmt.Sprintf("%s:%d:%d: %s", e.Pos.SrcName, e.Pos.LineNo, e.Pos.ColNo, e.Msg)
}

func NewParseError(pos SrcPos, msg string) *ParseError {
	return &ParseError{
		Pos: pos,
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
		return nil, NewParseError(SrcPos{}, fmt.Sprintf("unreadable input file %q", fileName))
	}
	return NewLexer(string(data), fileName), nil
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

func (l *Lexer) EatCh() rune {
	if l.EOF() {
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

func (l *Lexer) EatComment() {
	for !l.EOF() && l.EatCh() != '\n' {
	}
}

func (l *Lexer) EatMultiComment() error {
	for !l.EOF() {
		if l.MatchChars([]rune{'*', '/'}) {
			return nil
		}
		l.EatCh()
	}
	return NewParseError(l.GetPos(), "unexpected end of input inside multi-line comment")
}

func (l *Lexer) EatWS() error {
	for !l.EOF() {
		switch {
		case l.MatchChars([]rune{'/', '/'}):
			l.EatComment()
		case l.MatchChars([]rune{'/', '*'}):
			if err := l.EatMultiComment(); err != nil {
				return err
			}
		case strings.ContainsRune(" \t\n\r", l.PeekCh()):
			l.EatCh()
		default:
			return nil
		}
	}
	return nil
}

func (l *Lexer) NextToken() (Token, error) {
	if err := l.EatWS(); err != nil {
		return Token{}, err
	}
	pos := l.GetPos()
	if l.EOF() {
		return Token{Type: TokenEOF, Pos: pos}, nil
	}

	saved := *l
	startIdx := l.idx
	ch := l.PeekCh()

	state := daRoot
	for !l.EOF() {
		next, ok := defaultTrie.Step(state, l.PeekCh())
		if !ok {
			break
		}
		l.EatCh()
		state = next
	}

	if IsIdentStart(ch) {
		if (!l.EOF() && IsIdentCh(l.PeekCh())) || defaultTrie.output[state] == TokenNone {
			for !l.EOF() && IsIdentCh(l.PeekCh()) {
				l.EatCh()
			}
			return Token{Type: TokenIdent, Lexeme: l.input[startIdx:l.idx], Pos: pos}, nil
		}
		return Token{Type: defaultTrie.output[state], Lexeme: l.input[startIdx:l.idx], Pos: pos}, nil
	}

	if tt := defaultTrie.output[state]; tt != TokenNone {
		return Token{Type: tt, Lexeme: l.input[startIdx:l.idx], Pos: pos}, nil
	}

	if ch >= '0' && ch <= '9' {
		lexeme, err := l.ReadNumeric()
		return Token{Type: TokenInt, Lexeme: lexeme, Pos: pos}, err
	}
	if ch == '"' {
		strVal, err := l.ReadString()
		return Token{Type: TokenString, Lexeme: strVal, Pos: pos}, err
	}

	*l = saved
	return Token{}, NewParseError(pos, fmt.Sprintf("unexpected character %q", ch))
}

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

func (l *Lexer) ReadString() (string, error) {
	startIdx := l.idx

	pos := l.GetPos()
	if !l.MatchChar('"') {
		return "", NewParseError(pos, `expected opening quote '"'`)
	}

	for {
		pos = l.GetPos()

		if l.EOF() {
			return "", NewParseError(pos, "unexpected end of input inside string literal")
		}

		switch l.PeekCh() {
		case '\n':
			return "", NewParseError(pos, "unexpected new line inside string literal")
		case '"':
			l.EatCh()
			return l.input[startIdx:l.idx], nil
		case '\\':
			l.EatCh()
			if l.EOF() {
				return "", NewParseError(pos, "unexpected end of input after '\\'")
			}

			esc := l.EatCh()
			switch esc {
			case '"', '\\', 'n', 't':
			default:
				return "", NewParseError(pos, fmt.Sprintf("unknown escape sequence '\\%c'", esc))
			}
		default:
			l.EatCh()
		}
	}
}

func (l *Lexer) ReadNumeric() (string, error) {
	startIdx := l.idx
	for !l.EOF() && l.PeekCh() >= '0' && l.PeekCh() <= '9' {
		l.EatCh()
	}
	return l.input[startIdx:l.idx], nil
}
