package service

import (
	"context"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/auth"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/auth/authconnect"
)

// IdentityService implements the IdentityServiceHandler interface.
type IdentityService struct{}

// NewIdentityService creates a new IdentityService instance.
func NewIdentityService() *IdentityService {
	return &IdentityService{}
}

var _ authconnect.IdentityServiceHandler = (*IdentityService)(nil)

// UserInfo returns information about the currently logged in user.
// TODO: Wire with real auth to populate user info from the authenticated context.
func (s *IdentityService) UserInfo(
	ctx context.Context,
	req *connect.Request[auth.UserInfoRequest],
) (*connect.Response[auth.UserInfoResponse], error) {
	return connect.NewResponse(&auth.UserInfoResponse{}), nil
}
