package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeSettingsKey(t *testing.T) {
	cases := []struct {
		name    string
		domain  string
		project string
		want    string
	}{
		{name: "project level", domain: "dev", project: "recsys", want: "v1::dev:recsys"},
		{name: "domain level", domain: "dev", project: "", want: "v1::dev:"},
		{name: "instance level", domain: "", project: "", want: "v1:::"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, EncodeSettingsKey(tc.domain, tc.project))
		})
	}
}
