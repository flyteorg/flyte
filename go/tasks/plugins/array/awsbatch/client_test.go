/*
 * Copyright (c) 2018 Lyft. All rights reserved.
 */

package awsbatch

import (
	"context"

	stdConfig "github.com/flyteorg/flytestdlib/config"

	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/awsbatch/config"

	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/awsbatch/mocks"
	"github.com/flyteorg/flytestdlib/utils"

	"github.com/flyteorg/flytestdlib/promutils"

	"testing"

	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/stretchr/testify/assert"
)

func newClientWithMockBatch() *client {
	rateLimiter := utils.NewRateLimiter("Get", 1000, 1000)
	return NewCustomBatchClient(mocks.NewMockAwsBatchClient(), "account-id", "test-region", rateLimiter, rateLimiter).(*client)
}

func TestClient_SubmitJob(t *testing.T) {
	ctx := context.Background()
	rateLimiter := utils.NewRateLimiter("Get", 1000, 1000)
	c := NewCustomBatchClient(mocks.NewMockAwsBatchClient(), "account-id", "test-region", rateLimiter, rateLimiter).(*client)
	store, err := NewJobStore(ctx, c, config.JobStoreConfig{
		CacheSize:      1,
		Parallelizm:    1,
		BatchChunkSize: 1,
		ResyncPeriod:   stdConfig.Duration{Duration: 1000},
	}, EventHandler{}, promutils.NewTestScope())
	assert.NoError(t, err)

	o, err := c.SubmitJob(context.TODO(), &batch.SubmitJobInput{
		JobName: refStr("test-job"),
	})

	assert.NoError(t, err)
	assert.NotNil(t, o)

	_, err = store.GetOrCreate("test-job", &Job{
		ID: o,
	})
	assert.NoError(t, err)

	// Resubmit the same job
	o, err = c.SubmitJob(context.TODO(), &batch.SubmitJobInput{
		JobName: refStr("test-job"),
	})

	assert.NoError(t, err)
	assert.NotNil(t, o)
}

func TestClient_TerminateJob(t *testing.T) {
	c := newClientWithMockBatch()
	err := c.TerminateJob(context.TODO(), "1", "")
	assert.NoError(t, err)
}

func TestClient_GetJobDetailsBatch(t *testing.T) {
	c := newClientWithMockBatch()
	o, err := c.GetJobDetailsBatch(context.TODO(), []string{})
	assert.NoError(t, err)
	assert.NotNil(t, o)

	o, err = c.GetJobDetailsBatch(context.TODO(), []string{"fake_job_id"})
	assert.NoError(t, err)
	assert.NotNil(t, o)
	assert.Equal(t, 1, len(o))
	assert.Equal(t, "fake_job_id", *o[0].JobId)
}

func TestClient_RegisterJobDefinition(t *testing.T) {
	c := newClientWithMockBatch()
	j, err := c.RegisterJobDefinition(context.TODO(), "name-abc", "img", "admin-role", defaultComputeEngine)
	assert.NoError(t, err)
	assert.NotNil(t, j)
}
