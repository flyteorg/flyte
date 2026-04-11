package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/v2/runs/service/auth/config"
)

func TestIsPublicPath(t *testing.T) {
	cases := map[string]bool{
		"/healthz":                                        true,
		"/readyz":                                         true,
		"/healthcheck":                                    true,
		"/login":                                          true,
		"/login?redirect_url=/console":                    true,
		"/callback":                                       true,
		"/logout":                                         true,
		"/.well-known/openid-configuration":               true,
		"/.well-known/oauth-authorization-server":         true,
		"/flyteidl2.auth.AuthMetadataService/GetOAuth2Metadata": true,
		"/flyteidl2.workflow.RunService/CreateRun":        false,
		"/flyteidl2.auth.IdentityService/UserInfo":        false,
		"/":                                               false,
		"/api/v1/projects":                                false,
	}
	for path, want := range cases {
		got := IsPublicPath(path)
		assert.Equalf(t, want, got, "IsPublicPath(%q)", path)
	}
}

// servedBy wraps a boolean flag so tests can check that the next handler ran.
type servedBy struct{ called bool }

func (s *servedBy) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		s.called = true
		w.WriteHeader(http.StatusOK)
	})
}

func TestMiddleware_PublicPathBypassesAuth(t *testing.T) {
	// Even with a zero AuthHandlerConfig (no resource server, no cookie manager),
	// public paths must not touch any auth plumbing.
	h := &AuthHandlerConfig{AuthConfig: config.Config{}}
	mw := GetAuthenticationHTTPInterceptor(h)

	var sb servedBy
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	mw(sb.handler()).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, sb.called, "public path should have reached the next handler")
}

func TestMiddleware_DisabledForHTTPBypassesAuth(t *testing.T) {
	h := &AuthHandlerConfig{AuthConfig: config.Config{DisableForHTTP: true}}
	mw := GetAuthenticationHTTPInterceptor(h)

	var sb servedBy
	req := httptest.NewRequest(http.MethodGet, "/flyteidl2.workflow.RunService/CreateRun", nil)
	w := httptest.NewRecorder()
	mw(sb.handler()).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, sb.called)
}

func TestMiddleware_NoAuthReturns401(t *testing.T) {
	// AuthHandlerConfig missing a CookieManager will cause IdentityContextFromRequest
	// to fail when no bearer header is present. The middleware must convert that to 401.
	h := &AuthHandlerConfig{
		AuthConfig:     config.Config{},
		CookieManager:  CookieManager{},
		ResourceServer: nil,
	}
	mw := GetAuthenticationHTTPInterceptor(h)

	var sb servedBy
	req := httptest.NewRequest(http.MethodGet, "/flyteidl2.workflow.RunService/CreateRun", nil)
	w := httptest.NewRecorder()
	mw(sb.handler()).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.False(t, sb.called, "protected path must not reach next handler without auth")
}

