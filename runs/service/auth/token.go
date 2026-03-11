package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/metadata"

	authpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/auth"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

// GetRefreshedToken refreshes a JWT using the provided OAuth2 config and refresh token.
func GetRefreshedToken(ctx context.Context, oauth *oauth2.Config, accessToken, refreshToken string) (*oauth2.Token, error) {
	logger.Debugf(ctx, "Attempting to refresh token")
	originalToken := oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Expiry:       time.Now().Add(-1 * time.Minute), // force expired by setting to the past
	}

	tokenSource := oauth.TokenSource(ctx, &originalToken)
	newToken, err := tokenSource.Token()
	if err != nil {
		logger.Errorf(ctx, "Error refreshing token %s", err)
		return nil, fmt.Errorf("error refreshing token: %w", err)
	}

	return newToken, nil
}

// ParseIDTokenAndValidate parses and validates an ID token using the OIDC provider.
func ParseIDTokenAndValidate(ctx context.Context, clientID, rawIDToken string, provider *oidc.Provider) (*oidc.IDToken, error) {
	cfg := &oidc.Config{
		ClientID: clientID,
	}

	if len(clientID) == 0 {
		cfg.SkipClientIDCheck = true
		cfg.SkipIssuerCheck = true
		cfg.SkipExpiryCheck = true
	}

	verifier := provider.Verifier(cfg)

	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		logger.Debugf(ctx, "JWT parsing with claims failed %s", err)
		if strings.Contains(err.Error(), "token is expired") {
			return idToken, fmt.Errorf("token is expired: %w", err)
		}
		return idToken, fmt.Errorf("jwt parse with claims failed: %w", err)
	}
	return idToken, nil
}

// GRPCGetIdentityFromAccessToken attempts to extract a bearer token from gRPC metadata
// and validate it using the provided resource server.
func GRPCGetIdentityFromAccessToken(ctx context.Context, expectedAudience string, resourceServer OAuth2ResourceServer) (
	*IdentityContext, error) {

	tokenStr, err := bearerTokenFromMD(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve bearer token from metadata: %w", err)
	}

	return resourceServer.ValidateAccessToken(ctx, expectedAudience, tokenStr)
}

// GRPCGetIdentityFromIDToken attempts to extract an ID token from gRPC metadata and validate it.
func GRPCGetIdentityFromIDToken(ctx context.Context, clientID string, provider *oidc.Provider) (
	*IdentityContext, error) {

	tokenStr, err := idTokenFromMD(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve id token from metadata: %w", err)
	}

	return IdentityContextFromIDToken(ctx, tokenStr, clientID, provider, nil)
}

// IdentityContextFromIDToken creates an IdentityContext from a validated ID token.
func IdentityContextFromIDToken(ctx context.Context, tokenStr, clientID string, provider *oidc.Provider,
	userInfo *authpb.UserInfoResponse) (*IdentityContext, error) {

	idToken, err := ParseIDTokenAndValidate(ctx, clientID, tokenStr, provider)
	if err != nil {
		return nil, err
	}
	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		logger.Infof(ctx, "Failed to unmarshal claims from id token, err: %v", err)
	}

	return NewIdentityContext(idToken.Audience[0], idToken.Subject, "", idToken.IssuedAt,
		[]string{ScopeAll}, userInfo, claims), nil
}

func NewOAuthTokenFromRaw(accessToken, refreshToken, idToken string) *oauth2.Token {
	return (&oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}).WithExtra(map[string]interface{}{
		idTokenExtra: idToken,
	})
}

func ExtractTokensFromOauthToken(token *oauth2.Token) (idToken, accessToken, refreshToken string, err error) {
	if token == nil {
		return "", "", "", fmt.Errorf("attempting to set cookies with nil token")
	}

	idTokenRaw, converted := token.Extra(idTokenExtra).(string)
	if !converted {
		return "", "", "", fmt.Errorf("response does not contain an id_token")
	}

	return idTokenRaw, token.AccessToken, token.RefreshToken, nil
}

// bearerTokenFromMD extracts a Bearer token from gRPC incoming metadata.
func bearerTokenFromMD(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("no metadata in context")
	}

	vals := md.Get(DefaultAuthorizationHeader)
	if len(vals) == 0 {
		return "", fmt.Errorf("no authorization header in metadata")
	}

	header := vals[0]
	prefix := BearerScheme + " "
	if !strings.HasPrefix(header, prefix) {
		return "", fmt.Errorf("authorization header does not start with %q", BearerScheme)
	}

	token := strings.TrimPrefix(header, prefix)
	if token == "" {
		return "", fmt.Errorf("bearer token is blank")
	}

	return token, nil
}

// idTokenFromMD extracts an IDToken from gRPC incoming metadata.
func idTokenFromMD(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("no metadata in context")
	}

	vals := md.Get(DefaultAuthorizationHeader)
	if len(vals) == 0 {
		return "", fmt.Errorf("no authorization header in metadata")
	}

	header := vals[0]
	prefix := IDTokenScheme + " "
	if !strings.HasPrefix(header, prefix) {
		return "", fmt.Errorf("authorization header does not start with %q", IDTokenScheme)
	}

	token := strings.TrimPrefix(header, prefix)
	if token == "" {
		return "", fmt.Errorf("id token is blank")
	}

	return token, nil
}

