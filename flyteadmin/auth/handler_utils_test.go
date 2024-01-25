package auth

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteadmin/auth/config"
	flytestdconfig "github.com/flyteorg/flyte/flytestdlib/config"
)

func TestGetPublicURL(t *testing.T) {
	t.Run("Matching scheme and host", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "https://abc", nil)
		assert.NoError(t, err)
		u := GetPublicURL(context.Background(), req, &config.Config{
			AuthorizedURIs: []flytestdconfig.URL{
				{URL: *config.MustParseURL("https://xyz")},
				{URL: *config.MustParseURL("https://abc")},
			},
		})
		assert.Equal(t, "https://abc", u.String())
	})

	t.Run("Matching host and non-matching scheme", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "https://abc", nil)
		assert.NoError(t, err)
		u := GetPublicURL(context.Background(), req, &config.Config{
			AuthorizedURIs: []flytestdconfig.URL{
				{URL: *config.MustParseURL("https://xyz")},
				{URL: *config.MustParseURL("http://abc")},
			},
		})
		assert.Equal(t, "http://abc", u.String())
	})

	t.Run("No matching", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "https://abc", nil)
		assert.NoError(t, err)
		u := GetPublicURL(context.Background(), req, &config.Config{
			AuthorizedURIs: []flytestdconfig.URL{
				{URL: *config.MustParseURL("https://xyz")},
				{URL: *config.MustParseURL("http://xyz")},
			},
		})
		assert.Equal(t, "https://xyz", u.String())
	})

	t.Run("No config", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "https://abc", nil)
		assert.NoError(t, err)
		u := GetPublicURL(context.Background(), req, &config.Config{})
		assert.Equal(t, "https://abc", u.String())
	})

	t.Run("Matching testing setup but with ssl", func(t *testing.T) {
		// req has https so that it's not a scheme match with the list below
		req, err := http.NewRequest(http.MethodPost, "https://localhost:30081", nil)

		assert.NoError(t, err)
		u := GetPublicURL(context.Background(), req, &config.Config{
			AuthorizedURIs: []flytestdconfig.URL{
				{URL: *config.MustParseURL("http://flyteadmin:80")},
				{URL: *config.MustParseURL("http://localhost:30081")},
				{URL: *config.MustParseURL("http://localhost:8089")},
				{URL: *config.MustParseURL("http://localhost:8088")},
				{URL: *config.MustParseURL("http://flyteadmin.flyte.svc.cluster.local:80")},
			},
		})
		assert.Equal(t, "http://localhost:30081", u.String())
	})
}

func TestGetRedirectURLAllowed(t *testing.T) {
	ctx := context.TODO()
	t.Run("relative url", func(t *testing.T) {
		assert.True(t, GetRedirectURLAllowed(ctx, "/console", &config.Config{}))
	})
	t.Run("no redirect url", func(t *testing.T) {
		assert.True(t, GetRedirectURLAllowed(ctx, "", &config.Config{}))
	})
	cfg := &config.Config{
		AuthorizedURIs: []flytestdconfig.URL{
			{URL: *config.MustParseURL("https://example.com")},
			{URL: *config.MustParseURL("http://localhost:3008")},
		},
	}
	t.Run("authorized url", func(t *testing.T) {
		assert.True(t, GetRedirectURLAllowed(ctx, "https://example.com", cfg))
	})
	t.Run("authorized localhost url", func(t *testing.T) {
		assert.True(t, GetRedirectURLAllowed(ctx, "http://localhost:3008", cfg))
	})
	t.Run("unauthorized url", func(t *testing.T) {
		assert.False(t, GetRedirectURLAllowed(ctx, "https://flyte.com", cfg))
	})
}
