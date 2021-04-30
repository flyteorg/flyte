package auth

import (
	"context"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"

	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	emptyIdentityContext = IdentityContext{}
)

// IdentityContext is an abstract entity to enclose the authenticated identity of the user/app. Both gRPC and HTTP
// servers have interceptors to set the IdentityContext on the context.Context.
// To retrieve the current IdentityContext call auth.IdentityContextFromContext(ctx).
// To check whether there is an identity set, call auth.IdentityContextFromContext(ctx).IsEmpty()
type IdentityContext struct {
	audience        string
	userID          string
	appID           string
	authenticatedAt time.Time
	userInfo        *service.UserInfoResponse
	// Set to pointer just to keep this struct go-simple to support equal operator
	scopes *sets.String
}

func (c IdentityContext) Audience() string {
	return c.audience
}

func (c IdentityContext) UserID() string {
	return c.userID
}

func (c IdentityContext) AppID() string {
	return c.appID
}

func (c IdentityContext) UserInfo() *service.UserInfoResponse {
	if c.userInfo == nil {
		return &service.UserInfoResponse{}
	}

	return c.userInfo
}

func (c IdentityContext) IsEmpty() bool {
	return c == emptyIdentityContext
}

func (c IdentityContext) Scopes() sets.String {
	if c.scopes != nil {
		return *c.scopes
	}

	return sets.NewString()
}

func (c IdentityContext) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ContextKeyIdentityContext, c)
}

func (c IdentityContext) AuthenticatedAt() time.Time {
	return c.authenticatedAt
}

// NewIdentityContext creates a new IdentityContext.
func NewIdentityContext(audience, userID, appID string, authenticatedAt time.Time, scopes sets.String, userInfo *service.UserInfoResponse) IdentityContext {
	// For some reason, google IdP returns a subject in the ID Token but an empty subject in the /user_info endpoint
	if userInfo == nil {
		userInfo = &service.UserInfoResponse{}
	}

	if len(userInfo.Subject) == 0 {
		userInfo.Subject = userID
	}

	return IdentityContext{
		audience:        audience,
		userID:          userID,
		appID:           appID,
		userInfo:        userInfo,
		authenticatedAt: authenticatedAt,
		scopes:          &scopes,
	}
}

// IdentityContextFromContext retrieves the authenticated identity from context.Context.
func IdentityContextFromContext(ctx context.Context) IdentityContext {
	existing := ctx.Value(ContextKeyIdentityContext)
	if existing != nil {
		return existing.(IdentityContext)
	}

	return emptyIdentityContext
}
