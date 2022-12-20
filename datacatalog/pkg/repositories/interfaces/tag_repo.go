package interfaces

import (
	"context"

	"github.com/flyteorg/datacatalog/pkg/repositories/models"
)

//go:generate mockery -name=TagRepo -output=../mocks -case=underscore

type TagRepo interface {
	Create(ctx context.Context, in models.Tag) error
	Get(ctx context.Context, in models.TagKey) (models.Tag, error)
}
