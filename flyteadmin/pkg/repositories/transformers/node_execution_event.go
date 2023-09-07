package transformers

import (
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"
)

// Transforms a NodeExecutionEventRequest to a NodeExecutionEvent model
func CreateNodeExecutionEventModel(request admin.NodeExecutionEventRequest) (*models.NodeExecutionEvent, error) {
	occurredAt, err := ptypes.Timestamp(request.Event.OccurredAt)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to marshal occurred at timestamp")
	}
	return &models.NodeExecutionEvent{
		NodeExecutionKey: models.NodeExecutionKey{
			NodeID: request.Event.Id.NodeId,
			ExecutionKey: models.ExecutionKey{
				Project: request.Event.Id.ExecutionId.Project,
				Domain:  request.Event.Id.ExecutionId.Domain,
				Name:    request.Event.Id.ExecutionId.Name,
			},
		},
		RequestID:  request.RequestId,
		OccurredAt: occurredAt,
		Phase:      request.Event.Phase.String(),
	}, nil
}
