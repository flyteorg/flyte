package auth

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	authConfig "github.com/flyteorg/flyte/v2/runs/service/auth/config"
	"google.golang.org/grpc/metadata"
)

func TestURLFromRequest(t *testing.T) {
	t.Run("nil request", func(t *testing.T) {
		assert.Nil(t, URLFromRequest(nil))
	})

	t.Run("request with host", func(t *testing.T) {
		req := &http.Request{
			Host: "example.com:8080",
			URL:  &url.URL{Scheme: "https"},
		}
		u := URLFromRequest(req)
		assert.NotNil(t, u)
		assert.Equal(t, "example.com:8080", u.Host)
	})

	t.Run("request with empty host", func(t *testing.T) {
		req := &http.Request{
			URL: &url.URL{},
		}
		assert.Nil(t, URLFromRequest(req))
	})

	t.Run("request with absolute request URI", func(t *testing.T) {
		req := &http.Request{
			RequestURI: "https://absolute.example.com/path",
		}
		u := URLFromRequest(req)
		assert.NotNil(t, u)
		assert.Equal(t, "absolute.example.com", u.Hostname())
	})
}

func TestURLFromContext(t *testing.T) {
	t.Run("no metadata", func(t *testing.T) {
		assert.Nil(t, URLFromContext(context.Background()))
	})

	t.Run("x-forwarded-host", func(t *testing.T) {
		md := metadata.Pairs("x-forwarded-host", "forwarded.example.com")
		ctx := metadata.NewIncomingContext(context.Background(), md)
		u := URLFromContext(ctx)
		assert.NotNil(t, u)
		assert.Equal(t, "forwarded.example.com", u.Hostname())
		assert.Equal(t, "https", u.Scheme)
	})

	t.Run("authority fallback", func(t *testing.T) {
		md := metadata.Pairs(":authority", "authority.example.com")
		ctx := metadata.NewIncomingContext(context.Background(), md)
		u := URLFromContext(ctx)
		assert.NotNil(t, u)
		assert.Equal(t, "authority.example.com", u.Hostname())
	})
}

func TestFirstURL(t *testing.T) {
	u1, _ := url.Parse("https://first.example.com")
	u2, _ := url.Parse("https://second.example.com")

	assert.Nil(t, FirstURL())
	assert.Nil(t, FirstURL(nil, nil))
	assert.Equal(t, u1, FirstURL(u1, u2))
	assert.Equal(t, u2, FirstURL(nil, u2))
}

func TestWildcardMatch(t *testing.T) {
	assert.True(t, wildcardMatch("example.com", "example.com"))
	assert.True(t, wildcardMatch("tenant1.union.ai", "*.union.ai"))
	assert.False(t, wildcardMatch("union.ai", "*.union.ai"))
	assert.False(t, wildcardMatch("other.com", "example.com"))
	assert.False(t, wildcardMatch("sub.tenant1.union.ai", "*.union.ai"))
}

func TestBuildURL(t *testing.T) {
	authorized, _ := url.Parse("https://example.com:8080/path")
	result := buildURL(authorized, "tenant1.example.com")
	assert.Equal(t, "tenant1.example.com:8080", result.Host)
	assert.Equal(t, "/path", result.Path)

	noPort, _ := url.Parse("https://example.com/path")
	result = buildURL(noPort, "tenant1.example.com")
	assert.Equal(t, "tenant1.example.com", result.Host)
}

func TestGetPublicURL(t *testing.T) {
	t.Run("no request, returns first authorized URI", func(t *testing.T) {
		cfg := authConfig.Config{
			AuthorizedURIs: []authConfig.URL{
				{URL: *mustParseURL("https://flyte.example.com")},
			},
		}
		u := GetPublicURL(context.Background(), nil, cfg)
		assert.Equal(t, "flyte.example.com", u.Hostname())
	})

	t.Run("no request, no authorized URIs, returns nil", func(t *testing.T) {
		cfg := authConfig.Config{}
		u := GetPublicURL(context.Background(), nil, cfg)
		assert.Nil(t, u)
	})

	t.Run("wildcard match with request", func(t *testing.T) {
		cfg := authConfig.Config{
			AuthorizedURIs: []authConfig.URL{
				{URL: *mustParseURL("https://*.union.ai")},
			},
		}
		req := &http.Request{
			Host: "tenant1.union.ai",
			URL:  &url.URL{Scheme: "https"},
		}
		u := GetPublicURL(context.Background(), req, cfg)
		assert.Equal(t, "tenant1.union.ai", u.Hostname())
		assert.Equal(t, "https", u.Scheme)
	})

	t.Run("exact match with matching scheme", func(t *testing.T) {
		cfg := authConfig.Config{
			AuthorizedURIs: []authConfig.URL{
				{URL: *mustParseURL("https://flyte.example.com:8080")},
				{URL: *mustParseURL("http://flyte.example.com:8080")},
			},
		}
		req := &http.Request{
			Host: "flyte.example.com:8080",
			URL:  &url.URL{Scheme: "http"},
		}
		u := GetPublicURL(context.Background(), req, cfg)
		assert.Equal(t, "http", u.Scheme)
	})
}

func TestGetIssuer(t *testing.T) {
	t.Run("custom issuer", func(t *testing.T) {
		cfg := authConfig.Config{
			AppAuth: authConfig.OAuth2Options{
				SelfAuthServer: authConfig.AuthorizationServer{
					Issuer: "https://custom-issuer.example.com",
				},
			},
		}
		assert.Equal(t, "https://custom-issuer.example.com", GetIssuer(context.Background(), nil, cfg))
	})

	t.Run("falls back to public URL", func(t *testing.T) {
		cfg := authConfig.Config{
			AuthorizedURIs: []authConfig.URL{
				{URL: *mustParseURL("https://flyte.example.com")},
			},
		}
		assert.Equal(t, "https://flyte.example.com", GetIssuer(context.Background(), nil, cfg))
	})
}

func mustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return u
}
