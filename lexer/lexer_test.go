package lexer

import (
	"slices"
	"testing"
)

func TestGoldActiveLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "if statement",
			input:    "if/*condition*/{continue} else { break }",
			expected: []string{"if", "/*", "condition", "*/", "{", "continue", "}", "else", "{", "break", "}"},
		},
		{
			name:     "let statement",
			input:    "let i=0;",
			expected: []string{"let", "i", "=", "0", ";"},
		},
		{
			name:     "some closure",
			input:    "let _x = { let res = DoSomeComputation(); res.x };",
			expected: []string{"let", "_x", "=", "{", "let", "res", "=", "DoSomeComputation", "(", ")", ";", "res", ".", "x", "}", ";"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			al := New(tc.input)
			var got []string

			for {
				tok, ok := al.Next()
				if !ok {
					break
				}
				got = append(got, tok.Val)
			}

			if !slices.Equal(got, tc.expected) {
				t.Fatalf("Next() mismatch for input %q:\ngot:  %#v\nwant: %#v", tc.input, got, tc.expected)
			}

			// Ensure underlying string is exhausted or only whitespace remains
			if rem := al.Str(); rem != "" && al.Str() != tc.input {
				tok, ok := al.Next()
				if ok {
					t.Errorf("expected no more tokens, but got %q (remaining: %q)", tok, rem)
				}
			}
		})
	}
}

func TestGoldTokenizer(t *testing.T) {
	in := NewInterner()
	sym := func(s string) Symbol {
		return in.Intern(s)
	}

	tests := []struct {
		name     string
		src      string
		expected []Token
	}{{
		name: "main",
		src:  "// This is the main function.\nfn main() {\n\t// Print text to the console.\n\tprintln!(\"Hello World!\");\n}",
		expected: []Token{
			// fn main() {
			NewSymbolToken(KindKeyword, NewSpan(2, 1), SymFn),
			NewSymbolToken(KindIdent, NewSpan(2, 4), sym("main")),
			NewSymbolToken(KindDelimiter, NewSpan(2, 8), SymOpenParen),
			NewSymbolToken(KindDelimiter, NewSpan(2, 9), SymCloseParen),
			NewSymbolToken(KindDelimiter, NewSpan(2, 11), SymOpenBrace),

			// println!("Hello World!");
			NewSymbolToken(KindIdent, NewSpan(4, 2), sym("println")),
			NewSymbolToken(KindPunctuation, NewSpan(4, 9), SymNot),
			NewSymbolToken(KindDelimiter, NewSpan(4, 10), SymOpenParen),
			NewSymbolToken(KindString, NewSpan(4, 11), sym("Hello World!")),
			NewSymbolToken(KindDelimiter, NewSpan(4, 25), SymCloseParen),
			NewSymbolToken(KindPunctuation, NewSpan(4, 26), SymSemi),

			// }
			NewSymbolToken(KindDelimiter, NewSpan(5, 1), SymCloseBrace),
		},
	}, {
		name: "punctuation",
		src:  "=\n==\n!=\n<\n<=\n>\n>=\n+\n-\n*\n/\n%\n&&\n||\n!\n->\n(\n)\n{\n}\n[\n]\n;\n,\n:\n::\n.",
		expected: []Token{
			NewSymbolToken(KindPunctuation, NewSpan(1, 1), SymEq),
			NewSymbolToken(KindPunctuation, NewSpan(2, 1), SymEqEq),
			NewSymbolToken(KindPunctuation, NewSpan(3, 1), SymNotEq),
			NewSymbolToken(KindPunctuation, NewSpan(4, 1), SymLt),
			NewSymbolToken(KindPunctuation, NewSpan(5, 1), SymLtEq),
			NewSymbolToken(KindPunctuation, NewSpan(6, 1), SymGt),
			NewSymbolToken(KindPunctuation, NewSpan(7, 1), SymGtEq),
			NewSymbolToken(KindPunctuation, NewSpan(8, 1), SymPlus),
			NewSymbolToken(KindPunctuation, NewSpan(9, 1), SymMinus),
			NewSymbolToken(KindPunctuation, NewSpan(10, 1), SymStar),
			NewSymbolToken(KindPunctuation, NewSpan(11, 1), SymSlash),
			NewSymbolToken(KindPunctuation, NewSpan(12, 1), SymPercent),
			NewSymbolToken(KindPunctuation, NewSpan(13, 1), SymAndAnd),
			NewSymbolToken(KindPunctuation, NewSpan(14, 1), SymOrOr),
			NewSymbolToken(KindPunctuation, NewSpan(15, 1), SymNot),
			NewSymbolToken(KindPunctuation, NewSpan(16, 1), SymArrow),
			NewSymbolToken(KindDelimiter, NewSpan(17, 1), SymOpenParen),
			NewSymbolToken(KindDelimiter, NewSpan(18, 1), SymCloseParen),
			NewSymbolToken(KindDelimiter, NewSpan(19, 1), SymOpenBrace),
			NewSymbolToken(KindDelimiter, NewSpan(20, 1), SymCloseBrace),
			NewSymbolToken(KindDelimiter, NewSpan(21, 1), SymOpenBracket),
			NewSymbolToken(KindDelimiter, NewSpan(22, 1), SymCloseBracket),
			NewSymbolToken(KindPunctuation, NewSpan(23, 1), SymSemi),
			NewSymbolToken(KindPunctuation, NewSpan(24, 1), SymComma),
			NewSymbolToken(KindPunctuation, NewSpan(25, 1), SymColon),
			NewSymbolToken(KindPunctuation, NewSpan(26, 1), SymColonColon),
			NewSymbolToken(KindPunctuation, NewSpan(27, 1), SymDot),
		},
	}, {
		name: "idents",
		src:  "123foo",
		expected: []Token{
			NewSymbolToken(KindError, NewSpan(1, 1), SymError),
		},
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			tok := NewTokenizer(tc.src, in)

			var got []Token
			for {
				tok, ok := tok.Next()
				if !ok {
					break
				}
				got = append(got, tok)
			}

			if len(got) != len(tc.expected) {
				t.Fatalf("token count mismatch: got %d, want %d", len(got), len(tc.expected))
			}

			for i := range tc.expected {
				if got[i] != tc.expected[i] {
					t.Errorf("token[%d] mismatch:\ngot:  %+v\nwant: %+v", i, got[i], tc.expected[i])
				}
			}
		})
	}
}

func TestLexerArithmeticSnippet(t *testing.T) {
	src := `fn main() {
    // addition
    let sum = 5 + 10;

    // subtraction
    let difference = 95.5 - 4.3;

    // multiplication
    let product = 4 * 30;

    // division
    let quotient = 56.7 / 32.2;
    let truncated = -5 / 3; // Results in -1

    // remainder
    let remainder = 43 % 5;
}`

	in := NewInterner()
	sym := func(s string) Symbol {
		return in.Intern(s)
	}

	expected := []Token{
		// fn main() {
		NewSymbolToken(KindKeyword, NewSpan(1, 1), SymFn),
		NewSymbolToken(KindIdent, NewSpan(1, 4), sym("main")),
		NewSymbolToken(KindDelimiter, NewSpan(1, 8), SymOpenParen),
		NewSymbolToken(KindDelimiter, NewSpan(1, 9), SymCloseParen),
		NewSymbolToken(KindDelimiter, NewSpan(1, 11), SymOpenBrace),

		// let sum = 5 + 10;
		NewSymbolToken(KindKeyword, NewSpan(3, 5), SymLet),
		NewSymbolToken(KindIdent, NewSpan(3, 9), sym("sum")),
		NewSymbolToken(KindPunctuation, NewSpan(3, 13), SymEq),
		NewIntToken(NewSpan(3, 15), 5),
		NewSymbolToken(KindPunctuation, NewSpan(3, 17), SymPlus),
		NewIntToken(NewSpan(3, 19), 10),
		NewSymbolToken(KindPunctuation, NewSpan(3, 21), SymSemi),

		// let difference = 95.5 - 4.3;
		NewSymbolToken(KindKeyword, NewSpan(6, 5), SymLet),
		NewSymbolToken(KindIdent, NewSpan(6, 9), sym("difference")),
		NewSymbolToken(KindPunctuation, NewSpan(6, 20), SymEq),
		NewFloatToken(NewSpan(6, 22), 95.5),
		NewSymbolToken(KindPunctuation, NewSpan(6, 27), SymMinus),
		NewFloatToken(NewSpan(6, 29), 4.3),
		NewSymbolToken(KindPunctuation, NewSpan(6, 32), SymSemi),

		// let product = 4 * 30;
		NewSymbolToken(KindKeyword, NewSpan(9, 5), SymLet),
		NewSymbolToken(KindIdent, NewSpan(9, 9), sym("product")),
		NewSymbolToken(KindPunctuation, NewSpan(9, 17), SymEq),
		NewIntToken(NewSpan(9, 19), 4),
		NewSymbolToken(KindPunctuation, NewSpan(9, 21), SymStar),
		NewIntToken(NewSpan(9, 23), 30),
		NewSymbolToken(KindPunctuation, NewSpan(9, 25), SymSemi),

		// let quotient = 56.7 / 32.2;
		NewSymbolToken(KindKeyword, NewSpan(12, 5), SymLet),
		NewSymbolToken(KindIdent, NewSpan(12, 9), sym("quotient")),
		NewSymbolToken(KindPunctuation, NewSpan(12, 18), SymEq),
		NewFloatToken(NewSpan(12, 20), 56.7),
		NewSymbolToken(KindPunctuation, NewSpan(12, 25), SymSlash),
		NewFloatToken(NewSpan(12, 27), 32.2),
		NewSymbolToken(KindPunctuation, NewSpan(12, 31), SymSemi),

		// let truncated = -5 / 3;
		NewSymbolToken(KindKeyword, NewSpan(13, 5), SymLet),
		NewSymbolToken(KindIdent, NewSpan(13, 9), sym("truncated")),
		NewSymbolToken(KindPunctuation, NewSpan(13, 19), SymEq),
		NewSymbolToken(KindPunctuation, NewSpan(13, 21), SymMinus),
		NewIntToken(NewSpan(13, 22), 5),
		NewSymbolToken(KindPunctuation, NewSpan(13, 24), SymSlash),
		NewIntToken(NewSpan(13, 26), 3),
		NewSymbolToken(KindPunctuation, NewSpan(13, 27), SymSemi),

		// let remainder = 43 % 5;
		NewSymbolToken(KindKeyword, NewSpan(16, 5), SymLet),
		NewSymbolToken(KindIdent, NewSpan(16, 9), sym("remainder")),
		NewSymbolToken(KindPunctuation, NewSpan(16, 19), SymEq),
		NewIntToken(NewSpan(16, 21), 43),
		NewSymbolToken(KindPunctuation, NewSpan(16, 24), SymPercent),
		NewIntToken(NewSpan(16, 26), 5),
		NewSymbolToken(KindPunctuation, NewSpan(16, 27), SymSemi),

		// }
		NewSymbolToken(KindDelimiter, NewSpan(17, 1), SymCloseBrace),
	}

	tok := NewTokenizer(src, in)
	scanner := NewLexer(tok)

	var got []Token
	for {
		tkn, ok := scanner.Next()
		if !ok {
			break
		}
		got = append(got, tkn)
	}

	if len(got) != len(expected) {
		t.Fatalf("token count mismatch: got %d, want %d", len(got), len(expected))
	}

	for i := range expected {
		e := expected[i]
		g := got[i]

		if g.Kind != e.Kind || g.Span != e.Span || g.data != e.data {
			t.Errorf("token[%d] mismatch:\n got:  %s\n want: %s",
				i,
				g.Format(in),
				e.Format(in),
			)
		}
	}
}
