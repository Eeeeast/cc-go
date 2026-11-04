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

func tok(t TokenType, lex string, line, col int) Token {
	return Token{
		Type:   t,
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
			tok(TokenKwTrue, "true", 1, 1),
			tok(TokenKwFalse, "false", 1, 6),
		}},

		// регистр ключевых слов важен
		{"IF", []Token{tok(TokenIdent, "IF", 1, 1)}},
		{"True", []Token{tok(TokenIdent, "True", 1, 1)}},
		{"Return", []Token{tok(TokenIdent, "Return", 1, 1)}},

		// идентификаторы
		{"name", []Token{tok(TokenIdent, "name", 1, 1)}},
		{"foo_bar", []Token{tok(TokenIdent, "foo_bar", 1, 1)}},
		{"_hidden", []Token{tok(TokenIdent, "_hidden", 1, 1)}},
		{"x1", []Token{tok(TokenIdent, "x1", 1, 1)}},
		{"ifx", []Token{tok(TokenIdent, "ifx", 1, 1)}},
		{"int", []Token{tok(TokenIdent, "int", 1, 1)}},
		{"breakfast", []Token{tok(TokenIdent, "breakfast", 1, 1)}},

		// целочисленные литералы
		{"0 42 1000000", []Token{
			tok(TokenInt, "0", 1, 1),
			tok(TokenInt, "42", 1, 3),
			tok(TokenInt, "1000000", 1, 6),
		}},

		// строковые литералы, Lexeme содержит исходный текст
		{`"hello"`, []Token{tok(TokenString, `"hello"`, 1, 1)}},
		{`""`, []Token{tok(TokenString, `""`, 1, 1)}},
		{`"a\nb"`, []Token{tok(TokenString, `"a\nb"`, 1, 1)}},
		{`"a\tb"`, []Token{tok(TokenString, `"a\tb"`, 1, 1)}},
		{`"a\"b"`, []Token{tok(TokenString, `"a\"b"`, 1, 1)}},
		{`"a\\b"`, []Token{tok(TokenString, `"a\\b"`, 1, 1)}},
		{`"я"`, []Token{tok(TokenString, `"я"`, 1, 1)}},
		{`"я" x`, []Token{
			tok(TokenString, `"я"`, 1, 1),
			tok(TokenIdent, "x", 1, 5),
		}},

		// арифметика
		{"+ - * / %", []Token{
			tok(TokenPlus, "+", 1, 1),
			tok(TokenMinus, "-", 1, 3),
			tok(TokenStar, "*", 1, 5),
			tok(TokenSlash, "/", 1, 7),
			tok(TokenPercent, "%", 1, 9),
		}},

		// знаки перед числами
		{"-5", []Token{
			tok(TokenMinus, "-", 1, 1),
			tok(TokenInt, "5", 1, 2),
		}},
		{"+5", []Token{
			tok(TokenPlus, "+", 1, 1),
			tok(TokenInt, "5", 1, 2),
		}},
		{"-42", []Token{
			tok(TokenMinus, "-", 1, 1),
			tok(TokenInt, "42", 1, 2),
		}},

		// сочетание знаков подряд
		{"+-", []Token{
			tok(TokenPlus, "+", 1, 1),
			tok(TokenMinus, "-", 1, 2),
		}},
		{"-+", []Token{
			tok(TokenMinus, "-", 1, 1),
			tok(TokenPlus, "+", 1, 2),
		}},
		{"+-5", []Token{
			tok(TokenPlus, "+", 1, 1),
			tok(TokenMinus, "-", 1, 2),
			tok(TokenInt, "5", 1, 3),
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
		{"/* a */ /* b */ x", []Token{tok(TokenIdent, "x", 1, 17)}},

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

func TestLexerErrors(t *testing.T) {
	cases := []struct {
		name   string
		source string
	}{
		{"решётка", "#"},
		{"одиночный амперсанд", "&"},
		{"одиночная черта", "|"},
		{"одиночное двоеточие", ":"},
		{"одинарные кавычки", "'hi'"},
		{"незакрытый блочный комментарий", "/* never"},
		{"незакрытая строка", `"abc`},
		{"неизвестный escape", `"a\qb"`},
		{"escape hex", `"a\x41"`},
		{"escape unicode", `"a\u0041"`},
		{"перенос строки в строке", "\"a\nb\""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			lx := NewLexer(c.source, "")
			_, err := lx.TokenizeAll()
			require.Error(t, err)
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
			{SrcPos{"", 0, 0}, "boom"},
		}
		for _, c := range cases {
			require.Equal(t, c.want, NewParseError(c.pos, "boom").Error())
		}
	})

	t.Run("from file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "a.mg")
		require.NoError(t, os.WriteFile(path, []byte("abc"), 0o644))

		lx, err := FromFile(path)
		require.NoError(t, err)
		require.Equal(t, path, lx.SrcName)

		_, err = FromFile(filepath.Join(t.TempDir(), "nope.mg"))
		require.Error(t, err)
	})

	t.Run("read string", func(t *testing.T) {
		lx := NewLexer(`"hello"`, "")
		got, err := lx.ReadString()
		require.NoError(t, err)
		require.Equal(t, `"hello"`, got)
	})

	t.Run("read numeric", func(t *testing.T) {
		lx := NewLexer("123abc", "")
		got, err := lx.ReadNumeric()
		require.NoError(t, err)
		require.Equal(t, "123", got)

		lx = NewLexer("abc", "")
		got, err = lx.ReadNumeric()
		require.NoError(t, err)
		require.Equal(t, "", got)
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

	t.Run("multibyte rune takes one column", func(t *testing.T) {
		lx := NewLexer("яя", "")
		lx.EatCh()
		require.Equal(t, 2, lx.ColNo)
		lx.EatCh()
		require.Equal(t, 3, lx.ColNo)
	})
}

func TestEndOfFileToken(t *testing.T) {
	t.Run("empty source", func(t *testing.T) {
		lx := NewLexer("", "")
		tok, err := lx.NextToken()
		require.NoError(t, err)
		require.Equal(t, TokenEOF, tok.Type)
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
