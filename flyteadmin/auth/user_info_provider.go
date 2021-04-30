package auth

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
)

type UserInfoProvider struct {
}

// UserInfo returns user_info claims about the currently logged in user.
// See the OpenID Connect spec at https://openid.net/specs/openid-connect-core-1_0.html#UserInfoResponse for more information.
func (s UserInfoProvider) UserInfo(ctx context.Context, _ *service.UserInfoRequest) (*service.UserInfoResponse, error) {
	identityContext := IdentityContextFromContext(ctx)
	return identityContext.UserInfo(), nil
}

func NewUserInfoProvider() UserInfoProvider {
	return UserInfoProvider{}
}
