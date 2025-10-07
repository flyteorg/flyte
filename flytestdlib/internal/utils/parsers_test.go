package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustParseURL(t *testing.T) {
	t.Run("Valid URL", func(t *testing.T) {
		MustParseURL("http://something-profound-localhost.com")
	})

	t.Run("Invalid URL", func(t *testing.T) {
		assert.Panics(t, func() {
			MustParseURL("invalid_url:is_here\\")
		})
	})
}

func TestRefUint32(t *testing.T) {
	input := int(5)
	res := RefInt(input)
	assert.Equal(t, input, *res)
}

func TestMustCompileRegexp(t *testing.T) {
	t.Run("Valid regexp", func(t *testing.T) {
		MustCompileRegexp("^(?:[0-9]{1,3}\\.){3}[0-9]{1,3}$")
	})

	t.Run("Invalid regexp", func(t *testing.T) {
		assert.Panics(t, func() {
			MustCompileRegexp("^(")
		})
	})
}
