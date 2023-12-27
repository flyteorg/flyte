package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
)

//go:generate mockery -name=DatasetRepo -output=../mocks -case=underscore

type DatasetRepo interface {
	Create(ctx context.Context, id *datacatalog.DatasetID, in models.Dataset) error
	Get(ctx context.Context, id *datacatalog.DatasetID) (models.Dataset, error)
	List(ctx context.Context, in models.ListModelsInput) ([]models.Dataset, error)
}
