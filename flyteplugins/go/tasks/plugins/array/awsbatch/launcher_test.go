package awsbatch

import (
	"testing"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	core3 "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	mocks3 "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/array/arraystatus"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/array/awsbatch/config"
	mocks2 "github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/array/awsbatch/mocks"
	arrayCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/array/core"
	core2 "github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/array/core"
	"github.com/flyteorg/flyte/flytestdlib/bitarray"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/utils"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

func TestLaunchSubTasks(t *testing.T) {
	ctx := context.Background()
	tr := &mocks.TaskReader{}
	tr.EXPECT().Read(ctx).Return(&core.TaskTemplate{
		Target: &core.TaskTemplate_Container{
			Container: createSampleContainerTask(),
		},
	}, nil)

	tID := &mocks.TaskExecutionID{}
	tID.EXPECT().GetGeneratedName().Return("notfound")
	tID.EXPECT().GetID().Return(core.TaskExecutionIdentifier{
		TaskId: &core.Identifier{
			ResourceType: core.ResourceType_TASK,
			Project:      "a",
			Domain:       "d",
			Name:         "n",
			Version:      "abc",
		},
		NodeExecutionId: &core.NodeExecutionIdentifier{
			NodeId: "node1",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "a",
				Domain:  "d",
				Name:    "exec",
			},
		},
		RetryAttempt: 0,
	})

	overrides := &mocks.TaskOverrides{}
	overrides.EXPECT().GetConfig().Return(&v1.ConfigMap{Data: map[string]string{
		DynamicTaskQueueKey: "queue1",
	}})
	overrides.EXPECT().GetResources().Return(&v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU: resource.MustParse("10"),
		},
	})

	tMeta := &mocks.TaskExecutionMetadata{}
	tMeta.EXPECT().GetTaskExecutionID().Return(tID)
	tMeta.EXPECT().GetOverrides().Return(overrides)
	tMeta.EXPECT().GetPlatformResources().Return(&v1.ResourceRequirements{})
	tMeta.EXPECT().GetNamespace().Return("my-namespace")

	ow := &mocks3.OutputWriter{}
	ow.EXPECT().GetOutputPrefixPath().Return("/prefix/")
	ow.EXPECT().GetRawOutputPrefix().Return("s3://")
	ow.EXPECT().GetCheckpointPrefix().Return("/checkpoint")
	ow.EXPECT().GetPreviousCheckpointsPrefix().Return("/prev")

	ir := &mocks3.InputReader{}
	ir.EXPECT().GetInputPrefixPath().Return("/prefix/")
	ir.EXPECT().GetInputPath().Return("/prefix/inputs.pb")
	ir.EXPECT().Get(mock.Anything).Return(nil, nil)

	tCtx := &mocks.TaskExecutionContext{}
	tCtx.EXPECT().TaskReader().Return(tr)
	tCtx.EXPECT().TaskExecutionMetadata().Return(tMeta)
	tCtx.EXPECT().OutputWriter().Return(ow)
	tCtx.EXPECT().InputReader().Return(ir)

	batchClient := NewCustomBatchClient(mocks2.NewMockAwsBatchClient(), "", "",
		utils.NewRateLimiter("", 10, 20),
		utils.NewRateLimiter("", 10, 20))

	t.Run("Simple", func(t *testing.T) {
		currentState := &State{
			State: &core2.State{
				CurrentPhase:         core2.PhaseLaunch,
				ExecutionArraySize:   5,
				OriginalArraySize:    10,
				OriginalMinSuccesses: 5,
			},
			ExternalJobID:    nil,
			JobDefinitionArn: "arn",
		}

		retryAttemptsArray, err := bitarray.NewCompactArray(5, bitarray.Item(0))
		assert.NoError(t, err)

		expectedState := &State{
			State: &core2.State{
				CurrentPhase:         core2.PhaseCheckingSubTaskExecutions,
				Reason:               "Successfully launched subtasks.",
				ExecutionArraySize:   5,
				OriginalArraySize:    10,
				OriginalMinSuccesses: 5,
				ArrayStatus: arraystatus.ArrayStatus{
					Summary: map[core3.Phase]int64{
						core3.PhaseQueued: 5,
					},
					Detailed: arrayCore.NewPhasesCompactArray(5),
				},
				RetryAttempts: retryAttemptsArray,
			},

			ExternalJobID:    refStr("qpxyarq"),
			JobDefinitionArn: "arn",
		}

		newState, err := LaunchSubTasks(context.Background(), tCtx, batchClient, &config.Config{MaxArrayJobSize: 10}, currentState, getAwsBatchExecutorMetrics(promutils.NewTestScope()), 0)
		assert.NoError(t, err)
		assertEqual(t, expectedState, newState)
	})
}

func TestTerminateSubTasks(t *testing.T) {
	ctx := context.Background()
	pStateReader := &mocks.PluginStateReader{}
	pStateReader.EXPECT().Get(mock.Anything).Return(0, nil).Run(func(t interface{}) {
		s := t.(*State)
		s.ExternalJobID = refStr("abc-123")
	})

	tCtx := &mocks.TaskExecutionContext{}
	tCtx.EXPECT().PluginStateReader().Return(pStateReader)

	batchClient := &mocks2.Client{}
	batchClient.EXPECT().TerminateJob(ctx, "abc-123", "Test terminate").Return(nil).Once()

	t.Run("Simple", func(t *testing.T) {
		assert.NoError(t, TerminateSubTasks(ctx, tCtx, batchClient, "Test terminate", getAwsBatchExecutorMetrics(promutils.NewTestScope())))
	})

	batchClient.AssertExpectations(t)
	tCtx.AssertExpectations(t)
	pStateReader.AssertExpectations(t)
}

func assertEqual(t testing.TB, a, b interface{}) {
	if diff := deep.Equal(a, b); diff != nil {
		t.Error(diff)
	}
}
