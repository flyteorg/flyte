package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

type FlyteAdminDataProvider interface {
	GetClusterResourceAttributes(ctx context.Context, project, domain string) (*admin.ClusterResourceAttributes, error)
	GetProjects(ctx context.Context) (*admin.Projects, error)
}
