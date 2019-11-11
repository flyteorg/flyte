package awsbatch

import (
	"fmt"

	"github.com/lyft/flyteplugins/go/tasks/plugins/array/core"

	errors2 "github.com/lyft/flyteplugins/go/tasks/errors"
	"github.com/lyft/flytestdlib/errors"

	"github.com/lyft/flytestdlib/logger"

	idlCore "github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	pluginCore "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"
	"golang.org/x/net/context"
)

const (
	LogStreamFormatter = "https://console.aws.amazon.com/cloudwatch/home?region=%v#logEventViewer:group=/aws/batch/job;stream=%v"
	ArrayJobFormatter  = "https://console.aws.amazon.com/batch/home?region=%v#/jobs/%v"
	JobFormatter       = "https://console.aws.amazon.com/batch/home?region=%v#/jobs/queue/arn:aws:batch:%v:%v:job-queue~2F%v/job/%v"
)

func GetJobUri(jobSize int, accountID, region, queue, jobID string) string {
	if jobSize > 1 {
		return fmt.Sprintf(ArrayJobFormatter, region, jobID)
	}

	return fmt.Sprintf(JobFormatter, region, region, accountID, queue, jobID)
}

func GetJobTaskLog(jobSize int, accountID, region, queue, jobID string) *idlCore.TaskLog {
	return &idlCore.TaskLog{
		Name: fmt.Sprintf("AWS Batch Job"),
		Uri:  GetJobUri(jobSize, accountID, region, queue, jobID),
	}
}

func GetTaskLinks(ctx context.Context, taskMeta pluginCore.TaskExecutionMetadata, jobStore *JobStore, state *State) (
	[]*idlCore.TaskLog, error) {

	logLinks := make([]*idlCore.TaskLog, 0, 4)

	if state.GetExternalJobID() == nil {
		return logLinks, nil
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
		return nil, errors.Wrapf(errors2.DownstreamSystemError, err, "Failed to retrieve a job from job store.")
	}

	if job == nil {
		logger.Debugf(ctx, "Job [%v] not found in jobs store. It might have been evicted. If reasonable, bump the max "+
			"size of the LRU cache.", *state.GetExternalJobID())

		return logLinks, nil
	}

	for childIdx, subJob := range job.SubJobs {
		originalIndex := core.CalculateOriginalIndex(childIdx, state.GetIndexesToCache())

		for attemptIdx, attempt := range subJob.Attempts {
			if len(attempt.LogStream) > 0 {
				logLinks = append(logLinks, &idlCore.TaskLog{
					Name: fmt.Sprintf("AWS Batch #%v-%v (%v)", originalIndex, attemptIdx, subJob.Status.Phase),
					Uri:  fmt.Sprintf(LogStreamFormatter, jobStore.GetRegion(), attempt.LogStream),
				})
			}
		}
	}

	return logLinks, nil
}
