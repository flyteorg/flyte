package interfaces

import (
	"context"

	"github.com/lyft/datacatalog/pkg/repositories/models"
)

type DatasetRepo interface {
	Create(ctx context.Context, in models.Dataset) error
	Get(ctx context.Context, in models.DatasetKey) (models.Dataset, error)
}
