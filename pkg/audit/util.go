package audit

import (
	"fmt"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

type requestParameters = map[string]string

const (
	Project      = "project"
	Domain       = "domain"
	Name         = "name"
	Version      = "version"
	NodeID       = "node_id"
	RetryAttempt = "retry_attempt"
	ResourceType = "resource"

	TaskProject = "task_project"
	TaskDomain  = "task_domain"
	TaskName    = "task_name"
	TaskVersion = "task_version"
)

func ParametersFromIdentifier(identifier *core.Identifier) requestParameters {
	if identifier == nil {
		return requestParameters{}
	}
	return requestParameters{
		Project: identifier.Project,
		Domain:  identifier.Domain,
		Name:    identifier.Name,
		Version: identifier.Version,
	}
}

func ParametersFromNamedEntityIdentifier(identifier *admin.NamedEntityIdentifier) requestParameters {
	if identifier == nil {
		return requestParameters{}
	}
	return requestParameters{
		Project: identifier.Project,
		Domain:  identifier.Domain,
		Name:    identifier.Name,
	}
}

func ParametersFromNamedEntityIdentifierAndResource(identifier *admin.NamedEntityIdentifier, resourceType core.ResourceType) requestParameters {
	if identifier == nil {
		return requestParameters{}
	}
	parameters := ParametersFromNamedEntityIdentifier(identifier)
	parameters[ResourceType] = resourceType.String()
	return parameters
}

func ParametersFromExecutionIdentifier(identifier *core.WorkflowExecutionIdentifier) requestParameters {
	if identifier == nil {
		return requestParameters{}
	}
	return requestParameters{
		Project: identifier.Project,
		Domain:  identifier.Domain,
		Name:    identifier.Name,
	}
}

func ParametersFromNodeExecutionIdentifier(identifier *core.NodeExecutionIdentifier) requestParameters {
	if identifier == nil || identifier.ExecutionId == nil {
		return requestParameters{}
	}
	return requestParameters{
		Project: identifier.ExecutionId.Project,
		Domain:  identifier.ExecutionId.Domain,
		Name:    identifier.ExecutionId.Name,
		NodeID:  identifier.NodeId,
	}
}

func ParametersFromTaskExecutionIdentifier(identifier *core.TaskExecutionIdentifier) requestParameters {
	if identifier == nil || identifier.NodeExecutionId == nil || identifier.TaskId == nil {
		return requestParameters{}
	}
	params := ParametersFromNodeExecutionIdentifier(identifier.NodeExecutionId)
	params[RetryAttempt] = fmt.Sprint(identifier.RetryAttempt)
	params[TaskProject] = identifier.TaskId.Project
	params[TaskDomain] = identifier.TaskId.Domain
	params[TaskName] = identifier.TaskId.Name
	params[TaskVersion] = identifier.TaskId.Version
	return params
}
