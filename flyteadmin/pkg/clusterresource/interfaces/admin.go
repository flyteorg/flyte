package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

//go:generate mockery -name FlyteAdminDataProvider -output=../mocks -case=underscore

type FlyteAdminDataProvider interface {
	GetClusterResourceAttributes(ctx context.Context, identifier common.ResourceScope) (*admin.ClusterResourceAttributes, error)
	GetProjects(ctx context.Context) (*admin.Projects, error)
}
