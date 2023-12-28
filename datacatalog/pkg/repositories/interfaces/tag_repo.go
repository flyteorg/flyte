package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
)

//go:generate mockery -name=TagRepo -output=../mocks -case=underscore

type TagRepo interface {
	Create(ctx context.Context, datasetID *datacatalog.DatasetID, in models.Tag) error
	Get(ctx context.Context, datasetID *datacatalog.DatasetID, tagName string) (models.Tag, error)
}
