package transformers

import (
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

// Transforms a NodeExecutionEventRequest to a NodeExecutionEvent model
func CreateNodeExecutionEventModel(request *admin.NodeExecutionEventRequest) (*models.NodeExecutionEvent, error) {
	occurredAt, err := ptypes.Timestamp(request.GetEvent().GetOccurredAt())
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to marshal occurred at timestamp")
	}
	return &models.NodeExecutionEvent{
		NodeExecutionKey: models.NodeExecutionKey{
			NodeID: request.GetEvent().GetId().GetNodeId(),
			ExecutionKey: models.ExecutionKey{
				Project: request.GetEvent().GetId().GetExecutionId().GetProject(),
				Domain:  request.GetEvent().GetId().GetExecutionId().GetDomain(),
				Name:    request.GetEvent().GetId().GetExecutionId().GetName(),
			},
		},
		RequestID:  request.GetRequestId(),
		OccurredAt: occurredAt,
		Phase:      request.GetEvent().GetPhase().String(),
	}, nil
}
