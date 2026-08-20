package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeSettingsKey(t *testing.T) {
	cases := []struct {
		name    string
		org     string
		domain  string
		project string
		want    string
	}{
		{name: "all segments set", org: "acme", domain: "dev", project: "recsys", want: "v1:acme:dev:recsys"},
		{name: "domain level, empty org normalized", org: "", domain: "dev", project: "", want: "v1:flyte:dev:"},
		{name: "instance level, all empty", org: "", domain: "", project: "", want: "v1:flyte::"},
		{name: "empty org with domain and project", org: "", domain: "dev", project: "recsys", want: "v1:flyte:dev:recsys"},
		{name: "org flyte passes through unchanged", org: "flyte", domain: "dev", project: "", want: "v1:flyte:dev:"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, EncodeSettingsKey(tc.org, tc.domain, tc.project))
		})
	}
}
