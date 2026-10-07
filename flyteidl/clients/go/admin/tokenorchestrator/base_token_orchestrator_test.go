package tokenorchestrator

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/cache"
	cacheMocks "github.com/flyteorg/flyte/flyteidl/clients/go/admin/cache/mocks"
	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/oauth"
	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/utils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flyte/flytestdlib/config"
)

func TestRefreshTheToken(t *testing.T) {
	ctx := context.Background()
	clientConf := &oauth.Config{
		Config: &oauth2.Config{
			ClientID: "dummyClient",
		},
	}
	tokenCacheProvider := cache.NewTokenCacheInMemoryProvider()
	orchestrator := BaseTokenOrchestrator{
		ClientConfig: clientConf,
		TokenCache:   tokenCacheProvider,
	}

	t.Run("bad url in Config", func(t *testing.T) {
		tokenData := utils.GenTokenWithCustomExpiry(t, time.Now().Add(-20*time.Minute))
		refreshedToken, err := orchestrator.RefreshToken(ctx, tokenData)
		assert.Nil(t, refreshedToken)
		assert.NotNil(t, err)
	})
}

func TestFetchFromCache(t *testing.T) {
	ctx := context.Background()
	metadata := &service.OAuth2MetadataResponse{
		TokenEndpoint:   "/token",
		ScopesSupported: []string{"code", "all"},
	}
	clientMetatadata := &service.PublicClientAuthConfigResponse{
		AuthorizationMetadataKey: "flyte_authorization",
		RedirectUri:              "http://localhost:8089/redirect",
	}
	mockAuthClient := new(mocks.AuthMetadataServiceClient)
	mockAuthClient.EXPECT().GetOAuth2Metadata(mock.Anything, mock.Anything).Return(metadata, nil)
	mockAuthClient.EXPECT().GetPublicClientConfig(mock.Anything, mock.Anything).Return(clientMetatadata, nil)

	t.Run("no token in cache", func(t *testing.T) {
		tokenCacheProvider := cache.NewTokenCacheInMemoryProvider()

		orchestrator, err := NewBaseTokenOrchestrator(ctx, tokenCacheProvider, mockAuthClient)

		assert.NoError(t, err)
		refreshedToken, err := orchestrator.FetchTokenFromCacheOrRefreshIt(ctx, config.Duration{Duration: 5 * time.Minute})
		assert.Nil(t, refreshedToken)
		assert.NotNil(t, err)
	})

	t.Run("token in cache", func(t *testing.T) {
		tokenCacheProvider := cache.NewTokenCacheInMemoryProvider()
		orchestrator, err := NewBaseTokenOrchestrator(ctx, tokenCacheProvider, mockAuthClient)
		assert.NoError(t, err)
		tokenData := utils.GenTokenWithCustomExpiry(t, time.Now().Add(20*time.Minute))
		err = tokenCacheProvider.SaveToken(tokenData)
		assert.Nil(t, err)
		cachedToken, err := orchestrator.FetchTokenFromCacheOrRefreshIt(ctx, config.Duration{Duration: 5 * time.Minute})
		assert.Nil(t, err)
		assert.NotNil(t, cachedToken)
		assert.Equal(t, tokenData.AccessToken, cachedToken.AccessToken)
	})

	t.Run("expired token in cache", func(t *testing.T) {
		tokenCacheProvider := cache.NewTokenCacheInMemoryProvider()
		orchestrator, err := NewBaseTokenOrchestrator(ctx, tokenCacheProvider, mockAuthClient)
		assert.NoError(t, err)
		tokenData := utils.GenTokenWithCustomExpiry(t, time.Now().Add(-20*time.Minute))
		err = tokenCacheProvider.SaveToken(tokenData)
		assert.Nil(t, err)
		_, err = orchestrator.FetchTokenFromCacheOrRefreshIt(ctx, config.Duration{Duration: 5 * time.Minute})
		assert.NotNil(t, err)
	})

	t.Run("token fetch before grace period", func(t *testing.T) {
		mockTokenCacheProvider := new(cacheMocks.TokenCache)
		orchestrator, err := NewBaseTokenOrchestrator(ctx, mockTokenCacheProvider, mockAuthClient)
		assert.NoError(t, err)
		tokenData := utils.GenTokenWithCustomExpiry(t, time.Now().Add(20*time.Minute))
		mockTokenCacheProvider.EXPECT().GetToken().Return(tokenData, nil)
		mockTokenCacheProvider.EXPECT().SaveToken(mock.Anything).Return(nil)
		assert.Nil(t, err)
		refreshedToken, err := orchestrator.FetchTokenFromCacheOrRefreshIt(ctx, config.Duration{Duration: 5 * time.Minute})
		assert.Nil(t, err)
		assert.NotNil(t, refreshedToken)
		mockTokenCacheProvider.AssertNotCalled(t, "SaveToken")
	})

	t.Run("token fetch after grace period with refresh", func(t *testing.T) {
		mockTokenCacheProvider := new(cacheMocks.TokenCache)
		orchestrator, err := NewBaseTokenOrchestrator(ctx, mockTokenCacheProvider, mockAuthClient)
		assert.NoError(t, err)
		tokenData := utils.GenTokenWithCustomExpiry(t, time.Now().Add(20*time.Minute))
		mockTokenCacheProvider.EXPECT().GetToken().Return(tokenData, nil)
		assert.Nil(t, err)
		refreshedToken, err := orchestrator.FetchTokenFromCacheOrRefreshIt(ctx, config.Duration{Duration: 5 * time.Minute})
		assert.Nil(t, err)
		assert.NotNil(t, refreshedToken)
		mockTokenCacheProvider.AssertNotCalled(t, "SaveToken")
	})
}

func TestRefreshTheTokenWithIDToken(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NoError(t, r.ParseForm())
		assert.Equal(t, "refresh_token", r.Form.Get("grant_type"))
		assert.Equal(t, "old-refresh", r.Form.Get("refresh_token"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token": "new-access", "refresh_token": "new-refresh", "token_type": "bearer", "expires_in": 3600, "id_token": "new-id"}`))
	}))
	defer server.Close()

	tokenCacheProvider := cache.NewTokenCacheInMemoryProvider()
	orchestrator := BaseTokenOrchestrator{
		ClientConfig: &oauth.Config{
			Config:    &oauth2.Config{ClientID: "flytectl", Endpoint: oauth2.Endpoint{TokenURL: server.URL}},
			TokenType: oauth.TokenTypeIDToken,
		},
		TokenCache: tokenCacheProvider,
	}
	expired := &oauth2.Token{AccessToken: "old-id", RefreshToken: "old-refresh", TokenType: oauth.TokenTypeIDToken, Expiry: time.Now().Add(-time.Hour)}
	refreshed, err := orchestrator.RefreshToken(ctx, expired)
	assert.NoError(t, err)
	assert.Equal(t, "new-id", refreshed.AccessToken)
	assert.Equal(t, "new-refresh", refreshed.RefreshToken)
	assert.Equal(t, oauth.TokenTypeIDToken, refreshed.Type())
	cached, err := tokenCacheProvider.GetToken()
	assert.NoError(t, err)
	assert.Equal(t, "new-id", cached.AccessToken)
}
