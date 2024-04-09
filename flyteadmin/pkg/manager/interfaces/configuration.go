package interfaces

import (
	"context"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

type ConfigurationInterface interface {
	GetConfiguration(ctx context.Context, request admin.ConfigurationGetRequest) (*admin.ConfigurationGetResponse, error)
	UpdateConfiguration(ctx context.Context, request admin.ConfigurationUpdateRequest, mergeActive bool) (*admin.ConfigurationUpdateResponse, error)
}
