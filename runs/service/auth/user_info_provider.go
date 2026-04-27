package auth

import (
	"context"

	"connectrpc.com/connect"

	authpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/auth"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/auth/authconnect"
)

// UserInfoProvider serves user info claims about the currently logged in user.
// See the OpenID Connect spec at https://openid.net/specs/openid-connect-core-1_0.html#UserInfoResponse
type UserInfoProvider struct {
	authconnect.UnimplementedIdentityServiceHandler
}

func NewUserInfoProvider() *UserInfoProvider {
	return &UserInfoProvider{}
}

func (s *UserInfoProvider) UserInfo(ctx context.Context, _ *connect.Request[authpb.UserInfoRequest]) (*connect.Response[authpb.UserInfoResponse], error) {
	identityContext := IdentityContextFromContext(ctx)
	if identityContext != nil {
		return connect.NewResponse(identityContext.UserInfo()), nil
	}
	return connect.NewResponse(&authpb.UserInfoResponse{}), nil
}
