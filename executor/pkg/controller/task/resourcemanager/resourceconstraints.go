package resourcemanager

import (
	pluginCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

type ResourceConstraint interface {
	IsAllowed(int64) bool
}

func isAllowed(constraintValue int64, actualValue int64) bool {
	return constraintValue > actualValue
}

type BaseResourceConstraint struct {
	Value int64
}

func (brc *BaseResourceConstraint) IsAllowed(actualValue int64) bool {
	return isAllowed(brc.Value, actualValue)
}

type FullyQualifiedResourceConstraint struct {
	TargetedPrefixString string
	Value                int64
}

func (fqrc *FullyQualifiedResourceConstraint) IsAllowed(actualValue int64) bool {
	return isAllowed(fqrc.Value, actualValue)
}

func composeFullyQualifiedProjectScopeResourceConstraint(spec pluginCore.ResourceConstraintsSpec, id *core.TaskExecutionIdentifier) FullyQualifiedResourceConstraint {
	return FullyQualifiedResourceConstraint{
		TargetedPrefixString: string(composeProjectScopePrefix(id)),
		Value:                spec.ProjectScopeResourceConstraint.Value,
	}
}

func composeFullyQualifiedNamespaceScopeResourceConstraint(spec pluginCore.ResourceConstraintsSpec, id *core.TaskExecutionIdentifier) FullyQualifiedResourceConstraint {
	return FullyQualifiedResourceConstraint{
		TargetedPrefixString: string(composeNamespaceScopePrefix(id)),
		Value:                spec.NamespaceScopeResourceConstraint.Value,
	}
}
