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
	"github.com/flyteorg/flytestdlib/config"
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

func TestCustomTokenSource_GetTokenSource(t *testing.T) {
	ctx := context.Background()
	cfg := GetConfig(ctx)
	cfg.TokenRefreshWindow = config.Duration{Duration: time.Minute}
	cfg.ClientSecretLocation = ""

	hourAhead := time.Now().Add(time.Hour)
	validToken := oauth2.Token{AccessToken: "foo", Expiry: hourAhead}

	tests := []struct {
		name  string
		token *oauth2.Token
	}{
		{
			name:  "no token",
			token: nil,
		},
		{

			name:  "valid token",
			token: &validToken,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tokenCache := &tokenCacheMocks.TokenCache{}
			var tokenErr error = nil
			if test.token == nil {
				tokenErr = fmt.Errorf("no token")
			}
			tokenCache.OnGetToken().Return(test.token, tokenErr).Once()
			provider, err := NewClientCredentialsTokenSourceProvider(ctx, cfg, []string{}, "", tokenCache, "")
			assert.NoError(t, err)

			source, err := provider.GetTokenSource(ctx)
			assert.NoError(t, err)
			customSource, ok := source.(*customTokenSource)
			assert.True(t, ok)

			if test.token == nil {
				assert.Equal(t, time.Time{}, customSource.refreshTime)
			} else {
				assert.LessOrEqual(t, customSource.refreshTime.Unix(), test.token.Expiry.Unix())
				assert.GreaterOrEqual(t, customSource.refreshTime.Unix(), test.token.Expiry.Add(-cfg.TokenRefreshWindow.Duration).Unix())
			}
		})
	}
}

func TestCustomTokenSource_fetchTokenFromCache(t *testing.T) {
	ctx := context.Background()
	cfg := GetConfig(ctx)
	cfg.TokenRefreshWindow = config.Duration{Duration: time.Minute}
	cfg.ClientSecretLocation = ""

	minuteAgo := time.Now().Add(-time.Minute)
	hourAhead := time.Now().Add(time.Hour)
	invalidToken := oauth2.Token{AccessToken: "foo", Expiry: minuteAgo}
	validToken := oauth2.Token{AccessToken: "foo", Expiry: hourAhead}

	tests := []struct {
		name               string
		refreshTime        time.Time
		failedToRefresh    bool
		token              *oauth2.Token
		expectToken        bool
		expectNeedsRefresh bool
	}{
		{
			name:               "no token",
			refreshTime:        hourAhead,
			failedToRefresh:    false,
			token:              nil,
			expectToken:        false,
			expectNeedsRefresh: false,
		},
		{
			name:               "invalid token",
			refreshTime:        hourAhead,
			failedToRefresh:    false,
			token:              &invalidToken,
			expectToken:        false,
			expectNeedsRefresh: false,
		},
		{
			name:               "refresh exceeded",
			refreshTime:        minuteAgo,
			failedToRefresh:    false,
			token:              &validToken,
			expectToken:        true,
			expectNeedsRefresh: true,
		},
		{
			name:               "refresh exceeded failed",
			refreshTime:        minuteAgo,
			failedToRefresh:    true,
			token:              &validToken,
			expectToken:        true,
			expectNeedsRefresh: false,
		},
		{
			name:               "valid token",
			refreshTime:        hourAhead,
			failedToRefresh:    false,
			token:              &validToken,
			expectToken:        true,
			expectNeedsRefresh: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tokenCache := &tokenCacheMocks.TokenCache{}
			var tokenErr error = nil
			if test.token == nil {
				tokenErr = fmt.Errorf("no token")
			}
			tokenCache.OnGetToken().Return(test.token, tokenErr).Twice()
			provider, err := NewClientCredentialsTokenSourceProvider(ctx, cfg, []string{}, "", tokenCache, "")
			assert.NoError(t, err)
			source, err := provider.GetTokenSource(ctx)
			assert.NoError(t, err)
			customSource, ok := source.(*customTokenSource)
			assert.True(t, ok)

			customSource.refreshTime = test.refreshTime
			customSource.failedToRefresh = test.failedToRefresh
			token, needsRefresh := customSource.fetchTokenFromCache()
			if test.expectToken {
				assert.NotNil(t, token)
			} else {
				assert.Nil(t, token)
			}
			assert.Equal(t, test.expectNeedsRefresh, needsRefresh)
		})
	}
}

func TestCustomTokenSource_Token(t *testing.T) {
	ctx := context.Background()
	cfg := GetConfig(ctx)
	cfg.TokenRefreshWindow = config.Duration{Duration: time.Minute}
	cfg.ClientSecretLocation = ""

	minuteAgo := time.Now().Add(-time.Minute)
	hourAhead := time.Now().Add(time.Hour)
	twoHourAhead := time.Now().Add(2 * time.Hour)
	invalidToken := oauth2.Token{AccessToken: "foo", Expiry: minuteAgo}
	validToken := oauth2.Token{AccessToken: "foo", Expiry: hourAhead}
	newToken := oauth2.Token{AccessToken: "foo", Expiry: twoHourAhead}

	tests := []struct {
		name            string
		refreshTime     time.Time
		failedToRefresh bool
		token           *oauth2.Token
		newToken        *oauth2.Token
		expectedToken   *oauth2.Token
	}{
		{
			name:            "cached token",
			refreshTime:     hourAhead,
			failedToRefresh: false,
			token:           &validToken,
			newToken:        nil,
			expectedToken:   &validToken,
		},
		{
			name:            "failed refresh still valid",
			refreshTime:     minuteAgo,
			failedToRefresh: false,
			token:           &validToken,
			newToken:        nil,
			expectedToken:   &validToken,
		},
		{
			name:            "failed refresh invalid",
			refreshTime:     minuteAgo,
			failedToRefresh: false,
			token:           &invalidToken,
			newToken:        nil,
			expectedToken:   nil,
		},
		{
			name:            "refresh",
			refreshTime:     minuteAgo,
			failedToRefresh: false,
			token:           &invalidToken,
			newToken:        &newToken,
			expectedToken:   &newToken,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tokenCache := &tokenCacheMocks.TokenCache{}
			tokenCache.OnGetToken().Return(test.token, nil).Twice()
			provider, err := NewClientCredentialsTokenSourceProvider(ctx, cfg, []string{}, "", tokenCache, "")
			assert.NoError(t, err)
			source, err := provider.GetTokenSource(ctx)
			assert.NoError(t, err)
			customSource, ok := source.(*customTokenSource)
			assert.True(t, ok)

			mockSource := &adminMocks.TokenSource{}
			if test.newToken != nil {
				mockSource.OnToken().Return(test.newToken, nil)
			} else {
				mockSource.OnToken().Return(nil, fmt.Errorf("refresh token failed"))
			}
			customSource.new = mockSource
			customSource.refreshTime = test.refreshTime
			customSource.failedToRefresh = test.failedToRefresh
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
		})
	}
}
