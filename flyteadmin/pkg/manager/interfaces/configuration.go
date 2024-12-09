package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

//go:generate mockery -name ConfigurationInterface -output=../mocks -case=underscore

// Interface for managing project configuration, which includes attributes such as task resource attributes
type ConfigurationInterface interface {
	GetConfiguration(ctx context.Context, request *admin.ConfigurationGetRequest) (*admin.ConfigurationGetResponse, error)
	UpdateConfiguration(ctx context.Context, request *admin.ConfigurationUpdateRequest) (*admin.ConfigurationUpdateResponse, error)
	GetReadOnlyActiveDocument(ctx context.Context) (*admin.ConfigurationDocument, error)
	GetEditableActiveDocument(ctx context.Context) (*admin.ConfigurationDocument, error)
}
