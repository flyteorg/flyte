package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCmdIntegration(t *testing.T) {
	rootCmd := newRootCmd()
	assert.NotNil(t, rootCmd)
}
