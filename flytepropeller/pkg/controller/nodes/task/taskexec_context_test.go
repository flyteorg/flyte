package task

import (
	"bytes"
	"context"
	"testing"

	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/resourcemanager"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	pluginCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"

	mocks2 "github.com/flyteorg/flytepropeller/pkg/controller/executors/mocks"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/catalog/mocks"
	pluginCoreMocks "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	ioMocks "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/stretchr/testify/assert"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	flyteMocks "github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/handler"
	nodeMocks "github.com/flyteorg/flytepropeller/pkg/controller/nodes/handler/mocks"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/codex"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/secretmanager"
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
	ma := 5
	n.OnGetRetryStrategy().Return(&v1alpha1.RetryStrategy{MinAttempts: &ma})

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

	executionContext := &mocks2.ExecutionContext{}
	executionContext.OnGetExecutionConfig().Return(v1alpha1.ExecutionConfig{})
	executionContext.OnGetParentInfo().Return(nil)
	executionContext.OnGetEventVersion().Return(v1alpha1.EventVersion0)
	nCtx.OnExecutionContext().Return(executionContext)

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

	p := &pluginCoreMocks.Plugin{}
	p.On("GetID").Return("plugin1")
	p.OnGetProperties().Return(pluginCore.PluginProperties{})
	got, err := tk.newTaskExecutionContext(context.TODO(), nCtx, p)
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

	assert.EqualValues(t, got.ResourceManager().(resourcemanager.TaskResourceManager).GetResourcePoolInfo(), make([]*event.ResourcePoolInfo, 0))

	// TODO @kumare fix this test
	assert.NotNil(t, got.rm)

	_, err = got.rm.AllocateResource(context.TODO(), "foo", "token", pluginCore.ResourceConstraintsSpec{})
	assert.NoError(t, err)
	assert.EqualValues(t, []*event.ResourcePoolInfo{
		{
			Namespace:       "foo",
			AllocationToken: "token",
		},
	}, got.ResourceManager().(resourcemanager.TaskResourceManager).GetResourcePoolInfo())
	assert.Nil(t, got.Catalog())
	// assert.Equal(t, got.InputReader(), ir)

	anotherPlugin := &pluginCoreMocks.Plugin{}
	anotherPlugin.On("GetID").Return("plugin2")
	maxLength := 8
	anotherPlugin.OnGetProperties().Return(pluginCore.PluginProperties{
		GeneratedNameMaxLength: &maxLength,
	})
	anotherTaskExecCtx, _ := tk.newTaskExecutionContext(context.TODO(), nCtx, anotherPlugin)
	assert.Equal(t, anotherTaskExecCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), "fpmmhh6q")
}
