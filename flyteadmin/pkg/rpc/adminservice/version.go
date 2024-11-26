package adminservice

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

func (m *AdminService) GetVersion(ctx context.Context, request *admin.GetVersionRequest) (*admin.GetVersionResponse, error) {

	response, err := m.VersionManager.GetVersion(ctx, request)
	if err != nil {
		return nil, err
	}
	return response, nil
}
