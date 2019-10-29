package interfaces

import (
	"context"

	"github.com/lyft/datacatalog/pkg/repositories/models"
)

type PartitionRepo interface {
	Create(ctx context.Context, in models.Partition) error
}
