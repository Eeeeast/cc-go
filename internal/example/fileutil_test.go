package examplelib

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	unsortedPoem = "Играют волны — ветер свищет,\n" +
		"И мачта гнется и скрыпит…\n" +
		"Увы! он счастия не ищет\n" +
		"И не от счастия бежит!\n"

	sortedPoem = "И мачта гнется и скрыпит…\n" +
		"И не от счастия бежит!\n" +
		"Играют волны — ветер свищет,\n" +
		"Увы! он счастия не ищет"
)

func TestSortFileLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "multi-line poem",
			input:    unsortedPoem,
			expected: sortedPoem,
		},
		{
			name:     "single line",
			input:    "Играют волны — ветер свищет,",
			expected: "Играют волны — ветер свищет,",
		},
		{
			name:     "empty file",
			input:    "",
			expected: "",
		},
		{
			name:     "already sorted",
			input:    "a\nb\nc",
			expected: "a\nb\nc",
		},
		{
			name:     "trailing newline is not an extra empty line",
			input:    "b\na\n",
			expected: "a\nb",
		},
		{
			name:     "windows line endings are normalized to \\n",
			input:    "b\r\na\r\n",
			expected: "a\nb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "test.txt")
			require.NoError(t, os.WriteFile(path, []byte(tt.input), 0o644))

			require.NoError(t, SortFileLines(path))

			actual, err := os.ReadFile(path)
			require.NoError(t, err)
			require.Equal(t, tt.expected, string(actual))
		})
	}
}

func TestSortFileLines_MissingFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "does-not-exist.txt")
	require.Error(t, SortFileLines(path))
}
