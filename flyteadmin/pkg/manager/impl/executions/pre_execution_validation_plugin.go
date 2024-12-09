package executions

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

// PreExecutionValidationPlugin is an interface that allows for pre-execution validation.
type PreExecutionValidationPlugin interface {
	ValidateCreateExecutionRequest(ctx context.Context, request *admin.ExecutionCreateRequest) error
}

// NoopPreExecutionValidationPlugin is a noops implementation of the PreExecutionValidationPlugin interface.
type NoopPreExecutionValidationPlugin struct{}

// ValidateCreateExecutionRequest does nothing.
func (n *NoopPreExecutionValidationPlugin) ValidateCreateExecutionRequest(ctx context.Context, request *admin.ExecutionCreateRequest) error {
	return nil
}

func NewNoopPreExecutionValidationPlugin() *NoopPreExecutionValidationPlugin {
	return &NoopPreExecutionValidationPlugin{}
}
