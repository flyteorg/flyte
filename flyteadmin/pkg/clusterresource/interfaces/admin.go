package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/clusterresource/plugin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

//go:generate mockery -name FlyteAdminDataProvider -output=../mocks -case=underscore

type FlyteAdminDataProvider interface {
	GetClusterResourceAttributes(ctx context.Context, org, project, domain string) (*admin.ClusterResourceAttributes, error)
	GetProjects(ctx context.Context, clusterResourcePlugin plugin.ClusterResourcePlugin) (*admin.Projects, error)
	GetArchivedProjects(ctx context.Context, clusterResourcePlugin plugin.ClusterResourcePlugin) (*admin.Projects, error)
}
