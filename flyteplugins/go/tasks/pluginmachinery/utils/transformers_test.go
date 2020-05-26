package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {

	assert.True(t, Contains([]string{"a", "b", "c"}, "b"))

	assert.False(t, Contains([]string{"a", "b", "c"}, "spark"))

	assert.False(t, Contains([]string{}, "spark"))

	assert.False(t, Contains(nil, "b"))
}

func TestCopyMap(t *testing.T) {
	assert.Nil(t, CopyMap(nil))
	m := map[string]string{
		"l": "v",
	}
	assert.Equal(t, m, CopyMap(m))
}
