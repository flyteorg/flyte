package deviceflow

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/json"

	"github.com/flyteorg/flytestdlib/config"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"

	"github.com/flyteorg/flyteidl/clients/go/admin/cache"
	"github.com/flyteorg/flyteidl/clients/go/admin/oauth"
	"github.com/flyteorg/flyteidl/clients/go/admin/tokenorchestrator"
)

func TestFetchFromAuthFlow(t *testing.T) {
	ctx := context.Background()
	t.Run("fetch from auth flow", func(t *testing.T) {
		tokenCache := &cache.TokenCacheInMemoryProvider{}
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
			isTokReq := strings.Contains(string(body), deviceCode) && strings.Contains(string(body), grantType) && strings.Contains(string(body), cliendID)

			if isDeviceReq {
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
				dar := DeviceAccessTokenResponse{
					Token: oauth2.Token{
						AccessToken: "access_token",
					},
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

		tokenCache := &cache.TokenCacheInMemoryProvider{}
		orchestrator, err := NewDeviceFlowTokenOrchestrator(tokenorchestrator.BaseTokenOrchestrator{
			ClientConfig: &oauth.Config{
				Config: &oauth2.Config{
					ClientID:    cliendID,
					RedirectURL: "http://localhost:8089/redirect",
					Scopes:      []string{"code", "all"},
					Endpoint: oauth2.Endpoint{
						TokenURL: fakeServer.URL,
					},
				},
				DeviceEndpoint: fakeServer.URL,
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
	})
}
