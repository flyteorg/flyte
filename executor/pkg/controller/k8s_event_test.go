package controller

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruncateUTF8(t *testing.T) {
	longMessage := strings.Repeat("task failed ", 100)

	tests := []struct {
		name     string
		s        string
		max      int
		expected string
	}{
		{name: "empty input", s: "", max: 5, expected: ""},
		{name: "shorter than max", s: "task failed", max: 20, expected: "task failed"},
		{name: "exactly max", s: "task failed", max: 11, expected: "task failed"},
		{name: "word over max", s: "failed", max: 4, expected: "fail"},
		{name: "sentence over max", s: "the task failed after three retries", max: 15, expected: "the task failed"},
		{name: "zero max", s: "task failed", max: 0, expected: ""},
		{name: "message over note limit", s: longMessage, max: 1024, expected: longMessage[:1024]},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateUTF8(tt.s, tt.max)
			assert.Equal(t, tt.expected, got)
			assert.LessOrEqual(t, len(got), tt.max)
		})
	}
}
