package utils

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

// This regex is used to sanitize representations of protobufs in tests.
var sanitizeProtoRegex = regexp.MustCompile(`\s+`)

// This function is used to sanitize representations of protobufs in tests.
func AssertEqualWithSanitizedRegex(t *testing.T, expected string, actual string) {
	assert.Equal(t, sanitizeProtoRegex.ReplaceAllString(expected, ""), sanitizeProtoRegex.ReplaceAllString(actual, ""))
}
