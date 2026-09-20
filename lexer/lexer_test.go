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
				got = append(got, tok)
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
	tests := []struct {
		name     string
		src      string
		expected []Token
	}{{
		name: "main",
		src:  "// This is the main function.\nfn main() {\n\t// Print text to the console.\n\tprintln!(\"Hello World!\");\n}",
		expected: []Token{
			{Span: NewSpan(2, 1), Val: "fn"},
			{Span: NewSpan(2, 4), Val: "main"},
			{Span: NewSpan(2, 8), Val: "("},
			{Span: NewSpan(2, 9), Val: ")"},
			{Span: NewSpan(2, 11), Val: "{"},
			{Span: NewSpan(4, 2), Val: "println"},
			{Span: NewSpan(4, 9), Val: "!"},
			{Span: NewSpan(4, 10), Val: "("},
			{Span: NewSpan(4, 11), Val: "\"Hello World!\""},
			{Span: NewSpan(4, 25), Val: ")"},
			{Span: NewSpan(4, 26), Val: ";"},
			{Span: NewSpan(5, 1), Val: "}"},
		},
	}, {
		name: "punctuation",
		src:  "=\n==\n!=\n<\n<=\n>\n>=\n+\n-\n*\n/\n%\n&&\n||\n!\n->\n(\n)\n{\n}\n[\n]\n;\n,\n:\n::\n.",
		expected: []Token{
			{Span: NewSpan(1, 1), Val: "="},
			{Span: NewSpan(2, 1), Val: "=="},
			{Span: NewSpan(3, 1), Val: "!="},
			{Span: NewSpan(4, 1), Val: "<"},
			{Span: NewSpan(5, 1), Val: "<="},
			{Span: NewSpan(6, 1), Val: ">"},
			{Span: NewSpan(7, 1), Val: ">="},
			{Span: NewSpan(8, 1), Val: "+"},
			{Span: NewSpan(9, 1), Val: "-"},
			{Span: NewSpan(10, 1), Val: "*"},
			{Span: NewSpan(11, 1), Val: "/"},
			{Span: NewSpan(12, 1), Val: "%"},
			{Span: NewSpan(13, 1), Val: "&&"},
			{Span: NewSpan(14, 1), Val: "||"},
			{Span: NewSpan(15, 1), Val: "!"},
			{Span: NewSpan(16, 1), Val: "->"},
			{Span: NewSpan(17, 1), Val: "("},
			{Span: NewSpan(18, 1), Val: ")"},
			{Span: NewSpan(19, 1), Val: "{"},
			{Span: NewSpan(20, 1), Val: "}"},
			{Span: NewSpan(21, 1), Val: "["},
			{Span: NewSpan(22, 1), Val: "]"},
			{Span: NewSpan(23, 1), Val: ";"},
			{Span: NewSpan(24, 1), Val: ","},
			{Span: NewSpan(25, 1), Val: ":"},
			{Span: NewSpan(26, 1), Val: "::"},
			{Span: NewSpan(27, 1), Val: "."},
		},
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lex := NewTokenizer(tc.src)

			var got []Token
			for {
				tok, ok := lex.Next()
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
