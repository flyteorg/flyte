package task

import (
	"bytes"
	"context"
	"testing"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/catalog/mocks"
	ioMocks "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/storage"
	"github.com/stretchr/testify/assert"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	flyteMocks "github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
	nodeMocks "github.com/lyft/flytepropeller/pkg/controller/nodes/handler/mocks"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/task/codex"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/task/secretmanager"
)

func TestHandler_newTaskExecutionContext(t *testing.T) {
	wfExecID := &core.WorkflowExecutionIdentifier{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
	}

	nm := &nodeMocks.NodeExecutionMetadata{}
	nm.On("GetAnnotations").Return(map[string]string{})
	nm.On("GetExecutionID").Return(v1alpha1.WorkflowExecutionIdentifier{
		WorkflowExecutionIdentifier: wfExecID,
	})
	nm.On("GetK8sServiceAccount").Return("service-account")
	nm.On("GetLabels").Return(map[string]string{})
	nm.On("GetNamespace").Return("namespace")
	nm.On("GetOwnerID").Return(types.NamespacedName{Namespace: "namespace", Name: "name"})
	nm.On("GetOwnerReference").Return(v1.OwnerReference{
		Kind: "sample",
		Name: "name",
	})

	taskID := &core.Identifier{}
	tr := &nodeMocks.TaskReader{}
	tr.On("GetTaskID").Return(taskID)

	ns := &flyteMocks.ExecutableNodeStatus{}
	ns.On("GetDataDir").Return(storage.DataReference("data-dir"))

	res := &v12.ResourceRequirements{}
	n := &flyteMocks.ExecutableNode{}
	n.On("GetResources").Return(res)

	ir := &ioMocks.InputReader{}
	nCtx := &nodeMocks.NodeExecutionContext{}
	nCtx.On("NodeExecutionMetadata").Return(nm)
	nCtx.On("Node").Return(n)
	nCtx.On("InputReader").Return(ir)
	nCtx.On("DataStore").Return(storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope()))
	nCtx.On("CurrentAttempt").Return(uint32(1))
	nCtx.On("TaskReader").Return(tr)
	nCtx.On("MaxDatasetSizeBytes").Return(int64(1))
	nCtx.On("NodeStatus").Return(ns)
	nCtx.On("NodeID").Return("n1")
	nCtx.On("EventsRecorder").Return(nil)
	nCtx.On("EnqueueOwner").Return(nil)

	st := bytes.NewBuffer([]byte{})
	a := 45
	type test struct {
		A int
	}
	codex := codex.GobStateCodec{}
	assert.NoError(t, codex.Encode(test{A: a}, st))
	nr := &nodeMocks.NodeStateReader{}
	nr.On("GetTaskNodeState").Return(handler.TaskNodeState{
		PluginState: st.Bytes(),
	})
	nCtx.On("NodeStateReader").Return(nr)

	c := &mocks.Client{}
	tk := &Handler{
		catalog:       c,
		secretManager: secretmanager.NewFileEnvSecretManager(secretmanager.GetConfig()),
	}

	got, err := tk.newTaskExecutionContext(context.TODO(), nCtx, "plugin1")
	assert.NoError(t, err)
	assert.NotNil(t, got)

	f := &test{}
	v, err := got.PluginStateReader().Get(f)
	assert.NoError(t, err)
	assert.Equal(t, v, uint8(0))
	assert.Equal(t, f.A, a)

	// Try writing new state
	type test2 struct {
		G float32
	}
	s2 := &test2{G: 6.0}
	assert.NoError(t, got.PluginStateWriter().Put(10, s2))
	assert.Equal(t, got.psm.newStateVersion, uint8(10))
	assert.NotNil(t, got.psm.newState)

	assert.Equal(t, got.TaskReader(), tr)
	assert.Equal(t, got.MaxDatasetSizeBytes(), int64(1))
	assert.NotNil(t, got.SecretManager())

	assert.NotNil(t, got.OutputWriter())
	assert.Equal(t, got.TaskExecutionMetadata().GetOverrides().GetResources(), res)

	assert.Equal(t, got.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), "name-n1-1")
	assert.Equal(t, got.TaskExecutionMetadata().GetTaskExecutionID().GetID().TaskId, taskID)
	assert.Equal(t, got.TaskExecutionMetadata().GetTaskExecutionID().GetID().RetryAttempt, uint32(1))
	assert.Equal(t, got.TaskExecutionMetadata().GetTaskExecutionID().GetID().NodeExecutionId.GetNodeId(), "n1")
	assert.Equal(t, got.TaskExecutionMetadata().GetTaskExecutionID().GetID().NodeExecutionId.GetExecutionId(), wfExecID)

	// TODO @kumare fix this test
	assert.NotNil(t, got.ResourceManager())
	assert.Nil(t, got.Catalog())
	// assert.Equal(t, got.InputReader(), ir)
}
