package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/securecookie"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/runs/service/auth/config"
)

func TestHashCsrfState(t *testing.T) {
	h := HashCsrfState("hello")
	// sha256("hello") hex
	assert.Equal(t, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", h)
	assert.Equal(t, h, HashCsrfState("hello"), "deterministic")
	assert.NotEqual(t, h, HashCsrfState("world"))
}

func TestNewCsrfToken(t *testing.T) {
	a := NewCsrfToken(1)
	b := NewCsrfToken(1)
	assert.Equal(t, a, b, "same seed should produce identical token")
	assert.Len(t, a, 10)
	for _, r := range a {
		assert.Contains(t, string(AllowedChars), string(r))
	}
	assert.NotEqual(t, a, NewCsrfToken(2))
}

func TestNewCsrfCookie(t *testing.T) {
	c := NewCsrfCookie()
	assert.Equal(t, "flyte_csrf_state", c.Name)
	assert.Len(t, c.Value, 10)
	assert.Equal(t, http.SameSiteLaxMode, c.SameSite)
	assert.True(t, c.HttpOnly)
}

func newTestKeys() ([]byte, []byte) {
	return securecookie.GenerateRandomKey(64), securecookie.GenerateRandomKey(32)
}

func TestSecureCookie_RoundTrip(t *testing.T) {
	hashKey, blockKey := newTestKeys()
	cookie, err := NewSecureCookie("flyte_at", "super-secret", hashKey, blockKey, "", http.SameSiteLaxMode)
	require.NoError(t, err)
	assert.Equal(t, "flyte_at", cookie.Name)
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, "/", cookie.Path)

	out, err := ReadSecureCookie(context.Background(), cookie, hashKey, blockKey)
	require.NoError(t, err)
	assert.Equal(t, "super-secret", out)
}

func TestReadSecureCookie_WrongKey(t *testing.T) {
	hashKey, blockKey := newTestKeys()
	cookie, err := NewSecureCookie("flyte_at", "value", hashKey, blockKey, "", http.SameSiteLaxMode)
	require.NoError(t, err)

	wrongHash, wrongBlock := newTestKeys()
	_, err = ReadSecureCookie(context.Background(), cookie, wrongHash, wrongBlock)
	assert.Error(t, err)
}

func TestRetrieveSecureCookie_Missing(t *testing.T) {
	hashKey, blockKey := newTestKeys()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	_, err := retrieveSecureCookie(context.Background(), req, "missing", hashKey, blockKey)
	assert.Error(t, err)
}

func TestVerifyCsrfCookie(t *testing.T) {
	token := "abcdefghij"
	hashed := HashCsrfState(token)

	req := httptest.NewRequest(http.MethodPost, "/callback", strings.NewReader("state="+hashed))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "flyte_csrf_state", Value: token})
	require.NoError(t, req.ParseForm())
	require.NoError(t, VerifyCsrfCookie(context.Background(), req))
}

func TestVerifyCsrfCookie_Mismatch(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/callback", strings.NewReader("state=wrong"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "flyte_csrf_state", Value: "abcdefghij"})
	require.NoError(t, req.ParseForm())
	assert.Error(t, VerifyCsrfCookie(context.Background(), req))
}

func TestVerifyCsrfCookie_EmptyState(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/callback", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	require.NoError(t, req.ParseForm())
	assert.Error(t, VerifyCsrfCookie(context.Background(), req))
}

func TestVerifyCsrfCookie_MissingCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/callback", strings.NewReader("state=x"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	require.NoError(t, req.ParseForm())
	assert.Error(t, VerifyCsrfCookie(context.Background(), req))
}

func TestNewRedirectCookie(t *testing.T) {
	c := NewRedirectCookie(context.Background(), "/console/projects")
	require.NotNil(t, c)
	assert.Equal(t, "flyte_redirect_location", c.Name)
	assert.Equal(t, "/console/projects", c.Value)
	assert.True(t, c.HttpOnly)
}

func TestNewRedirectCookie_Invalid(t *testing.T) {
	assert.Nil(t, NewRedirectCookie(context.Background(), ""))
}

func TestGetAuthFlowEndRedirect_QueryAllowed(t *testing.T) {
	authorized := []config.URL{{URL: *mustURL(t, "https://flyte.mycorp.com")}}
	req := httptest.NewRequest(http.MethodGet, "https://flyte.mycorp.com/callback?redirect_url=https://flyte.mycorp.com/console", nil)

	got := GetAuthFlowEndRedirect(context.Background(), "/default", authorized, req)
	assert.Equal(t, "https://flyte.mycorp.com/console", got)
}

func TestGetAuthFlowEndRedirect_QueryUnauthorizedFallsBack(t *testing.T) {
	authorized := []config.URL{{URL: *mustURL(t, "https://flyte.mycorp.com")}}
	req := httptest.NewRequest(http.MethodGet, "https://flyte.mycorp.com/callback?redirect_url=https://evil.example.com", nil)

	got := GetAuthFlowEndRedirect(context.Background(), "/default", authorized, req)
	assert.Equal(t, "/default", got)
}

func TestGetAuthFlowEndRedirect_CookieFallback(t *testing.T) {
	authorized := []config.URL{{URL: *mustURL(t, "https://flyte.mycorp.com")}}
	req := httptest.NewRequest(http.MethodGet, "https://flyte.mycorp.com/callback", nil)
	req.AddCookie(&http.Cookie{Name: "flyte_redirect_location", Value: "https://flyte.mycorp.com/console"})

	got := GetAuthFlowEndRedirect(context.Background(), "/default", authorized, req)
	assert.Equal(t, "https://flyte.mycorp.com/console", got)
}

func TestGetAuthFlowEndRedirect_NoCookieReturnsDefault(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://flyte.mycorp.com/callback", nil)
	got := GetAuthFlowEndRedirect(context.Background(), "/default", nil, req)
	assert.Equal(t, "/default", got)
}

func mustURL(t *testing.T, s string) *url.URL {
	t.Helper()
	u, err := url.Parse(s)
	require.NoError(t, err)
	return u
}
