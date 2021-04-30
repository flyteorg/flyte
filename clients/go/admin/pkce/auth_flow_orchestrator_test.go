package pkce

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/flyteorg/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"
)

func TestRefreshTheToken(t *testing.T) {
	ctx := context.Background()
	clientConf := &oauth2.Config{
		ClientID: "dummyClient",
	}
	orchestrator := TokenOrchestrator{
		clientConfig: clientConf,
	}
	plan, _ := ioutil.ReadFile("testdata/token.json")
	var tokenData oauth2.Token
	err := json.Unmarshal(plan, &tokenData)
	assert.Nil(t, err)
	t.Run("bad url in config", func(t *testing.T) {
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
	orchestrator, err := NewTokenOrchestrator(ctx, Config{}, &TokenCacheInMemoryProvider{}, mockAuthClient)
	assert.NoError(t, err)

	t.Run("no token in cache", func(t *testing.T) {
		refreshedToken, err := orchestrator.FetchTokenFromCacheOrRefreshIt(ctx)
		assert.Nil(t, refreshedToken)
		assert.NotNil(t, err)
	})
}

func TestFetchFromAuthFlow(t *testing.T) {
	ctx := context.Background()
	t.Run("fetch from auth flow", func(t *testing.T) {
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
		tokenCache := &TokenCacheInMemoryProvider{}
		orchestrator, err := NewTokenOrchestrator(ctx, Config{}, tokenCache, mockAuthClient)
		assert.NoError(t, err)
		refreshedToken, err := orchestrator.FetchTokenFromAuthFlow(ctx)
		assert.Nil(t, refreshedToken)
		assert.NotNil(t, err)
	})
}
