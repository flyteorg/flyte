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
}
