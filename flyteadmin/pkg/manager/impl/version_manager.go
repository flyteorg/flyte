package impl

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	adminversion "github.com/flyteorg/flytestdlib/version"
)

type VersionManager struct {
	Version   string
	Build     string
	BuildTime string
}

func (v *VersionManager) GetVersion(ctx context.Context, r *admin.GetVersionRequest) (*admin.GetVersionResponse, error) {
	return &admin.GetVersionResponse{
		ControlPlaneVersion: &admin.Version{
			Version:   v.Version,
			Build:     v.Build,
			BuildTime: v.BuildTime,
		},
	}, nil
}

func NewVersionManager() interfaces.VersionInterface {
	return &VersionManager{
		Build:     adminversion.Build,
		Version:   adminversion.Version,
		BuildTime: adminversion.BuildTime,
	}
}
