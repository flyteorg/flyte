package kubernetes

import (
	"testing"
	"gotest.tools/assert"
)

func TestNewPlugin(t *testing.T) {
	p := NewPlugin()
	t.Run("New default plugin", func(t *testing.T) {
		assert.NotNil(t, p)
	})
}