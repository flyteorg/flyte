package auth

import "context"

// OAuth2ResourceServer represents a resource server that can validate access tokens.
type OAuth2ResourceServer interface {
	ValidateAccessToken(ctx context.Context, expectedAudience, tokenStr string) (*IdentityContext, error)
}
