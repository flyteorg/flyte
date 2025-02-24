package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

//go:generate mockery-v2 --name=MetricsInterface --output=../mocks --case=underscore --with-expecter

// Interface for managing Flyte execution metrics
type MetricsInterface interface {
	GetExecutionMetrics(ctx context.Context, request *admin.WorkflowExecutionGetMetricsRequest) (
		*admin.WorkflowExecutionGetMetricsResponse, error)
}
