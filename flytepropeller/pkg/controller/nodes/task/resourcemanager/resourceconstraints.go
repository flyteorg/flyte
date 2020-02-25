package resourcemanager

import (
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	pluginCore "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"
)

type ComposedResourceConstraint struct {
	TargetedPrefixString string
	Value                int64
}

func composeProjectScopeResourceConstraint(spec pluginCore.ResourceConstraintsSpec, id *core.TaskExecutionIdentifier) ComposedResourceConstraint {
	return ComposedResourceConstraint{
		TargetedPrefixString: string(composeProjectScopePrefix(id)),
		Value:                spec.ProjectScopeResourceConstraint.Value,
	}
}

func composeNamespaceScopeResourceConstraint(spec pluginCore.ResourceConstraintsSpec, id *core.TaskExecutionIdentifier) ComposedResourceConstraint {
	return ComposedResourceConstraint{
		TargetedPrefixString: string(composeNamespaceScopePrefix(id)),
		Value:                spec.NamespaceScopeResourceConstraint.Value,
	}
}
