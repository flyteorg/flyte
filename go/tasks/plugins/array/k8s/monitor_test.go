package k8s

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flyteplugins/go/tasks/logs"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/workqueue"

	core2 "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	mocks2 "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/arraystatus"
	"github.com/flyteorg/flytestdlib/bitarray"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"

	arrayCore "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/core"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func createSampleContainerTask() *core2.Container {
	return &core2.Container{
		Command: []string{"cmd"},
		Args:    []string{"{{$inputPrefix}}"},
		Image:   "img1",
	}
}

func getMockTaskExecutionContext(ctx context.Context) *mocks.TaskExecutionContext {
	tr := &mocks.TaskReader{}
	tr.OnRead(ctx).Return(&core2.TaskTemplate{
		Target: &core2.TaskTemplate_Container{
			Container: createSampleContainerTask(),
		},
	}, nil)

	tID := &mocks.TaskExecutionID{}
	tID.OnGetGeneratedName().Return("notfound")
	tID.OnGetID().Return(core2.TaskExecutionIdentifier{
		TaskId: &core2.Identifier{
			ResourceType: core2.ResourceType_TASK,
			Project:      "a",
			Domain:       "d",
			Name:         "n",
			Version:      "abc",
		},
		NodeExecutionId: &core2.NodeExecutionIdentifier{
			NodeId: "node1",
			ExecutionId: &core2.WorkflowExecutionIdentifier{
				Project: "a",
				Domain:  "d",
				Name:    "exec",
			},
		},
		RetryAttempt: 0,
	})

	overrides := &mocks.TaskOverrides{}
	overrides.OnGetResources().Return(&v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU: resource.MustParse("10"),
		},
	})

	tMeta := &mocks.TaskExecutionMetadata{}
	tMeta.OnGetTaskExecutionID().Return(tID)
	tMeta.OnGetOverrides().Return(overrides)
	tMeta.OnIsInterruptible().Return(false)
	tMeta.OnGetK8sServiceAccount().Return("s")

	tMeta.OnGetNamespace().Return("n")
	tMeta.OnGetLabels().Return(nil)
	tMeta.OnGetAnnotations().Return(nil)
	tMeta.OnGetOwnerReference().Return(v12.OwnerReference{})

	ow := &mocks2.OutputWriter{}
	ow.OnGetOutputPrefixPath().Return("/prefix/")
	ow.OnGetRawOutputPrefix().Return("/raw_prefix/")

	ir := &mocks2.InputReader{}
	ir.OnGetInputPrefixPath().Return("/prefix/")
	ir.OnGetInputPath().Return("/prefix/inputs.pb")
	ir.OnGetMatch(mock.Anything).Return(&core2.LiteralMap{}, nil)

	tCtx := &mocks.TaskExecutionContext{}
	tCtx.OnTaskReader().Return(tr)
	tCtx.OnTaskExecutionMetadata().Return(tMeta)
	tCtx.OnOutputWriter().Return(ow)
	tCtx.OnInputReader().Return(ir)
	return tCtx
}

func TestGetNamespaceForExecution(t *testing.T) {
	ctx := context.Background()
	tCtx := getMockTaskExecutionContext(ctx)

	assert.Equal(t, GetNamespaceForExecution(tCtx, ""), tCtx.TaskExecutionMetadata().GetNamespace())
	assert.Equal(t, GetNamespaceForExecution(tCtx, "abcd"), "abcd")
	assert.Equal(t, GetNamespaceForExecution(tCtx, "a-{{.namespace}}-b"), fmt.Sprintf("a-%s-b", tCtx.TaskExecutionMetadata().GetNamespace()))
}

func testSubTaskIDs(t *testing.T, actual []*string) {
	var expected = make([]*string, 5)
	for i := 0; i < len(expected); i++ {
		subTaskID := fmt.Sprintf("notfound-%d", i)
		expected[i] = &subTaskID
	}
	assert.EqualValues(t, expected, actual)
}

func TestCheckSubTasksState(t *testing.T) {
	ctx := context.Background()

	tCtx := getMockTaskExecutionContext(ctx)
	kubeClient := mocks.KubeClient{}
	kubeClient.OnGetClient().Return(mocks.NewFakeKubeClient())
	kubeClient.OnGetCache().Return(mocks.NewFakeKubeCache())

	resourceManager := mocks.ResourceManager{}
	resourceManager.OnAllocateResourceMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(core.AllocationStatusExhausted, nil)
	tCtx.OnResourceManager().Return(&resourceManager)

	t.Run("Happy case", func(t *testing.T) {
		config := Config{
			MaxArrayJobSize:      100,
			MaxErrorStringLength: 200,
			NamespaceTemplate:    "a-{{.namespace}}-b",
			OutputAssembler: workqueue.Config{
				Workers:            2,
				MaxRetries:         0,
				IndexCacheMaxItems: 100,
			},
			ErrorAssembler: workqueue.Config{
				Workers:            2,
				MaxRetries:         0,
				IndexCacheMaxItems: 100,
			},
			LogConfig: LogConfig{
				Config: logs.LogConfig{
					IsCloudwatchEnabled:   true,
					CloudwatchTemplateURI: "https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logStream:group=/kubernetes/flyte;prefix=var.log.containers.{{ .podName }};streamFilter=typeLogStreamPrefix",
					IsKubernetesEnabled:   true,
					KubernetesTemplateURI: "k8s/log/{{.namespace}}/{{.podName}}/pod?namespace={{.namespace}}",
				}},
		}
		cacheIndexes := bitarray.NewBitSet(5)
		cacheIndexes.Set(0)
		cacheIndexes.Set(1)
		cacheIndexes.Set(2)
		cacheIndexes.Set(3)
		cacheIndexes.Set(4)

		newState, logLinks, subTaskIDs, err := LaunchAndCheckSubTasksState(ctx, tCtx, &kubeClient, &config, nil, "/prefix/", "/prefix-sand/", &arrayCore.State{
			CurrentPhase:         arrayCore.PhaseCheckingSubTaskExecutions,
			ExecutionArraySize:   5,
			OriginalArraySize:    10,
			OriginalMinSuccesses: 5,
			IndexesToCache:       cacheIndexes,
		})

		assert.Nil(t, err)
		assert.NotEmpty(t, logLinks)
		assert.Equal(t, 10, len(logLinks))
		for i := 0; i < 10; i = i + 2 {
			assert.Equal(t, fmt.Sprintf("Kubernetes Logs #0-%d (PhaseRunning)", i/2), logLinks[i].Name)
			assert.Equal(t, fmt.Sprintf("k8s/log/a-n-b/notfound-%d/pod?namespace=a-n-b", i/2), logLinks[i].Uri)

			assert.Equal(t, fmt.Sprintf("Cloudwatch Logs #0-%d (PhaseRunning)", i/2), logLinks[i+1].Name)
			assert.Equal(t, fmt.Sprintf("https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logStream:group=/kubernetes/flyte;prefix=var.log.containers.notfound-%d;streamFilter=typeLogStreamPrefix", i/2), logLinks[i+1].Uri)
		}

		p, _ := newState.GetPhase()
		assert.Equal(t, arrayCore.PhaseCheckingSubTaskExecutions.String(), p.String())
		resourceManager.AssertNumberOfCalls(t, "AllocateResource", 0)
		testSubTaskIDs(t, subTaskIDs)
	})

	t.Run("Resource exhausted", func(t *testing.T) {
		config := Config{
			MaxArrayJobSize: 100,
			ResourceConfig: ResourceConfig{
				PrimaryLabel: "p",
				Limit:        10,
			},
		}

		newState, _, subTaskIDs, err := LaunchAndCheckSubTasksState(ctx, tCtx, &kubeClient, &config, nil, "/prefix/", "/prefix-sand/", &arrayCore.State{
			CurrentPhase:         arrayCore.PhaseCheckingSubTaskExecutions,
			ExecutionArraySize:   5,
			OriginalArraySize:    10,
			OriginalMinSuccesses: 5,
			ArrayStatus: arraystatus.ArrayStatus{
				Detailed: arrayCore.NewPhasesCompactArray(uint(5)),
			},
		})

		assert.Nil(t, err)
		p, _ := newState.GetPhase()
		assert.Equal(t, arrayCore.PhaseWaitingForResources.String(), p.String())
		resourceManager.AssertNumberOfCalls(t, "AllocateResource", 5)
		assert.Empty(t, subTaskIDs, "subtask ids are only populated when monitor is called for a successfully launched task")
	})
}

func TestCheckSubTasksStateResourceGranted(t *testing.T) {
	ctx := context.Background()

	tCtx := getMockTaskExecutionContext(ctx)
	kubeClient := mocks.KubeClient{}
	kubeClient.OnGetClient().Return(mocks.NewFakeKubeClient())
	kubeClient.OnGetCache().Return(mocks.NewFakeKubeCache())

	resourceManager := mocks.ResourceManager{}

	resourceManager.OnAllocateResourceMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(core.AllocationStatusGranted, nil)
	resourceManager.OnReleaseResourceMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil)
	tCtx.OnResourceManager().Return(&resourceManager)

	t.Run("Resource granted", func(t *testing.T) {
		config := Config{
			MaxArrayJobSize: 100,
			ResourceConfig: ResourceConfig{
				PrimaryLabel: "p",
				Limit:        10,
			},
		}

		cacheIndexes := bitarray.NewBitSet(5)
		newState, _, subTaskIDs, err := LaunchAndCheckSubTasksState(ctx, tCtx, &kubeClient, &config, nil, "/prefix/", "/prefix-sand/", &arrayCore.State{
			CurrentPhase:         arrayCore.PhaseCheckingSubTaskExecutions,
			ExecutionArraySize:   5,
			OriginalArraySize:    10,
			OriginalMinSuccesses: 5,
			IndexesToCache:       cacheIndexes,
			ArrayStatus: arraystatus.ArrayStatus{
				Detailed: arrayCore.NewPhasesCompactArray(uint(5)),
			},
		})

		assert.Nil(t, err)
		p, _ := newState.GetPhase()
		assert.Equal(t, arrayCore.PhaseCheckingSubTaskExecutions.String(), p.String())
		resourceManager.AssertNumberOfCalls(t, "AllocateResource", 5)
		testSubTaskIDs(t, subTaskIDs)
	})

	t.Run("All tasks success", func(t *testing.T) {
		config := Config{
			MaxArrayJobSize: 100,
			ResourceConfig: ResourceConfig{
				PrimaryLabel: "p",
				Limit:        10,
			},
		}

		arrayStatus := &arraystatus.ArrayStatus{
			Summary:  arraystatus.ArraySummary{},
			Detailed: arrayCore.NewPhasesCompactArray(uint(5)),
		}
		for childIdx := range arrayStatus.Detailed.GetItems() {
			arrayStatus.Detailed.SetItem(childIdx, bitarray.Item(core.PhaseSuccess))

		}
		cacheIndexes := bitarray.NewBitSet(5)
		newState, _, subTaskIDs, err := LaunchAndCheckSubTasksState(ctx, tCtx, &kubeClient, &config, nil, "/prefix/", "/prefix-sand/", &arrayCore.State{
			CurrentPhase:         arrayCore.PhaseCheckingSubTaskExecutions,
			ExecutionArraySize:   5,
			OriginalArraySize:    10,
			OriginalMinSuccesses: 5,
			ArrayStatus:          *arrayStatus,
			IndexesToCache:       cacheIndexes,
		})

		assert.Nil(t, err)
		p, _ := newState.GetPhase()
		assert.Equal(t, arrayCore.PhaseWriteToDiscovery.String(), p.String())
		resourceManager.AssertNumberOfCalls(t, "ReleaseResource", 5)
		assert.Empty(t, subTaskIDs, "terminal phases don't need to collect subtask IDs")
	})
}
