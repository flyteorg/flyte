package task

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog/mocks"
	pluginCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	pluginCoreMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	ioMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	flyteMocks "github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors"
	mocks2 "github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/handler"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"
	nodeMocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/task/codex"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/task/resourcemanager"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/task/secretmanager"
	"github.com/flyteorg/flyte/flytepropeller/pkg/utils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

type dummyPluginState struct {
	A int
}

var (
	wfExecID = &core.WorkflowExecutionIdentifier{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
	}
	taskID    = &core.Identifier{}
	nodeID    = "n1"
	resources = &corev1.ResourceRequirements{
		Requests: make(corev1.ResourceList),
		Limits:   make(corev1.ResourceList),
	}
	dummyPluginStateA = 45
)

func dummyNodeExecutionContext(t *testing.T, parentInfo executors.ImmutableParentInfo, eventVersion v1alpha1.EventVersion) interfaces.NodeExecutionContext {
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

	tr := &nodeMocks.TaskReader{}
	tr.OnGetTaskID().Return(taskID)

	ns := &flyteMocks.ExecutableNodeStatus{}
	ns.OnGetDataDir().Return("data-dir")
	ns.OnGetOutputDir().Return("output-dir")

	n := &flyteMocks.ExecutableNode{}
	n.OnGetResources().Return(resources)
	ma := 5
	n.OnGetRetryStrategy().Return(&v1alpha1.RetryStrategy{MinAttempts: &ma})

	ir := &ioMocks.InputReader{}
	nCtx := &nodeMocks.NodeExecutionContext{}
	nCtx.OnNodeExecutionMetadata().Return(nm)
	nCtx.OnNode().Return(n)
	nCtx.OnInputReader().Return(ir)
	nCtx.OnCurrentAttempt().Return(uint32(1))
	nCtx.OnTaskReader().Return(tr)
	nCtx.OnNodeStatus().Return(ns)
	nCtx.OnNodeID().Return(nodeID)
	nCtx.OnEventsRecorder().Return(nil)
	nCtx.OnEnqueueOwnerFunc().Return(nil)

	executionContext := &mocks2.ExecutionContext{}
	executionContext.EXPECT().GetExecutionConfig().Return(v1alpha1.ExecutionConfig{})
	executionContext.EXPECT().GetParentInfo().Return(parentInfo)
	executionContext.EXPECT().GetEventVersion().Return(eventVersion)
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
	codex := codex.GobStateCodec{}
	assert.NoError(t, codex.Encode(dummyPluginState{A: dummyPluginStateA}, st))
	nr := &nodeMocks.NodeStateReader{}
	nr.OnGetTaskNodeState().Return(handler.TaskNodeState{
		PluginState: st.Bytes(),
	})
	nCtx.OnNodeStateReader().Return(nr)
	nCtx.OnRawOutputPrefix().Return("s3://sandbox/")
	nCtx.OnOutputShardSelector().Return(ioutils.NewConstantShardSelector([]string{"x"}))
	return nCtx
}

func dummyPlugin() pluginCore.Plugin {
	p := &pluginCoreMocks.Plugin{}
	p.On("GetID").Return("plugin1")
	p.OnGetProperties().Return(pluginCore.PluginProperties{})
	return p
}

func dummyHandler() *Handler {
	noopRm := CreateNoopResourceManager(context.TODO(), promutils.NewTestScope())
	c := &mocks.Client{}
	return &Handler{
		catalog:         c,
		secretManager:   secretmanager.NewFileEnvSecretManager(secretmanager.GetConfig()),
		resourceManager: noopRm,
	}
}

func TestHandler_newTaskExecutionContext(t *testing.T) {
	nCtx := dummyNodeExecutionContext(t, nil, v1alpha1.EventVersion0)
	p := dummyPlugin()
	tk := dummyHandler()
	got, err := tk.newTaskExecutionContext(context.TODO(), nCtx, p)
	assert.NoError(t, err)
	assert.NotNil(t, got)

	f := &dummyPluginState{}
	v, err := got.PluginStateReader().Get(f)
	assert.NoError(t, err)
	assert.Equal(t, v, uint8(0))
	assert.Equal(t, f.A, dummyPluginStateA)

	// Try writing new state
	type test2 struct {
		G float32
	}
	s2 := &test2{G: 6.0}
	assert.NoError(t, got.PluginStateWriter().Put(10, s2))
	assert.Equal(t, got.psm.newStateVersion, uint8(10))
	assert.NotNil(t, got.psm.newState)

	assert.NotNil(t, got.TaskReader())
	assert.NotNil(t, got.SecretManager())

	assert.NotNil(t, got.OutputWriter())
	assert.Equal(t, got.TaskExecutionMetadata().GetOverrides().GetResources(), resources)

	assert.Equal(t, got.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), "name-n1-1")
	assert.Equal(t, got.TaskExecutionMetadata().GetTaskExecutionID().GetID().TaskId, taskID)
	assert.Equal(t, got.TaskExecutionMetadata().GetTaskExecutionID().GetID().RetryAttempt, uint32(1))
	assert.Equal(t, got.TaskExecutionMetadata().GetTaskExecutionID().GetID().NodeExecutionId.GetNodeId(), nodeID)
	assert.Equal(t, got.TaskExecutionMetadata().GetTaskExecutionID().GetID().NodeExecutionId.GetExecutionId(), wfExecID)
	assert.Equal(t, got.TaskExecutionMetadata().GetTaskExecutionID().GetUniqueNodeID(), nodeID)

	assert.EqualValues(t, got.ResourceManager().(resourcemanager.TaskResourceManager).GetResourcePoolInfo(), make([]*event.ResourcePoolInfo, 0))

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
	anotherPlugin.OnGetID().Return("plugin2")
	maxLength := 8
	anotherPlugin.OnGetProperties().Return(pluginCore.PluginProperties{
		GeneratedNameMaxLength: &maxLength,
	})
	anotherTaskExecCtx, err := tk.newTaskExecutionContext(context.TODO(), nCtx, anotherPlugin)
	assert.NoError(t, err)
	assert.Equal(t, anotherTaskExecCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), "fpmmhh6q")
	assert.NotNil(t, anotherTaskExecCtx.ow)
	assert.Equal(t, storage.DataReference("s3://sandbox/x/fpmmhh6q"), anotherTaskExecCtx.ow.GetRawOutputPrefix())
	assert.Equal(t, storage.DataReference("s3://sandbox/x/fpmmhh6q/_flytecheckpoints"), anotherTaskExecCtx.ow.GetCheckpointPrefix())
	assert.Equal(t, storage.DataReference("s3://sandbox/x/fpqmhlei/_flytecheckpoints"), anotherTaskExecCtx.ow.GetPreviousCheckpointsPrefix())
	assert.NotNil(t, anotherTaskExecCtx.psm)
	assert.NotNil(t, anotherTaskExecCtx.ber)
	assert.NotNil(t, anotherTaskExecCtx.rm)
	assert.NotNil(t, anotherTaskExecCtx.sm)
	assert.NotNil(t, anotherTaskExecCtx.tm)
	assert.NotNil(t, anotherTaskExecCtx.tr)
}

func TestHandler_newTaskExecutionContext_taskExecutionID_WithParentInfo(t *testing.T) {
	parentInfo := &mocks2.ImmutableParentInfo{}
	parentInfo.OnGetUniqueID().Return("n0")
	parentInfo.OnCurrentAttempt().Return(uint32(2))

	nCtx := dummyNodeExecutionContext(t, parentInfo, v1alpha1.EventVersion1)
	p := dummyPlugin()
	tk := dummyHandler()
	got, err := tk.newTaskExecutionContext(context.TODO(), nCtx, p)
	assert.NoError(t, err)
	assert.NotNil(t, got)

	assert.Equal(t, got.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), "name-n0-2-n1-1")
	assert.Equal(t, got.TaskExecutionMetadata().GetTaskExecutionID().GetUniqueNodeID(), "n0-2-n1")
}

func TestGetGeneratedNameWith(t *testing.T) {
	t.Run("length 0", func(t *testing.T) {
		tCtx := taskExecutionID{
			execName: "exec name",
		}

		_, err := tCtx.GetGeneratedNameWith(0, 0)
		assert.Error(t, err)
	})

	t.Run("within range", func(t *testing.T) {
		tCtx := taskExecutionID{
			execName: "exec name name",
		}

		name, err := tCtx.GetGeneratedNameWith(10, 100)
		assert.NoError(t, err)
		assert.Equal(t, tCtx.execName, name)
	})

	t.Run("needs padding", func(t *testing.T) {
		tCtx := taskExecutionID{
			execName: "exec",
		}

		name, err := tCtx.GetGeneratedNameWith(10, 100)
		assert.NoError(t, err)
		assert.Equal(t, "exec000000", name)
	})
}

func TestAssignResource(t *testing.T) {
	type testCase struct {
		name              string
		execConfigRequest resource.Quantity
		execConfigLimit   resource.Quantity
		requests          corev1.ResourceList
		limits            corev1.ResourceList
		expectedRequests  corev1.ResourceList
		expectedLimits    corev1.ResourceList
	}
	var testCases = []testCase{
		{
			name:              "nothing to do",
			execConfigRequest: resource.MustParse("10"),
			execConfigLimit:   resource.MustParse("100"),
			requests: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("1"),
			},
			limits: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("1"),
			},
			expectedRequests: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("1"),
			},
			expectedLimits: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("1"),
			},
		},
		{
			name:              "assign request",
			execConfigRequest: resource.MustParse("10"),
			execConfigLimit:   resource.MustParse("100"),
			requests:          corev1.ResourceList{},
			limits: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("100"),
			},
			expectedRequests: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("10"),
			},
			expectedLimits: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("100"),
			},
		},
		{
			name:              "adjust request",
			execConfigRequest: resource.MustParse("10"),
			execConfigLimit:   resource.MustParse("100"),
			requests: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("1000"),
			},
			limits: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("100"),
			},
			expectedRequests: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("100"),
			},
			expectedLimits: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("100"),
			},
		},
		{
			name:              "assign limit",
			execConfigRequest: resource.MustParse("10"),
			execConfigLimit:   resource.MustParse("100"),
			requests: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("10"),
			},
			limits: corev1.ResourceList{},
			expectedRequests: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("10"),
			},
			expectedLimits: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("10"),
			},
		},
		{
			name:              "adjust limit based on exec config",
			execConfigRequest: resource.MustParse("10"),
			execConfigLimit:   resource.MustParse("100"),
			requests: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("10"),
			},
			limits: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("1000"),
			},
			expectedRequests: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("10"),
			},
			expectedLimits: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("100"),
			},
		},
		{
			name:              "assigned request should not exceed limit",
			execConfigRequest: resource.MustParse("10"),
			execConfigLimit:   resource.MustParse("100"),
			requests:          corev1.ResourceList{},
			limits: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("1"),
			},
			expectedRequests: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("1"),
			},
			expectedLimits: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("1"),
			},
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			assignResource(corev1.ResourceCPU, test.execConfigRequest, test.execConfigLimit, test.requests, test.limits)
			assert.EqualValues(t, test.requests, test.expectedRequests)
			assert.EqualValues(t, test.limits, test.expectedLimits)
		})
	}
}

func TestConvertTaskResourcesToRequirements(t *testing.T) {
	resourceRequirements := convertTaskResourcesToRequirements(v1alpha1.TaskResources{
		Requests: v1alpha1.TaskResourceSpec{
			CPU:              resource.MustParse("1"),
			Memory:           resource.MustParse("2"),
			EphemeralStorage: resource.MustParse("3"),
			GPU:              resource.MustParse("5"),
		},
		Limits: v1alpha1.TaskResourceSpec{
			CPU:              resource.MustParse("10"),
			Memory:           resource.MustParse("20"),
			EphemeralStorage: resource.MustParse("30"),
			GPU:              resource.MustParse("50"),
		},
	})
	assert.EqualValues(t, &corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:              resource.MustParse("1"),
			corev1.ResourceMemory:           resource.MustParse("2"),
			corev1.ResourceEphemeralStorage: resource.MustParse("3"),
			utils.ResourceNvidiaGPU:         resource.MustParse("5"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:              resource.MustParse("10"),
			corev1.ResourceMemory:           resource.MustParse("20"),
			corev1.ResourceEphemeralStorage: resource.MustParse("30"),
			utils.ResourceNvidiaGPU:         resource.MustParse("50"),
		},
	}, resourceRequirements)
}

func TestComputeRawOutputPrefix(t *testing.T) {

	nCtx := &nodeMocks.NodeExecutionContext{}
	nm := &nodeMocks.NodeExecutionMetadata{}
	nm.OnGetOwnerID().Return(types.NamespacedName{Namespace: "namespace", Name: "name"})
	nCtx.OnOutputShardSelector().Return(ioutils.NewConstantShardSelector([]string{"x"}))
	nCtx.OnRawOutputPrefix().Return("s3://sandbox/")
	ds, err := storage.NewDataStore(
		&storage.Config{
			Type: storage.TypeMemory,
		},
		promutils.NewTestScope(),
	)
	assert.NoError(t, err)
	nCtx.OnDataStore().Return(ds)
	nCtx.OnNodeExecutionMetadata().Return(nm)

	pre, uid, err := ComputeRawOutputPrefix(context.TODO(), 100, nCtx, "n1", 0)
	assert.NoError(t, err)
	assert.Equal(t, "name-n1-0", uid)
	assert.NotNil(t, pre)
	assert.Equal(t, storage.DataReference("s3://sandbox/x/name-n1-0"), pre.GetRawOutputPrefix())

	pre, uid, err = ComputeRawOutputPrefix(context.TODO(), 8, nCtx, "n1", 0)
	assert.NoError(t, err)
	assert.Equal(t, "fpqmhlei", uid)
	assert.NotNil(t, pre)
	assert.Equal(t, storage.DataReference("s3://sandbox/x/fpqmhlei"), pre.GetRawOutputPrefix())

	_, _, err = ComputeRawOutputPrefix(context.TODO(), 5, nCtx, "n1", 0)
	assert.Error(t, err)
}

func TestComputePreviousCheckpointPath(t *testing.T) {
	nCtx := &nodeMocks.NodeExecutionContext{}
	nm := &nodeMocks.NodeExecutionMetadata{}
	nm.OnGetOwnerID().Return(types.NamespacedName{Namespace: "namespace", Name: "name"})
	nCtx.OnOutputShardSelector().Return(ioutils.NewConstantShardSelector([]string{"x"}))
	nCtx.OnRawOutputPrefix().Return("s3://sandbox/")
	ds, err := storage.NewDataStore(
		&storage.Config{
			Type: storage.TypeMemory,
		},
		promutils.NewTestScope(),
	)
	assert.NoError(t, err)
	nCtx.OnDataStore().Return(ds)
	nCtx.OnNodeExecutionMetadata().Return(nm)
	reader := &nodeMocks.NodeStateReader{}
	reader.OnGetTaskNodeState().Return(handler.TaskNodeState{})
	nCtx.OnNodeStateReader().Return(reader)

	t.Run("attempt-0-nCtx", func(t *testing.T) {
		c, err := ComputePreviousCheckpointPath(context.TODO(), 100, nCtx, "n1", 0)
		assert.NoError(t, err)
		assert.Equal(t, storage.DataReference(""), c)
	})

	t.Run("attempt-1-nCtx", func(t *testing.T) {
		c, err := ComputePreviousCheckpointPath(context.TODO(), 100, nCtx, "n1", 1)
		assert.NoError(t, err)
		assert.Equal(t, storage.DataReference("s3://sandbox/x/name-n1-0/_flytecheckpoints"), c)
	})
}

func TestComputePreviousCheckpointPath_Recovery(t *testing.T) {
	nCtx := &nodeMocks.NodeExecutionContext{}
	nm := &nodeMocks.NodeExecutionMetadata{}
	nm.OnGetOwnerID().Return(types.NamespacedName{Namespace: "namespace", Name: "name"})
	nCtx.OnOutputShardSelector().Return(ioutils.NewConstantShardSelector([]string{"x"}))
	nCtx.OnRawOutputPrefix().Return("s3://sandbox/")
	ds, err := storage.NewDataStore(
		&storage.Config{
			Type: storage.TypeMemory,
		},
		promutils.NewTestScope(),
	)
	assert.NoError(t, err)
	nCtx.OnDataStore().Return(ds)
	nCtx.OnNodeExecutionMetadata().Return(nm)
	reader := &nodeMocks.NodeStateReader{}
	reader.OnGetTaskNodeState().Return(handler.TaskNodeState{
		PreviousNodeExecutionCheckpointURI: storage.DataReference("s3://sandbox/x/prevname-n1-0/_flytecheckpoints"),
	})
	nCtx.OnNodeStateReader().Return(reader)

	t.Run("recovery-attempt-0-nCtx", func(t *testing.T) {
		c, err := ComputePreviousCheckpointPath(context.TODO(), 100, nCtx, "n1", 0)
		assert.NoError(t, err)
		assert.Equal(t, storage.DataReference("s3://sandbox/x/prevname-n1-0/_flytecheckpoints"), c)
	})
	t.Run("recovery-attempt-1-nCtx", func(t *testing.T) {
		c, err := ComputePreviousCheckpointPath(context.TODO(), 100, nCtx, "n1", 1)
		assert.NoError(t, err)
		assert.Equal(t, storage.DataReference("s3://sandbox/x/name-n1-0/_flytecheckpoints"), c)
	})
}
