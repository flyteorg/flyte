package interfaces

import (
	"context"

	"github.com/flyteorg/datacatalog/pkg/repositories/models"
)

//go:generate mockery -name=DatasetRepo -output=../mocks -case=underscore

type DatasetRepo interface {
	Create(ctx context.Context, in models.Dataset) error
	Get(ctx context.Context, in models.DatasetKey) (models.Dataset, error)
	List(ctx context.Context, in models.ListModelsInput) ([]models.Dataset, error)
}
