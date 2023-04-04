package authzserver

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	config2 "github.com/flyteorg/flytestdlib/config"

	"github.com/flyteorg/flyteadmin/auth"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyteadmin/auth/interfaces/mocks"

	"github.com/flyteorg/flyteadmin/auth/config"

	"github.com/ory/fosite"

	"github.com/stretchr/testify/assert"
)

func TestAuthEndpoint(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		originalURL := "http://localhost:8088/oauth2/authorize?client_id=my-client&redirect_uri=http%3A%2F%2Flocalhost%3A3846%2Fcallback&response_type=code&scope=photos+openid+offline&state=some-random-state-foobar&nonce=some-random-nonce&code_challenge=p0v_UR0KrXl4--BpxM2BQa7qIW5k3k4WauBhjmkVQw8&code_challenge_method=S256"
		req := httptest.NewRequest(http.MethodGet, originalURL, nil)
		w := httptest.NewRecorder()

		authCtx := &mocks.AuthenticationContext{}
		oauth2Provider := &mocks.OAuth2Provider{}
		oauth2Provider.OnNewAuthorizeRequest(req.Context(), req).Return(fosite.NewAuthorizeRequest(), nil)
		authCtx.OnOAuth2Provider().Return(oauth2Provider)

		cookieManager := &mocks.CookieHandler{}
		cookieManager.OnSetAuthCodeCookie(req.Context(), w, originalURL).Return(nil)
		authCtx.OnCookieManager().Return(cookieManager)

		authEndpoint(authCtx, w, req)
		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	})

	t.Run("Fail to write cookie", func(t *testing.T) {
		originalURL := "http://localhost:8088/oauth2/authorize?client_id=my-client&redirect_uri=http%3A%2F%2Flocalhost%3A3846%2Fcallback&response_type=code&scope=photos+openid+offline&state=some-random-state-foobar&nonce=some-random-nonce&code_challenge=p0v_UR0KrXl4--BpxM2BQa7qIW5k3k4WauBhjmkVQw8&code_challenge_method=S256"
		req := httptest.NewRequest(http.MethodGet, originalURL, nil)
		w := httptest.NewRecorder()

		authCtx := &mocks.AuthenticationContext{}
		oauth2Provider := &mocks.OAuth2Provider{}
		requester := fosite.NewAuthorizeRequest()
		oauth2Provider.OnNewAuthorizeRequest(req.Context(), req).Return(requester, nil)
		oauth2Provider.On("WriteAuthorizeError", w, requester, mock.Anything).Run(func(args mock.Arguments) {
			rw := args.Get(0).(http.ResponseWriter)
			rw.WriteHeader(http.StatusForbidden)
		})
		authCtx.OnOAuth2Provider().Return(oauth2Provider)

		cookieManager := &mocks.CookieHandler{}
		cookieManager.OnSetAuthCodeCookie(req.Context(), w, originalURL).Return(fmt.Errorf("failure injection"))
		authCtx.OnCookieManager().Return(cookieManager)

		authEndpoint(authCtx, w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

//func TestAuthCallbackEndpoint(t *testing.T) {
//	originalURL := "http://localhost:8088/oauth2/authorize?client_id=my-client&redirect_uri=http%3A%2F%2Flocalhost%3A3846%2Fcallback&response_type=code&scope=photos+openid+offline&state=some-random-state-foobar&nonce=some-random-nonce&code_challenge=p0v_UR0KrXl4--BpxM2BQa7qIW5k3k4WauBhjmkVQw8&code_challenge_method=S256"
//	req := httptest.NewRequest(http.MethodGet, originalURL, nil)
//	w := httptest.NewRecorder()
//
//	authCtx := &mocks.AuthenticationContext{}
//
//	oauth2Provider := &mocks.OAuth2Provider{}
//	requester := fosite.NewAuthorizeRequest()
//	oauth2Provider.OnNewAuthorizeRequest(req.Context(), req).Return(requester, nil)
//	oauth2Provider.On("WriteAuthorizeError", w, requester, mock.Anything).Run(func(args mock.Arguments) {
//		rw := args.Get(0).(http.ResponseWriter)
//		rw.WriteHeader(http.StatusForbidden)
//	})
//
//	authCtx.OnOAuth2Provider().Return(oauth2Provider)
//
//	cookieManager := &mocks.CookieHandler{}
//	cookieManager.OnSetAuthCodeCookie(req.Context(), w, originalURL).Return(nil)
//	cookieManager.OnRetrieveTokenValues(req.Context(), req).Return(sampleIDToken, "", "", nil)
//	cookieManager.OnRetrieveUserInfo(req.Context(), req).Return(&service.UserInfoResponse{Subject: "abc"}, nil)
//	authCtx.OnCookieManager().Return(cookieManager)
//
//	authCtx.OnOptions().Return(&config.Config{
//		UserAuth: config.UserAuthConfig{
//			OpenID: config.OpenIDOptions{
//				//ClientID: "http://localhost",
//			},
//		},
//	})
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	var issuer string
//	hf := func(w http.ResponseWriter, r *http.Request) {
//		if r.URL.Path == "/.well-known/openid-configuration" {
//			w.Header().Set("Content-Type", "application/json")
//			io.WriteString(w, strings.ReplaceAll(`{
//				"issuer": "ISSUER",
//				"authorization_endpoint": "https://example.com/auth",
//				"token_endpoint": "https://example.com/token",
//				"jwks_uri": "ISSUER/keys",
//				"id_token_signing_alg_values_supported": ["RS256"]
//			}`, "ISSUER", issuer))
//			return
//		} else if r.URL.Path == "/keys" {
//			w.Header().Set("Content-Type", "application/json")
//			io.WriteString(w, `{"keys":[{"kty":"RSA","alg":"RS256","kid":"Z6dmZ_TXhduw-jUBZ6uEEzvnh-jhNO0YhemB7qa_LOc","use":"sig","e":"AQAB","n":"jyMcudBiz7XqeDIvxfMlmG4fvAUU7cl3R4iSIv_ahHanCcVRvqcXOsIknwn7i4rOUjP6MlH45uIYsaj6MuLYgoaIbC-Z823Tu4asoC-rGbpZgf-bMcJLxtZVBNsSagr_M0n8xA1oogHRF1LGRiD93wNr2b9OkKVbWnyNdASk5_xui024nVzakm2-RAEyaC048nHfnjVBvwo4BdJVDgBEK03fbkBCyuaZyE1ZQF545MTbD4keCv58prSCmbDRJgRk48FzaFnQeYTho-pUxXxM9pvhMykeI62WZ7diDfIc9isOpv6ALFOHgKy7Ihhve6pLIylLRTnn2qhHFkGPtU3djQ"}]}`)
//			return
//		}
//
//		http.NotFound(w, r)
//		return
//
//	}
//
//	s := httptest.NewServer(http.HandlerFunc(hf))
//	defer s.Close()
//
//	issuer = s.URL
//	mockOidcProvider, err := oidc.NewProvider(ctx, issuer)
//	if !assert.NoError(t, err) {
//		t.FailNow()
//	}
//
//	authCtx.OnOidcProvider().Return(mockOidcProvider)
//
//	authCallbackEndpoint(authCtx, w, req)
//	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
//}

func TestGetIssuer(t *testing.T) {
	t.Run("SelfAuthServerIssuer wins", func(t *testing.T) {
		issuer := GetIssuer(context.Background(), nil, &config.Config{
			AppAuth: config.OAuth2Options{
				SelfAuthServer: config.AuthorizationServer{
					Issuer: "my_issuer",
				},
			},
			AuthorizedURIs: []config2.URL{{URL: *config.MustParseURL("http://localhost/")}},
		})

		assert.Equal(t, "my_issuer", issuer)
	})

	t.Run("Fallback to http public uri", func(t *testing.T) {
		issuer := GetIssuer(context.Background(), nil, &config.Config{
			AuthorizedURIs: []config2.URL{{URL: *config.MustParseURL("http://localhost/")}},
		})

		assert.Equal(t, "http://localhost/", issuer)
	})
}

func TestEncryptDecrypt(t *testing.T) {
	cookieHashKey := [auth.SymmetricKeyLength]byte{}
	_, err := rand.Read(cookieHashKey[:])
	assert.NoError(t, err)

	input := "hello world"
	encrypted, err := encryptString(input, cookieHashKey)
	assert.NoError(t, err)

	decrypted, err := decryptString(encrypted, cookieHashKey)
	assert.NoError(t, err)

	assert.Equal(t, input, decrypted)
	assert.NotEqual(t, input, encrypted)
}
