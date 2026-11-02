package namespaceutils

import (
	"strings"
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

func TestMappedNamespace(t *testing.T) {
	t.Run("empty template uses the default namespace", func(t *testing.T) {
		got, err := MappedNamespace("", "flyte", "", "flytesnacks", "development")
		assert.NoError(t, err)
		assert.Equal(t, "flyte", got)
	})

	t.Run("template is rendered", func(t *testing.T) {
		got, err := MappedNamespace("{{ project }}-{{ domain }}", "flyte", "", "flytesnacks", "development")
		assert.NoError(t, err)
		assert.Equal(t, "flytesnacks-development", got)
	})

	invalid := []struct {
		name    string
		project string
	}{
		{"uppercase", "FlyteSnacks"},
		{"underscore", "flyte_snacks"},
		{"too long", strings.Repeat("p", 60)},
	}
	for _, tc := range invalid {
		t.Run("rejects "+tc.name, func(t *testing.T) {
			_, err := MappedNamespace("{{ project }}-{{ domain }}", "flyte", "", tc.project, "development")
			assert.ErrorContains(t, err, "not a valid Kubernetes namespace name")
		})
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
