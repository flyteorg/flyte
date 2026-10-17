package oauth

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
)

func TestGenerateClientConfig(t *testing.T) {
	ctx := context.Background()
	mockAuthClient := new(mocks.AuthMetadataServiceClient)
	flyteClientResp := &service.PublicClientAuthConfigResponse{
		ClientId:    "dummyClient",
		RedirectUri: "dummyRedirectUri",
		Scopes:      []string{"dummyScopes"},
		Audience:    "dummyAudience",
	}
	oauthMetaDataResp := &service.OAuth2MetadataResponse{
		Issuer:                        "dummyIssuer",
		AuthorizationEndpoint:         "dummyAuthEndPoint",
		TokenEndpoint:                 "dummyTokenEndpoint",
		CodeChallengeMethodsSupported: []string{"dummyCodeChallenege"},
		DeviceAuthorizationEndpoint:   "dummyDeviceEndpoint",
	}
	mockAuthClient.EXPECT().GetPublicClientConfig(ctx, mock.Anything).Return(flyteClientResp, nil)
	mockAuthClient.EXPECT().GetOAuth2Metadata(ctx, mock.Anything).Return(oauthMetaDataResp, nil)
	oauthConfig, err := BuildConfigFromMetadataService(ctx, mockAuthClient)
	assert.Nil(t, err)
	assert.NotNil(t, oauthConfig)
	assert.Equal(t, "dummyClient", oauthConfig.ClientID)
	assert.Equal(t, "dummyRedirectUri", oauthConfig.RedirectURL)
	assert.Equal(t, "dummyTokenEndpoint", oauthConfig.Endpoint.TokenURL)
	assert.Equal(t, "dummyAuthEndPoint", oauthConfig.Endpoint.AuthURL)
	assert.Equal(t, "dummyDeviceEndpoint", oauthConfig.DeviceEndpoint)
	assert.Equal(t, "dummyAudience", oauthConfig.Audience)
}

const (
	testAccessToken = "access"
	testIDToken     = "id"
	testBearerLower = "bearer"
)

func TestNormalizeTokenType(t *testing.T) {
	tests := []struct {
		in       string
		expected string
		wantErr  bool
	}{
		{in: "", expected: TokenTypeBearer},
		{in: testBearerLower, expected: TokenTypeBearer},
		{in: " Bearer ", expected: TokenTypeBearer},
		{in: "IDToken", expected: TokenTypeIDToken},
		{in: "idtoken", expected: TokenTypeIDToken},
		{in: idTokenKey, expected: TokenTypeIDToken},
		{in: "MAC", wantErr: true},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%q", test.in), func(t *testing.T) {
			got, err := NormalizeTokenType(test.in)
			if test.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestPrepareToken(t *testing.T) {
	t.Run("bearer leaves the token alone", func(t *testing.T) {
		cfg := &Config{Config: &oauth2.Config{}, TokenType: TokenTypeBearer}
		token := (&oauth2.Token{AccessToken: testAccessToken, TokenType: testBearerLower}).WithExtra(map[string]interface{}{idTokenKey: testIDToken})
		got, err := cfg.PrepareToken(token)
		assert.NoError(t, err)
		assert.Equal(t, testAccessToken, got.AccessToken)
		assert.Equal(t, "Bearer", got.Type())
	})

	t.Run("id token replaces the access token", func(t *testing.T) {
		cfg := &Config{Config: &oauth2.Config{}, TokenType: TokenTypeIDToken}
		token := (&oauth2.Token{AccessToken: testAccessToken, RefreshToken: "refresh", TokenType: testBearerLower}).WithExtra(map[string]interface{}{idTokenKey: testIDToken})
		got, err := cfg.PrepareToken(token)
		assert.NoError(t, err)
		assert.Equal(t, testIDToken, got.AccessToken)
		assert.Equal(t, "refresh", got.RefreshToken)
		assert.Equal(t, TokenTypeIDToken, got.TokenType)
		assert.Equal(t, TokenTypeIDToken, got.Type())
	})

	t.Run("id token missing from the response", func(t *testing.T) {
		cfg := &Config{Config: &oauth2.Config{}, TokenType: TokenTypeIDToken}
		got, err := cfg.PrepareToken(&oauth2.Token{AccessToken: testAccessToken})
		assert.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("explicit id token", func(t *testing.T) {
		cfg := &Config{Config: &oauth2.Config{}, TokenType: TokenTypeIDToken}
		got, err := cfg.PrepareTokenWithIDToken(&oauth2.Token{AccessToken: testAccessToken}, testIDToken)
		assert.NoError(t, err)
		assert.Equal(t, testIDToken, got.AccessToken)
		assert.Equal(t, TokenTypeIDToken, got.TokenType)
	})

	t.Run("nil token", func(t *testing.T) {
		cfg := &Config{Config: &oauth2.Config{}, TokenType: TokenTypeIDToken}
		got, err := cfg.PrepareToken(nil)
		assert.NoError(t, err)
		assert.Nil(t, got)
	})
}
