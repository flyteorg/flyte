package validation

import (
	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func ValidateTaskExecutionRequest(request *admin.TaskExecutionEventRequest, maxOutputSizeInBytes int64) error {
	if request.GetEvent() == nil {
		return shared.GetMissingArgumentError(shared.Event)
	}
	if request.GetEvent().GetOccurredAt() == nil {
		return shared.GetMissingArgumentError(shared.OccurredAt)
	}
	if err := ValidateOutputData(request.GetEvent().GetOutputData(), maxOutputSizeInBytes); err != nil {
		return err
	}

	return ValidateTaskExecutionIdentifier(&core.TaskExecutionIdentifier{
		TaskId:          request.GetEvent().GetTaskId(),
		NodeExecutionId: request.GetEvent().GetParentNodeExecutionId(),
		RetryAttempt:    request.GetEvent().GetRetryAttempt(),
	})
}

func ValidateTaskExecutionIdentifier(identifier *core.TaskExecutionIdentifier) error {
	if identifier == nil {
		return shared.GetMissingArgumentError(shared.ID)
	}
	if identifier.GetNodeExecutionId() == nil {
		return shared.GetMissingArgumentError(shared.NodeExecutionID)
	}

	if err := ValidateNodeExecutionIdentifier(identifier.GetNodeExecutionId()); err != nil {
		return err
	}

	if identifier.GetTaskId() == nil {
		return shared.GetMissingArgumentError(shared.TaskID)
	}

	if err := ValidateIdentifier(identifier.GetTaskId(), common.Task); err != nil {
		return err
	}

	return nil
}

func ValidateTaskExecutionListRequest(request *admin.TaskExecutionListRequest) error {
	if err := ValidateNodeExecutionIdentifier(request.GetNodeExecutionId()); err != nil {
		return err
	}
	if err := ValidateLimit(request.GetLimit()); err != nil {
		return err
	}
	return nil
}
