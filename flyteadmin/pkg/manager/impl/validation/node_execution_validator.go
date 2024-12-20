package validation

import (
	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func ValidateNodeExecutionIdentifier(identifier *core.NodeExecutionIdentifier) error {
	if identifier == nil {
		return shared.GetMissingArgumentError(shared.ID)
	}
	if identifier.GetExecutionId() == nil {
		return shared.GetMissingArgumentError(shared.ExecutionID)
	}
	if identifier.GetNodeId() == "" {
		return shared.GetMissingArgumentError(shared.NodeID)
	}

	return ValidateWorkflowExecutionIdentifier(identifier.GetExecutionId())
}

// Validates that NodeExecutionEventRequests handled by admin include a valid node execution identifier.
// In the case the event specifies a DynamicWorkflow in the TaskNodeMetadata, this method also validates the contents of
// the dynamic workflow.
func ValidateNodeExecutionEventRequest(request *admin.NodeExecutionEventRequest, maxOutputSizeInBytes int64) error {
	if request.GetEvent() == nil {
		return shared.GetMissingArgumentError(shared.Event)
	}
	err := ValidateNodeExecutionIdentifier(request.GetEvent().GetId())
	if err != nil {
		return err
	}
	if request.GetEvent().GetTaskNodeMetadata() != nil && request.GetEvent().GetTaskNodeMetadata().GetDynamicWorkflow() != nil {
		dynamicWorkflowNodeMetadata := request.GetEvent().GetTaskNodeMetadata().GetDynamicWorkflow()
		if err := ValidateIdentifier(dynamicWorkflowNodeMetadata.GetId(), common.Workflow); err != nil {
			return err
		}
		if dynamicWorkflowNodeMetadata.GetCompiledWorkflow() == nil {
			return shared.GetMissingArgumentError("compiled dynamic workflow")
		}
		if dynamicWorkflowNodeMetadata.GetCompiledWorkflow().GetPrimary() == nil {
			return shared.GetMissingArgumentError("primary dynamic workflow")
		}
		if dynamicWorkflowNodeMetadata.GetCompiledWorkflow().GetPrimary().GetTemplate() == nil {
			return shared.GetMissingArgumentError("primary dynamic workflow template")
		}
		if err := ValidateIdentifier(dynamicWorkflowNodeMetadata.GetCompiledWorkflow().GetPrimary().GetTemplate().GetId(), common.Workflow); err != nil {
			return err
		}
	}
	if err := ValidateOutputData(request.GetEvent().GetOutputData(), maxOutputSizeInBytes); err != nil {
		return err
	}
	return nil
}

func ValidateNodeExecutionListRequest(request *admin.NodeExecutionListRequest) error {
	if err := ValidateWorkflowExecutionIdentifier(request.GetWorkflowExecutionId()); err != nil {
		return shared.GetMissingArgumentError(shared.ExecutionID)
	}
	if err := ValidateLimit(request.GetLimit()); err != nil {
		return err
	}
	return nil
}

func ValidateNodeExecutionForTaskListRequest(request *admin.NodeExecutionForTaskListRequest) error {
	if err := ValidateTaskExecutionIdentifier(request.GetTaskExecutionId()); err != nil {
		return err
	}
	if err := ValidateLimit(request.GetLimit()); err != nil {
		return err
	}
	return nil
}
