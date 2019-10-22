package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNamespaceName(t *testing.T) {
	testCases := []struct {
		mapping NamespaceMapping
		project string
		domain  string
		want    string
	}{
		{ProjectDomain, "project", "production", "project-production"},
		{20 /*Dummy enum value that is not supported*/, "project", "development", "project-development"},
		{Domain, "project", "production", "production"},
	}

	for _, tc := range testCases {
		got := GetNamespaceName(tc.mapping, tc.project, tc.domain)
		assert.Equal(t, got, tc.want)
	}
}
