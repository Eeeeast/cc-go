package examplelib

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractWords(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name: "russian words",
			text: `Играют волны — ветер свищет,
И мачта гнётся и скрыпит…
Увы! он счастия не ищет
И не от счастия бежит!`,
			expected: []string{
				"Играют",
				"волны",
				"ветер",
				"свищет",
				"И",
				"мачта",
				"гнётся",
				"и",
				"скрыпит",
				"Увы",
				"он",
				"счастия",
				"не",
				"ищет",
				"И",
				"не",
				"от",
				"счастия",
				"бежит",
			},
		},
		{
			name: "words with hyphens",
			text: "Что-нибудь да как-нибудь, и +/- что- то ещё",
			expected: []string{
				"Что-нибудь",
				"да",
				"как-нибудь",
				"и",
				"что",
				"то",
				"ещё",
			},
		},
		{
			name: "words with apostrophes",
			text: "Children's toys and three cats' toys",
			expected: []string{
				"Children's",
				"toys",
				"and",
				"three",
				"cats'",
				"toys",
			},
		},
		{
			name: "words with grave accent",
			text: "Children`s toys and three cats` toys, all of''them are green",
			expected: []string{
				"Children`s",
				"toys",
				"and",
				"three",
				"cats`",
				"toys",
				"all",
				"of'",
				"them",
				"are",
				"green",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := ExtractWords(tt.text)
			require.Equal(t, tt.expected, actual)
		})
	}
}
