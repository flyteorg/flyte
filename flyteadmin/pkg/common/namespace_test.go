package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNamespaceName(t *testing.T) {
	testCases := []struct {
		template string
		project  string
		domain   string
		want     string
	}{
		{"prefix-{{ project }}-{{ domain }}", "flytesnacks", "production", "prefix-flytesnacks-production"},
		{"{{ domain }}", "flytesnacks", "production", "production"},
		{"{{ project }}", "flytesnacks", "production", "flytesnacks"},
	}

	for _, tc := range testCases {
		got := GetNamespaceName(tc.template, tc.project, tc.domain)
		assert.Equal(t, got, tc.want)
	}
}
