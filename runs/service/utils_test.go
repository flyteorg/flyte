package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruncateShortDescription_UnderLimit(t *testing.T) {
	desc := "Short description"
	result := truncateShortDescription(desc)
	assert.Equal(t, desc, result)
}

func TestTruncateShortDescription_AtLimit(t *testing.T) {
	desc := strings.Repeat("a", 255)
	result := truncateShortDescription(desc)
	assert.Equal(t, desc, result)
	assert.Len(t, result, 255)
}

func TestTruncateShortDescription_OverLimit(t *testing.T) {
	desc := strings.Repeat("a", 300)
	result := truncateShortDescription(desc)
	assert.Len(t, result, 255)
	assert.Equal(t, strings.Repeat("a", 255), result)
}

func TestTruncateLongDescription_UnderLimit(t *testing.T) {
	desc := "Long description"
	result := truncateLongDescription(desc)
	assert.Equal(t, desc, result)
}

func TestTruncateLongDescription_AtLimit(t *testing.T) {
	desc := strings.Repeat("b", 2048)
	result := truncateLongDescription(desc)
	assert.Equal(t, desc, result)
	assert.Len(t, result, 2048)
}

func TestTruncateLongDescription_OverLimit(t *testing.T) {
	desc := strings.Repeat("b", 3000)
	result := truncateLongDescription(desc)
	assert.Len(t, result, 2048)
	assert.Equal(t, strings.Repeat("b", 2048), result)
}
