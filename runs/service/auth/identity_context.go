package auth

import (
	"context"
	"time"

	authpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/auth"
)

type contextKey string

const contextKeyIdentityContext contextKey = "identity_context"

// ScopeAll is the default scope granted to user-only access tokens.
const ScopeAll = "all"

// IdentityContext encloses the authenticated identity of the user/app. Both gRPC and HTTP
// servers have interceptors to set the IdentityContext on the context.AuthenticationContext.
type IdentityContext struct {
	audience        string
	userID          string
	appID           string
	authenticatedAt time.Time
	userInfo        *authpb.UserInfoResponse
	scopes          []string
	claims          map[string]interface{}
}

// NewIdentityContext creates a new IdentityContext.
func NewIdentityContext(audience, userID, appID string, authenticatedAt time.Time, scopes []string, userInfo *authpb.UserInfoResponse, claims map[string]interface{}) *IdentityContext {
	if userInfo == nil {
		userInfo = &authpb.UserInfoResponse{}
	}

	if len(userInfo.Subject) == 0 {
		userInfo.Subject = userID
	}

	return &IdentityContext{
		audience:        audience,
		userID:          userID,
		appID:           appID,
		authenticatedAt: authenticatedAt,
		userInfo:        userInfo,
		scopes:          scopes,
		claims:          claims,
	}
}

func (c *IdentityContext) Audience() string               { return c.audience }
func (c *IdentityContext) UserID() string                 { return c.userID }
func (c *IdentityContext) AppID() string                  { return c.appID }
func (c *IdentityContext) AuthenticatedAt() time.Time     { return c.authenticatedAt }
func (c *IdentityContext) Scopes() []string               { return c.scopes }
func (c *IdentityContext) Claims() map[string]interface{} { return c.claims }

func (c *IdentityContext) UserInfo() *authpb.UserInfoResponse {
	if c.userInfo == nil {
		return &authpb.UserInfoResponse{}
	}
	return c.userInfo
}

func (c *IdentityContext) IsEmpty() bool {
	return c == nil || (c.audience == "" && c.userID == "" && c.appID == "")
}

// WithContext stores the IdentityContext in the given context.
func (c *IdentityContext) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKeyIdentityContext, c)
}

// IdentityContextFromContext retrieves the authenticated identity from context.AuthenticationContext.
func IdentityContextFromContext(ctx context.Context) *IdentityContext {
	existing := ctx.Value(contextKeyIdentityContext)
	if existing != nil {
		return existing.(*IdentityContext)
	}
	return nil
}
