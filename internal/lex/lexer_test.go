package lex

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

type testCase struct {
	source   string
	expected []Token
}

func (c *testCase) name() string {
	return c.source
}

func pieceFor(t TokenType) PuzzlePiece {
	switch t {
	case TokenIdent, TokenInt, TokenString, TokenBool:
		return PieceOperand
	case TokenKwFor, TokenKwIf, TokenKwElse, TokenKwFunc, TokenKwStruct,
		TokenKwReturn, TokenKwType, TokenKwVar, TokenNot:
		return PieceAction
	case TokenKwBreak, TokenKwContinue, TokenSemicolon,
		TokenLParen, TokenRParen, TokenLBrace, TokenRBrace,
		TokenLBracket, TokenRBracket:
		return PieceTerminator
	}
	return PieceConnector
}

func tok(t TokenType, lex string, line, col int) Token {
	return Token{
		Type:   t,
		Piece:  pieceFor(t),
		Lexeme: lex,
		Pos:    SrcPos{LineNo: line, ColNo: col},
	}
}

func tokenize(source string) []Token {
	lx := NewLexer(source, "")
	tokens, err := lx.TokenizeAll()
	if err != nil {
		panic(err)
	}
	if len(tokens) > 0 && tokens[len(tokens)-1].Type == TokenEOF {
		tokens = tokens[:len(tokens)-1]
	}
	return tokens
}

func TestLexer(t *testing.T) {
	cases := []testCase{
		// ключевые слова
		{"if", []Token{tok(TokenKwIf, "if", 1, 1)}},
		{"for", []Token{tok(TokenKwFor, "for", 1, 1)}},
		{"else", []Token{tok(TokenKwElse, "else", 1, 1)}},
		{"func", []Token{tok(TokenKwFunc, "func", 1, 1)}},
		{"struct", []Token{tok(TokenKwStruct, "struct", 1, 1)}},
		{"return", []Token{tok(TokenKwReturn, "return", 1, 1)}},
		{"type", []Token{tok(TokenKwType, "type", 1, 1)}},
		{"var", []Token{tok(TokenKwVar, "var", 1, 1)}},
		{"break", []Token{tok(TokenKwBreak, "break", 1, 1)}},
		{"continue", []Token{tok(TokenKwContinue, "continue", 1, 1)}},

		// булевы литералы
		{"true false", []Token{
			tok(TokenBool, "true", 1, 1),
			tok(TokenBool, "false", 1, 6),
		}},

		// идентификаторы
		{"name", []Token{tok(TokenIdent, "name", 1, 1)}},
		{"foo_bar", []Token{tok(TokenIdent, "foo_bar", 1, 1)}},
		{"_hidden", []Token{tok(TokenIdent, "_hidden", 1, 1)}},
		{"x1", []Token{tok(TokenIdent, "x1", 1, 1)}},
		{"ifx", []Token{tok(TokenIdent, "ifx", 1, 1)}},
		{"int", []Token{tok(TokenIdent, "int", 1, 1)}},

		// целочисленные литералы
		{"0 42 1000000", []Token{
			tok(TokenInt, "0", 1, 1),
			tok(TokenInt, "42", 1, 3),
			tok(TokenInt, "1000000", 1, 6),
		}},
		{"0b1010", []Token{tok(TokenInt, "10", 1, 1)}},
		{"0B1010", []Token{tok(TokenInt, "10", 1, 1)}},
		{"0o17", []Token{tok(TokenInt, "15", 1, 1)}},
		{"0O17", []Token{tok(TokenInt, "15", 1, 1)}},
		{"0xFF", []Token{tok(TokenInt, "255", 1, 1)}},
		{"1_000", []Token{tok(TokenInt, "1000", 1, 1)}},

		// строковые литералы
		{`"hello"`, []Token{tok(TokenString, "hello", 1, 1)}},
		{`""`, []Token{tok(TokenString, "", 1, 1)}},
		{`"a\nb"`, []Token{tok(TokenString, "a\nb", 1, 1)}},
		{`'hi'`, []Token{tok(TokenString, "hi", 1, 1)}},

		// арифметика
		{"+ - * / %", []Token{
			tok(TokenPlus, "+", 1, 1),
			tok(TokenMinus, "-", 1, 3),
			tok(TokenStar, "*", 1, 5),
			tok(TokenSlash, "/", 1, 7),
			tok(TokenPercent, "%", 1, 9),
		}},

		// сравнение
		{"== != < <= > >=", []Token{
			tok(TokenEq, "==", 1, 1),
			tok(TokenNotEq, "!=", 1, 4),
			tok(TokenLt, "<", 1, 7),
			tok(TokenLtEq, "<=", 1, 9),
			tok(TokenGt, ">", 1, 12),
			tok(TokenGtEq, ">=", 1, 14),
		}},

		// логика
		{"&& || !", []Token{
			tok(TokenAnd, "&&", 1, 1),
			tok(TokenOr, "||", 1, 4),
			tok(TokenNot, "!", 1, 7),
		}},

		// присваивание
		{"=", []Token{tok(TokenAssign, "=", 1, 1)}},
		{":=", []Token{tok(TokenColonEq, ":=", 1, 1)}},

		// пунктуация
		{"( ) { } [ ] ; , .", []Token{
			tok(TokenLParen, "(", 1, 1),
			tok(TokenRParen, ")", 1, 3),
			tok(TokenLBrace, "{", 1, 5),
			tok(TokenRBrace, "}", 1, 7),
			tok(TokenLBracket, "[", 1, 9),
			tok(TokenRBracket, "]", 1, 11),
			tok(TokenSemicolon, ";", 1, 13),
			tok(TokenComma, ",", 1, 15),
			tok(TokenDot, ".", 1, 17),
		}},

		// комментарии
		{"// comment\nx", []Token{tok(TokenIdent, "x", 2, 1)}},
		{"/* comment */ x", []Token{tok(TokenIdent, "x", 1, 15)}},
		{"/* a /* b */ c */ x", []Token{tok(TokenIdent, "x", 1, 19)}},

		// разбиение двойных символов
		{"++", []Token{tok(TokenPlus, "+", 1, 1), tok(TokenPlus, "+", 1, 2)}},
		{"--", []Token{tok(TokenMinus, "-", 1, 1), tok(TokenMinus, "-", 1, 2)}},
		{"<-", []Token{tok(TokenLt, "<", 1, 1), tok(TokenMinus, "-", 1, 2)}},
		{"a < -b", []Token{
			tok(TokenIdent, "a", 1, 1),
			tok(TokenLt, "<", 1, 3),
			tok(TokenMinus, "-", 1, 5),
			tok(TokenIdent, "b", 1, 6),
		}},

		// переводы строк
		{"a\nb", []Token{tok(TokenIdent, "a", 1, 1), tok(TokenIdent, "b", 2, 1)}},
		{"a\rb", []Token{tok(TokenIdent, "a", 1, 1), tok(TokenIdent, "b", 1, 3)}},
		{"a\r\nb", []Token{tok(TokenIdent, "a", 1, 1), tok(TokenIdent, "b", 2, 1)}},

		// пустой и пробельный вход
		{"", []Token{}},
		{"   \t\n  ", []Token{}},
		{"x\n\n\ny", []Token{tok(TokenIdent, "x", 1, 1), tok(TokenIdent, "y", 4, 1)}},

		// программа
		{
			"func main() {\n" +
				"  var x int = 10\n" +
				"  if x >= 0 {\n" +
				"    return x\n" +
				"  }\n" +
				"}",
			[]Token{
				tok(TokenKwFunc, "func", 1, 1),
				tok(TokenIdent, "main", 1, 6),
				tok(TokenLParen, "(", 1, 10),
				tok(TokenRParen, ")", 1, 11),
				tok(TokenLBrace, "{", 1, 13),
				tok(TokenKwVar, "var", 2, 3),
				tok(TokenIdent, "x", 2, 7),
				tok(TokenIdent, "int", 2, 9),
				tok(TokenAssign, "=", 2, 13),
				tok(TokenInt, "10", 2, 15),
				tok(TokenKwIf, "if", 3, 3),
				tok(TokenIdent, "x", 3, 6),
				tok(TokenGtEq, ">=", 3, 8),
				tok(TokenInt, "0", 3, 11),
				tok(TokenLBrace, "{", 3, 13),
				tok(TokenKwReturn, "return", 4, 5),
				tok(TokenIdent, "x", 4, 12),
				tok(TokenRBrace, "}", 5, 3),
				tok(TokenRBrace, "}", 6, 1),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name(), func(t *testing.T) {
			require.Equal(t, c.expected, tokenize(c.source))
		})
	}
}

func TestLexerAPI(t *testing.T) {
	t.Run("error format", func(t *testing.T) {
		cases := []struct {
			pos  SrcPos
			want string
		}{
			{SrcPos{"m", 0, 0}, "m: boom"},
			{SrcPos{"m", 3, 0}, "m:3: boom"},
			{SrcPos{"m", 3, 7}, "m:3:7: boom"},
		}
		for _, c := range cases {
			require.Equal(t, c.want, (&ParseError{Pos: c.pos, Msg: "boom"}).Error())
		}
	})

	t.Run("new parse error", func(t *testing.T) {
		err := NewParseError(NewLexer("abc", "f"), "boom")
		require.Equal(t, "f:1:1: boom", err.Error())
	})

	t.Run("rune to digit", func(t *testing.T) {
		require.Equal(t, 0, runeToDigit('0', 10))
		require.Equal(t, 9, runeToDigit('9', 10))
		require.Equal(t, -1, runeToDigit('a', 10))
		require.Equal(t, 10, runeToDigit('a', 16))
		require.Equal(t, 15, runeToDigit('F', 16))
		require.Equal(t, -1, runeToDigit('g', 16))
		require.Equal(t, -1, runeToDigit('2', 2))
	})

	t.Run("reserved", func(t *testing.T) {
		for _, s := range []string{"if", "for", "func", "return", "var", "break", "continue"} {
			require.True(t, IsReserved(s), s)
		}
		for _, s := range []string{"", "go", "defer", "true", "main", "IF"} {
			require.False(t, IsReserved(s), s)
		}
	})

	t.Run("peek str", func(t *testing.T) {
		lx := NewLexer("hello", "")
		require.True(t, lx.PeekStr("hell"))
		require.True(t, lx.PeekStr(""))
		require.False(t, lx.PeekStr("world"))
	})

	t.Run("match token", func(t *testing.T) {
		lx := NewLexer("abc", "")
		ok, err := lx.MatchToken("abc")
		require.NoError(t, err)
		require.True(t, ok)
		require.True(t, lx.EOF())

		lx = NewLexer("abc", "")
		ok, err = lx.MatchToken("xyz")
		require.NoError(t, err)
		require.False(t, ok)
	})

	t.Run("match keyword", func(t *testing.T) {
		lx := NewLexer("if x", "")
		ok, err := lx.MatchKeyword("if")
		require.NoError(t, err)
		require.True(t, ok)

		lx = NewLexer("ifx", "")
		ok, err = lx.MatchKeyword("if")
		require.NoError(t, err)
		require.False(t, ok)
	})

	t.Run("expect token", func(t *testing.T) {
		lx := NewLexer("{x}", "")
		require.NoError(t, lx.ExpectToken("{"))
		require.Equal(t, 'x', lx.PeekCh())

		lx = NewLexer("(x)", "")
		require.Error(t, lx.ExpectToken("{"))
	})

	t.Run("from file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "a.mg")
		require.NoError(t, os.WriteFile(path, []byte("abc"), 0o644))

		lx, err := FromFile(path)
		require.NoError(t, err)
		require.Equal(t, path, lx.GetSrcName())

		_, err = FromFile(filepath.Join(t.TempDir(), "nope.mg"))
		require.Error(t, err)
	})

	t.Run("parse ident", func(t *testing.T) {
		lx := NewLexer("foo_bar baz", "")
		got, err := lx.ParseIdent()
		require.NoError(t, err)
		require.Equal(t, "foo_bar", got)

		lx = NewLexer("1abc", "")
		_, err = lx.ParseIdent()
		require.Error(t, err)
	})

	t.Run("parse str errors", func(t *testing.T) {
		for _, src := range []string{`"abc`, `"a\qb"`, `"\xZZ"`, `"\x80"`, `"\uZZ"`, `"\uD800"`, `"\u110000"`} {
			lx := NewLexer(src, "")
			_, err := lx.ParseStr('"')
			require.Error(t, err, src)
		}
	})

	t.Run("parse int", func(t *testing.T) {
		lx := NewLexer("1010", "")
		got, err := lx.ParseInt(2)
		require.NoError(t, err)
		require.Equal(t, "10", got.String())

		lx = NewLexer("abc", "")
		_, err = lx.ParseInt(10)
		require.Error(t, err)
	})

	t.Run("read numeric errors", func(t *testing.T) {
		for _, src := range []string{"1_", "1__0", "0x"} {
			lx := NewLexer(src, "")
			_, err := lx.ReadNumeric()
			require.Error(t, err, src)
		}
	})

	t.Run("eat ws error", func(t *testing.T) {
		lx := NewLexer("/* never", "")
		require.Error(t, lx.EatWS())
	})

	t.Run("tokenize error", func(t *testing.T) {
		lx := NewLexer("#", "")
		_, err := lx.TokenizeAll()
		require.Error(t, err)
	})

	t.Run("peek at eof", func(t *testing.T) {
		require.Equal(t, rune(0), NewLexer("", "").PeekCh())
		require.Equal(t, rune(0), NewLexer("", "").PeekChAt(0))
		require.Equal(t, rune(0), NewLexer("a", "").PeekChAt(1))
	})

	t.Run("match char fail", func(t *testing.T) {
		require.False(t, NewLexer("abc", "").MatchChar('z'))
	})

	t.Run("peek chars fail", func(t *testing.T) {
		require.False(t, NewLexer("ab", "").PeekChars([]rune{'a', 'b', 'c'}))
	})

	t.Run("eat at eof", func(t *testing.T) {
		require.Equal(t, rune(0), NewLexer("", "").EatCh())
	})

	t.Run("eat multi comment error", func(t *testing.T) {
		lx := NewLexer("/* never", "")
		lx.MatchChars([]rune{'/', '*'})
		require.Error(t, lx.EatMultiComment())
	})
}

func TestEndOfFileToken(t *testing.T) {
	t.Run("empty source", func(t *testing.T) {
		lx := NewLexer("", "")
		tok, err := lx.NextToken()
		require.NoError(t, err)
		require.Equal(t, TokenEOF, tok.Type)
		require.Equal(t, PieceTerminator, tok.Piece)
		require.Equal(t, "", tok.Lexeme)
		require.Equal(t, 1, tok.Pos.LineNo)
		require.Equal(t, 1, tok.Pos.ColNo)
	})

	t.Run("after identifier", func(t *testing.T) {
		lx := NewLexer("x", "")
		_, err := lx.NextToken()
		require.NoError(t, err)
		tok, err := lx.NextToken()
		require.NoError(t, err)
		require.Equal(t, TokenEOF, tok.Type)
	})
}

func TestTokenizeAll(t *testing.T) {
	lx := NewLexer("x + 1", "")
	tokens, err := lx.TokenizeAll()
	require.NoError(t, err)
	require.Len(t, tokens, 4)
	require.Equal(t, TokenEOF, tokens[3].Type)
}
