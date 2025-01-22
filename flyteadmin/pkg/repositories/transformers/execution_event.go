package transformers

import (
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

// Transforms a ExecutionEventCreateRequest to a ExecutionEvent model
func CreateExecutionEventModel(request *admin.WorkflowExecutionEventRequest) (*models.ExecutionEvent, error) {
	occurredAt, err := ptypes.Timestamp(request.GetEvent().GetOccurredAt())
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to marshal occurred at timestamp")
	}
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
