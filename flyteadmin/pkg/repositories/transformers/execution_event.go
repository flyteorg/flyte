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
	occurredAt, err := ptypes.Timestamp(request.Event.OccurredAt)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to marshal occurred at timestamp")
	}
	return &models.ExecutionEvent{
		ExecutionKey: models.ExecutionKey{
			Project: request.Event.ExecutionId.Project,
			Domain:  request.Event.ExecutionId.Domain,
			Name:    request.Event.ExecutionId.Name,
		},
		RequestID:  request.RequestId,
		OccurredAt: occurredAt,
		Phase:      request.Event.Phase.String(),
	}, nil
}
