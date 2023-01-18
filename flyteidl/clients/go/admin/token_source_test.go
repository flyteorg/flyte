package admin

import (
	"context"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"

	tokenCacheMocks "github.com/flyteorg/flyteidl/clients/go/admin/cache/mocks"
	adminMocks "github.com/flyteorg/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
)

type DummyTestTokenSource struct {
	oauth2.TokenSource
}

func (d DummyTestTokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{
		AccessToken: "abc",
	}, nil
}

func TestNewTokenSource(t *testing.T) {
	tokenSource := DummyTestTokenSource{}
	flyteTokenSource := NewCustomHeaderTokenSource(tokenSource, true, "test")
	metadata, err := flyteTokenSource.GetRequestMetadata(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "Bearer abc", metadata["test"])
}

func TestNewTokenSourceProvider(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name                     string
		audienceCfg              string
		scopesCfg                []string
		useAudienceFromAdmin     bool
		clientConfigResponse     service.PublicClientAuthConfigResponse
		expectedAudience         string
		expectedScopes           []string
		expectedCallsPubEndpoint int
	}{
		{
			name:                     "audience from client config",
			audienceCfg:              "clientConfiguredAud",
			scopesCfg:                []string{"all"},
			clientConfigResponse:     service.PublicClientAuthConfigResponse{},
			expectedAudience:         "clientConfiguredAud",
			expectedScopes:           []string{"all"},
			expectedCallsPubEndpoint: 0,
		},
		{
			name:                     "audience from public client response",
			audienceCfg:              "clientConfiguredAud",
			useAudienceFromAdmin:     true,
			scopesCfg:                []string{"all"},
			clientConfigResponse:     service.PublicClientAuthConfigResponse{Audience: "AdminConfiguredAud", Scopes: []string{}},
			expectedAudience:         "AdminConfiguredAud",
			expectedScopes:           []string{"all"},
			expectedCallsPubEndpoint: 1,
		},

		{
			name:                     "audience from client with useAudience from admin false",
			audienceCfg:              "clientConfiguredAud",
			useAudienceFromAdmin:     false,
			scopesCfg:                []string{"all"},
			clientConfigResponse:     service.PublicClientAuthConfigResponse{Audience: "AdminConfiguredAud", Scopes: []string{}},
			expectedAudience:         "clientConfiguredAud",
			expectedScopes:           []string{"all"},
			expectedCallsPubEndpoint: 0,
		},
	}
	for _, test := range tests {
		cfg := GetConfig(ctx)
		tokenCache := &tokenCacheMocks.TokenCache{}
		metadataClient := &adminMocks.AuthMetadataServiceClient{}
		metadataClient.OnGetOAuth2MetadataMatch(mock.Anything, mock.Anything).Return(&service.OAuth2MetadataResponse{}, nil)
		metadataClient.OnGetPublicClientConfigMatch(mock.Anything, mock.Anything).Return(&test.clientConfigResponse, nil)
		cfg.AuthType = AuthTypeClientSecret
		cfg.Audience = test.audienceCfg
		cfg.Scopes = test.scopesCfg
		cfg.UseAudienceFromAdmin = test.useAudienceFromAdmin
		flyteTokenSource, err := NewTokenSourceProvider(ctx, cfg, tokenCache, metadataClient)
		assert.True(t, metadataClient.AssertNumberOfCalls(t, "GetPublicClientConfig", test.expectedCallsPubEndpoint))
		assert.NoError(t, err)
		assert.NotNil(t, flyteTokenSource)
		clientCredSourceProvider, ok := flyteTokenSource.(ClientCredentialsTokenSourceProvider)
		assert.True(t, ok)
		assert.Equal(t, test.expectedScopes, clientCredSourceProvider.ccConfig.Scopes)
		assert.Equal(t, url.Values{audienceKey: {test.expectedAudience}}, clientCredSourceProvider.ccConfig.EndpointParams)
	}
}
