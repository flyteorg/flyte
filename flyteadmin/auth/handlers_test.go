package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/flyteorg/flyte/flyteadmin/auth/config"
	"github.com/flyteorg/flyte/flyteadmin/auth/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/auth/interfaces/mocks"
	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteadmin/plugins"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	stdConfig "github.com/flyteorg/flyte/flytestdlib/config"
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
	mockCookieHandler.OnSetTokenCookiesMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	mockCookieHandler.OnSetUserInfoCookieMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
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
	mockAuthCtx.OnOptions().Return(&config.Config{
		UserAuth: config.UserAuthConfig{
			IDPQueryParameter: "idp",
		},
	})
	mockAuthCtx.OnOAuth2ClientConfigMatch(mock.Anything).Return(&dummyOAuth2Config)
	handler := GetLoginHandler(ctx, &mockAuthCtx)

	type test struct {
		name               string
		url                string
		expectedStatusCode int
		expectedLocation   string
		expectedSetCookie  string
	}
	tests := []test{
		{
			name:               "no idp parameter",
			url:                "/login",
			expectedStatusCode: 307,
			expectedLocation:   "response_type=code&scope=openid+other",
			expectedSetCookie:  "flyte_csrf_state=",
		},
		{
			name:               "with idp parameter config",
			url:                "/login?idp=dummyIDP",
			expectedStatusCode: 307,
			expectedLocation:   "dummyIDP",
			expectedSetCookie:  "flyte_csrf_state=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.url, nil)
			assert.NoError(t, err)
			w := httptest.NewRecorder()
			handler(w, req)
			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.True(t, strings.Contains(w.Header().Get("Location"), tt.expectedLocation))
			assert.True(t, strings.Contains(w.Header().Get("Set-Cookie"), tt.expectedSetCookie))
		})
	}
}

func TestGetLogoutHandler(t *testing.T) {
	ctx := context.Background()

	t.Run("no_hook_no_redirect", func(t *testing.T) {
		cookieHandler := &CookieManager{}
		authCtx := mocks.AuthenticationContext{}
		authCtx.OnCookieManager().Return(cookieHandler).Once()
		w := httptest.NewRecorder()
		r := plugins.NewRegistry()
		req, err := http.NewRequest(http.MethodGet, "/logout", nil)
		require.NoError(t, err)

		GetLogoutEndpointHandler(ctx, &authCtx, r)(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		require.Len(t, w.Result().Cookies(), 5)
		authCtx.AssertExpectations(t)
	})

	t.Run("no_hook_with_redirect", func(t *testing.T) {
		ctx := context.Background()
		cookieHandler := &CookieManager{}
		authCtx := mocks.AuthenticationContext{}
		authCtx.OnCookieManager().Return(cookieHandler).Once()
		w := httptest.NewRecorder()
		r := plugins.NewRegistry()
		req, err := http.NewRequest(http.MethodGet, "/logout?redirect_url=/foo", nil)
		require.NoError(t, err)

		GetLogoutEndpointHandler(ctx, &authCtx, r)(w, req)

		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		authCtx.AssertExpectations(t)
		require.Len(t, w.Result().Cookies(), 5)
	})

	t.Run("with_hook_with_redirect", func(t *testing.T) {
		ctx := context.Background()
		cookieHandler := &CookieManager{}
		authCtx := mocks.AuthenticationContext{}
		authCtx.OnCookieManager().Return(cookieHandler).Once()
		w := httptest.NewRecorder()
		r := plugins.NewRegistry()
		hook := new(mock.Mock)
		err := r.Register(plugins.PluginIDLogoutHook, LogoutHookFunc(func(
			ctx context.Context,
			authCtx interfaces.AuthenticationContext,
			request *http.Request,
			w http.ResponseWriter) error {
			return hook.MethodCalled("hook").Error(0)
		}))
		hook.On("hook").Return(nil).Once()
		require.NoError(t, err)
		req, err := http.NewRequest(http.MethodGet, "/logout?redirect_url=/foo", nil)
		require.NoError(t, err)

		GetLogoutEndpointHandler(ctx, &authCtx, r)(w, req)

		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		require.Len(t, w.Result().Cookies(), 5)
		authCtx.AssertExpectations(t)
		hook.AssertExpectations(t)
	})

	t.Run("hook_error", func(t *testing.T) {
		ctx := context.Background()
		authCtx := mocks.AuthenticationContext{}
		w := httptest.NewRecorder()
		r := plugins.NewRegistry()
		hook := new(mock.Mock)
		err := r.Register(plugins.PluginIDLogoutHook, LogoutHookFunc(func(
			ctx context.Context,
			authCtx interfaces.AuthenticationContext,
			request *http.Request,
			w http.ResponseWriter) error {
			return hook.MethodCalled("hook").Error(0)
		}))
		hook.On("hook").Return(errors.New("fail")).Once()
		require.NoError(t, err)
		req, err := http.NewRequest(http.MethodGet, "/logout?redirect_url=/foo", nil)
		require.NoError(t, err)

		GetLogoutEndpointHandler(ctx, &authCtx, r)(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Empty(t, w.Result().Cookies())
		authCtx.AssertExpectations(t)
		hook.AssertExpectations(t)
	})
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

	accessTokenCookie, err := NewSecureCookie(accessTokenCookieNameSplitFirst, "a.b.c", cookieManager.hashKey, cookieManager.blockKey, "localhost", http.SameSiteDefaultMode)
	assert.NoError(t, err)
	req.AddCookie(&accessTokenCookie)

	accessTokenCookieSplit, err := NewSecureCookie(accessTokenCookieNameSplitSecond, ".d.e.f", cookieManager.hashKey, cookieManager.blockKey, "localhost", http.SameSiteDefaultMode)
	assert.NoError(t, err)
	req.AddCookie(&accessTokenCookieSplit)

	idCookie, err := NewSecureCookie(idTokenCookieName, "a.b.c.d.e.f", cookieManager.hashKey, cookieManager.blockKey, "localhost", http.SameSiteDefaultMode)
	assert.NoError(t, err)
	req.AddCookie(&idCookie)

	assert.Equal(t, "IDToken a.b.c.d.e.f", handler(ctx, req)["authorization"][0])
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
	type test struct {
		name                     string
		baseURL                  string
		metadataPath             string
		expectedRedirectLocation string
	}
	tests := []test{
		{
			name:                     "base_url_without_path",
			baseURL:                  "http://www.google.com",
			metadataPath:             OIdCMetadataEndpoint,
			expectedRedirectLocation: "http://www.google.com/.well-known/openid-configuration",
		},
		{
			name:                     "base_url_with_path",
			baseURL:                  "https://login.microsoftonline.com/abc/v2.0",
			metadataPath:             OIdCMetadataEndpoint,
			expectedRedirectLocation: "https://login.microsoftonline.com/abc/v2.0/.well-known/openid-configuration",
		},
		{
			name:                     "base_url_with_trailing_slash_path",
			baseURL:                  "https://login.microsoftonline.com/abc/v2.0/",
			metadataPath:             OIdCMetadataEndpoint,
			expectedRedirectLocation: "https://login.microsoftonline.com/abc/v2.0/.well-known/openid-configuration",
		},
		{
			name:                     "absolute_metadata_path",
			baseURL:                  "https://login.microsoftonline.com/abc/v2.0/",
			metadataPath:             "/.well-known/openid-configuration",
			expectedRedirectLocation: "https://login.microsoftonline.com/.well-known/openid-configuration",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadataPath := mustParseURL(t, tt.metadataPath)
			mockAuthCtx := mocks.AuthenticationContext{}
			mockAuthCtx.OnOptions().Return(&config.Config{
				UserAuth: config.UserAuthConfig{
					OpenID: config.OpenIDOptions{
						BaseURL: stdConfig.URL{URL: mustParseURL(t, tt.baseURL)},
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
			assert.Equal(t, tt.expectedRedirectLocation, w.Header()["Location"][0])
		})
	}
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
