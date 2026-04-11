package auth

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	stdconfig "github.com/flyteorg/flyte/v2/flytestdlib/config"
	"github.com/flyteorg/flyte/v2/runs/service/auth/config"
)

func mustParse(t *testing.T, raw string) stdconfig.URL {
	t.Helper()
	u, err := url.Parse(raw)
	require.NoError(t, err)
	return stdconfig.URL{URL: *u}
}

func TestComputeOIDCRedirectURL(t *testing.T) {
	cases := []struct {
		name string
		cfg  config.Config
		want string
	}{
		{
			name: "no authorizedUris falls back to relative path",
			cfg:  config.Config{},
			want: "callback",
		},
		{
			name: "simple https host",
			cfg: config.Config{
				AuthorizedURIs: []stdconfig.URL{mustParse(t, "https://flyte.example.com")},
			},
			want: "https://flyte.example.com/callback",
		},
		{
			name: "host with trailing slash does not duplicate separator",
			cfg: config.Config{
				AuthorizedURIs: []stdconfig.URL{mustParse(t, "https://flyte.example.com/")},
			},
			want: "https://flyte.example.com/callback",
		},
		{
			name: "picks first uri when multiple",
			cfg: config.Config{
				AuthorizedURIs: []stdconfig.URL{
					mustParse(t, "https://flyte.example.com"),
					mustParse(t, "http://flyte2.flyte:8080"),
				},
			},
			want: "https://flyte.example.com/callback",
		},
		{
			name: "host with path prefix appends callback",
			cfg: config.Config{
				AuthorizedURIs: []stdconfig.URL{mustParse(t, "https://flyte.example.com/v2")},
			},
			want: "https://flyte.example.com/v2/callback",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, computeOIDCRedirectURL(tc.cfg))
		})
	}
}
