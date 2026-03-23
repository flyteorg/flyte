package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizePath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "tilde expansion",
			input:    "~/foo/bar",
			expected: filepath.Join(home, "foo/bar"),
		},
		{
			name:     "tilde with dot-dot",
			input:    "~/foo/../bar",
			expected: filepath.Join(home, "bar"),
		},
		{
			name:     "absolute path with dot-dot",
			input:    "/etc/flyte/../config.yaml",
			expected: "/etc/config.yaml",
		},
		{
			name:     "absolute path with multiple dot-dot",
			input:    "/a/b/c/../../d",
			expected: "/a/d",
		},
		{
			name:     "relative path with dot-dot",
			input:    "foo/../bar",
			expected: "bar",
		},
		{
			name:     "dot in path",
			input:    "/foo/./bar",
			expected: "/foo/bar",
		},
		{
			name:     "trailing slash removed",
			input:    "/foo/bar/",
			expected: "/foo/bar",
		},
		{
			name:     "plain absolute path unchanged",
			input:    "/etc/flyte/config.yaml",
			expected: "/etc/flyte/config.yaml",
		},
		{
			name:     "plain relative path unchanged",
			input:    "config.yaml",
			expected: "config.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, NormalizePath(tt.input))
		})
	}
}
