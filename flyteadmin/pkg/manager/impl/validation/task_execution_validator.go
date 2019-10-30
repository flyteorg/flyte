package validation

import (
	"github.com/lyft/flyteadmin/pkg/common"
	"github.com/lyft/flyteadmin/pkg/manager/impl/shared"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
)

func ValidateTaskExecutionRequest(request admin.TaskExecutionEventRequest) error {
	if request.Event == nil {
		return shared.GetMissingArgumentError(shared.Event)
	}
	if request.Event.OccurredAt == nil {
		return shared.GetMissingArgumentError(shared.OccurredAt)
	}

	return ValidateTaskExecutionIdentifier(&core.TaskExecutionIdentifier{
		TaskId:          request.Event.TaskId,
		NodeExecutionId: request.Event.ParentNodeExecutionId,
		RetryAttempt:    request.Event.RetryAttempt,
	})
}

func ValidateTaskExecutionIdentifier(identifier *core.TaskExecutionIdentifier) error {
	if identifier == nil {
		return shared.GetMissingArgumentError(shared.ID)
	}
	if identifier.NodeExecutionId == nil {
		return shared.GetMissingArgumentError(shared.NodeExecutionID)
	}

	if err := ValidateNodeExecutionIdentifier(identifier.NodeExecutionId); err != nil {
		return err
	}

	if identifier.TaskId == nil {
		return shared.GetMissingArgumentError(shared.TaskID)
	}

	if err := ValidateIdentifier(identifier.TaskId, common.Task); err != nil {
		return err
	}

	return nil
}

func ValidateTaskExecutionListRequest(request admin.TaskExecutionListRequest) error {
	if err := ValidateNodeExecutionIdentifier(request.NodeExecutionId); err != nil {
		return err
	}
	if err := ValidateLimit(request.Limit); err != nil {
		return err
	}
	return nil
}
