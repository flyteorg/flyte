package admin

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"

	tokenCacheMocks "github.com/flyteorg/flyteidl/clients/go/admin/cache/mocks"
	adminMocks "github.com/flyteorg/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
)

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

func TestCustomTokenSource_Token(t *testing.T) {
	ctx := context.Background()
	cfg := GetConfig(ctx)
	cfg.ClientSecretLocation = ""

	minuteAgo := time.Now().Add(-time.Minute)
	hourAhead := time.Now().Add(time.Hour)
	twoHourAhead := time.Now().Add(2 * time.Hour)
	invalidToken := oauth2.Token{AccessToken: "foo", Expiry: minuteAgo}
	validToken := oauth2.Token{AccessToken: "foo", Expiry: hourAhead}
	newToken := oauth2.Token{AccessToken: "foo", Expiry: twoHourAhead}

	tests := []struct {
		name          string
		token         *oauth2.Token
		newToken      *oauth2.Token
		expectedToken *oauth2.Token
	}{
		{
			name:          "no cached token",
			token:         nil,
			newToken:      &newToken,
			expectedToken: &newToken,
		},
		{
			name:          "cached token valid",
			token:         &validToken,
			newToken:      nil,
			expectedToken: &validToken,
		},
		{
			name:          "cached token expired",
			token:         &invalidToken,
			newToken:      &newToken,
			expectedToken: &newToken,
		},
		{
			name:          "failed new token",
			token:         &invalidToken,
			newToken:      nil,
			expectedToken: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tokenCache := &tokenCacheMocks.TokenCache{}
			tokenCache.OnGetToken().Return(test.token, nil).Once()
			provider, err := NewClientCredentialsTokenSourceProvider(ctx, cfg, []string{}, "", tokenCache, "")
			assert.NoError(t, err)
			source, err := provider.GetTokenSource(ctx)
			assert.NoError(t, err)
			customSource, ok := source.(*customTokenSource)
			assert.True(t, ok)

			mockSource := &adminMocks.TokenSource{}
			if test.token != &validToken {
				if test.newToken != nil {
					mockSource.OnToken().Return(test.newToken, nil)
				} else {
					mockSource.OnToken().Return(nil, fmt.Errorf("refresh token failed"))
				}
			}
			customSource.new = mockSource
			if test.newToken != nil {
				tokenCache.OnSaveToken(test.newToken).Return(nil).Once()
			}
			token, err := source.Token()
			if test.expectedToken != nil {
				assert.Equal(t, test.expectedToken, token)
				assert.NoError(t, err)
			} else {
				assert.Nil(t, token)
				assert.Error(t, err)
			}
			tokenCache.AssertExpectations(t)
			mockSource.AssertExpectations(t)
		})
	}
}
