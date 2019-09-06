package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyMap(t *testing.T) {
	m := map[string]string{
		"k1": "v1",
		"k2": "v2",
	}
	co := CopyMap(m)
	assert.NotNil(t, co)
	assert.Equal(t, m, co)

	assert.Nil(t, CopyMap(nil))
}
