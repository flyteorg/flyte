package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/models"
)


type PartitionRepo interface {
	Create(ctx context.Context, in models.Partition) error
}
