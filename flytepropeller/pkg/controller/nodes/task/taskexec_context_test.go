package task

import (
	"bytes"
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"

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
	corev1 "k8s.io/api/core/v1"
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

	res := &corev1.ResourceRequirements{
		Requests: make(corev1.ResourceList),
		Limits:   make(corev1.ResourceList),
	}
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

	assert.NotNil(t, got.TaskReader())
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

func TestDetermineResourceRequirements(t *testing.T) {
	node := &flyteMocks.ExecutableNode{}
	node.OnGetResources().Return(&corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("1"),
			corev1.ResourceMemory: resource.MustParse("10"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:              resource.MustParse("1"),
			corev1.ResourceMemory:           resource.MustParse("100"),
			corev1.ResourceEphemeralStorage: resource.MustParse("10"),
			corev1.ResourceStorage:          resource.MustParse("20"),
		},
	})
	nodeExecutionContext := &nodeMocks.NodeExecutionContext{}
	nodeExecutionContext.OnNode().Return(node)

	taskResources := v1alpha1.TaskResources{
		Requests: v1alpha1.TaskResourceSpec{
			CPU:              resource.MustParse("1"),
			Memory:           resource.MustParse("20"),
			EphemeralStorage: resource.MustParse("50"),
		},
		Limits: v1alpha1.TaskResourceSpec{
			CPU:              resource.MustParse("2"),
			Memory:           resource.MustParse("50"),
			EphemeralStorage: resource.MustParse("100"),
			Storage:          resource.MustParse("15"),
		},
	}
	resources := determineResourceRequirements(nodeExecutionContext, taskResources)
	assert.EqualValues(t, resources.Requests, corev1.ResourceList{
		corev1.ResourceCPU:              resource.MustParse("1"),
		corev1.ResourceMemory:           resource.MustParse("10"),
		corev1.ResourceEphemeralStorage: resource.MustParse("10"),
	})
	assert.EqualValues(t, resources.Limits, corev1.ResourceList{
		corev1.ResourceCPU:              resource.MustParse("1"),
		corev1.ResourceMemory:           resource.MustParse("50"),
		corev1.ResourceEphemeralStorage: resource.MustParse("10"),
		corev1.ResourceStorage:          resource.MustParse("15"),
	})
}

func TestGetResources(t *testing.T) {
	node := &flyteMocks.ExecutableNode{}
	node.OnGetResources().Return(&corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("1"),
			corev1.ResourceMemory: resource.MustParse("10"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:              resource.MustParse("1"),
			corev1.ResourceMemory:           resource.MustParse("100"),
			corev1.ResourceEphemeralStorage: resource.MustParse("10"),
		},
	})

	computedRequirements := &corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("2"),
			corev1.ResourceMemory: resource.MustParse("20"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:              resource.MustParse("2"),
			corev1.ResourceMemory:           resource.MustParse("200"),
			corev1.ResourceEphemeralStorage: resource.MustParse("20"),
		},
	}

	overrides := newTaskOverrides(node, computedRequirements)
	assert.EqualValues(t, overrides.GetResources(), computedRequirements)
}
