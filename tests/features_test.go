package tests

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"compiler/internal/driver"
	"compiler/tests/shared"
)

func TestFeatures(t *testing.T) {
	for _, c := range shared.MustCollect("features", ".mgo", ".stdout", ".stdin") {
		t.Run(c.Name, func(t *testing.T) {
			var in string
			if c.Input != "" {
				data, err := os.ReadFile(c.Input)
				require.NoError(t, err)
				in = string(data)
			}

			var out strings.Builder
			err := driver.Interpret(strings.NewReader(in), &out, c.Source)
			require.NoError(t, err)

			expected, err := os.ReadFile(c.Expected)
			require.NoError(t, err)
			require.Equal(t, normalize(string(expected)), normalize(out.String()))
		})
	}
}

func normalize(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}
