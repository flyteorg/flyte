package interfaces

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

//go:generate mockery -name=MetricsInterface -output=../mocks -case=underscore

// Interface for managing Flyte execution metrics
type MetricsInterface interface {
	GetExecutionMetrics(ctx context.Context, request admin.WorkflowExecutionGetMetricsRequest) (
		*admin.WorkflowExecutionGetMetricsResponse, error)
}
