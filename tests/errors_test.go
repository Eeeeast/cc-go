package tests

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"compiler/internal/driver"
	"compiler/tests/shared"
)

func TestErrors(t *testing.T) {
	for _, c := range shared.MustCollect("errors", ".mgo", ".stderr", ".stdin") {
		t.Run(c.Name, func(t *testing.T) {
			var in string
			if c.Input != "" {
				data, err := os.ReadFile(c.Input)
				require.NoError(t, err)
				in = string(data)
			}

			var out strings.Builder
			interpretErr := driver.Interpret(strings.NewReader(in), &out, c.Source)
			require.Error(t, interpretErr)

			expected, err := os.ReadFile(c.Expected)
			require.NoError(t, err)
			require.Contains(t, interpretErr.Error(), strings.TrimSpace(string(expected)))
		})
	}
}
