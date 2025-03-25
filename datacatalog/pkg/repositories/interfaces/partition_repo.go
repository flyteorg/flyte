package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/models"
)

//go:generate mockery --name=PartitionRepo --output=../mocks --case=underscore --with-expecter

type PartitionRepo interface {
	Create(ctx context.Context, in models.Partition) error
}
