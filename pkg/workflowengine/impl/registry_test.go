package impl

import (
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/workflowengine/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/workflowengine/mocks"
	"github.com/stretchr/testify/assert"
)

func getMockK8sWorkflowExecutor(id string) interfaces.WorkflowExecutor {
	exec := mocks.WorkflowExecutor{}
	exec.OnID().Return(id)
	return &exec
}

var testExecID = "foo"
var defaultExecID = "default"

func TestRegister(t *testing.T) {
	registry := workflowExecutorRegistry{}
	exec := getMockK8sWorkflowExecutor(testExecID)
	registry.Register(exec)
	assert.Equal(t, testExecID, registry.GetExecutor().ID())
}

func TestRegisterDefault(t *testing.T) {
	registry := workflowExecutorRegistry{}

	defaultExec := getMockK8sWorkflowExecutor(defaultExecID)
	registry.RegisterDefault(defaultExec)
	assert.Equal(t, defaultExecID, registry.GetExecutor().ID())

	exec := getMockK8sWorkflowExecutor(testExecID)
	registry.Register(exec)
	assert.Equal(t, testExecID, registry.GetExecutor().ID())
}
