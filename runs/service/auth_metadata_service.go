package service

import (
	"context"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/auth"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/auth/authconnect"
)

// AuthMetadataService implements the AuthMetadataServiceHandler interface.
type AuthMetadataService struct {
	authconnect.UnimplementedAuthMetadataServiceHandler
	dataplaneDomain string
}

// NewAuthMetadataService creates a new AuthMetadataService instance.
func NewAuthMetadataService(dataplaneDomain string) *AuthMetadataService {
	return &AuthMetadataService{
		dataplaneDomain: dataplaneDomain,
	}
}

var _ authconnect.AuthMetadataServiceHandler = (*AuthMetadataService)(nil)

func (s *AuthMetadataService) GetPublicClientConfig(
	ctx context.Context,
	req *connect.Request[auth.GetPublicClientConfigRequest],
) (*connect.Response[auth.GetPublicClientConfigResponse], error) {
	return connect.NewResponse(&auth.GetPublicClientConfigResponse{
		DataplaneDomain: s.dataplaneDomain,
	}), nil
}