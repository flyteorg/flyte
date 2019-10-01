package dynamic

import (
	"context"
	"testing"

	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/storage"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

func createInmemoryStore(t testing.TB) *storage.DataStore {
	cfg := storage.Config{
		Type: storage.TypeMemory,
	}

	d, err := storage.NewDataStore(&cfg, promutils.NewTestScope())
	assert.NoError(t, err)

	return d
}

func Test_cacheFlyteWorkflow(t *testing.T) {
	store := createInmemoryStore(t)
	expected := &v1alpha1.FlyteWorkflow{
		TypeMeta:   v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{},
		WorkflowSpec: &v1alpha1.WorkflowSpec{
			ID: "abc",
			Connections: v1alpha1.Connections{
				DownstreamEdges: map[v1alpha1.NodeID][]v1alpha1.NodeID{},
				UpstreamEdges:   map[v1alpha1.NodeID][]v1alpha1.NodeID{},
			},
		},
		Inputs: &v1alpha1.Inputs{LiteralMap: &core.LiteralMap{}},
	}

	location := storage.DataReference("somekey/file.json")
	assert.NoError(t, cacheFlyteWorkflow(context.TODO(), store, expected, location))
	actual, err := loadCachedFlyteWorkflow(context.TODO(), store, location)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
