package workflowengine

import (
	"github.com/flyteorg/flyteadmin/pkg/workflowengine/impl"
	"github.com/flyteorg/flyteadmin/pkg/workflowengine/interfaces"
)

var registry = impl.NewRegistry()

func GetRegistry() interfaces.WorkflowExecutorRegistry {
	return registry
}
