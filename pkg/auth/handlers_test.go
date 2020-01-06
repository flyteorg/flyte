package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/lyft/flyteadmin/pkg/auth/config"
	"github.com/lyft/flyteadmin/pkg/common"

	"github.com/lyft/flyteadmin/pkg/auth/interfaces/mocks"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"

	"testing"
)

func TestWithUserEmail(t *testing.T) {
	ctx := WithUserEmail(context.Background(), "abc")
	assert.Equal(t, "abc", ctx.Value(common.PrincipalContextKey))
}

func TestGetLoginHandler(t *testing.T) {
	ctx := context.Background()
	dummyOAuth2Config := oauth2.Config{
		ClientID: "abc",
		Scopes:   []string{"openid", "other"},
	}
	mockAuthCtx := mocks.AuthenticationContext{}
	mockAuthCtx.On("OAuth2Config").Return(&dummyOAuth2Config)
	handler := GetLoginHandler(ctx, &mockAuthCtx)
	req, err := http.NewRequest("GET", "/login", nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	handler(w, req)
	assert.Equal(t, 307, w.Code)
	assert.True(t, strings.Contains(w.Header().Get("Location"), "response_type=code&scope=openid+other"))
	assert.True(t, strings.Contains(w.Header().Get("Set-Cookie"), "flyte_csrf_state="))
}

func TestGetHTTPRequestCookieToMetadataHandler(t *testing.T) {
	ctx := context.Background()
	// These were generated for unit testing only.
	hashKeyEncoded := "wG4pE1ccdw/pHZ2ml8wrD5VJkOtLPmBpWbKHmezWXktGaFbRoAhXidWs8OpbA3y7N8vyZhz1B1E37+tShWC7gA" //nolint:goconst
	blockKeyEncoded := "afyABVgGOvWJFxVyOvCWCupoTn6BkNl4SOHmahho16Q"                                           //nolint:goconst
	cookieManager, err := NewCookieManager(ctx, hashKeyEncoded, blockKeyEncoded)
	assert.NoError(t, err)
	mockAuthCtx := mocks.AuthenticationContext{}
	mockAuthCtx.On("CookieManager").Return(&cookieManager)
	handler := GetHTTPRequestCookieToMetadataHandler(&mockAuthCtx)
	req, err := http.NewRequest("GET", "/api/v1/projects", nil)
	assert.NoError(t, err)
	jwtCookie, err := NewSecureCookie(accessTokenCookieName, "a.b.c", cookieManager.hashKey, cookieManager.blockKey)
	assert.NoError(t, err)
	req.AddCookie(&jwtCookie)
	assert.Equal(t, "Bearer a.b.c", handler(ctx, req)["authorization"][0])
}

func TestGetHTTPMetadataTaggingHandler(t *testing.T) {
	ctx := context.Background()
	mockAuthCtx := mocks.AuthenticationContext{}
	annotator := GetHTTPMetadataTaggingHandler(&mockAuthCtx)
	request, err := http.NewRequest("GET", "/api", nil)
	assert.NoError(t, err)
	md := annotator(ctx, request)
	assert.Equal(t, FromHTTPVal, md.Get(FromHTTPKey)[0])
}

func TestGetHTTPRequestCookieToMetadataHandler_CustomHeader(t *testing.T) {
	ctx := context.Background()
	// These were generated for unit testing only.
	hashKeyEncoded := "wG4pE1ccdw/pHZ2ml8wrD5VJkOtLPmBpWbKHmezWXktGaFbRoAhXidWs8OpbA3y7N8vyZhz1B1E37+tShWC7gA" //nolint:goconst
	blockKeyEncoded := "afyABVgGOvWJFxVyOvCWCupoTn6BkNl4SOHmahho16Q"                                           //nolint:goconst
	cookieManager, err := NewCookieManager(ctx, hashKeyEncoded, blockKeyEncoded)
	assert.NoError(t, err)
	mockAuthCtx := mocks.AuthenticationContext{}
	mockAuthCtx.On("CookieManager").Return(&cookieManager)
	mockConfig := config.OAuthOptions{
		HTTPAuthorizationHeader: "Custom-Header",
	}
	mockAuthCtx.On("Options").Return(mockConfig)
	handler := GetHTTPRequestCookieToMetadataHandler(&mockAuthCtx)
	req, err := http.NewRequest("GET", "/api/v1/projects", nil)
	assert.NoError(t, err)
	req.Header.Set("Custom-Header", "Bearer a.b.c")
	assert.Equal(t, "Bearer a.b.c", handler(ctx, req)["authorization"][0])
}

func TestGetMetadataEndpointRedirectHandler(t *testing.T) {
	ctx := context.Background()
	baseURL, err := url.Parse("http://www.google.com")
	assert.NoError(t, err)
	metadataPath, err := url.Parse(MetadataEndpoint)
	assert.NoError(t, err)
	mockAuthCtx := mocks.AuthenticationContext{}
	mockAuthCtx.On("GetBaseURL").Return(baseURL)
	mockAuthCtx.On("GetMetadataURL").Return(metadataPath)
	handler := GetMetadataEndpointRedirectHandler(ctx, &mockAuthCtx)
	req, err := http.NewRequest("GET", "/xyz", nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	handler(w, req)
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "http://www.google.com/.well-known/oauth-authorization-server", w.Header()["Location"][0])
}
