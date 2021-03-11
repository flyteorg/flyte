package validation

import (
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

func ValidateNodeExecutionIdentifier(identifier *core.NodeExecutionIdentifier) error {
	if identifier == nil {
		return shared.GetMissingArgumentError(shared.ID)
	}
	if identifier.ExecutionId == nil {
		return shared.GetMissingArgumentError(shared.ExecutionID)
	}
	if identifier.NodeId == "" {
		return shared.GetMissingArgumentError(shared.NodeID)
	}

	return ValidateWorkflowExecutionIdentifier(identifier.ExecutionId)
}

func ValidateNodeExecutionListRequest(request admin.NodeExecutionListRequest) error {
	if err := ValidateWorkflowExecutionIdentifier(request.WorkflowExecutionId); err != nil {
		return shared.GetMissingArgumentError(shared.ExecutionID)
	}
	if err := ValidateLimit(request.Limit); err != nil {
		return err
	}
	return nil
}

func ValidateNodeExecutionForTaskListRequest(request admin.NodeExecutionForTaskListRequest) error {
	if err := ValidateTaskExecutionIdentifier(request.TaskExecutionId); err != nil {
		return err
	}
	if err := ValidateLimit(request.Limit); err != nil {
		return err
	}
	return nil
}
