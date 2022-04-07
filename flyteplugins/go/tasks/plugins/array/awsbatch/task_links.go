package awsbatch

import (
	"fmt"

	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/core"

	errors2 "github.com/flyteorg/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flytestdlib/errors"

	"github.com/flyteorg/flytestdlib/logger"

	idlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	pluginCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"golang.org/x/net/context"
)

const (
	LogStreamFormatter = "https://console.aws.amazon.com/cloudwatch/home?region=%v#logEventViewer:group=/aws/batch/job;stream=%v"
	ArrayJobFormatter  = "https://console.aws.amazon.com/batch/home?region=%v#/jobs/%v"
	JobFormatter       = "https://console.aws.amazon.com/batch/home?region=%v#/jobs/queue/arn:aws:batch:%v:%v:job-queue~2F%v/job/%v"
)

func GetJobURI(jobSize int, accountID, region, queue, jobID string) string {
	if jobSize > 1 {
		return fmt.Sprintf(ArrayJobFormatter, region, jobID)
	}

	return fmt.Sprintf(JobFormatter, region, region, accountID, queue, jobID)
}

func GetJobTaskLog(jobSize int, accountID, region, queue, jobID string) *idlCore.TaskLog {
	return &idlCore.TaskLog{
		Name: "AWS Batch Job",
		Uri:  GetJobURI(jobSize, accountID, region, queue, jobID),
	}
}

func GetTaskLinks(ctx context.Context, taskMeta pluginCore.TaskExecutionMetadata, jobStore *JobStore, state *State) (
	[]*idlCore.TaskLog, []*pluginCore.ExternalResource, error) {

	logLinks := make([]*idlCore.TaskLog, 0, 4)
	externalResources := make([]*pluginCore.ExternalResource, 0)

	if state.GetExternalJobID() == nil {
		return logLinks, externalResources, nil
	}

	// TODO: Add tasktemplate container config to job config
	jobConfig := newJobConfig().
		MergeFromConfigMap(taskMeta.GetOverrides().GetConfig())
	logLinks = append(logLinks, GetJobTaskLog(state.GetExecutionArraySize(), jobStore.Client.GetAccountID(),
		jobStore.Client.GetRegion(), jobConfig.DynamicTaskQueue, *state.GetExternalJobID()))

	jobName := taskMeta.GetTaskExecutionID().GetGeneratedName()
	job, err := jobStore.GetOrCreate(jobName, &Job{
		ID:      *state.GetExternalJobID(),
		SubJobs: createSubJobList(state.GetExecutionArraySize()),
	})

	if err != nil {
		return logLinks, externalResources, errors.Wrapf(errors2.DownstreamSystemError, err, "Failed to retrieve a job from job store.")
	}

	if job == nil {
		logger.Debugf(ctx, "Job [%v] not found in jobs store. It might have been evicted. If reasonable, bump the max "+
			"size of the LRU cache.", *state.GetExternalJobID())

		return logLinks, externalResources, nil
	}

	detailedArrayStatus := state.GetArrayStatus().Detailed
	for childIdx, subJob := range job.SubJobs {
		originalIndex := core.CalculateOriginalIndex(childIdx, state.GetIndexesToCache())
		finalPhaseIdx := detailedArrayStatus.GetItem(childIdx)
		finalPhase := pluginCore.Phases[finalPhaseIdx]

		// The caveat here is that we will mark all attempts with the final phase we are tracking in the state.
		subTaskLogLinks := make([]*idlCore.TaskLog, 0)
		retryAttempt := 0
		for attemptIdx, attempt := range subJob.Attempts {
			if len(attempt.LogStream) > 0 {
				subTaskLogLinks = append(subTaskLogLinks, &idlCore.TaskLog{
					Name: fmt.Sprintf("AWS Batch #%v-%v (%v)", originalIndex, attemptIdx, finalPhase),
					Uri:  fmt.Sprintf(LogStreamFormatter, jobStore.GetRegion(), attempt.LogStream),
				})
			}

			retryAttempt = attemptIdx
		}

		externalResources = append(externalResources, &pluginCore.ExternalResource{
			ExternalID:   subJob.ID,
			Index:        uint32(originalIndex),
			Logs:         subTaskLogLinks,
			RetryAttempt: uint32(retryAttempt),
			Phase:        finalPhase,
		})
	}

	return logLinks, externalResources, nil
}
