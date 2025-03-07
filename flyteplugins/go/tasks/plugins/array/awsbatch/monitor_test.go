package awsbatch

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/types"

	flyteIdl "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	ioMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/array/arraystatus"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/array/awsbatch/config"
	batchMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/array/awsbatch/mocks"
	arrayCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/array/core"
	"github.com/flyteorg/flyte/flytestdlib/bitarray"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/flytestdlib/storage"
	"github.com/flyteorg/flyte/flytestdlib/utils"
)

func init() {
	labeled.SetMetricKeys(contextutils.RoutineLabelKey)
}

func TestCheckSubTasksState(t *testing.T) {
	ctx := context.Background()
	tCtx := &mocks.TaskExecutionContext{}
	tID := &mocks.TaskExecutionID{}
	tID.EXPECT().GetGeneratedName().Return("generated-name")
	tMeta := &mocks.TaskExecutionMetadata{}
	tMeta.EXPECT().GetOwnerID().Return(types.NamespacedName{
		Namespace: "domain",
		Name:      "name",
	})
	tMeta.EXPECT().GetTaskExecutionID().Return(tID)
	inMemDatastore, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)

	outputWriter := &ioMocks.OutputWriter{}
	outputWriter.EXPECT().GetOutputPrefixPath().Return("")
	outputWriter.EXPECT().GetRawOutputPrefix().Return("")

	taskReader := &mocks.TaskReader{}
	task := &flyteIdl.TaskTemplate{
		Type: "test",
		Target: &flyteIdl.TaskTemplate_Container{
			Container: &flyteIdl.Container{
				Command: []string{"command"},
				Args:    []string{"{{.Input}}"},
			},
		},
		Metadata: &flyteIdl.TaskMetadata{Retries: &flyteIdl.RetryStrategy{Retries: 3}},
	}
	taskReader.On("Read", mock.Anything).Return(task, nil)

	tCtx.EXPECT().OutputWriter().Return(outputWriter)
	tCtx.EXPECT().TaskReader().Return(taskReader)
	tCtx.EXPECT().DataStore().Return(inMemDatastore)
	tCtx.EXPECT().TaskExecutionMetadata().Return(tMeta)

	t.Run("Not in cache", func(t *testing.T) {
		mBatchClient := batchMocks.NewMockAwsBatchClient()
		batchClient := NewCustomBatchClient(mBatchClient, "", "",
			utils.NewRateLimiter("", 10, 20),
			utils.NewRateLimiter("", 10, 20))

		jobStore := newJobsStore(t, batchClient)
		newState, err := CheckSubTasksState(ctx, tCtx, jobStore, &config.Config{}, &State{
			State: &arrayCore.State{
				CurrentPhase:         arrayCore.PhaseCheckingSubTaskExecutions,
				ExecutionArraySize:   5,
				OriginalArraySize:    10,
				OriginalMinSuccesses: 5,
			},
			ExternalJobID:    refStr("job-id"),
			JobDefinitionArn: "",
		}, getAwsBatchExecutorMetrics(promutils.NewTestScope()))

		assert.NoError(t, err)
		p, _ := newState.GetPhase()
		assert.Equal(t, arrayCore.PhaseCheckingSubTaskExecutions.String(), p.String())
	})

	t.Run("Succeeded", func(t *testing.T) {
		mBatchClient := batchMocks.NewMockAwsBatchClient()
		mBatchClient.DescribeJobsWithContextCb =
			func(ctx context.Context, input *batch.DescribeJobsInput, opts ...request.Option) (
				output *batch.DescribeJobsOutput, e error) {
				return &batch.DescribeJobsOutput{
					Jobs: []*batch.JobDetail{
						{
							JobId:   refStr("job-id"),
							JobName: refStr(tID.GetGeneratedName()),
							Status:  refStr(batch.JobStatusSucceeded),
						},
					},
				}, nil
			}

		batchClient := NewCustomBatchClient(mBatchClient, "", "",
			utils.NewRateLimiter("", 10, 20),
			utils.NewRateLimiter("", 10, 20))

		jobStore := newJobsStore(t, batchClient)
		_, err := jobStore.GetOrCreate(tID.GetGeneratedName(), &Job{
			ID: "job-id",
			Status: JobStatus{
				Phase: core.PhaseSuccess,
			},
		})

		assert.NoError(t, err)

		newState, err := CheckSubTasksState(ctx, tCtx, jobStore, &config.Config{}, &State{
			State: &arrayCore.State{
				CurrentPhase:         arrayCore.PhaseCheckingSubTaskExecutions,
				ExecutionArraySize:   5,
				OriginalArraySize:    10,
				OriginalMinSuccesses: 5,
			},
			ExternalJobID:    refStr("job-id"),
			JobDefinitionArn: "",
		}, getAwsBatchExecutorMetrics(promutils.NewTestScope()))

		assert.NoError(t, err)
		p, _ := newState.GetPhase()
		assert.Equal(t, arrayCore.PhaseWriteToDiscovery.String(), p.String())
	})

	t.Run("queued", func(t *testing.T) {
		mBatchClient := batchMocks.NewMockAwsBatchClient()
		batchClient := NewCustomBatchClient(mBatchClient, "", "",
			utils.NewRateLimiter("", 10, 20),
			utils.NewRateLimiter("", 10, 20))

		jobStore := newJobsStore(t, batchClient)
		_, err := jobStore.GetOrCreate(tID.GetGeneratedName(), &Job{
			ID: "job-id",
			Status: JobStatus{
				Phase: core.PhaseRunning,
			},
			SubJobs: []*Job{
				{Status: JobStatus{Phase: core.PhaseQueued}},
			},
		})

		assert.NoError(t, err)

		retryAttemptsArray, err := bitarray.NewCompactArray(1, bitarray.Item(1))
		assert.NoError(t, err)

		newState, err := CheckSubTasksState(ctx, tCtx, jobStore, &config.Config{}, &State{
			State: &arrayCore.State{
				CurrentPhase:         arrayCore.PhaseCheckingSubTaskExecutions,
				ExecutionArraySize:   1,
				OriginalArraySize:    1,
				OriginalMinSuccesses: 1,
				ArrayStatus: arraystatus.ArrayStatus{
					Detailed: arrayCore.NewPhasesCompactArray(1),
				},
				IndexesToCache: bitarray.NewBitSet(1),
				RetryAttempts:  retryAttemptsArray,
			},
			ExternalJobID:    refStr("job-id"),
			JobDefinitionArn: "",
		}, getAwsBatchExecutorMetrics(promutils.NewTestScope()))

		assert.NoError(t, err)
		p, _ := newState.GetPhase()
		assert.Equal(t, arrayCore.PhaseCheckingSubTaskExecutions.String(), p.String())

	})

	t.Run("Still running", func(t *testing.T) {
		mBatchClient := batchMocks.NewMockAwsBatchClient()
		batchClient := NewCustomBatchClient(mBatchClient, "", "",
			utils.NewRateLimiter("", 10, 20),
			utils.NewRateLimiter("", 10, 20))

		jobStore := newJobsStore(t, batchClient)
		_, err := jobStore.GetOrCreate(tID.GetGeneratedName(), &Job{
			ID: "job-id",
			Status: JobStatus{
				Phase: core.PhaseRunning,
			},
			SubJobs: []*Job{
				{Status: JobStatus{Phase: core.PhaseRunning}},
				{Status: JobStatus{Phase: core.PhaseSuccess}},
			},
		})

		assert.NoError(t, err)

		retryAttemptsArray, err := bitarray.NewCompactArray(2, bitarray.Item(1))
		assert.NoError(t, err)

		newState, err := CheckSubTasksState(ctx, tCtx, jobStore, &config.Config{}, &State{
			State: &arrayCore.State{
				CurrentPhase:         arrayCore.PhaseCheckingSubTaskExecutions,
				ExecutionArraySize:   2,
				OriginalArraySize:    2,
				OriginalMinSuccesses: 2,
				ArrayStatus: arraystatus.ArrayStatus{
					Detailed: arrayCore.NewPhasesCompactArray(2),
				},
				IndexesToCache: bitarray.NewBitSet(2),
				RetryAttempts:  retryAttemptsArray,
			},
			ExternalJobID:    refStr("job-id"),
			JobDefinitionArn: "",
		}, getAwsBatchExecutorMetrics(promutils.NewTestScope()))

		assert.NoError(t, err)
		p, _ := newState.GetPhase()
		assert.Equal(t, arrayCore.PhaseCheckingSubTaskExecutions.String(), p.String())
	})

	t.Run("retry limit exceeded", func(t *testing.T) {
		mBatchClient := batchMocks.NewMockAwsBatchClient()
		batchClient := NewCustomBatchClient(mBatchClient, "", "",
			utils.NewRateLimiter("", 10, 20),
			utils.NewRateLimiter("", 10, 20))

		jobStore := newJobsStore(t, batchClient)
		_, err := jobStore.GetOrCreate(tID.GetGeneratedName(), &Job{
			ID: "job-id",
			Status: JobStatus{
				Phase: core.PhaseRunning,
			},
			SubJobs: []*Job{
				{Status: JobStatus{Phase: core.PhaseRetryableFailure}, Attempts: []Attempt{{LogStream: "failed"}}},
				{Status: JobStatus{Phase: core.PhaseSuccess}},
			},
		})

		assert.NoError(t, err)

		retryAttemptsArray, err := bitarray.NewCompactArray(2, bitarray.Item(1))
		assert.NoError(t, err)

		newState, err := CheckSubTasksState(ctx, tCtx, jobStore, &config.Config{}, &State{
			State: &arrayCore.State{
				CurrentPhase:         arrayCore.PhaseCheckingSubTaskExecutions,
				ExecutionArraySize:   2,
				OriginalArraySize:    2,
				OriginalMinSuccesses: 2,
				ArrayStatus: arraystatus.ArrayStatus{
					Detailed: arrayCore.NewPhasesCompactArray(2),
				},
				IndexesToCache: bitarray.NewBitSet(2),
				RetryAttempts:  retryAttemptsArray,
			},
			ExternalJobID:    refStr("job-id"),
			JobDefinitionArn: "",
		}, getAwsBatchExecutorMetrics(promutils.NewTestScope()))

		assert.NoError(t, err)
		p, _ := newState.GetPhase()
		assert.Equal(t, arrayCore.PhaseAbortSubTasks, p)
	})
}
