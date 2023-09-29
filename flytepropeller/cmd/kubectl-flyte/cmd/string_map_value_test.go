package cmd

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func formatArg(values map[string]string) string {
	res := ""
	for key, value := range values {
		res += fmt.Sprintf(",%v%v%v=%v%v%v", randSpaces(), key, randSpaces(), randSpaces(), value, randSpaces())
	}

	if len(values) > 0 {
		return res[1:]
	}

	return res
}

func randSpaces() string {
	res := ""
	for cnt := rand.Int()%10 + 1; cnt > 0; cnt-- { // nolint: gas
		res += " "
	}

	return res
}

func runPositiveTest(t *testing.T, expected map[string]string) {
	v := newStringMapValue()
	assert.NoError(t, v.Set(formatArg(expected)))

	assert.Equal(t, len(expected), len(*v.value))
	assert.Equal(t, expected, *v.value)
}

func TestSet(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		expected := map[string]string{
			"a":     "1",
			"b":     "2",
			"c":     "3",
			"d.sub": "x.y",
			"e":     "4",
		}
		runPositiveTest(t, expected)
	})

	t.Run("Empty arg", func(t *testing.T) {
		expected := map[string]string{
			"":      "",
			"a":     "1",
			"b":     "2",
			"c":     "3",
			"d.sub": "x.y",
			"e":     "4",
		}

		runPositiveTest(t, expected)
	})
}
