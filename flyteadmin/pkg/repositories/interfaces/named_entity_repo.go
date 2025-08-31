package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

type GetNamedEntityInput struct {
	ResourceType core.ResourceType
	Project      string
	Domain       string
	Name         string
}

// Parameters for querying multiple resources.
type ListNamedEntityInput struct {
	ListResourceInput
	Project      string
	Domain       string
	ResourceType core.ResourceType
}

type NamedEntityCollectionOutput struct {
	Entities []models.NamedEntity
}

//go:generate mockery --name=NamedEntityRepoInterface --output=../mocks --case=underscore --with-expecter

// Defines the interface for interacting with NamedEntity models
type NamedEntityRepoInterface interface {
	// Returns NamedEntity objects matching the provided query. A limit is
	// required
	List(ctx context.Context, input ListNamedEntityInput) (NamedEntityCollectionOutput, error)
	// Updates NamedEntity record, will create metadata if it does not exist
	Update(ctx context.Context, input models.NamedEntity) error
	// Gets metadata (if available) associated with a NamedEntity
	Get(ctx context.Context, input GetNamedEntityInput) (models.NamedEntity, error)
}
