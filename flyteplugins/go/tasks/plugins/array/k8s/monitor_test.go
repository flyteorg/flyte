package k8s

import (
	"testing"

	core2 "github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"
	mocks2 "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/lyft/flyteplugins/go/tasks/plugins/array/arraystatus"
	"github.com/lyft/flytestdlib/bitarray"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"

	arrayCore "github.com/lyft/flyteplugins/go/tasks/plugins/array/core"

	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core/mocks"
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

func TestCheckSubTasksState(t *testing.T) {
	ctx := context.Background()

	tCtx := getMockTaskExecutionContext(ctx)
	kubeClient := mocks.KubeClient{}
	kubeClient.OnGetClient().Return(mocks.NewFakeKubeClient())
	resourceManager := mocks.ResourceManager{}
	resourceManager.OnAllocateResourceMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(core.AllocationStatusExhausted, nil)
	tCtx.OnResourceManager().Return(&resourceManager)

	t.Run("Happy case", func(t *testing.T) {
		config := Config{MaxArrayJobSize: 100}
		newState, _, err := LaunchAndCheckSubTasksState(ctx, tCtx, &kubeClient, &config, nil, "/prefix/", "/prefix-sand/", &arrayCore.State{
			CurrentPhase:         arrayCore.PhaseCheckingSubTaskExecutions,
			ExecutionArraySize:   5,
			OriginalArraySize:    10,
			OriginalMinSuccesses: 5,
		})

		assert.Nil(t, err)
		//assert.NotEmpty(t, logLinks)
		p, _ := newState.GetPhase()
		assert.Equal(t, arrayCore.PhaseCheckingSubTaskExecutions.String(), p.String())
		resourceManager.AssertNumberOfCalls(t, "AllocateResource", 0)
	})

	t.Run("Resource exhausted", func(t *testing.T) {
		config := Config{
			MaxArrayJobSize: 100,
			ResourceConfig: ResourceConfig{
				PrimaryLabel: "p",
				Limit:        10,
			},
		}

		newState, _, err := LaunchAndCheckSubTasksState(ctx, tCtx, &kubeClient, &config, nil, "/prefix/", "/prefix-sand/", &arrayCore.State{
			CurrentPhase:         arrayCore.PhaseCheckingSubTaskExecutions,
			ExecutionArraySize:   5,
			OriginalArraySize:    10,
			OriginalMinSuccesses: 5,
		})

		assert.Nil(t, err)
		p, _ := newState.GetPhase()
		assert.Equal(t, arrayCore.PhaseWaitingForResources.String(), p.String())
		resourceManager.AssertNumberOfCalls(t, "AllocateResource", 5)
	})
}

func TestCheckSubTasksStateResourceGranted(t *testing.T) {
	ctx := context.Background()

	tCtx := getMockTaskExecutionContext(ctx)
	kubeClient := mocks.KubeClient{}
	kubeClient.OnGetClient().Return(mocks.NewFakeKubeClient())
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

		newState, _, err := LaunchAndCheckSubTasksState(ctx, tCtx, &kubeClient, &config, nil, "/prefix/", "/prefix-sand/", &arrayCore.State{
			CurrentPhase:         arrayCore.PhaseCheckingSubTaskExecutions,
			ExecutionArraySize:   5,
			OriginalArraySize:    10,
			OriginalMinSuccesses: 5,
		})

		assert.Nil(t, err)
		p, _ := newState.GetPhase()
		assert.Equal(t, arrayCore.PhaseCheckingSubTaskExecutions.String(), p.String())
		resourceManager.AssertNumberOfCalls(t, "AllocateResource", 5)
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
		newState, _, err := LaunchAndCheckSubTasksState(ctx, tCtx, &kubeClient, &config, nil, "/prefix/", "/prefix-sand/", &arrayCore.State{
			CurrentPhase:         arrayCore.PhaseCheckingSubTaskExecutions,
			ExecutionArraySize:   5,
			OriginalArraySize:    10,
			OriginalMinSuccesses: 5,
			ArrayStatus:          *arrayStatus,
		})

		assert.Nil(t, err)
		p, _ := newState.GetPhase()
		assert.Equal(t, arrayCore.PhaseWriteToDiscovery.String(), p.String())
		resourceManager.AssertNumberOfCalls(t, "ReleaseResource", 5)
	})
}
