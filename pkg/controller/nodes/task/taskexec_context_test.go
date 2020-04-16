package task

import (
	"bytes"
	"context"
	"testing"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/catalog/mocks"
	ioMocks "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/storage"
	"github.com/stretchr/testify/assert"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

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

	nodeID := "n1"

	nm := &nodeMocks.NodeExecutionMetadata{}
	nm.OnGetAnnotations().Return(map[string]string{})
	nm.OnGetNodeExecutionID().Return(&core.NodeExecutionIdentifier{
		NodeId:      nodeID,
		ExecutionId: wfExecID,
	})
	nm.OnGetK8sServiceAccount().Return("service-account")
	nm.OnGetLabels().Return(map[string]string{})
	nm.OnGetNamespace().Return("namespace")
	nm.OnGetOwnerID().Return(types.NamespacedName{Namespace: "namespace", Name: "name"})
	nm.OnGetOwnerReference().Return(v1.OwnerReference{
		Kind: "sample",
		Name: "name",
	})

	taskID := &core.Identifier{}
	tr := &nodeMocks.TaskReader{}
	tr.OnGetTaskID().Return(taskID)

	ns := &flyteMocks.ExecutableNodeStatus{}
	ns.OnGetDataDir().Return(storage.DataReference("data-dir"))
	ns.OnGetOutputDir().Return(storage.DataReference("output-dir"))

	res := &v12.ResourceRequirements{}
	n := &flyteMocks.ExecutableNode{}
	n.OnGetResources().Return(res)

	ir := &ioMocks.InputReader{}
	nCtx := &nodeMocks.NodeExecutionContext{}
	nCtx.OnNodeExecutionMetadata().Return(nm)
	nCtx.OnNode().Return(n)
	nCtx.OnInputReader().Return(ir)
	nCtx.OnCurrentAttempt().Return(uint32(1))
	nCtx.OnTaskReader().Return(tr)
	nCtx.OnMaxDatasetSizeBytes().Return(int64(1))
	nCtx.OnNodeStatus().Return(ns)
	nCtx.OnNodeID().Return(nodeID)
	nCtx.OnEventsRecorder().Return(nil)
	nCtx.OnEnqueueOwnerFunc().Return(nil)

	ds, err := storage.NewDataStore(
		&storage.Config{
			Type: storage.TypeMemory,
		},
		promutils.NewTestScope(),
	)
	assert.NoError(t, err)
	nCtx.OnDataStore().Return(ds)

	st := bytes.NewBuffer([]byte{})
	a := 45
	type test struct {
		A int
	}
	codex := codex.GobStateCodec{}
	assert.NoError(t, codex.Encode(test{A: a}, st))
	nr := &nodeMocks.NodeStateReader{}
	nr.OnGetTaskNodeState().Return(handler.TaskNodeState{
		PluginState: st.Bytes(),
	})
	nCtx.OnNodeStateReader().Return(nr)
	nCtx.OnRawOutputPrefix().Return("s3://sandbox/")
	nCtx.OnOutputShardSelector().Return(ioutils.NewConstantShardSelector([]string{"x"}))

	noopRm := CreateNoopResourceManager(context.TODO(), promutils.NewTestScope())

	c := &mocks.Client{}
	tk := &Handler{
		catalog:         c,
		secretManager:   secretmanager.NewFileEnvSecretManager(secretmanager.GetConfig()),
		resourceManager: noopRm,
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
	assert.Equal(t, got.TaskExecutionMetadata().GetTaskExecutionID().GetID().NodeExecutionId.GetNodeId(), nodeID)
	assert.Equal(t, got.TaskExecutionMetadata().GetTaskExecutionID().GetID().NodeExecutionId.GetExecutionId(), wfExecID)

	// TODO @kumare fix this test
	assert.NotNil(t, got.ResourceManager())
	assert.Nil(t, got.Catalog())
	// assert.Equal(t, got.InputReader(), ir)
}
