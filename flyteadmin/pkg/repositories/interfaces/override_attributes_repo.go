package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
)

type OverrideAttributesInterface interface {
	GetActive(ctx context.Context) (models.OverrideAttributes, error)
	EraseActive(ctx context.Context) error
	Create(ctx context.Context, input models.OverrideAttributes) error
}
