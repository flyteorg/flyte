package dynamic

import (
	"testing"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"
)

func TestHierarchicalNodeID(t *testing.T) {
	t.Run("empty parent", func(t *testing.T) {
		actual, err := hierarchicalNodeID("", "abc")
		assert.NoError(t, err)
		assert.Equal(t, "-abc", actual)
	})

	t.Run("long result", func(t *testing.T) {
		actual, err := hierarchicalNodeID("abcdefghijklmnopqrstuvwxyz", "abc")
		assert.NoError(t, err)
		assert.Equal(t, "fpa3kc3y", actual)
	})
}

func TestUnderlyingInterface(t *testing.T) {
	expectedIface := &core.TypedInterface{
		Outputs: &core.VariableMap{
			Variables: map[string]*core.Variable{
				"in": {
					Type: &core.LiteralType{
						Type: &core.LiteralType_Simple{
							Simple: core.SimpleType_INTEGER,
						},
					},
				},
			},
		},
	}
	wf := &mocks.ExecutableWorkflow{}

	subWF := &mocks.ExecutableSubWorkflow{}
	wf.On("FindSubWorkflow", mock.Anything).Return(subWF)
	subWF.On("GetOutputs").Return(&v1alpha1.OutputVarMap{VariableMap: expectedIface.Outputs})

	task := &mocks.ExecutableTask{}
	wf.On("GetTask", mock.Anything).Return(task, nil)
	task.On("CoreTask").Return(&core.TaskTemplate{
		Interface: expectedIface,
	})

	n := &mocks.ExecutableNode{}
	wf.On("GetNode", mock.Anything).Return(n)
	emptyStr := ""
	n.On("GetTaskID").Return(&emptyStr)

	iface, err := underlyingInterface(wf, n)
	assert.NoError(t, err)
	assert.NotNil(t, iface)
	assert.Equal(t, expectedIface, iface)

	n = &mocks.ExecutableNode{}
	n.On("GetTaskID").Return(nil)

	wfNode := &mocks.ExecutableWorkflowNode{}
	n.On("GetWorkflowNode").Return(wfNode)
	wfNode.On("GetSubWorkflowRef").Return(&emptyStr)

	iface, err = underlyingInterface(wf, n)
	assert.NoError(t, err)
	assert.NotNil(t, iface)
	assert.Equal(t, expectedIface, iface)
}
