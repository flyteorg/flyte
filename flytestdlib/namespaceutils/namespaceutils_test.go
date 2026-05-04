package namespaceutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNamespaceName(t *testing.T) {
	testCases := []struct {
		template string
		org      string
		project  string
		domain   string
		want     string
	}{
		{"prefix-{{ project }}-{{ domain }}", "", "flytesnacks", "production", "prefix-flytesnacks-production"},
		{"{{ domain }}", "", "flytesnacks", "production", "production"},
		{"{{ project }}", "", "flytesnacks", "production", "flytesnacks"},
		{"{{ org }}-{{ project }}-{{ domain }}", "acme", "flytesnacks", "production", "acme-flytesnacks-production"},
		{"{{ org }}", "acme", "flytesnacks", "production", "acme"},
	}

	for _, tc := range testCases {
		got := GetNamespaceName(tc.template, tc.org, tc.project, tc.domain)
		assert.Equal(t, tc.want, got)
	}
}

func TestGetNameWithNamespacedPrefix(t *testing.T) {
	testCases := []struct {
		prefix string
		name   string
		want   string
	}{
		{"project-domain-", "app", "project-domain-app"},
		{"", "app", "app"},
		{"prefix-", "", "prefix-"},
	}

	for _, tc := range testCases {
		got := GetNameWithNamespacedPrefix(tc.prefix, tc.name)
		assert.Equal(t, tc.want, got)
	}
}
