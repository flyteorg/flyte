package controller

import (
	"context"
	"testing"
	"time"

	config2 "github.com/flyteorg/flytepropeller/pkg/controller/config"

	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

func TestNewCompositeWorkQueue(t *testing.T) {
	ctx := context.TODO()

	t.Run("simple", func(t *testing.T) {
		testScope := promutils.NewScope("test1")
		cfg := config2.CompositeQueueConfig{}
		q, err := NewCompositeWorkQueue(ctx, cfg, testScope)
		assert.NoError(t, err)
		assert.NotNil(t, q)
		switch q.(type) {
		case *SimpleWorkQueue:
			return
		default:
			assert.FailNow(t, "SimpleWorkQueue expected")
		}
	})

	t.Run("batch", func(t *testing.T) {
		testScope := promutils.NewScope("test2")
		cfg := config2.CompositeQueueConfig{
			Type:             config2.CompositeQueueBatch,
			BatchSize:        -1,
			BatchingInterval: config.Duration{Duration: time.Second * 1},
		}
		q, err := NewCompositeWorkQueue(ctx, cfg, testScope)
		assert.NoError(t, err)
		assert.NotNil(t, q)
		switch bq := q.(type) {
		case *BatchingWorkQueue:
			assert.Equal(t, -1, bq.batchSize)
			assert.Equal(t, time.Second*1, bq.batchingInterval)
			return
		default:
			assert.FailNow(t, "BatchWorkQueue expected")
		}
	})
}

func TestSimpleWorkQueue(t *testing.T) {
	ctx := context.TODO()
	testScope := promutils.NewScope("test")
	cfg := config2.CompositeQueueConfig{}
	q, err := NewCompositeWorkQueue(ctx, cfg, testScope)
	assert.NoError(t, err)
	assert.NotNil(t, q)

	t.Run("AddSubQueue", func(t *testing.T) {
		q.AddToSubQueue("x")
		i, s := q.Get()
		assert.False(t, s)
		assert.Equal(t, "x", i.(string))
		q.Done(i)
	})

	t.Run("AddAfterSubQueue", func(t *testing.T) {
		q.AddToSubQueueAfter("y", time.Nanosecond*0)
		i, s := q.Get()
		assert.False(t, s)
		assert.Equal(t, "y", i.(string))
		q.Done(i)
	})

	t.Run("AddRateLimitedSubQueue", func(t *testing.T) {
		q.AddToSubQueueRateLimited("z")
		i, s := q.Get()
		assert.False(t, s)
		assert.Equal(t, "z", i.(string))
		q.Done(i)
	})

	t.Run("shutdown", func(t *testing.T) {
		q.ShutdownAll()
		_, s := q.Get()
		assert.True(t, s)
	})
}

func TestBatchingQueue(t *testing.T) {
	ctx := context.TODO()
	testScope := promutils.NewScope("test_batch")
	cfg := config2.CompositeQueueConfig{
		Type:             config2.CompositeQueueBatch,
		BatchSize:        -1,
		BatchingInterval: config.Duration{Duration: time.Nanosecond * 1},
	}
	q, err := NewCompositeWorkQueue(ctx, cfg, testScope)
	assert.NoError(t, err)
	assert.NotNil(t, q)

	batchQueue := q.(*BatchingWorkQueue)

	t.Run("AddSubQueue", func(t *testing.T) {
		q.AddToSubQueue("x")
		assert.Equal(t, 0, q.Len())
		batchQueue.runSubQueueHandler(ctx)
		i, s := q.Get()
		assert.False(t, s)
		assert.Equal(t, "x", i.(string))
		q.Done(i)
	})

	t.Run("AddAfterSubQueue", func(t *testing.T) {
		q.AddToSubQueueAfter("y", time.Nanosecond*0)
		assert.Equal(t, 0, q.Len())
		batchQueue.runSubQueueHandler(ctx)
		i, s := q.Get()
		assert.False(t, s)
		assert.Equal(t, "y", i.(string))
		q.Done(i)
	})

	t.Run("AddRateLimitedSubQueue", func(t *testing.T) {
		q1, err := NewCompositeWorkQueue(ctx, cfg, promutils.NewScope("test_batch_inner"))
		assert.NoError(t, err)
		assert.NotNil(t, q1)

		batchQueue := q1.(*BatchingWorkQueue)
		q1.AddToSubQueueRateLimited("z")
		assert.Equal(t, 0, q1.Len())
		batchQueue.Start(ctx)
		i, s := q1.Get()
		assert.False(t, s)
		assert.Equal(t, "z", i.(string))
		q1.Done(i)
	})

	t.Run("shutdown", func(t *testing.T) {
		q.AddToSubQueue("g")
		q.ShutdownAll()
		assert.Equal(t, 0, q.Len())
		batchQueue.runSubQueueHandler(ctx)
		i, s := q.Get()
		assert.True(t, s)
		assert.Nil(t, i)
		q.Done(i)
	})
}
