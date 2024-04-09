package interfaces

import (
	"context"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

type ConfigurationInterface interface {
	GetConfiguration(ctx context.Context, request admin.ConfigurationGetRequest) (*admin.ConfigurationGetResponse, error)
	UpdateConfiguration(ctx context.Context, request admin.ConfigurationUpdateRequest, mergeActive bool) (*admin.ConfigurationUpdateResponse, error)
	GetUpdateInput(ctx context.Context, inputConfiguration *admin.Configuration, project, domain, workflow, launchPlan string) (*interfaces.UpdateConfigurationInput, error)
}
