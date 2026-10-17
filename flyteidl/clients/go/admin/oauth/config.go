package oauth

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/oauth2"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
)

const (
	// TokenTypeBearer sends the OAuth2 access token to admin as "Bearer <token>". This is the default.
	TokenTypeBearer = "Bearer"
	// TokenTypeIDToken sends the OIDC id_token from the token response to admin as "IDToken <token>". Admin validates
	// tokens sent with this scheme against its configured userAuth OIDC provider rather than its authorization server.
	TokenTypeIDToken = "IDToken"

	idTokenKey = "id_token"
)

// Config oauth2.Config overridden with device endpoint for supporting Device Authorization Grant flow [RFC8268]
type Config struct {
	*oauth2.Config
	DeviceEndpoint string
	// Audience value to be passed when requesting access token using device flow.This needs to be passed in the first request of the device flow currently and is configured in admin public client config.Required when auth server hasn't been configured with default audience"`
	Audience string
	// TokenType is the scheme used when sending the token to admin: TokenTypeBearer (default) or TokenTypeIDToken.
	TokenType string
}

// NormalizeTokenType returns the canonical token type for a configured value. An empty value means TokenTypeBearer.
func NormalizeTokenType(tokenType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(tokenType)) {
	case "", "bearer":
		return TokenTypeBearer, nil
	case "idtoken", "id_token":
		return TokenTypeIDToken, nil
	default:
		return "", fmt.Errorf("unsupported tokenType %q, expected %v or %v", tokenType, TokenTypeBearer, TokenTypeIDToken)
	}
}

// PrepareToken adjusts a token returned by the authorization server so that it is sent to admin with the configured
// TokenType. For TokenTypeIDToken the id_token carried in the token response replaces the access token.
func (c *Config) PrepareToken(token *oauth2.Token) (*oauth2.Token, error) {
	if token == nil {
		return nil, nil
	}

	idToken, _ := token.Extra(idTokenKey).(string)
	return c.PrepareTokenWithIDToken(token, idToken)
}

// PrepareTokenWithIDToken is PrepareToken for callers that parsed the token response themselves and hold the id_token.
func (c *Config) PrepareTokenWithIDToken(token *oauth2.Token, idToken string) (*oauth2.Token, error) {
	if token == nil || c == nil || c.TokenType != TokenTypeIDToken {
		return token, nil
	}

	if len(idToken) == 0 {
		return nil, fmt.Errorf("tokenType is %v but the token response has no id_token; make sure the openid scope is requested", TokenTypeIDToken)
	}

	token.AccessToken = idToken
	token.TokenType = TokenTypeIDToken
	return token, nil
}

// BuildConfigFromMetadataService builds OAuth2 config from information retrieved through the anonymous auth metadata service.
func BuildConfigFromMetadataService(ctx context.Context, authMetadataClient service.AuthMetadataServiceClient) (clientConf *Config, err error) {
	var clientResp *service.PublicClientAuthConfigResponse
	if clientResp, err = authMetadataClient.GetPublicClientConfig(ctx, &service.PublicClientAuthConfigRequest{}); err != nil {
		return nil, err
	}

	var oauthMetaResp *service.OAuth2MetadataResponse
	if oauthMetaResp, err = authMetadataClient.GetOAuth2Metadata(ctx, &service.OAuth2MetadataRequest{}); err != nil {
		return nil, err
	}

	clientConf = &Config{
		Config: &oauth2.Config{
			ClientID:    clientResp.GetClientId(),
			RedirectURL: clientResp.GetRedirectUri(),
			Scopes:      clientResp.GetScopes(),
			Endpoint: oauth2.Endpoint{
				TokenURL: oauthMetaResp.GetTokenEndpoint(),
				AuthURL:  oauthMetaResp.GetAuthorizationEndpoint(),
			},
		},
		DeviceEndpoint: oauthMetaResp.GetDeviceAuthorizationEndpoint(),
		Audience:       clientResp.GetAudience(),
		TokenType:      TokenTypeBearer,
	}

	return clientConf, nil
}
