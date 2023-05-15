package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"

	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	emptyIdentityContext = IdentityContext{}
)

type claimsType = map[string]interface{}

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
	// Raw JWT token from the IDP. Set to a pointer to support the equal operator for this struct.
	claims *claimsType
	// executionIdentity stores a unique string that can be used to identify the user associated with a given task.
	// This identifier is passed down to the ExecutionSpec and can be used for various purposes, such as setting the user identifier on a pod label.
	// By default, the execution user identifier is filled with the value of IdentityContext.userID. However, you can customize your middleware to assign other values if needed.
	// Providing a user identifier can be useful for tracking tasks and associating them with specific users, especially in multi-user environments.
	executionIdentity string
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

func (c IdentityContext) Claims() map[string]interface{} {
	if c.claims != nil {
		return *c.claims
	}
	return make(map[string]interface{})
}

func (c IdentityContext) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ContextKeyIdentityContext, c)
}

func (c IdentityContext) AuthenticatedAt() time.Time {
	return c.authenticatedAt
}

func (c IdentityContext) ExecutionIdentity() string {
	return c.executionIdentity
}

// WithExecutionUserIdentifier creates a copy of the original identity context and attach ExecutionIdentity
func (c IdentityContext) WithExecutionUserIdentifier(euid string) IdentityContext {
	c.executionIdentity = euid
	return c
}

// NewIdentityContext creates a new IdentityContext.
func NewIdentityContext(audience, userID, appID string, authenticatedAt time.Time, scopes sets.String, userInfo *service.UserInfoResponse, claims map[string]interface{}) (
	IdentityContext, error) {
	// For some reason, google IdP returns a subject in the ID Token but an empty subject in the /user_info endpoint
	if userInfo == nil {
		userInfo = &service.UserInfoResponse{}
	}

	if len(userInfo.Subject) == 0 {
		userInfo.Subject = userID
	}

	if len(claims) > 0 {
		claimsStruct, err := utils.MarshalObjToStruct(claims)
		if err != nil {
			return IdentityContext{}, fmt.Errorf("failed to marshal claims [%+v] to struct: %w", claims, err)
		}
		userInfo.AdditionalClaims = claimsStruct
	}

	return IdentityContext{
		audience:        audience,
		userID:          userID,
		appID:           appID,
		userInfo:        userInfo,
		authenticatedAt: authenticatedAt,
		scopes:          &scopes,
		claims:          &claims,
	}, nil
}

// IdentityContextFromContext retrieves the authenticated identity from context.Context.
func IdentityContextFromContext(ctx context.Context) IdentityContext {
	existing := ctx.Value(ContextKeyIdentityContext)
	if existing != nil {
		return existing.(IdentityContext)
	}

	return emptyIdentityContext
}
