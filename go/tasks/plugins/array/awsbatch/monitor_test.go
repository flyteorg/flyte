package awsbatch

import (
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"

	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/storage"

	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/arraystatus"

	flyteIdl "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/batch"

	arrayCore "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/core"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	ioMocks "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/awsbatch/config"
	batchMocks "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/awsbatch/mocks"
	"github.com/flyteorg/flytestdlib/utils"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/types"

	"github.com/flyteorg/flytestdlib/bitarray"
)

func init() {
	labeled.SetMetricKeys(contextutils.RoutineLabelKey)
}

func TestCheckSubTasksState(t *testing.T) {
	ctx := context.Background()
	tCtx := &mocks.TaskExecutionContext{}
	tID := &mocks.TaskExecutionID{}
	tID.OnGetGeneratedName().Return("generated-name")
	tMeta := &mocks.TaskExecutionMetadata{}
	tMeta.OnGetOwnerID().Return(types.NamespacedName{
		Namespace: "domain",
		Name:      "name",
	})
	tMeta.OnGetTaskExecutionID().Return(tID)
	inMemDatastore, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)

	outputWriter := &ioMocks.OutputWriter{}
	outputWriter.OnGetOutputPrefixPath().Return("")
	outputWriter.OnGetRawOutputPrefix().Return("")

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

	tCtx.OnOutputWriter().Return(outputWriter)
	tCtx.OnTaskReader().Return(taskReader)
	tCtx.OnDataStore().Return(inMemDatastore)
	tCtx.OnTaskExecutionMetadata().Return(tMeta)

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
				CurrentPhase:         arrayCore.PhaseWriteToDiscoveryThenFail,
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
		assert.Equal(t, arrayCore.PhaseWriteToDiscoveryThenFail, p)
	})
}
