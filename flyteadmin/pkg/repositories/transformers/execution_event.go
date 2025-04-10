package transformers

import (
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

// Transforms a ExecutionEventCreateRequest to a ExecutionEvent model
func CreateExecutionEventModel(request *admin.WorkflowExecutionEventRequest) (*models.ExecutionEvent, error) {
	occurredAt := request.GetEvent().GetOccurredAt().AsTime()
	return &models.ExecutionEvent{
		ExecutionKey: models.ExecutionKey{
			Project: request.GetEvent().GetExecutionId().GetProject(),
			Domain:  request.GetEvent().GetExecutionId().GetDomain(),
			Name:    request.GetEvent().GetExecutionId().GetName(),
		},
		RequestID:  request.GetRequestId(),
		OccurredAt: occurredAt,
		Phase:      request.GetEvent().GetPhase().String(),
	}, nil
}
