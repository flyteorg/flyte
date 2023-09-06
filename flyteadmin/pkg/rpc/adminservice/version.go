package adminservice

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

func (m *AdminService) GetVersion(ctx context.Context, request *admin.GetVersionRequest) (*admin.GetVersionResponse, error) {

	defer m.interceptPanic(ctx, request)
	response, err := m.VersionManager.GetVersion(ctx, request)
	if err != nil {
		return nil, err
	}
	return response, nil
}
