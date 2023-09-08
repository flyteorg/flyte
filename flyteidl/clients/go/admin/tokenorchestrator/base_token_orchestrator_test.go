package tokenorchestrator

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"

	"github.com/flyteorg/flyteidl/clients/go/admin/cache"
	cacheMocks "github.com/flyteorg/flyteidl/clients/go/admin/cache/mocks"
	"github.com/flyteorg/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyteidl/clients/go/admin/oauth"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flytestdlib/config"
)

func TestRefreshTheToken(t *testing.T) {
	ctx := context.Background()
	clientConf := &oauth.Config{
		Config: &oauth2.Config{
			ClientID: "dummyClient",
		},
	}
	tokenCacheProvider := &cache.TokenCacheInMemoryProvider{}
	orchestrator := BaseTokenOrchestrator{
		ClientConfig: clientConf,
		TokenCache:   tokenCacheProvider,
	}

	plan, _ := os.ReadFile("testdata/token.json")
	var tokenData oauth2.Token
	err := json.Unmarshal(plan, &tokenData)
	assert.Nil(t, err)
	t.Run("bad url in Config", func(t *testing.T) {
		refreshedToken, err := orchestrator.RefreshToken(ctx, &tokenData)
		assert.Nil(t, refreshedToken)
		assert.NotNil(t, err)
	})
}

func TestFetchFromCache(t *testing.T) {
	ctx := context.Background()
	metatdata := &service.OAuth2MetadataResponse{
		TokenEndpoint:   "/token",
		ScopesSupported: []string{"code", "all"},
	}
	clientMetatadata := &service.PublicClientAuthConfigResponse{
		AuthorizationMetadataKey: "flyte_authorization",
		RedirectUri:              "http://localhost:8089/redirect",
	}
	mockAuthClient := new(mocks.AuthMetadataServiceClient)
	mockAuthClient.OnGetOAuth2MetadataMatch(mock.Anything, mock.Anything).Return(metatdata, nil)
	mockAuthClient.OnGetPublicClientConfigMatch(mock.Anything, mock.Anything).Return(clientMetatadata, nil)

	t.Run("no token in cache", func(t *testing.T) {
		tokenCacheProvider := &cache.TokenCacheInMemoryProvider{}

		orchestrator, err := NewBaseTokenOrchestrator(ctx, tokenCacheProvider, mockAuthClient)

		assert.NoError(t, err)
		refreshedToken, err := orchestrator.FetchTokenFromCacheOrRefreshIt(ctx, config.Duration{Duration: 5 * time.Minute})
		assert.Nil(t, refreshedToken)
		assert.NotNil(t, err)
	})

	t.Run("token in cache", func(t *testing.T) {
		tokenCacheProvider := &cache.TokenCacheInMemoryProvider{}
		orchestrator, err := NewBaseTokenOrchestrator(ctx, tokenCacheProvider, mockAuthClient)
		assert.NoError(t, err)
		fileData, _ := os.ReadFile("testdata/token.json")
		var tokenData oauth2.Token
		err = json.Unmarshal(fileData, &tokenData)
		assert.Nil(t, err)
		tokenData.Expiry = time.Now().Add(20 * time.Minute)
		err = tokenCacheProvider.SaveToken(&tokenData)
		assert.Nil(t, err)
		cachedToken, err := orchestrator.FetchTokenFromCacheOrRefreshIt(ctx, config.Duration{Duration: 5 * time.Minute})
		assert.Nil(t, err)
		assert.NotNil(t, cachedToken)
		assert.Equal(t, tokenData.AccessToken, cachedToken.AccessToken)
	})

	t.Run("expired token in cache", func(t *testing.T) {
		tokenCacheProvider := &cache.TokenCacheInMemoryProvider{}
		orchestrator, err := NewBaseTokenOrchestrator(ctx, tokenCacheProvider, mockAuthClient)
		assert.NoError(t, err)
		fileData, _ := os.ReadFile("testdata/token.json")
		var tokenData oauth2.Token
		err = json.Unmarshal(fileData, &tokenData)
		assert.Nil(t, err)
		tokenData.Expiry = time.Now().Add(-20 * time.Minute)
		err = tokenCacheProvider.SaveToken(&tokenData)
		assert.Nil(t, err)
		_, err = orchestrator.FetchTokenFromCacheOrRefreshIt(ctx, config.Duration{Duration: 5 * time.Minute})
		assert.NotNil(t, err)
	})

	t.Run("token fetch before grace period", func(t *testing.T) {
		mockTokenCacheProvider := new(cacheMocks.TokenCache)
		orchestrator, err := NewBaseTokenOrchestrator(ctx, mockTokenCacheProvider, mockAuthClient)
		assert.NoError(t, err)
		fileData, _ := os.ReadFile("testdata/token.json")
		var tokenData oauth2.Token
		err = json.Unmarshal(fileData, &tokenData)
		assert.Nil(t, err)
		tokenData.Expiry = time.Now().Add(20 * time.Minute)
		mockTokenCacheProvider.OnGetTokenMatch(mock.Anything).Return(&tokenData, nil)
		mockTokenCacheProvider.OnSaveTokenMatch(mock.Anything).Return(nil)
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
		fileData, _ := os.ReadFile("testdata/token.json")
		var tokenData oauth2.Token
		err = json.Unmarshal(fileData, &tokenData)
		assert.Nil(t, err)
		tokenData.Expiry = time.Now().Add(20 * time.Minute)
		mockTokenCacheProvider.OnGetTokenMatch(mock.Anything).Return(&tokenData, nil)
		assert.Nil(t, err)
		refreshedToken, err := orchestrator.FetchTokenFromCacheOrRefreshIt(ctx, config.Duration{Duration: 5 * time.Minute})
		assert.Nil(t, err)
		assert.NotNil(t, refreshedToken)
		mockTokenCacheProvider.AssertNotCalled(t, "SaveToken")
	})
}
