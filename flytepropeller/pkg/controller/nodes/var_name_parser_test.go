package nodes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseVarName(t *testing.T) {
	t.Run("JustVarName", func(t *testing.T) {
		idx, name, err := ParseVarName("someVar")
		assert.NoError(t, err)
		assert.Nil(t, idx)
		assert.NotNil(t, name)
		assert.Equal(t, "someVar", name)
	})

	t.Run("WithIndex", func(t *testing.T) {
		idx, name, err := ParseVarName("[10].someVar")
		assert.NoError(t, err)
		assert.Equal(t, 10, *idx)
		assert.Equal(t, "someVar", name)
	})

	t.Run("Invalid no dot", func(t *testing.T) {
		_, _, err := ParseVarName("[10]someVar")
		assert.Error(t, err)
	})

	t.Run("Invalid not int", func(t *testing.T) {
		_, _, err := ParseVarName("[asdf].someVar")
		assert.Error(t, err)
	})
}
