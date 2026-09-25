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

	tokenCacheMocks "github.com/flyteorg/flyte/flyteidl/clients/go/admin/cache/mocks"
	adminMocks "github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/oauth"
	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/utils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
)

func TestNewTokenSourceProvider(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name                     string
		audienceCfg              string
		scopesCfg                []string
		useAudienceFromAdmin     bool
		clientConfigResponse     *service.PublicClientAuthConfigResponse
		expectedAudience         string
		expectedScopes           []string
		expectedCallsPubEndpoint int
	}{
		{
			name:                     "audience from client config",
			audienceCfg:              "clientConfiguredAud",
			scopesCfg:                []string{"all"},
			clientConfigResponse:     &service.PublicClientAuthConfigResponse{},
			expectedAudience:         "clientConfiguredAud",
			expectedScopes:           []string{"all"},
			expectedCallsPubEndpoint: 0,
		},
		{
			name:                     "audience from public client response",
			audienceCfg:              "clientConfiguredAud",
			useAudienceFromAdmin:     true,
			scopesCfg:                []string{"all"},
			clientConfigResponse:     &service.PublicClientAuthConfigResponse{Audience: "AdminConfiguredAud", Scopes: []string{}},
			expectedAudience:         "AdminConfiguredAud",
			expectedScopes:           []string{"all"},
			expectedCallsPubEndpoint: 1,
		},

		{
			name:                     "audience from client with useAudience from admin false",
			audienceCfg:              "clientConfiguredAud",
			useAudienceFromAdmin:     false,
			scopesCfg:                []string{"all"},
			clientConfigResponse:     &service.PublicClientAuthConfigResponse{Audience: "AdminConfiguredAud", Scopes: []string{}},
			expectedAudience:         "clientConfiguredAud",
			expectedScopes:           []string{"all"},
			expectedCallsPubEndpoint: 0,
		},
	}
	for _, test := range tests {
		cfg := GetConfig(ctx)
		tokenCache := &tokenCacheMocks.TokenCache{}
		metadataClient := &adminMocks.AuthMetadataServiceClient{}
		metadataClient.EXPECT().GetOAuth2Metadata(mock.Anything, mock.Anything).Return(&service.OAuth2MetadataResponse{}, nil)
		metadataClient.EXPECT().GetPublicClientConfig(mock.Anything, mock.Anything).Return(test.clientConfigResponse, nil)
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
	invalidToken := utils.GenTokenWithCustomExpiry(t, minuteAgo)
	validToken := utils.GenTokenWithCustomExpiry(t, hourAhead)
	newToken := utils.GenTokenWithCustomExpiry(t, twoHourAhead)

	tests := []struct {
		name          string
		token         *oauth2.Token
		newToken      *oauth2.Token
		expectedToken *oauth2.Token
	}{
		{
			name:          "no cached token",
			token:         nil,
			newToken:      newToken,
			expectedToken: newToken,
		},
		{
			name:          "cached token valid",
			token:         validToken,
			newToken:      nil,
			expectedToken: validToken,
		},
		{
			name:          "cached token expired",
			token:         invalidToken,
			newToken:      newToken,
			expectedToken: newToken,
		},
		{
			name:          "failed new token",
			token:         invalidToken,
			newToken:      nil,
			expectedToken: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tokenCache := &tokenCacheMocks.TokenCache{}
			tokenCache.EXPECT().GetToken().Return(test.token, nil).Maybe()
			tokenCache.EXPECT().Lock().Return().Maybe()
			tokenCache.EXPECT().Unlock().Return().Maybe()
			provider, err := NewClientCredentialsTokenSourceProvider(ctx, cfg, []string{}, "", tokenCache, "")
			assert.NoError(t, err)
			source, err := provider.GetTokenSource(ctx)
			assert.NoError(t, err)
			customSource, ok := source.(*customTokenSource)
			assert.True(t, ok)

			mockSource := &adminMocks.TokenSource{}
			if test.token != validToken {
				if test.newToken != nil {
					mockSource.EXPECT().Token().Return(test.newToken, nil)
				} else {
					mockSource.EXPECT().Token().Return(nil, fmt.Errorf("refresh token failed"))
				}
			}
			customSource.new = mockSource
			if test.newToken != nil {
				tokenCache.EXPECT().SaveToken(test.newToken).Return(nil).Once()
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

const (
	testIdpClientID     = "flytectl"
	testAdminAuthURL    = "https://admin/oauth2/authorize"
	testAdminTokenURL   = "https://admin/oauth2/token" //nolint:gosec // test fixture URL, not a credential
	testScopeOpenID     = "openid"
	testScopeOffline    = "offline"
	testCommandArgument = "token-from-command"
	testEcho            = "echo"
	testIdpAuthURL      = "https://idp/auth"
	testIdpTokenURL     = "https://idp/token" //nolint:gosec // test fixture URL, not a credential
	testIdpDeviceURL    = "https://idp/device/code"
	testIgnored         = "ignored"
)

func TestNewTokenSourceProvider_TokenType(t *testing.T) {
	ctx := context.Background()

	t.Run("external command defaults to bearer", func(t *testing.T) {
		cfg := &Config{AuthType: AuthTypeExternalCommand, Command: []string{testEcho, testCommandArgument}}
		provider, err := NewTokenSourceProvider(ctx, cfg, nil, nil)
		assert.NoError(t, err)
		tokenSource, err := provider.GetTokenSource(ctx)
		assert.NoError(t, err)
		token, err := tokenSource.Token()
		assert.NoError(t, err)
		assert.Equal(t, testCommandArgument, token.AccessToken)
		assert.Equal(t, "Bearer", token.Type())
	})

	t.Run("external command with IDToken", func(t *testing.T) {
		cfg := &Config{AuthType: AuthTypeExternalCommand, Command: []string{testEcho, "id-token-from-command"}, TokenType: oauth.TokenTypeIDToken}
		provider, err := NewTokenSourceProvider(ctx, cfg, nil, nil)
		assert.NoError(t, err)
		tokenSource, err := provider.GetTokenSource(ctx)
		assert.NoError(t, err)
		token, err := tokenSource.Token()
		assert.NoError(t, err)
		assert.Equal(t, "id-token-from-command", token.AccessToken)
		assert.Equal(t, oauth.TokenTypeIDToken, token.Type())
	})

	t.Run("unsupported token type", func(t *testing.T) {
		cfg := &Config{AuthType: AuthTypeExternalCommand, Command: []string{testEcho, "x"}, TokenType: "MAC"}
		_, err := NewTokenSourceProvider(ctx, cfg, nil, nil)
		assert.Error(t, err)
	})

	t.Run("device flow against explicitly configured endpoints", func(t *testing.T) {
		metadataClient := &adminMocks.AuthMetadataServiceClient{}
		metadataClient.EXPECT().GetOAuth2Metadata(mock.Anything, mock.Anything).Return(&service.OAuth2MetadataResponse{
			AuthorizationEndpoint: testAdminAuthURL, TokenEndpoint: testAdminTokenURL}, nil)
		metadataClient.EXPECT().GetPublicClientConfig(mock.Anything, mock.Anything).Return(&service.PublicClientAuthConfigResponse{
			ClientId: "flytectl-on-admin", RedirectUri: "http://localhost:53593/callback", Scopes: []string{testScopeOffline, "all"}, Audience: "admin-audience"}, nil)

		cfg := &Config{
			AuthType:               AuthTypeDeviceFlow,
			DeviceAuthorizationURL: testIdpDeviceURL,
			TokenURL:               testIdpTokenURL,
			ClientID:               "flytectl-on-idp",
			Scopes:                 []string{testScopeOpenID, "offline_access"},
			TokenType:              oauth.TokenTypeIDToken,
		}
		provider, err := NewTokenSourceProvider(ctx, cfg, &tokenCacheMocks.TokenCache{}, metadataClient)
		assert.NoError(t, err)
		deviceFlowProvider, ok := provider.(DeviceFlowTokenSourceProvider)
		assert.True(t, ok)
		clientConfig := deviceFlowProvider.tokenOrchestrator.ClientConfig
		assert.Equal(t, testIdpDeviceURL, clientConfig.DeviceEndpoint)
		assert.Equal(t, testIdpTokenURL, clientConfig.Endpoint.TokenURL)
		assert.Equal(t, "flytectl-on-idp", clientConfig.ClientID)
		assert.Equal(t, []string{testScopeOpenID, "offline_access"}, clientConfig.Scopes)
		assert.Equal(t, "admin-audience", clientConfig.Audience)
		assert.Equal(t, "http://localhost:53593/callback", clientConfig.RedirectURL)
		assert.Equal(t, oauth.TokenTypeIDToken, clientConfig.TokenType)
	})

	t.Run("pkce against explicitly configured endpoints", func(t *testing.T) {
		metadataClient := &adminMocks.AuthMetadataServiceClient{}
		metadataClient.EXPECT().GetOAuth2Metadata(mock.Anything, mock.Anything).Return(&service.OAuth2MetadataResponse{}, nil)
		metadataClient.EXPECT().GetPublicClientConfig(mock.Anything, mock.Anything).Return(&service.PublicClientAuthConfigResponse{}, nil)
		cfg := &Config{AuthType: AuthTypePkce, AuthorizationURL: testIdpAuthURL, TokenURL: testIdpTokenURL, ClientID: testIdpClientID, Scopes: []string{testScopeOpenID}, Audience: "idp-audience"}
		provider, err := NewTokenSourceProvider(ctx, cfg, &tokenCacheMocks.TokenCache{}, metadataClient)
		assert.NoError(t, err)
		clientConfig := provider.(PKCETokenSourceProvider).tokenOrchestrator.ClientConfig
		assert.Equal(t, testIdpAuthURL, clientConfig.Endpoint.AuthURL)
		assert.Equal(t, testIdpTokenURL, clientConfig.Endpoint.TokenURL)
		assert.Equal(t, testIdpClientID, clientConfig.ClientID)
		assert.Equal(t, "idp-audience", clientConfig.Audience)
		assert.Equal(t, oauth.TokenTypeBearer, clientConfig.TokenType)
	})

	t.Run("client secret rejects IDToken", func(t *testing.T) {
		cfg := &Config{AuthType: AuthTypeClientSecret, TokenType: oauth.TokenTypeIDToken, TokenURL: testAdminTokenURL, Scopes: []string{"all"}}
		_, err := NewTokenSourceProvider(ctx, cfg, &tokenCacheMocks.TokenCache{}, &adminMocks.AuthMetadataServiceClient{})
		assert.Error(t, err)
	})

	t.Run("explicit endpoints require tokenUrl", func(t *testing.T) {
		metadataClient := &adminMocks.AuthMetadataServiceClient{}
		metadataClient.EXPECT().GetOAuth2Metadata(mock.Anything, mock.Anything).Return(&service.OAuth2MetadataResponse{}, nil)
		metadataClient.EXPECT().GetPublicClientConfig(mock.Anything, mock.Anything).Return(&service.PublicClientAuthConfigResponse{}, nil)
		cfg := &Config{AuthType: AuthTypeDeviceFlow, DeviceAuthorizationURL: testIdpDeviceURL, ClientID: testIdpClientID, Scopes: []string{testScopeOpenID}}
		_, err := NewTokenSourceProvider(ctx, cfg, &tokenCacheMocks.TokenCache{}, metadataClient)
		assert.ErrorContains(t, err, "tokenUrl")
	})

	t.Run("explicit endpoints require scopes", func(t *testing.T) {
		metadataClient := &adminMocks.AuthMetadataServiceClient{}
		metadataClient.EXPECT().GetOAuth2Metadata(mock.Anything, mock.Anything).Return(&service.OAuth2MetadataResponse{}, nil)
		metadataClient.EXPECT().GetPublicClientConfig(mock.Anything, mock.Anything).Return(&service.PublicClientAuthConfigResponse{}, nil)
		cfg := &Config{AuthType: AuthTypePkce, AuthorizationURL: testIdpAuthURL, TokenURL: testIdpTokenURL, ClientID: testIdpClientID}
		_, err := NewTokenSourceProvider(ctx, cfg, &tokenCacheMocks.TokenCache{}, metadataClient)
		assert.ErrorContains(t, err, "scopes")
	})

	t.Run("a stale tokenUrl alone does not change pkce or device flow", func(t *testing.T) {
		metadataClient := &adminMocks.AuthMetadataServiceClient{}
		metadataClient.EXPECT().GetOAuth2Metadata(mock.Anything, mock.Anything).Return(&service.OAuth2MetadataResponse{
			TokenEndpoint: testAdminTokenURL, DeviceAuthorizationEndpoint: "https://admin/oauth2/device"}, nil)
		metadataClient.EXPECT().GetPublicClientConfig(mock.Anything, mock.Anything).Return(&service.PublicClientAuthConfigResponse{
			ClientId: testIdpClientID, Scopes: []string{testScopeOffline, "all"}}, nil)
		cfg := &Config{AuthType: AuthTypeDeviceFlow, TokenURL: testIdpTokenURL, ClientID: testIgnored, Scopes: []string{testIgnored}}
		provider, err := NewTokenSourceProvider(ctx, cfg, &tokenCacheMocks.TokenCache{}, metadataClient)
		assert.NoError(t, err)
		clientConfig := provider.(DeviceFlowTokenSourceProvider).tokenOrchestrator.ClientConfig
		assert.Equal(t, testAdminTokenURL, clientConfig.Endpoint.TokenURL)
		assert.Equal(t, "https://admin/oauth2/device", clientConfig.DeviceEndpoint)
		assert.Equal(t, testIdpClientID, clientConfig.ClientID)
	})

	t.Run("admin discovery is unchanged without a separate authorization server", func(t *testing.T) {
		metadataClient := &adminMocks.AuthMetadataServiceClient{}
		metadataClient.EXPECT().GetOAuth2Metadata(mock.Anything, mock.Anything).Return(&service.OAuth2MetadataResponse{
			TokenEndpoint: testAdminTokenURL, DeviceAuthorizationEndpoint: "https://admin/oauth2/device"}, nil)
		metadataClient.EXPECT().GetPublicClientConfig(mock.Anything, mock.Anything).Return(&service.PublicClientAuthConfigResponse{
			ClientId: testIdpClientID, Scopes: []string{testScopeOffline, "all"}}, nil)
		cfg := &Config{AuthType: AuthTypeDeviceFlow, ClientID: testIgnored, Scopes: []string{testIgnored}}
		provider, err := NewTokenSourceProvider(ctx, cfg, &tokenCacheMocks.TokenCache{}, metadataClient)
		assert.NoError(t, err)
		clientConfig := provider.(DeviceFlowTokenSourceProvider).tokenOrchestrator.ClientConfig
		assert.Equal(t, testIdpClientID, clientConfig.ClientID)
		assert.Equal(t, []string{testScopeOffline, "all"}, clientConfig.Scopes)
		assert.Equal(t, "https://admin/oauth2/device", clientConfig.DeviceEndpoint)
		assert.Equal(t, oauth.TokenTypeBearer, clientConfig.TokenType)
	})
}
