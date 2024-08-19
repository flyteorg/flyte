package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	t.Run("New scheduler plugin config", func(t *testing.T) {
		config := NewConfig()
		assert.Equal(t, "", config.GetScheduler())
		assert.Equal(t, "", config.GetParameters())
	})
}
