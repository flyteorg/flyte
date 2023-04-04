package auth

import (
	"context"
	"net/http"

	"testing"

	config2 "github.com/flyteorg/flytestdlib/config"

	"github.com/flyteorg/flyteadmin/auth/config"
	"github.com/stretchr/testify/assert"
)

func TestGetPublicURL(t *testing.T) {
	t.Run("Matching scheme and host", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "https://abc", nil)
		assert.NoError(t, err)
		u := GetPublicURL(context.Background(), req, &config.Config{
			AuthorizedURIs: []config2.URL{
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
			AuthorizedURIs: []config2.URL{
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
			AuthorizedURIs: []config2.URL{
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
			AuthorizedURIs: []config2.URL{
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
