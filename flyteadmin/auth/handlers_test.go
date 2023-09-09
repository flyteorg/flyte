package auth

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/coreos/go-oidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/flyteorg/flyteadmin/auth/config"
	"github.com/flyteorg/flyteadmin/auth/interfaces"
	"github.com/flyteorg/flyteadmin/auth/interfaces/mocks"
	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/plugins"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	stdConfig "github.com/flyteorg/flytestdlib/config"
)

const (
	oauth2TokenURL = "/oauth2/token" // #nosec
)

func TestWithUserEmail(t *testing.T) {
	ctx := WithUserEmail(context.Background(), "abc")
	assert.Equal(t, "abc", ctx.Value(common.PrincipalContextKey))
}

func setupMockedAuthContextAtEndpoint(endpoint string) *mocks.AuthenticationContext {
	mockAuthCtx := &mocks.AuthenticationContext{}
	mockAuthCtx.OnOptions().Return(&config.Config{})
	mockCookieHandler := new(mocks.CookieHandler)
	dummyOAuth2Config := oauth2.Config{
		ClientID: "abc",
		Endpoint: oauth2.Endpoint{
			AuthURL:  endpoint + "/oauth2/authorize",
			TokenURL: endpoint + oauth2TokenURL,
		},
		Scopes: []string{"openid", "other"},
	}
	dummyHTTPClient := &http.Client{
		Timeout: IdpConnectionTimeout,
	}
	mockAuthCtx.OnCookieManagerMatch().Return(mockCookieHandler)
	mockCookieHandler.OnSetTokenCookiesMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockCookieHandler.OnSetUserInfoCookieMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockAuthCtx.OnOAuth2ClientConfigMatch(mock.Anything).Return(&dummyOAuth2Config)
	mockAuthCtx.OnGetHTTPClient().Return(dummyHTTPClient)
	return mockAuthCtx
}

func addStateString(request *http.Request) {
	v := url.Values{
		"state": []string{"b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"},
	}
	request.Form = v
}

func addCsrfCookie(request *http.Request) {
	cookie := NewCsrfCookie()
	cookie.Value = "hello world"
	request.AddCookie(&cookie)
}

func TestGetCallbackHandlerWithErrorOnToken(t *testing.T) {
	ctx := context.Background()
	hf := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == oauth2TokenURL {
			w.WriteHeader(403)
			return
		}
	}
	localServer := httptest.NewServer(http.HandlerFunc(hf))
	defer localServer.Close()
	http.DefaultClient = localServer.Client()
	mockAuthCtx := setupMockedAuthContextAtEndpoint(localServer.URL)
	r := plugins.NewRegistry()
	callbackHandlerFunc := GetCallbackHandler(ctx, mockAuthCtx, r)
	request := httptest.NewRequest("GET", localServer.URL+"/callback", nil)
	addCsrfCookie(request)
	addStateString(request)
	writer := httptest.NewRecorder()
	callbackHandlerFunc(writer, request)
	assert.Equal(t, "403 Forbidden", writer.Result().Status)
}

func TestGetCallbackHandlerWithUnAuthorized(t *testing.T) {
	ctx := context.Background()
	hf := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == oauth2TokenURL {
			w.WriteHeader(403)
			return
		}
	}
	localServer := httptest.NewServer(http.HandlerFunc(hf))
	defer localServer.Close()
	http.DefaultClient = localServer.Client()
	mockAuthCtx := setupMockedAuthContextAtEndpoint(localServer.URL)
	r := plugins.NewRegistry()
	callbackHandlerFunc := GetCallbackHandler(ctx, mockAuthCtx, r)
	request := httptest.NewRequest("GET", localServer.URL+"/callback", nil)
	writer := httptest.NewRecorder()
	callbackHandlerFunc(writer, request)
	assert.Equal(t, "401 Unauthorized", writer.Result().Status)
}

func TestGetCallbackHandler(t *testing.T) {
	var openIDConfigJSON string
	var userInfoJSON string
	ctx := context.Background()
	hf := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == oauth2TokenURL {
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"access_token":"Sample.Access.Token",
									"issued_token_type":"urn:ietf:params:oauth:token-type:access_token",	
									"token_type":"Bearer",	
									"expires_in":3600,	
									"scope":"all"}`)
			return
		}
		if r.URL.Path == "/.well-known/openid-configuration" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, openIDConfigJSON)
			return
		}
		if r.URL.Path == "/userinfo" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, userInfoJSON)
			return
		}
	}
	localServer := httptest.NewServer(http.HandlerFunc(hf))
	defer localServer.Close()
	http.DefaultClient = localServer.Client()
	issuer := localServer.URL
	userInfoJSON = `{
							"subject" : "dummySubject",
							"profile" : "dummyProfile",
							"email"   : "dummyEmail"
					}`
	openIDConfigJSON = fmt.Sprintf(`{
				"issuer": "%v",
				"authorization_endpoint": "%v/auth",
				"token_endpoint": "%v/token",
				"jwks_uri": "%v/keys",
				"id_token_signing_alg_values_supported": ["RS256"]
			}`, issuer, issuer, issuer, issuer)

	t.Run("forbidden request when accessing user info", func(t *testing.T) {
		mockAuthCtx := setupMockedAuthContextAtEndpoint(localServer.URL)
		r := plugins.NewRegistry()
		callbackHandlerFunc := GetCallbackHandler(ctx, mockAuthCtx, r)
		request := httptest.NewRequest("GET", localServer.URL+"/callback", nil)
		addCsrfCookie(request)
		addStateString(request)
		writer := httptest.NewRecorder()
		openIDConfigJSON = fmt.Sprintf(`{
				"issuer": "%v",
				"authorization_endpoint": "%v/auth",
				"token_endpoint": "%v/token",
				"jwks_uri": "%v/keys",
				"id_token_signing_alg_values_supported": ["RS256"]
			}`, issuer, issuer, issuer, issuer)
		oidcProvider, err := oidc.NewProvider(ctx, issuer)
		assert.Nil(t, err)
		mockAuthCtx.OnOidcProviderMatch().Return(oidcProvider)
		callbackHandlerFunc(writer, request)
		assert.Equal(t, "403 Forbidden", writer.Result().Status)
	})

	t.Run("successful callback with redirect and successful preredirect hook call", func(t *testing.T) {
		mockAuthCtx := setupMockedAuthContextAtEndpoint(localServer.URL)
		r := plugins.NewRegistry()
		var redirectFunc PreRedirectHookFunc = func(redirectContext context.Context, authCtx interfaces.AuthenticationContext, request *http.Request, responseWriter http.ResponseWriter) *PreRedirectHookError {
			return nil
		}

		r.RegisterDefault(plugins.PluginIDPreRedirectHook, redirectFunc)
		callbackHandlerFunc := GetCallbackHandler(ctx, mockAuthCtx, r)
		request := httptest.NewRequest("GET", localServer.URL+"/callback", nil)
		addCsrfCookie(request)
		addStateString(request)
		writer := httptest.NewRecorder()
		openIDConfigJSON = fmt.Sprintf(`{
				"userinfo_endpoint": "%v/userinfo",
				"issuer": "%v",
				"authorization_endpoint": "%v/auth",
				"token_endpoint": "%v/token",
				"jwks_uri": "%v/keys",
				"id_token_signing_alg_values_supported": ["RS256"]
			}`, issuer, issuer, issuer, issuer, issuer)
		oidcProvider, err := oidc.NewProvider(ctx, issuer)
		assert.Nil(t, err)
		mockAuthCtx.OnOidcProviderMatch().Return(oidcProvider)
		callbackHandlerFunc(writer, request)
		assert.Equal(t, "307 Temporary Redirect", writer.Result().Status)
	})

	t.Run("successful callback with pre-redirecthook failure", func(t *testing.T) {
		mockAuthCtx := setupMockedAuthContextAtEndpoint(localServer.URL)
		r := plugins.NewRegistry()
		var redirectFunc PreRedirectHookFunc = func(redirectContext context.Context, authCtx interfaces.AuthenticationContext, request *http.Request, responseWriter http.ResponseWriter) *PreRedirectHookError {
			return &PreRedirectHookError{
				Code:    http.StatusPreconditionFailed,
				Message: "precondition error",
			}
		}

		r.RegisterDefault(plugins.PluginIDPreRedirectHook, redirectFunc)
		callbackHandlerFunc := GetCallbackHandler(ctx, mockAuthCtx, r)
		request := httptest.NewRequest("GET", localServer.URL+"/callback", nil)
		addCsrfCookie(request)
		addStateString(request)
		writer := httptest.NewRecorder()
		openIDConfigJSON = fmt.Sprintf(`{
				"userinfo_endpoint": "%v/userinfo",
				"issuer": "%v",
				"authorization_endpoint": "%v/auth",
				"token_endpoint": "%v/token",
				"jwks_uri": "%v/keys",
				"id_token_signing_alg_values_supported": ["RS256"]
			}`, issuer, issuer, issuer, issuer, issuer)
		oidcProvider, err := oidc.NewProvider(ctx, issuer)
		assert.Nil(t, err)
		mockAuthCtx.OnOidcProviderMatch().Return(oidcProvider)
		callbackHandlerFunc(writer, request)
		assert.Equal(t, "412 Precondition Failed", writer.Result().Status)
	})
}

func TestGetLoginHandler(t *testing.T) {
	ctx := context.Background()
	dummyOAuth2Config := oauth2.Config{
		ClientID: "abc",
		Scopes:   []string{"openid", "other"},
	}
	mockAuthCtx := mocks.AuthenticationContext{}
	mockAuthCtx.OnOptions().Return(&config.Config{})
	mockAuthCtx.OnOAuth2ClientConfigMatch(mock.Anything).Return(&dummyOAuth2Config)
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
	cookieSetting := config.CookieSettings{
		SameSitePolicy: config.SameSiteDefaultMode,
		Domain:         "default",
	}
	cookieManager, err := NewCookieManager(ctx, hashKeyEncoded, blockKeyEncoded, cookieSetting)
	assert.NoError(t, err)
	mockAuthCtx := mocks.AuthenticationContext{}
	mockAuthCtx.OnCookieManager().Return(&cookieManager)
	mockAuthCtx.OnOptions().Return(&config.Config{})
	handler := GetHTTPRequestCookieToMetadataHandler(&mockAuthCtx)
	req, err := http.NewRequest("GET", "/api/v1/projects", nil)
	assert.NoError(t, err)

	accessTokenCookie, err := NewSecureCookie(accessTokenCookieName, "a.b.c", cookieManager.hashKey, cookieManager.blockKey, "localhost", http.SameSiteDefaultMode)
	assert.NoError(t, err)
	req.AddCookie(&accessTokenCookie)

	idCookie, err := NewSecureCookie(idTokenCookieName, "a.b.c", cookieManager.hashKey, cookieManager.blockKey, "localhost", http.SameSiteDefaultMode)
	assert.NoError(t, err)
	req.AddCookie(&idCookie)

	assert.Equal(t, "IDToken a.b.c", handler(ctx, req)["authorization"][0])
}

func TestGetHTTPMetadataTaggingHandler(t *testing.T) {
	ctx := context.Background()
	annotator := GetHTTPMetadataTaggingHandler()
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
	cookieSetting := config.CookieSettings{
		SameSitePolicy: config.SameSiteDefaultMode,
		Domain:         "default",
	}
	cookieManager, err := NewCookieManager(ctx, hashKeyEncoded, blockKeyEncoded, cookieSetting)
	assert.NoError(t, err)
	mockAuthCtx := mocks.AuthenticationContext{}
	mockAuthCtx.On("CookieManager").Return(&cookieManager)
	mockConfig := &config.Config{
		HTTPAuthorizationHeader: "Custom-Header",
	}
	mockAuthCtx.OnOptions().Return(mockConfig)
	handler := GetHTTPRequestCookieToMetadataHandler(&mockAuthCtx)
	req, err := http.NewRequest("GET", "/api/v1/projects", nil)
	assert.NoError(t, err)
	req.Header.Set("Custom-Header", "Bearer a.b.c")
	assert.Equal(t, "Bearer a.b.c", handler(ctx, req)["authorization"][0])
}

func TestGetOIdCMetadataEndpointRedirectHandler(t *testing.T) {
	ctx := context.Background()
	metadataPath := mustParseURL(t, OIdCMetadataEndpoint)
	mockAuthCtx := mocks.AuthenticationContext{}
	mockAuthCtx.OnOptions().Return(&config.Config{
		UserAuth: config.UserAuthConfig{
			OpenID: config.OpenIDOptions{
				BaseURL: stdConfig.URL{URL: mustParseURL(t, "http://www.google.com")},
			},
		},
	})

	mockAuthCtx.OnGetOIdCMetadataURL().Return(&metadataPath)
	handler := GetOIdCMetadataEndpointRedirectHandler(ctx, &mockAuthCtx)
	req, err := http.NewRequest("GET", "/xyz", nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	handler(w, req)
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "http://www.google.com/.well-known/openid-configuration", w.Header()["Location"][0])
}

func TestUserInfoForwardResponseHander(t *testing.T) {
	ctx := context.Background()
	handler := GetUserInfoForwardResponseHandler()
	w := httptest.NewRecorder()
	additionalClaims := map[string]interface{}{
		"cid": "cid-id",
		"ver": 1,
	}
	additionalClaimsStruct, err := structpb.NewStruct(additionalClaims)
	assert.NoError(t, err)
	resp := service.UserInfoResponse{
		Subject:          "user-id",
		AdditionalClaims: additionalClaimsStruct,
	}
	assert.NoError(t, handler(ctx, w, &resp))
	assert.Contains(t, w.Result().Header, "X-User-Subject")
	assert.Equal(t, w.Result().Header["X-User-Subject"], []string{"user-id"})
	assert.Contains(t, w.Result().Header, "X-User-Claim-Cid")
	assert.Equal(t, w.Result().Header["X-User-Claim-Cid"], []string{"\"cid-id\""})
	assert.Contains(t, w.Result().Header, "X-User-Claim-Ver")
	assert.Equal(t, w.Result().Header["X-User-Claim-Ver"], []string{"1"})

	w = httptest.NewRecorder()
	unrelatedResp := service.OAuth2MetadataResponse{}
	assert.NoError(t, handler(ctx, w, &unrelatedResp))
	assert.NotContains(t, w.Result().Header, "X-User-Subject")
}
