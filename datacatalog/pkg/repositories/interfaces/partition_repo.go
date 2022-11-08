package interfaces

import (
	"context"

	"github.com/flyteorg/datacatalog/pkg/repositories/models"
)

//go:generate mockery -name=PartitionRepo -output=../mocks -case=underscore

type PartitionRepo interface {
	Create(ctx context.Context, in models.Partition) error
}
