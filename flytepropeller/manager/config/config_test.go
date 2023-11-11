package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	assert.Equal(t, DefaultConfig, GetConfig())
}
