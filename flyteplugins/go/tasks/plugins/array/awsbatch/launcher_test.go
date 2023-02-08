package awsbatch

import (
	"testing"

	"github.com/flyteorg/flytestdlib/bitarray"
	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/stretchr/testify/mock"

	"k8s.io/apimachinery/pkg/api/resource"

	core3 "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/arraystatus"
	arrayCore "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/core"

	"github.com/go-test/deep"

	mocks3 "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io/mocks"

	v1 "k8s.io/api/core/v1"

	core2 "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/core"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/awsbatch/config"
	mocks2 "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/awsbatch/mocks"
	"github.com/flyteorg/flytestdlib/utils"
	"golang.org/x/net/context"
)

func TestLaunchSubTasks(t *testing.T) {
	ctx := context.Background()
	tr := &mocks.TaskReader{}
	tr.OnRead(ctx).Return(&core.TaskTemplate{
		Target: &core.TaskTemplate_Container{
			Container: createSampleContainerTask(),
		},
	}, nil)

	tID := &mocks.TaskExecutionID{}
	tID.OnGetGeneratedName().Return("notfound")
	tID.OnGetID().Return(core.TaskExecutionIdentifier{
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
	overrides.OnGetConfig().Return(&v1.ConfigMap{Data: map[string]string{
		DynamicTaskQueueKey: "queue1",
	}})
	overrides.OnGetResources().Return(&v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU: resource.MustParse("10"),
		},
	})

	tMeta := &mocks.TaskExecutionMetadata{}
	tMeta.OnGetTaskExecutionID().Return(tID)
	tMeta.OnGetOverrides().Return(overrides)
	tMeta.OnGetPlatformResources().Return(&v1.ResourceRequirements{})

	ow := &mocks3.OutputWriter{}
	ow.OnGetOutputPrefixPath().Return("/prefix/")
	ow.OnGetRawOutputPrefix().Return("s3://")
	ow.OnGetCheckpointPrefix().Return("/checkpoint")
	ow.OnGetPreviousCheckpointsPrefix().Return("/prev")

	ir := &mocks3.InputReader{}
	ir.OnGetInputPrefixPath().Return("/prefix/")
	ir.OnGetInputPath().Return("/prefix/inputs.pb")
	ir.OnGetMatch(mock.Anything).Return(nil, nil)

	tCtx := &mocks.TaskExecutionContext{}
	tCtx.OnTaskReader().Return(tr)
	tCtx.OnTaskExecutionMetadata().Return(tMeta)
	tCtx.OnOutputWriter().Return(ow)
	tCtx.OnInputReader().Return(ir)

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

		newState, err := LaunchSubTasks(context.TODO(), tCtx, batchClient, &config.Config{MaxArrayJobSize: 10}, currentState, getAwsBatchExecutorMetrics(promutils.NewTestScope()), 0)
		assert.NoError(t, err)
		assertEqual(t, expectedState, newState)
	})
}

func TestTerminateSubTasks(t *testing.T) {
	ctx := context.Background()
	pStateReader := &mocks.PluginStateReader{}
	pStateReader.OnGetMatch(mock.Anything).Return(0, nil).Run(func(args mock.Arguments) {
		s := args.Get(0).(*State)
		s.ExternalJobID = refStr("abc-123")
	})

	tCtx := &mocks.TaskExecutionContext{}
	tCtx.OnPluginStateReader().Return(pStateReader)

	batchClient := &mocks2.Client{}
	batchClient.OnTerminateJob(ctx, "abc-123", "Test terminate").Return(nil).Once()

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
