package k8s

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
	structpb "google.golang.org/protobuf/types/known/structpb"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	core2 "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/logs"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	mocks2 "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/workqueue"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/array/arraystatus"
	arrayCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/array/core"
	"github.com/flyteorg/flyte/flytestdlib/bitarray"
	"github.com/flyteorg/flyte/flytestdlib/storage"
	stdmocks "github.com/flyteorg/flyte/flytestdlib/storage/mocks"
)

type metadata struct {
	exists     bool
	size       int64
	etag       string
	contentMD5 string
}

func (m metadata) Exists() bool {
	return m.exists
}

func (m metadata) Size() int64 {
	return m.size
}

func (m metadata) Etag() string {
	return m.etag
}

func (m metadata) ContentMD5() string {
	return m.contentMD5
}

func createSampleContainerTask() *core2.Container {
	return &core2.Container{
		Command: []string{"cmd"},
		Args:    []string{"{{$inputPrefix}}"},
		Image:   "img1",
	}
}

func getMockTaskExecutionContext(ctx context.Context, parallelism int) *mocks.TaskExecutionContext {
	customStruct, _ := structpb.NewStruct(map[string]interface{}{
		"parallelism": fmt.Sprintf("%d", parallelism),
	})

	tr := &mocks.TaskReader{}
	tr.EXPECT().Read(ctx).Return(&core2.TaskTemplate{
		Custom: customStruct,
		Target: &core2.TaskTemplate_Container{
			Container: createSampleContainerTask(),
		},
	}, nil)

	tID := &mocks.TaskExecutionID{}
	tID.EXPECT().GetGeneratedName().Return("notfound")
	tID.On("GetUniqueNodeID").Return("an-unique-id")
	tID.EXPECT().GetID().Return(core2.TaskExecutionIdentifier{
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
	overrides.EXPECT().GetResources().Return(&v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU: resource.MustParse("10"),
		},
	})
	overrides.EXPECT().GetExtendedResources().Return(nil)
	overrides.EXPECT().GetContainerImage().Return("")
	overrides.EXPECT().GetPodTemplate().Return(nil)

	tMeta := &mocks.TaskExecutionMetadata{}
	tMeta.EXPECT().GetTaskExecutionID().Return(tID)
	tMeta.EXPECT().GetOverrides().Return(overrides)
	tMeta.EXPECT().IsInterruptible().Return(false)
	tMeta.EXPECT().GetK8sServiceAccount().Return("s")
	tMeta.EXPECT().GetSecurityContext().Return(core2.SecurityContext{})

	tMeta.EXPECT().GetMaxAttempts().Return(2)
	tMeta.EXPECT().GetNamespace().Return("n")
	tMeta.EXPECT().GetLabels().Return(nil)
	tMeta.EXPECT().GetAnnotations().Return(nil)
	tMeta.EXPECT().GetOwnerReference().Return(metav1.OwnerReference{})
	tMeta.EXPECT().GetPlatformResources().Return(&v1.ResourceRequirements{})
	tMeta.EXPECT().GetInterruptibleFailureThreshold().Return(2)
	tMeta.EXPECT().GetEnvironmentVariables().Return(nil)
	tMeta.EXPECT().GetConsoleURL().Return("")
	tMeta.EXPECT().GetOnOOMConfig().Return(nil)

	ow := &mocks2.OutputWriter{}
	ow.EXPECT().GetOutputPrefixPath().Return("/prefix/")
	ow.EXPECT().GetRawOutputPrefix().Return("/raw_prefix/")
	ow.EXPECT().GetCheckpointPrefix().Return("/checkpoint")
	ow.EXPECT().GetPreviousCheckpointsPrefix().Return("/prev")

	ir := &mocks2.InputReader{}
	ir.EXPECT().GetInputPrefixPath().Return("/prefix/")
	ir.EXPECT().GetInputPath().Return("/prefix/inputs.pb")
	ir.EXPECT().Get(mock.Anything).Return(&core2.LiteralMap{}, nil)

	composedProtobufStore := &stdmocks.ComposedProtobufStore{}
	matchedBy := mock.MatchedBy(func(s storage.DataReference) bool {
		return true
	})
	composedProtobufStore.On("Head", mock.Anything, matchedBy).Return(metadata{true, 0, "", ""}, nil)
	dataStore := &storage.DataStore{
		ComposedProtobufStore: composedProtobufStore,
		ReferenceConstructor:  &storage.URLPathConstructor{},
	}

	pluginStateReader := &mocks.PluginStateReader{}
	pluginStateReader.EXPECT().Get(mock.Anything).Return(0, nil)

	tCtx := &mocks.TaskExecutionContext{}
	tCtx.EXPECT().TaskReader().Return(tr)
	tCtx.EXPECT().TaskExecutionMetadata().Return(tMeta)
	tCtx.EXPECT().OutputWriter().Return(ow)
	tCtx.EXPECT().InputReader().Return(ir)
	tCtx.EXPECT().DataStore().Return(dataStore)
	tCtx.EXPECT().PluginStateReader().Return(pluginStateReader)
	return tCtx
}

func TestCheckSubTasksState(t *testing.T) {
	ctx := context.Background()
	subtaskCount := 5

	config := Config{
		MaxArrayJobSize: int64(subtaskCount * 10),
		ResourceConfig: ResourceConfig{
			PrimaryLabel: "p",
			Limit:        subtaskCount,
		},
	}

	fakeKubeClient := mocks.NewFakeKubeClient()
	fakeKubeCache := mocks.NewFakeKubeCache()

	for i := 0; i < subtaskCount; i++ {
		pod := flytek8s.BuildIdentityPod()
		pod.SetName(fmt.Sprintf("notfound-%d", i))
		pod.SetNamespace("a-n-b")
		pod.Spec.Containers = append(pod.Spec.Containers, v1.Container{Name: "foo"})

		pod.Status.Phase = v1.PodRunning
		pod.Status.ContainerStatuses = []v1.ContainerStatus{
			v1.ContainerStatus{
				State: v1.ContainerState{
					Running: &v1.ContainerStateRunning{},
				},
			},
		}
		_ = fakeKubeClient.Create(ctx, pod)
		_ = fakeKubeCache.Create(ctx, pod)
	}

	failureFakeKubeClient := mocks.NewFakeKubeClient()
	failureFakeKubeCache := mocks.NewFakeKubeCache()

	for i := 0; i < subtaskCount; i++ {
		pod := flytek8s.BuildIdentityPod()
		pod.SetName(fmt.Sprintf("notfound-%d", i))
		pod.SetNamespace("a-n-b")
		pod.Spec.Containers = append(pod.Spec.Containers, v1.Container{Name: "foo"})

		pod.Status.Phase = v1.PodFailed
		_ = failureFakeKubeClient.Create(ctx, pod)
		_ = failureFakeKubeCache.Create(ctx, pod)
	}

	t.Run("Launch", func(t *testing.T) {
		// initialize metadata
		kubeClient := mocks.KubeClient{}
		kubeClient.EXPECT().GetClient().Return(mocks.NewFakeKubeClient())
		kubeClient.EXPECT().GetCache().Return(mocks.NewFakeKubeCache())

		resourceManager := mocks.ResourceManager{}
		resourceManager.EXPECT().AllocateResource(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(core.AllocationStatusGranted, nil)

		tCtx := getMockTaskExecutionContext(ctx, 0)
		tCtx.EXPECT().ResourceManager().Return(&resourceManager)

		currentState := &arrayCore.State{
			CurrentPhase:         arrayCore.PhaseCheckingSubTaskExecutions,
			ExecutionArraySize:   subtaskCount,
			OriginalArraySize:    int64(subtaskCount),
			OriginalMinSuccesses: int64(subtaskCount),
			ArrayStatus: arraystatus.ArrayStatus{
				// #nosec G115
				Detailed: arrayCore.NewPhasesCompactArray(uint(subtaskCount)), // set all tasks to core.PhaseUndefined
			},
			// #nosec G115
			IndexesToCache: arrayCore.InvertBitSet(bitarray.NewBitSet(uint(subtaskCount)), uint(subtaskCount)), // set all tasks to be cached
		}

		// execute
		newState, _, err := LaunchAndCheckSubTasksState(ctx, tCtx, &kubeClient, &config, nil, "/prefix/", "/prefix-sand/", currentState)

		// validate results
		assert.Nil(t, err)
		p, _ := newState.GetPhase()
		assert.Equal(t, arrayCore.PhaseCheckingSubTaskExecutions.String(), p.String())
		resourceManager.AssertNumberOfCalls(t, "AllocateResource", subtaskCount)
		for _, subtaskPhaseIndex := range newState.GetArrayStatus().Detailed.GetItems() {
			assert.Equal(t, core.PhaseQueued, core.Phases[subtaskPhaseIndex])
		}
	})

	for i := 1; i <= subtaskCount; i++ {
		t.Run(fmt.Sprintf("LaunchParallelism%d", i), func(t *testing.T) {
			// initialize metadata
			kubeClient := mocks.KubeClient{}
			kubeClient.EXPECT().GetClient().Return(mocks.NewFakeKubeClient())
			kubeClient.EXPECT().GetCache().Return(mocks.NewFakeKubeCache())

			resourceManager := mocks.ResourceManager{}
			resourceManager.EXPECT().AllocateResource(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(core.AllocationStatusGranted, nil)

			tCtx := getMockTaskExecutionContext(ctx, i)
			tCtx.EXPECT().ResourceManager().Return(&resourceManager)

			currentState := &arrayCore.State{
				CurrentPhase:         arrayCore.PhaseCheckingSubTaskExecutions,
				ExecutionArraySize:   subtaskCount,
				OriginalArraySize:    int64(subtaskCount),
				OriginalMinSuccesses: int64(subtaskCount),
				ArrayStatus: arraystatus.ArrayStatus{
					// #nosec G115
					Detailed: arrayCore.NewPhasesCompactArray(uint(subtaskCount)), // set all tasks to core.PhaseUndefined
				},
				// #nosec G115
				IndexesToCache: arrayCore.InvertBitSet(bitarray.NewBitSet(uint(subtaskCount)), uint(subtaskCount)), // set all tasks to be cached
			}

			// execute
			newState, _, err := LaunchAndCheckSubTasksState(ctx, tCtx, &kubeClient, &config, nil, "/prefix/", "/prefix-sand/", currentState)

			// validate results
			assert.Nil(t, err)
			p, _ := newState.GetPhase()
			assert.Equal(t, arrayCore.PhaseCheckingSubTaskExecutions.String(), p.String())

			executed := 0
			for _, existingPhaseIdx := range newState.GetArrayStatus().Detailed.GetItems() {
				if core.Phases[existingPhaseIdx] != core.PhaseUndefined {
					executed++
				}
			}

			assert.Equal(t, i, executed)
		})
	}

	t.Run("LaunchResourcesExhausted", func(t *testing.T) {
		// initialize metadata
		kubeClient := mocks.KubeClient{}
		kubeClient.EXPECT().GetClient().Return(mocks.NewFakeKubeClient())
		kubeClient.EXPECT().GetCache().Return(mocks.NewFakeKubeCache())

		resourceManager := mocks.ResourceManager{}
		resourceManager.EXPECT().AllocateResource(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(core.AllocationStatusExhausted, nil)

		tCtx := getMockTaskExecutionContext(ctx, 0)
		tCtx.EXPECT().ResourceManager().Return(&resourceManager)

		currentState := &arrayCore.State{
			CurrentPhase:         arrayCore.PhaseCheckingSubTaskExecutions,
			ExecutionArraySize:   subtaskCount,
			OriginalArraySize:    int64(subtaskCount),
			OriginalMinSuccesses: int64(subtaskCount),
			ArrayStatus: arraystatus.ArrayStatus{
				// #nosec G115
				Detailed: arrayCore.NewPhasesCompactArray(uint(subtaskCount)), // set all tasks to core.PhaseUndefined
			},
			// #nosec G115
			IndexesToCache: arrayCore.InvertBitSet(bitarray.NewBitSet(uint(subtaskCount)), uint(subtaskCount)), // set all tasks to be cached
		}

		// execute
		newState, _, err := LaunchAndCheckSubTasksState(ctx, tCtx, &kubeClient, &config, nil, "/prefix/", "/prefix-sand/", currentState)

		// validate results
		assert.Nil(t, err)
		p, _ := newState.GetPhase()
		assert.Equal(t, arrayCore.PhaseCheckingSubTaskExecutions.String(), p.String())
		resourceManager.AssertNumberOfCalls(t, "AllocateResource", subtaskCount)
		for _, subtaskPhaseIndex := range newState.GetArrayStatus().Detailed.GetItems() {
			assert.Equal(t, core.PhaseWaitingForResources, core.Phases[subtaskPhaseIndex])
		}

		// execute again - with resources available and validate results
		nresourceManager := mocks.ResourceManager{}
		nresourceManager.EXPECT().AllocateResource(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(core.AllocationStatusGranted, nil)

		ntCtx := getMockTaskExecutionContext(ctx, 0)
		ntCtx.EXPECT().ResourceManager().Return(&nresourceManager)

		lastState, _, err := LaunchAndCheckSubTasksState(ctx, ntCtx, &kubeClient, &config, nil, "/prefix/", "/prefix-sand/", newState)
		assert.Nil(t, err)
		np, _ := lastState.GetPhase()
		assert.Equal(t, arrayCore.PhaseCheckingSubTaskExecutions.String(), np.String())
		resourceManager.AssertNumberOfCalls(t, "AllocateResource", subtaskCount)
		for _, subtaskPhaseIndex := range lastState.GetArrayStatus().Detailed.GetItems() {
			assert.Equal(t, core.PhaseQueued, core.Phases[subtaskPhaseIndex])
		}
	})

	t.Run("LaunchRetryableFailures", func(t *testing.T) {
		// initialize metadata
		kubeClient := mocks.KubeClient{}
		kubeClient.EXPECT().GetClient().Return(fakeKubeClient)
		kubeClient.EXPECT().GetCache().Return(fakeKubeCache)

		resourceManager := mocks.ResourceManager{}
		resourceManager.EXPECT().AllocateResource(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(core.AllocationStatusGranted, nil)

		tCtx := getMockTaskExecutionContext(ctx, 0)
		tCtx.EXPECT().ResourceManager().Return(&resourceManager)

		detailed := arrayCore.NewPhasesCompactArray(uint(subtaskCount)) // #nosec G115
		for i := 0; i < subtaskCount; i++ {
			detailed.SetItem(i, bitarray.Item(core.PhaseRetryableFailure)) // set all tasks to core.PhaseRetryableFailure
		}

		retryAttemptsArray, err := bitarray.NewCompactArray(uint(subtaskCount), bitarray.Item(1)) // #nosec G115
		assert.NoError(t, err)

		currentState := &arrayCore.State{
			CurrentPhase:         arrayCore.PhaseCheckingSubTaskExecutions,
			ExecutionArraySize:   subtaskCount,
			OriginalArraySize:    int64(subtaskCount),
			OriginalMinSuccesses: int64(subtaskCount),
			ArrayStatus: arraystatus.ArrayStatus{
				Detailed: detailed,
			},
			// #nosec G115
			IndexesToCache: arrayCore.InvertBitSet(bitarray.NewBitSet(uint(subtaskCount)), uint(subtaskCount)), // set all tasks to be cached
			RetryAttempts:  retryAttemptsArray,
		}

		// execute
		newState, _, err := LaunchAndCheckSubTasksState(ctx, tCtx, &kubeClient, &config, nil, "/prefix/", "/prefix-sand/", currentState)

		// validate results
		assert.Nil(t, err)
		p, _ := newState.GetPhase()
		assert.Equal(t, arrayCore.PhaseCheckingSubTaskExecutions.String(), p.String())
		resourceManager.AssertNumberOfCalls(t, "AllocateResource", subtaskCount)
		for i, subtaskPhaseIndex := range newState.GetArrayStatus().Detailed.GetItems() {
			assert.Equal(t, core.PhaseQueued, core.Phases[subtaskPhaseIndex])
			assert.Equal(t, bitarray.Item(1), newState.RetryAttempts.GetItem(i))
		}
	})

	t.Run("RunningLogLinksAndSubtaskIDs", func(t *testing.T) {
		// initialize metadata
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

		kubeClient := mocks.KubeClient{}
		kubeClient.EXPECT().GetClient().Return(fakeKubeClient)
		kubeClient.EXPECT().GetCache().Return(fakeKubeCache)

		resourceManager := mocks.ResourceManager{}
		resourceManager.EXPECT().AllocateResource(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(core.AllocationStatusExhausted, nil)

		tCtx := getMockTaskExecutionContext(ctx, 0)
		tCtx.EXPECT().ResourceManager().Return(&resourceManager)

		detailed := arrayCore.NewPhasesCompactArray(uint(subtaskCount)) // #nosec G115
		for i := 0; i < subtaskCount; i++ {
			// #nosec G115
			detailed.SetItem(i, bitarray.Item(core.PhaseRunning)) // set all tasks to core.PhaseRunning
		}

		currentState := &arrayCore.State{
			CurrentPhase:         arrayCore.PhaseCheckingSubTaskExecutions,
			ExecutionArraySize:   subtaskCount,
			OriginalArraySize:    int64(subtaskCount),
			OriginalMinSuccesses: int64(subtaskCount),
			ArrayStatus: arraystatus.ArrayStatus{
				Detailed: detailed,
			},
			// #nosec G115
			IndexesToCache: arrayCore.InvertBitSet(bitarray.NewBitSet(uint(subtaskCount)), uint(subtaskCount)), // set all tasks to be cached
		}

		// execute
		newState, externalResources, err := LaunchAndCheckSubTasksState(ctx, tCtx, &kubeClient, &config, nil, "/prefix/", "/prefix-sand/", currentState)

		// validate results
		assert.Nil(t, err)
		p, _ := newState.GetPhase()
		assert.Equal(t, arrayCore.PhaseCheckingSubTaskExecutions.String(), p.String())

		resourceManager.AssertNumberOfCalls(t, "AllocateResource", 0)
		resourceManager.AssertNumberOfCalls(t, "ReleaseResource", 0)

		assert.Equal(t, subtaskCount, len(externalResources))
		for i := 0; i < subtaskCount; i++ {
			externalResource := externalResources[i]
			assert.Equal(t, fmt.Sprintf("notfound-%d", i), externalResource.ExternalID)

			logLinks := externalResource.Logs
			assert.Equal(t, 2, len(logLinks))
			assert.Equal(t, fmt.Sprintf("Kubernetes Logs #0-%d", i), logLinks[0].GetName())
			assert.Equal(t, fmt.Sprintf("k8s/log/a-n-b/notfound-%d/pod?namespace=a-n-b", i), logLinks[0].GetUri())
			assert.Equal(t, fmt.Sprintf("Cloudwatch Logs #0-%d", i), logLinks[1].GetName())
			assert.Equal(t, fmt.Sprintf("https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logStream:group=/kubernetes/flyte;prefix=var.log.containers.notfound-%d;streamFilter=typeLogStreamPrefix", i), logLinks[1].GetUri())
		}
	})

	t.Run("RunningRetryableFailures", func(t *testing.T) {
		// initialize metadata
		kubeClient := mocks.KubeClient{}
		kubeClient.EXPECT().GetClient().Return(failureFakeKubeClient)
		kubeClient.EXPECT().GetCache().Return(failureFakeKubeCache)

		resourceManager := mocks.ResourceManager{}
		resourceManager.EXPECT().ReleaseResource(mock.Anything, mock.Anything, mock.Anything).Return(nil)

		tCtx := getMockTaskExecutionContext(ctx, 0)
		tCtx.EXPECT().ResourceManager().Return(&resourceManager)

		detailed := arrayCore.NewPhasesCompactArray(uint(subtaskCount)) // #nosec G115
		for i := 0; i < subtaskCount; i++ {
			// #nosec G115
			detailed.SetItem(i, bitarray.Item(core.PhaseRunning)) // set all tasks to core.PhaseRunning
		}

		retryAttemptsArray, err := bitarray.NewCompactArray(uint(subtaskCount), bitarray.Item(1)) // #nosec G115
		assert.NoError(t, err)

		currentState := &arrayCore.State{
			CurrentPhase:         arrayCore.PhaseCheckingSubTaskExecutions,
			ExecutionArraySize:   subtaskCount,
			OriginalArraySize:    int64(subtaskCount),
			OriginalMinSuccesses: int64(subtaskCount),
			ArrayStatus: arraystatus.ArrayStatus{
				Detailed: detailed,
			},
			// #nosec G115
			IndexesToCache: arrayCore.InvertBitSet(bitarray.NewBitSet(uint(subtaskCount)), uint(subtaskCount)), // set all tasks to be cached
			RetryAttempts:  retryAttemptsArray,
		}

		// execute
		newState, _, err := LaunchAndCheckSubTasksState(ctx, tCtx, &kubeClient, &config, tCtx.DataStore(), "/prefix/", "/prefix-sand/", currentState)

		// validate results
		assert.Nil(t, err)
		p, _ := newState.GetPhase()
		assert.Equal(t, arrayCore.PhaseCheckingSubTaskExecutions.String(), p.String())
		resourceManager.AssertNumberOfCalls(t, "ReleaseResource", subtaskCount)
		for _, subtaskPhaseIndex := range newState.GetArrayStatus().Detailed.GetItems() {
			assert.Equal(t, core.PhaseRetryableFailure, core.Phases[subtaskPhaseIndex])
		}
	})

	t.Run("RunningPermanentFailures", func(t *testing.T) {
		// initialize metadata
		kubeClient := mocks.KubeClient{}
		kubeClient.EXPECT().GetClient().Return(failureFakeKubeClient)
		kubeClient.EXPECT().GetCache().Return(failureFakeKubeCache)

		resourceManager := mocks.ResourceManager{}
		resourceManager.EXPECT().ReleaseResource(mock.Anything, mock.Anything, mock.Anything).Return(nil)

		tCtx := getMockTaskExecutionContext(ctx, 0)
		tCtx.EXPECT().ResourceManager().Return(&resourceManager)

		// #nosec G115
		detailed := arrayCore.NewPhasesCompactArray(uint(subtaskCount))
		for i := 0; i < subtaskCount; i++ {
			detailed.SetItem(i, bitarray.Item(core.PhaseRunning)) // set all tasks to core.PhaseRunning
		}

		// #nosec G115
		retryAttemptsArray, err := bitarray.NewCompactArray(uint(subtaskCount), bitarray.Item(1))
		assert.NoError(t, err)

		for i := 0; i < subtaskCount; i++ {
			retryAttemptsArray.SetItem(i, bitarray.Item(1))
		}

		currentState := &arrayCore.State{
			CurrentPhase:         arrayCore.PhaseCheckingSubTaskExecutions,
			ExecutionArraySize:   subtaskCount,
			OriginalArraySize:    int64(subtaskCount),
			OriginalMinSuccesses: int64(subtaskCount),
			ArrayStatus: arraystatus.ArrayStatus{
				Detailed: detailed,
			},
			// #nosec G115
			IndexesToCache: arrayCore.InvertBitSet(bitarray.NewBitSet(uint(subtaskCount)), uint(subtaskCount)), // set all tasks to be cached
			RetryAttempts:  retryAttemptsArray,
		}

		// execute
		newState, _, err := LaunchAndCheckSubTasksState(ctx, tCtx, &kubeClient, &config, tCtx.DataStore(), "/prefix/", "/prefix-sand/", currentState)

		// validate results
		assert.Nil(t, err)
		p, _ := newState.GetPhase()
		assert.Equal(t, arrayCore.PhaseAbortSubTasks.String(), p.String())
		resourceManager.AssertNumberOfCalls(t, "ReleaseResource", subtaskCount)
		for _, subtaskPhaseIndex := range newState.GetArrayStatus().Detailed.GetItems() {
			assert.Equal(t, core.PhasePermanentFailure, core.Phases[subtaskPhaseIndex])
		}
	})
}

func TestTerminateSubTasksOnAbort(t *testing.T) {
	ctx := context.Background()
	subtaskCount := 3
	config := Config{
		MaxArrayJobSize: int64(subtaskCount * 10),
		ResourceConfig: ResourceConfig{
			PrimaryLabel: "p",
			Limit:        subtaskCount,
		},
	}
	kubeClient := mocks.KubeClient{}
	kubeClient.EXPECT().GetClient().Return(mocks.NewFakeKubeClient())
	kubeClient.EXPECT().GetCache().Return(mocks.NewFakeKubeCache())

	compactArray := arrayCore.NewPhasesCompactArray(uint(subtaskCount)) // #nosec G115
	for i := 0; i < subtaskCount; i++ {
		compactArray.SetItem(i, 5)
	}

	currentState := &arrayCore.State{
		CurrentPhase:         arrayCore.PhaseCheckingSubTaskExecutions,
		ExecutionArraySize:   subtaskCount,
		OriginalArraySize:    int64(subtaskCount),
		OriginalMinSuccesses: int64(subtaskCount),
		ArrayStatus: arraystatus.ArrayStatus{
			Detailed: compactArray,
		},
		// #nosec G115
		IndexesToCache: arrayCore.InvertBitSet(bitarray.NewBitSet(uint(subtaskCount)), uint(subtaskCount)),
	}

	t.Run("SuccessfulTermination", func(t *testing.T) {
		eventRecorder := mocks.EventsRecorder{}
		eventRecorder.EXPECT().RecordRaw(mock.Anything, mock.Anything).Return(nil)
		tCtx := getMockTaskExecutionContext(ctx, 0)
		tCtx.EXPECT().EventsRecorder().Return(&eventRecorder)

		mockTerminateFunction := func(ctx context.Context, subTaskCtx SubTaskExecutionContext, cfg *Config, kubeClient core.KubeClient) error {
			return nil
		}

		err := TerminateSubTasksOnAbort(ctx, tCtx, &kubeClient, &config, mockTerminateFunction, currentState)

		assert.Nil(t, err)
		eventRecorder.AssertCalled(t, "RecordRaw", mock.Anything, mock.Anything)
	})

	t.Run("TerminationWithError", func(t *testing.T) {
		eventRecorder := mocks.EventsRecorder{}
		eventRecorder.EXPECT().RecordRaw(mock.Anything, mock.Anything).Return(nil)
		tCtx := getMockTaskExecutionContext(ctx, 0)
		tCtx.EXPECT().EventsRecorder().Return(&eventRecorder)

		mockTerminateFunction := func(ctx context.Context, subTaskCtx SubTaskExecutionContext, cfg *Config, kubeClient core.KubeClient) error {
			return fmt.Errorf("termination error")
		}

		err := TerminateSubTasksOnAbort(ctx, tCtx, &kubeClient, &config, mockTerminateFunction, currentState)

		assert.NotNil(t, err)
		eventRecorder.AssertNotCalled(t, "RecordRaw", mock.Anything, mock.Anything)
	})
}

func TestTerminateSubTasks(t *testing.T) {
	ctx := context.Background()
	subtaskCount := 3
	config := Config{
		MaxArrayJobSize: int64(subtaskCount * 10),
		ResourceConfig: ResourceConfig{
			PrimaryLabel: "p",
			Limit:        subtaskCount,
		},
	}
	kubeClient := mocks.KubeClient{}
	kubeClient.EXPECT().GetClient().Return(mocks.NewFakeKubeClient())
	kubeClient.EXPECT().GetCache().Return(mocks.NewFakeKubeCache())

	tests := []struct {
		name                string
		initialPhaseIndices []int
		expectedAbortCount  int
		terminateError      error
	}{
		{
			name:                "AllSubTasksRunning",
			initialPhaseIndices: []int{5, 5, 5},
			expectedAbortCount:  3,
			terminateError:      nil,
		},
		{
			name:                "MixedSubTaskStates",
			initialPhaseIndices: []int{8, 0, 5},
			expectedAbortCount:  1,
			terminateError:      nil,
		},
		{
			name:                "TerminateFunctionFails",
			initialPhaseIndices: []int{5, 5, 5},
			expectedAbortCount:  3,
			terminateError:      fmt.Errorf("error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// #nosec G115
			compactArray := arrayCore.NewPhasesCompactArray(uint(subtaskCount))
			for i, phaseIdx := range test.initialPhaseIndices {
				compactArray.SetItem(i, bitarray.Item(phaseIdx)) // #nosec G115
			}
			currentState := &arrayCore.State{
				CurrentPhase:         arrayCore.PhaseCheckingSubTaskExecutions,
				PhaseVersion:         0,
				ExecutionArraySize:   subtaskCount,
				OriginalArraySize:    int64(subtaskCount),
				OriginalMinSuccesses: int64(subtaskCount),
				ArrayStatus: arraystatus.ArrayStatus{
					Detailed: compactArray,
				},
				// #nosec G115
				IndexesToCache: arrayCore.InvertBitSet(bitarray.NewBitSet(uint(subtaskCount)), uint(subtaskCount)),
			}

			tCtx := getMockTaskExecutionContext(ctx, 0)
			terminateCounter := 0
			mockTerminateFunction := func(ctx context.Context, subTaskCtx SubTaskExecutionContext, cfg *Config, kubeClient core.KubeClient) error {
				terminateCounter++
				return test.terminateError
			}

			nextState, externalResources, err := TerminateSubTasks(ctx, tCtx, &kubeClient, &config, mockTerminateFunction, currentState)

			assert.Equal(t, test.expectedAbortCount, terminateCounter)

			if test.terminateError != nil {
				assert.NotNil(t, err)
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, uint32(1), nextState.PhaseVersion)
			assert.Equal(t, arrayCore.PhaseWriteToDiscoveryThenFail, nextState.CurrentPhase)
			assert.Len(t, externalResources, terminateCounter)

			for _, externalResource := range externalResources {
				phase := core.Phases[externalResource.Phase]
				assert.True(t, phase.IsAborted())
			}
		})
	}
}
