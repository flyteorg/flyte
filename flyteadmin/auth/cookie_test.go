package auth

import (
	"bytes"
	"context"
	"encoding/base64"
	"net/http"
	"net/url"
	"testing"

	"github.com/gorilla/securecookie"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteadmin/auth/config"
	"github.com/flyteorg/flyte/flyteadmin/auth/interfaces/mocks"
	serverConfig "github.com/flyteorg/flyte/flyteadmin/pkg/config"
	stdConfig "github.com/flyteorg/flyte/flytestdlib/config"
)

func mustParseURL(t testing.TB, u string) url.URL {
	res, err := url.Parse(u)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	return *res
}

func TestSecureCookieLifecycle(t *testing.T) {
	tests := []struct {
		name                 string
		insecureCookieHeader bool
		expectedSecure       bool
	}{
		{
			name:                 "secure_cookie",
			insecureCookieHeader: false,
			expectedSecure:       true,
		},
		{
			name:                 "insecure_cookie",
			insecureCookieHeader: true,
			expectedSecure:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate hash and block keys for secure cookie
			hashKey := securecookie.GenerateRandomKey(64)
			assert.True(t, base64.RawStdEncoding.EncodeToString(hashKey) != "")

			blockKey := securecookie.GenerateRandomKey(32)
			assert.True(t, base64.RawStdEncoding.EncodeToString(blockKey) != "")

			// Set up server configuration with insecureCookieHeader option
			serverConfig.SetConfig(&serverConfig.ServerConfig{
				Security: serverConfig.ServerSecurityOptions{
					InsecureCookieHeader: tt.insecureCookieHeader,
				},
			})

			// Create a secure cookie
			cookie, err := NewSecureCookie("choc", "chip", hashKey, blockKey, "localhost", http.SameSiteLaxMode)
			assert.NoError(t, err)

			// Validate the Secure attribute of the cookie
			assert.Equal(t, tt.expectedSecure, cookie.Secure)

			// Read and validate the secure cookie value
			value, err := ReadSecureCookie(context.Background(), cookie, hashKey, blockKey)
			assert.NoError(t, err)
			assert.Equal(t, "chip", value)
		})
	}
}

func TestNewCsrfToken(t *testing.T) {
	csrf := NewCsrfToken(5)
	assert.Equal(t, "5qz3p9w8qo", csrf)
}

func TestNewCsrfCookie(t *testing.T) {
	tests := []struct {
		name                 string
		insecureCookieHeader bool
		expectedSecure       bool
	}{
		{
			name:                 "secure_csrf_cookie",
			insecureCookieHeader: false,
			expectedSecure:       true,
		},
		{
			name:                 "insecure_csrf_cookie",
			insecureCookieHeader: true,
			expectedSecure:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up server configuration with insecureCookieHeader option
			serverConfig.SetConfig(&serverConfig.ServerConfig{
				Security: serverConfig.ServerSecurityOptions{
					InsecureCookieHeader: tt.insecureCookieHeader,
				},
			})

			// Generate CSRF cookie
			cookie := NewCsrfCookie()

			// Validate CSRF cookie properties
			assert.Equal(t, "flyte_csrf_state", cookie.Name)
			assert.True(t, cookie.HttpOnly)
			assert.Equal(t, tt.expectedSecure, cookie.Secure)
		})
	}
}

func TestHashCsrfState(t *testing.T) {
	h := HashCsrfState("hello world")
	assert.Equal(t, "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9", h)
}

func TestVerifyCsrfCookie(t *testing.T) {
	t.Run("test no state", func(t *testing.T) {
		var buf bytes.Buffer
		request, err := http.NewRequest(http.MethodPost, "/test", &buf)
		assert.NoError(t, err)
		err = VerifyCsrfCookie(context.Background(), request)
		assert.Error(t, err)
	})

	t.Run("test incorrect token", func(t *testing.T) {
		var buf bytes.Buffer
		request, err := http.NewRequest(http.MethodPost, "/test", &buf)
		assert.NoError(t, err)
		v := url.Values{
			"state": []string{"b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"},
		}
		cookie := NewCsrfCookie()
		cookie.Value = "helloworld"
		request.Form = v
		request.AddCookie(&cookie)
		err = VerifyCsrfCookie(context.Background(), request)
		assert.Error(t, err)
	})

	t.Run("test correct token", func(t *testing.T) {
		var buf bytes.Buffer
		request, err := http.NewRequest(http.MethodPost, "/test", &buf)
		assert.NoError(t, err)
		v := url.Values{
			"state": []string{"b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"},
		}
		cookie := NewCsrfCookie()
		cookie.Value = "hello world"
		request.Form = v
		request.AddCookie(&cookie)
		err = VerifyCsrfCookie(context.Background(), request)
		assert.NoError(t, err)
	})
}

func TestNewRedirectCookie(t *testing.T) {
	t.Run("test local path", func(t *testing.T) {
		ctx := context.Background()
		cookie := NewRedirectCookie(ctx, "/console")
		assert.NotNil(t, cookie)
		assert.Equal(t, "/console", cookie.Value)
	})

	t.Run("test external domain", func(t *testing.T) {
		ctx := context.Background()
		cookie := NewRedirectCookie(ctx, "http://www.example.com/postLogin")
		assert.NotNil(t, cookie)
		assert.Equal(t, "http://www.example.com/postLogin", cookie.Value)
	})

	t.Run("uses same-site lax policy", func(t *testing.T) {
		ctx := context.Background()
		cookie := NewRedirectCookie(ctx, "http://www.example.com/postLogin")
		assert.NotNil(t, cookie)
		assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
	})

	tests := []struct {
		name                 string
		insecureCookieHeader bool
		expectedSecure       bool
	}{
		{
			name:                 "secure_cookies",
			insecureCookieHeader: false,
			expectedSecure:       true,
		},
		{
			name:                 "insecure_cookies",
			insecureCookieHeader: true,
			expectedSecure:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverConfig.SetConfig(&serverConfig.ServerConfig{
				Security: serverConfig.ServerSecurityOptions{
					InsecureCookieHeader: tt.insecureCookieHeader,
				},
			})
			ctx := context.Background()
			cookie := NewRedirectCookie(ctx, "http://www.example.com/postLogin")
			assert.NotNil(t, cookie)
			assert.Equal(t, cookie.Secure, tt.expectedSecure)
		})
	}
}

func TestGetAuthFlowEndRedirect(t *testing.T) {
	t.Run("in request", func(t *testing.T) {
		ctx := context.Background()
		request, err := http.NewRequest(http.MethodGet, "/test", nil)
		assert.NoError(t, err)
		cookie := NewRedirectCookie(ctx, "/console")
		assert.NotNil(t, cookie)
		request.AddCookie(cookie)
		mockAuthCtx := &mocks.AuthenticationContext{}
		redirect := GetAuthFlowEndRedirect(ctx, mockAuthCtx, request)
		assert.Equal(t, "/console", redirect)
	})

	t.Run("not in request", func(t *testing.T) {
		ctx := context.Background()
		request, err := http.NewRequest(http.MethodGet, "/test", nil)
		assert.NoError(t, err)
		mockAuthCtx := &mocks.AuthenticationContext{}
		mockAuthCtx.OnOptions().Return(&config.Config{
			UserAuth: config.UserAuthConfig{
				RedirectURL: stdConfig.URL{URL: mustParseURL(t, "/api/v1/projects")},
			},
		})
		redirect := GetAuthFlowEndRedirect(ctx, mockAuthCtx, request)
		assert.Equal(t, "/api/v1/projects", redirect)
	})
}
