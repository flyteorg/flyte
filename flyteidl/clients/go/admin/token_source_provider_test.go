package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"

	tokenCacheMocks "github.com/flyteorg/flyte/flyteidl/clients/go/admin/cache/mocks"
	adminMocks "github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
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
		metadataClient.On("GetOAuth2Metadata", mock.Anything, mock.Anything).Return(&service.OAuth2MetadataResponse{}, nil)
		metadataClient.On("GetPublicClientConfig", mock.Anything, mock.Anything).Return(test.clientConfigResponse, nil)
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
			tokenCache.On("GetToken").Return(test.token, nil).Maybe()
			tokenCache.On("Lock").Return().Maybe()
			tokenCache.On("Unlock").Return().Maybe()
			provider, err := NewClientCredentialsTokenSourceProvider(ctx, cfg, []string{}, "", tokenCache, "")
			assert.NoError(t, err)
			source, err := provider.GetTokenSource(ctx)
			assert.NoError(t, err)
			customSource, ok := source.(*customTokenSource)
			assert.True(t, ok)

			fetchCalled := false
			customSource.fetchNewToken = func() (*oauth2.Token, error) {
				fetchCalled = true
				if test.newToken != nil {
					return test.newToken, nil
				}
				return nil, fmt.Errorf("refresh token failed")
			}
			if test.newToken != nil {
				tokenCache.On("SaveToken", test.newToken).Return(nil).Once()
			}
			token, err := source.Token()
			if test.expectedToken != nil {
				assert.Equal(t, test.expectedToken, token)
				assert.NoError(t, err)
			} else {
				assert.Nil(t, token)
				assert.Error(t, err)
			}
			if test.token == validToken {
				assert.False(t, fetchCalled)
			} else {
				assert.True(t, fetchCalled)
			}
			tokenCache.AssertExpectations(t)
		})
	}
}

func TestCustomTokenSource_RefreshPathBypassesInnerReuseCache(t *testing.T) {
	var callCount atomic.Int32
	transport := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		nextCall := callCount.Add(1)

		token := jwt.New(jwt.SigningMethodHS256)
		claims := token.Claims.(jwt.MapClaims)
		// Keep JWT exp within the 15m validity buffer so utils.Valid() rejects it
		// immediately and forces refresh logic on every Token() call.
		claims["exp"] = time.Now().Add(5 * time.Minute).Unix()
		claims["jti"] = fmt.Sprintf("token-%d", nextCall)
		tokenString, err := token.SignedString([]byte("test-secret"))
		if err != nil {
			return nil, err
		}

		response := map[string]interface{}{
			"access_token": tokenString,
			"token_type":   "bearer",
			// Long expires_in keeps oauth2's internal Token.Expiry far in the future.
			"expires_in": 24 * 60 * 60,
		}
		raw, err := json.Marshal(response)
		if err != nil {
			return nil, err
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body:    io.NopCloser(bytes.NewReader(raw)),
			Request: r,
		}, nil
	})
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{Transport: transport})

	cfg := GetConfig(ctx)
	cfg.ClientID = "test-client"
	cfg.ClientSecretLocation = ""
	cfg.ClientSecretEnvVar = "FLYTE_ADMIN_CLIENT_SECRET_FOR_TEST"
	cfg.PerRetryTimeout.Duration = time.Millisecond
	cfg.MaxRetries = 0
	t.Setenv(cfg.ClientSecretEnvVar, "super-secret")

	tokenCache := &tokenCacheMocks.TokenCache{}
	var cachedToken *oauth2.Token
	tokenCache.On("GetToken").Return(
		func() *oauth2.Token {
			return cachedToken
		},
		func() error {
			if cachedToken == nil {
				return fmt.Errorf("cannot find token in cache")
			}
			return nil
		},
	).Twice()
	tokenCache.On("SaveToken", mock.AnythingOfType("*oauth2.Token")).Return(
		func(token *oauth2.Token) error {
			cachedToken = token
			return nil
		},
	).Twice()

	provider, err := NewClientCredentialsTokenSourceProvider(ctx, cfg, []string{"all"}, "https://example.test/oauth2/token", tokenCache, "")
	assert.NoError(t, err)

	source, err := provider.GetTokenSource(ctx)
	assert.NoError(t, err)

	_, err = source.Token()
	assert.NoError(t, err)
	_, err = source.Token()
	assert.NoError(t, err)

	// Regression assertion: second Token() should force a fresh token request.
	// Old implementation (ccConfig.TokenSource + oauth2.ReuseTokenSource) only
	// called token endpoint once and replayed stale token from inner cache.
	assert.Equal(t, int32(2), callCount.Load())
	tokenCache.AssertExpectations(t)
}

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
