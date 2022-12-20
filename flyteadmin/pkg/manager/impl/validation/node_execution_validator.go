package validation

import (
	"github.com/flyteorg/flyteadmin/pkg/common"
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

// Validates that NodeExecutionEventRequests handled by admin include a valid node execution identifier.
// In the case the event specifies a DynamicWorkflow in the TaskNodeMetadata, this method also validates the contents of
// the dynamic workflow.
func ValidateNodeExecutionEventRequest(request *admin.NodeExecutionEventRequest, maxOutputSizeInBytes int64) error {
	if request.Event == nil {
		return shared.GetMissingArgumentError(shared.Event)
	}
	err := ValidateNodeExecutionIdentifier(request.Event.Id)
	if err != nil {
		return err
	}
	if request.Event.GetTaskNodeMetadata() != nil && request.Event.GetTaskNodeMetadata().DynamicWorkflow != nil {
		dynamicWorkflowNodeMetadata := request.Event.GetTaskNodeMetadata().DynamicWorkflow
		if err := ValidateIdentifier(dynamicWorkflowNodeMetadata.Id, common.Workflow); err != nil {
			return err
		}
		if dynamicWorkflowNodeMetadata.CompiledWorkflow == nil {
			return shared.GetMissingArgumentError("compiled dynamic workflow")
		}
		if dynamicWorkflowNodeMetadata.CompiledWorkflow.Primary == nil {
			return shared.GetMissingArgumentError("primary dynamic workflow")
		}
		if dynamicWorkflowNodeMetadata.CompiledWorkflow.Primary.Template == nil {
			return shared.GetMissingArgumentError("primary dynamic workflow template")
		}
		if err := ValidateIdentifier(dynamicWorkflowNodeMetadata.CompiledWorkflow.Primary.Template.Id, common.Workflow); err != nil {
			return err
		}
	}
	if err := ValidateOutputData(request.Event.GetOutputData(), maxOutputSizeInBytes); err != nil {
		return err
	}
	return nil
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
