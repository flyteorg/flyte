package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

type OverrideAttributesInterface interface {
	GetOverrideAttributes(ctx context.Context, request admin.OverrideAttributesGetRequest) (*admin.OverrideAttributesGetResponse, error)
	UpdateOverrideAttributes(ctx context.Context, request admin.OverrideAttributesUpdateRequest) (*admin.OverrideAttributesUpdateResponse, error)
}
