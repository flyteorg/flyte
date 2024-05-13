package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
)

//go:generate mockery -name=ConfigurationRepoInterface -output=../mocks -case=underscore

type UpdateConfigurationInput struct {
	VersionToUpdate          string
	NewConfigurationMetadata *models.ConfigurationDocumentMetadata
}

// ConfigurationRepoInterface defines the interface for interacting with configuration_document_metadata models.
type ConfigurationRepoInterface interface {
	GetActive(ctx context.Context) (models.ConfigurationDocumentMetadata, error)
	Update(ctx context.Context, input *UpdateConfigurationInput) error
	Create(ctx context.Context, input *models.ConfigurationDocumentMetadata) error
}
