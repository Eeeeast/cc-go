package activeline

import (
	"slices"
	"testing"
)

func TestGold(t *testing.T) {
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
			input:    "let x = { let res = DoSomeComputation(); res.x };",
			expected: []string{"let", "x", "=", "{", "let", "res", "=", "DoSomeComputation", "(", ")", ";", "res", ".", "x", "}", ";"},
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
