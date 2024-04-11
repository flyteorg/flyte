package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
)

type UpdateConfigurationInput struct {
	VersionToUpdate          string
	NewConfigurationDocument *models.ConfigurationDocument
}

type ConfigurationDocumentRepoInterface interface {
	GetActive(ctx context.Context) (models.ConfigurationDocument, error)
	Update(ctx context.Context, input *UpdateConfigurationInput) error
	Create(ctx context.Context, input *models.ConfigurationDocument) error
}
