package deviceflow

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
	"k8s.io/apimachinery/pkg/util/json"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/cache"
	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/oauth"
	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/tokenorchestrator"
	"github.com/flyteorg/flyte/flytestdlib/config"
)

func TestFetchFromAuthFlow(t *testing.T) {
	ctx := context.Background()
	t.Run("fetch from auth flow", func(t *testing.T) {
		tokenCache := cache.NewTokenCacheInMemoryProvider()
		orchestrator, err := NewDeviceFlowTokenOrchestrator(tokenorchestrator.BaseTokenOrchestrator{
			ClientConfig: &oauth.Config{
				Config: &oauth2.Config{
					RedirectURL: "http://localhost:8089/redirect",
					Scopes:      []string{"code", "all"},
				},
				DeviceEndpoint: "http://dummyDeviceEndpoint",
			},
			TokenCache: tokenCache,
		}, Config{})
		assert.NoError(t, err)
		refreshedToken, err := orchestrator.FetchTokenFromAuthFlow(ctx)
		assert.Nil(t, refreshedToken)
		assert.NotNil(t, err)
	})

	t.Run("fetch from auth flow", func(t *testing.T) {
		fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			assert.Nil(t, err)
			isDeviceReq := strings.Contains(string(body), scope)
			isTokReq := strings.Contains(string(body), deviceCode) && strings.Contains(string(body), grantType) && strings.Contains(string(body), clientID)

			if isDeviceReq {
				for _, urlParm := range strings.Split(string(body), "&") {
					paramKeyValue := strings.Split(urlParm, "=")
					switch paramKeyValue[0] {
					case audience:
						assert.Equal(t, "abcd", paramKeyValue[1])
					case clientID:
						assert.Equal(t, clientID, paramKeyValue[1])
					}
				}
				dar := DeviceAuthorizationResponse{
					DeviceCode:              "e1db31fe-3b23-4fce-b759-82bf8ea323d6",
					UserCode:                "RPBQZNRX",
					VerificationURI:         "https://oauth-server/activate",
					VerificationURIComplete: "https://oauth-server/activate?user_code=RPBQZNRX",
					Interval:                5,
				}
				darBytes, err := json.Marshal(dar)
				assert.Nil(t, err)
				_, err = w.Write(darBytes)
				assert.Nil(t, err)
				return
			} else if isTokReq {
				for _, urlParm := range strings.Split(string(body), "&") {
					paramKeyValue := strings.Split(urlParm, "=")
					switch paramKeyValue[0] {
					case grantType:
						assert.Equal(t, url.QueryEscape(grantTypeValue), paramKeyValue[1])
					case deviceCode:
						assert.Equal(t, "e1db31fe-3b23-4fce-b759-82bf8ea323d6", paramKeyValue[1])
					case clientID:
						assert.Equal(t, clientID, paramKeyValue[1])
					}
				}
				dar := DeviceAccessTokenResponse{
					Token: oauth2.Token{
						AccessToken: "access_token",
					},
					ExpiresIn: 300,
				}
				darBytes, err := json.Marshal(dar)
				assert.Nil(t, err)
				_, err = w.Write(darBytes)
				assert.Nil(t, err)
				return
			}
			t.Fatal("unknown request")
		}))
		defer fakeServer.Close()

		tokenCache := cache.NewTokenCacheInMemoryProvider()
		orchestrator, err := NewDeviceFlowTokenOrchestrator(tokenorchestrator.BaseTokenOrchestrator{
			ClientConfig: &oauth.Config{
				Config: &oauth2.Config{
					ClientID:    clientID,
					RedirectURL: "http://localhost:8089/redirect",
					Scopes:      []string{"code", "all"},
					Endpoint: oauth2.Endpoint{
						TokenURL: fakeServer.URL,
					},
				},
				DeviceEndpoint: fakeServer.URL,
				Audience:       "abcd",
			},
			TokenCache: tokenCache,
		}, Config{
			Timeout: config.Duration{Duration: 1 * time.Minute},
		})
		assert.NoError(t, err)
		authToken, err := orchestrator.FetchTokenFromAuthFlow(ctx)
		assert.Nil(t, err)
		assert.NotNil(t, authToken)
		assert.Equal(t, "access_token", authToken.AccessToken)
		assert.True(t, authToken.Expiry.After(time.Now().Add(time.Second*200)))
	})
}

func TestFetchFromAuthFlowWithIDToken(t *testing.T) {
	ctx := context.Background()
	newServer := func(t *testing.T, tokenResponse string) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			assert.Nil(t, err)
			if strings.Contains(string(body), deviceCode) {
				_, err = w.Write([]byte(tokenResponse))
				assert.Nil(t, err)
				return
			}
			dar := DeviceAuthorizationResponse{DeviceCode: "device-code", UserCode: "USER-CODE", VerificationURI: "https://idp/device", Interval: 1}
			darBytes, err := json.Marshal(dar)
			assert.Nil(t, err)
			_, err = w.Write(darBytes)
			assert.Nil(t, err)
		}))
	}
	newOrchestrator := func(t *testing.T, server *httptest.Server, tokenType string) TokenOrchestrator {
		orchestrator, err := NewDeviceFlowTokenOrchestrator(tokenorchestrator.BaseTokenOrchestrator{
			ClientConfig: &oauth.Config{
				Config: &oauth2.Config{
					ClientID: clientID,
					Scopes:   []string{"openid", "offline_access"},
					Endpoint: oauth2.Endpoint{TokenURL: server.URL},
				},
				DeviceEndpoint: server.URL,
				TokenType:      tokenType,
			},
			TokenCache: cache.NewTokenCacheInMemoryProvider(),
		}, Config{Timeout: config.Duration{Duration: 1 * time.Minute}})
		assert.NoError(t, err)
		return orchestrator
	}

	t.Run("id token is sent as IDToken", func(t *testing.T) {
		server := newServer(t, `{"access_token": "access", "refresh_token": "refresh", "token_type": "bearer", "expires_in": 300, "id_token": "id"}`)
		defer server.Close()
		orchestrator := newOrchestrator(t, server, oauth.TokenTypeIDToken)
		authToken, err := orchestrator.FetchTokenFromAuthFlow(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "id", authToken.AccessToken)
		assert.Equal(t, "refresh", authToken.RefreshToken)
		assert.Equal(t, oauth.TokenTypeIDToken, authToken.Type())
		cached, err := orchestrator.TokenCache.GetToken()
		assert.NoError(t, err)
		assert.Equal(t, "id", cached.AccessToken)
	})

	t.Run("bearer keeps the access token", func(t *testing.T) {
		server := newServer(t, `{"access_token": "access", "token_type": "bearer", "expires_in": 300, "id_token": "id"}`)
		defer server.Close()
		orchestrator := newOrchestrator(t, server, oauth.TokenTypeBearer)
		authToken, err := orchestrator.FetchTokenFromAuthFlow(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "access", authToken.AccessToken)
		assert.Equal(t, "Bearer", authToken.Type())
	})

	t.Run("id token missing", func(t *testing.T) {
		server := newServer(t, `{"access_token": "access", "token_type": "bearer", "expires_in": 300}`)
		defer server.Close()
		orchestrator := newOrchestrator(t, server, oauth.TokenTypeIDToken)
		authToken, err := orchestrator.FetchTokenFromAuthFlow(ctx)
		assert.Error(t, err)
		assert.Nil(t, authToken)
	})
}
