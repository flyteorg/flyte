package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
)

type ConfigurationInterface interface {
	GetActive(ctx context.Context) (models.Configuration, error)
	EraseActiveAndCreate(ctx context.Context, activeVersion string, input models.Configuration) error
}
