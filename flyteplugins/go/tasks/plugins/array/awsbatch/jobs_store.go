/*
 * Copyright (c) 2018 Lyft. All rights reserved.
 */

package awsbatch

import (
	"context"
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/awsbatch/config"

	"k8s.io/client-go/util/workqueue"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flytestdlib/logger"

	"k8s.io/apimachinery/pkg/types"

	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/flyteorg/flytestdlib/cache"

	"github.com/flyteorg/flytestdlib/promutils"
)

type JobName = string
type JobID = string
type JobPhaseType = core.Phase
type ArrayJobSummary map[JobPhaseType]int64

type Event struct {
	OldJob *Job
	NewJob *Job
}

type EventHandler struct {
	Updated func(ctx context.Context, event Event)
}

type Job struct {
	ID             JobID                `json:"id,omitempty"`
	OwnerReference types.NamespacedName `json:"owner.omitempty"`
	Attempts       []Attempt            `json:"attempts,omitempty"`
	Status         JobStatus            `json:"status,omitempty"`
	SubJobs        []*Job               `json:"array,omitempty"`
}

type Attempt struct {
	LogStream string    `json:"logStream,omitempty"`
	StartedAt time.Time `json:"startedAt,omitempty"`
	StoppedAt time.Time `json:"stoppedAt,omitempty"`
}

type JobStatus struct {
	Phase   JobPhaseType `json:"phase,omitempty"`
	Message string       `json:"msg,omitempty"`
}

func (j Job) String() string {
	return fmt.Sprintf("(ID: %v)", j.ID)
}

func GetJobID(id JobID, index int) JobID {
	return fmt.Sprintf(arrayJobIDFormatter, id, index)
}

func batchJobsForSync(_ context.Context, batchChunkSize int) cache.CreateBatchesFunc {
	return func(ctx context.Context, items []cache.ItemWrapper) (batches []cache.Batch, err error) {
		batches = make([]cache.Batch, 0, 100)
		currentBatch := make(cache.Batch, 0, batchChunkSize)
		currentBatchSize := 0
		for _, item := range items {
			j := item.GetItem().(*Job)
			if j.Status.Phase.IsTerminal() {
				// If the job has already been terminated, do not include it in any batch.
				continue
			}

			if currentBatchSize > 0 && currentBatchSize+len(j.SubJobs)+1 > batchChunkSize {
				batches = append(batches, currentBatch)
				currentBatchSize = 0
				currentBatch = make(cache.Batch, 0, batchChunkSize)
			}

			currentBatchSize += len(j.SubJobs) + 1
			currentBatch = append(currentBatch, item)
		}

		if len(currentBatch) != 0 {
			batches = append(batches, currentBatch)
		}

		logger.Debugf(ctx, "Created batches from [%v] item(s). Batches [%v]", len(items), len(batches))

		return batches, nil
	}
}

func updateJob(ctx context.Context, source *batch.JobDetail, target *Job) (updated bool) {
	msg := make([]string, 0, 2)
	if source.Status == nil {
		logger.Warnf(ctx, "No status received for job [%v]", *source.JobId)
		msg = append(msg, "JobID in AWS BATCH has no Status")
	} else {
		newPhase := jobPhaseToPluginsPhase(*source.Status)
		if target.Status.Phase != newPhase {
			updated = true
		}

		target.Status.Phase = newPhase
	}

	if source.StatusReason != nil {
		msg = append(msg, *source.StatusReason)
	}

	logger.Debugf(ctx, "Job [%v] has (%v) attempts.", *source.JobId, len(source.Attempts))
	uniqueLogStreams := sets.String{}
	target.Attempts = make([]Attempt, 0, len(source.Attempts))
	lastStatusReason := ""
	for _, attempt := range source.Attempts {
		var a Attempt
		a, lastStatusReason = convertBatchAttemptToAttempt(attempt)
		if len(a.LogStream) > 0 {
			uniqueLogStreams.Insert(a.LogStream)
		}

		target.Attempts = append(target.Attempts, a)
	}

	// Add the "current" log stream to log links if one exists.
	attempt, exitReason := createAttemptFromJobDetail(ctx, source)
	if !uniqueLogStreams.Has(attempt.LogStream) {
		target.Attempts = append(target.Attempts, attempt)
	}

	if len(lastStatusReason) == 0 {
		lastStatusReason = exitReason
	}

	msg = append(msg, lastStatusReason)

	target.Status.Message = strings.Join(msg, " - ")
	return updated
}

func convertBatchAttemptToAttempt(attempt *batch.AttemptDetail) (a Attempt, exitReason string) {
	if attempt.StartedAt != nil {
		a.StartedAt = time.Unix(*attempt.StartedAt, 0)
	}

	if attempt.StoppedAt != nil {
		a.StoppedAt = time.Unix(*attempt.StoppedAt, 0)
	}

	if container := attempt.Container; container != nil {
		if container.LogStreamName != nil {
			a.LogStream = *container.LogStreamName
		}

		if container.Reason != nil {
			exitReason = *container.Reason
		}

		if container.ExitCode != nil {
			exitReason += fmt.Sprintf(" exit(%v)", *container.ExitCode)
		}
	}

	return a, exitReason
}

func createAttemptFromJobDetail(ctx context.Context, source *batch.JobDetail) (a Attempt, exitReason string) {
	if source.StartedAt != nil {
		a.StartedAt = time.Unix(*source.StartedAt, 0)
	}

	if source.StoppedAt != nil {
		a.StoppedAt = time.Unix(*source.StoppedAt, 0)
	}

	if container := source.Container; container != nil {
		if container.LogStreamName != nil {
			logger.Debug(ctx, "Using log stream from container info.")
			a.LogStream = *container.LogStreamName
		}

		if container.Reason != nil {
			exitReason = *container.Reason
		}

		if container.ExitCode != nil {
			exitReason += fmt.Sprintf(" exit(%v)", *container.ExitCode)
		}
	}

	return a, exitReason
}

func minInt(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func toRanges(totalSize, chunkSize int) (startIdx, endIdx []int) {
	startIdx = make([]int, 0, totalSize/chunkSize+1)
	endIdx = make([]int, 0, totalSize/chunkSize+1)
	for i := 0; i < totalSize; i += chunkSize {
		endI := minInt(chunkSize+i, totalSize)
		startIdx = append(startIdx, i)
		endIdx = append(endIdx, endI)
	}

	return
}

func syncBatches(_ context.Context, client Client, handler EventHandler, batchChunkSize int) cache.SyncFunc {
	return func(ctx context.Context, batch cache.Batch) ([]cache.ItemSyncResponse, error) {
		jobIDsMap := make(map[JobID]*Job, len(batch))
		jobIds := make([]JobID, 0, len(batch))
		jobNames := make(map[JobID]string, len(batch))

		// Build a flat list of JobIds to query batch for their status. Also build a reverse lookup to find these jobs
		// and update them in the cache.
		for _, item := range batch {
			j := item.GetItem().(*Job)
			if len(j.SubJobs) == 0 {
				logger.Errorf(ctx, "A job with no subjobs found job id [%v].", j.ID)
				continue
			}

			if j.Status.Phase.IsTerminal() {
				continue
			}

			jobIds = append(jobIds, j.ID)
			jobIDsMap[j.ID] = j
			jobNames[j.ID] = item.GetID()

			for idx, subJob := range j.SubJobs {
				if !subJob.Status.Phase.IsTerminal() {
					fullJobID := GetJobID(j.ID, idx)
					jobIds = append(jobIds, fullJobID)
					jobIDsMap[fullJobID] = subJob
				}
			}
		}

		if len(jobIds) == 0 {
			logger.Debug(ctx, "All jobs in batch have terminated, skipping sync call.")
			return []cache.ItemSyncResponse{}, nil
		}

		logger.Debugf(ctx, "Syncing jobs [%v].", len(jobIds))

		res := make([]cache.ItemSyncResponse, 0, len(jobIds))
		startIdx, endIdx := toRanges(len(jobIds), batchChunkSize)
		for i := 0; i < len(startIdx); i++ {
			logger.Debugf(ctx, "Syncing chunk [%v, %v) out of [%v] job ids.", startIdx[i], endIdx[i])
			response, err := client.GetJobDetailsBatch(ctx, jobIds[startIdx[i]:endIdx[i]])
			if err != nil {
				logger.Errorf(ctx, "Failed to get job details from AWS. Error: %v", err)
				return nil, err
			}

			for _, jobDetail := range response {
				job, found := jobIDsMap[*jobDetail.JobId]
				if !found {
					logger.Warn(ctx, "Received an update for unrequested job id [%v]", jobDetail.JobId)
					continue
				}

				changed := updateJob(ctx, jobDetail, job)

				if changed {
					handler.Updated(ctx, Event{
						NewJob: job,
					})
				}

				action := cache.Unchanged
				if changed {
					action = cache.Update
				}

				// If it's a single job, AWS Batch doesn't support arrays of size 1 so this workaround will ensure the rest
				// of the code doesn't have to deal with this limitation.
				if len(job.SubJobs) == 1 {
					subJob := job.SubJobs[0]
					subJob.Status = job.Status
					subJob.Attempts = job.Attempts
				}

				if jobName, found := jobNames[job.ID]; found {
					res = append(res, cache.ItemSyncResponse{
						ID:     jobName,
						Item:   job,
						Action: action,
					})
				}
			}
		}

		return res, nil
	}
}

type JobStore struct {
	Client
	cache.AutoRefresh

	started bool
}

// Submits a new job to AWS Batch and retrieves job info. Note that submitted jobs will not have status populated.
func (s JobStore) SubmitJob(ctx context.Context, input *batch.SubmitJobInput) (jobID string, err error) {
	name := *input.JobName
	if item, err := s.AutoRefresh.Get(name); err == nil {
		logger.Infof(ctx, "Job already found in cache with the same name [%v]. Will not submit a new job.",
			name)
		return item.(*Job).ID, nil
	}

	return s.Client.SubmitJob(ctx, input)
}

func (s *JobStore) Start(ctx context.Context) error {
	err := s.AutoRefresh.Start(ctx)
	if err != nil {
		return err
	}

	s.started = true
	return nil
}

func (s JobStore) GetOrCreate(jobName string, job *Job) (*Job, error) {
	j, err := s.AutoRefresh.GetOrCreate(jobName, job)
	if err != nil {
		return nil, err
	}

	return j.(*Job), err
}

func (s JobStore) Get(jobName string) *Job {
	j, err := s.AutoRefresh.Get(jobName)
	if err != nil {
		return nil
	}

	return j.(*Job)
}

func (s JobStore) IsStarted() bool {
	return s.started
}

// Constructs a new in-memory store.
func NewJobStore(ctx context.Context, batchClient Client, cfg config.JobStoreConfig,
	handler EventHandler, scope promutils.Scope) (JobStore, error) {

	store := JobStore{
		Client: batchClient,
	}

	autoCache, err := cache.NewAutoRefreshBatchedCache("aws-batch-jobs", batchJobsForSync(ctx, cfg.BatchChunkSize),
		syncBatches(ctx, store, handler, cfg.BatchChunkSize), workqueue.DefaultControllerRateLimiter(), cfg.ResyncPeriod.Duration,
		cfg.Parallelizm, cfg.CacheSize, scope)

	store.AutoRefresh = autoCache
	return store, err
}
