/*
 * Copyright (c) 2018 Lyft. All rights reserved.
 */

package awsbatch

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/service/batch"

	mocks2 "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/awsbatch/mocks"
	mocks3 "github.com/flyteorg/flytestdlib/cache/mocks"
	"github.com/flyteorg/flytestdlib/utils"

	config2 "github.com/flyteorg/flytestdlib/config"

	"github.com/flyteorg/flytestdlib/cache"

	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/awsbatch/config"

	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/stretchr/testify/assert"
)

func createJobWithID(id JobID) *Job {
	return &Job{
		ID: id,
	}
}

func newJobsStore(t testing.TB, batchClient Client) *JobStore {
	return newJobsStoreWithSize(t, batchClient, 1)
}

func newJobsStoreWithSize(t testing.TB, batchClient Client, size int) *JobStore {
	store, err := NewJobStore(context.TODO(), batchClient, config.JobStoreConfig{
		CacheSize:      size,
		Parallelizm:    1,
		BatchChunkSize: 2,
		ResyncPeriod:   config2.Duration{Duration: 1000},
	}, EventHandler{}, promutils.NewTestScope())
	assert.NoError(t, err)
	return &store
}
func TestGetJobsStore(t *testing.T) {
	s := newJobsStore(t, nil)
	assert.NotNil(t, s)
	assert.NotNil(t, s)
}

func TestJobStore_GetOrCreate(t *testing.T) {
	s := newJobsStore(t, nil)
	assert.NotNil(t, s)
	ok, err := s.GetOrCreate("RandomId", createJobWithID("RandomId"))
	assert.NoError(t, err)
	assert.NotNil(t, ok)
}

func TestStore_Get(t *testing.T) {
	s := newJobsStore(t, nil)
	assert.NotNil(t, s)
	_, err := s.GetOrCreate("Id1", createJobWithID("Id1"))
	assert.NoError(t, err)
	_, err = s.GetOrCreate("Id2", createJobWithID("Id2"))
	assert.NoError(t, err)

	j := s.Get("Id2")
	assert.NotNil(t, j)
	assert.Equal(t, "Id2", j.ID)

	j = s.Get("Id3")
	assert.Nil(t, j)
}

// Current values:
// BenchmarkStore_AddOrUpdate-8   	  500000	      2677 ns/op
func BenchmarkStore_GetOrUpdate(b *testing.B) {
	s := newJobsStore(b, nil)
	assert.NotNil(b, s)
	for i := 0; i < b.N; i++ {
		_, err := s.GetOrCreate("Id1", createJobWithID("Id1"))
		assert.NoError(b, err)
	}
}

type mockItem struct {
	id   cache.ItemID
	item cache.Item
}

func (m mockItem) GetID() cache.ItemID {
	return m.id
}

func (m mockItem) GetItem() cache.Item {
	return m.item
}

func TestBatchJobsForSync(t *testing.T) {
	t.Run("jobs < chunk size", func(t *testing.T) {
		f := batchJobsForSync(context.TODO(), 2)
		batches, err := f(context.TODO(), []cache.ItemWrapper{
			mockItem{
				id:   "id1",
				item: &Job{ID: "id1"},
			},
		})

		assert.NoError(t, err)
		assert.Len(t, batches, 1)
		assert.Len(t, batches[0], 1)
		assert.Equal(t, "id1", batches[0][0].GetID())
	})

	t.Run("jobs > chunk size", func(t *testing.T) {
		f := batchJobsForSync(context.TODO(), 2)
		batches, err := f(context.TODO(), []cache.ItemWrapper{
			mockItem{
				id:   "id1",
				item: &Job{ID: "id1"},
			},
			mockItem{
				id:   "id2",
				item: &Job{ID: "id2"},
			},
			mockItem{
				id:   "id3",
				item: &Job{ID: "id3"},
			},
		})

		assert.NoError(t, err)
		assert.Len(t, batches, 2)
		assert.Len(t, batches[0], 2)
		assert.Len(t, batches[1], 1)
		assert.Equal(t, "id1", batches[0][0].GetID())
	})

	t.Run("sub jobs > chunk size", func(t *testing.T) {
		f := batchJobsForSync(context.TODO(), 2)
		batches, err := f(context.TODO(), []cache.ItemWrapper{
			mockItem{
				id:   "id1",
				item: &Job{ID: "id1"},
			},
			mockItem{
				id:   "id2",
				item: &Job{ID: "id2", SubJobs: createSubJobList(3)},
			},
			mockItem{
				id:   "id3",
				item: &Job{ID: "id3"},
			},
		})

		assert.NoError(t, err)
		assert.Len(t, batches, 3)
		assert.Len(t, batches[0], 1)
		assert.Len(t, batches[1], 1)
		assert.Len(t, batches[2], 1)
		assert.Equal(t, "id1", batches[0][0].GetID())
	})
}

// Current values:
// BenchmarkStore_Get-8           	  200000	     11400 ns/op
func BenchmarkStore_Get(b *testing.B) {
	n := b.N
	s := newJobsStoreWithSize(b, nil, b.N)
	assert.NotNil(b, s)
	createName := func(i int) string {
		return fmt.Sprintf("Identifier%v", i)
	}

	for i := 0; i < n; i++ {
		_, err := s.GetOrCreate(createName(i), createJobWithID(createName(i)))
		assert.NoError(b, err)
	}

	b.ResetTimer()

	for i := 0; i < n; i++ {
		j := s.Get(createName(i))
		assert.NotNil(b, j)
	}
}

func Test_syncBatches(t *testing.T) {
	t.Run("chunksize < length", func(t *testing.T) {
		ctx := context.Background()
		c := NewCustomBatchClient(mocks2.NewMockAwsBatchClient(), "", "",
			utils.NewRateLimiter("", 10, 20),
			utils.NewRateLimiter("", 10, 20))

		for i := 0; i < 10; i++ {
			_, err := c.SubmitJob(ctx, &batch.SubmitJobInput{
				JobName: refStr(fmt.Sprintf("job-%v", i)),
			})

			assert.NoError(t, err)
		}

		f := syncBatches(ctx, c, EventHandler{Updated: func(ctx context.Context, event Event) {
		}}, 2)

		i := &mocks3.ItemWrapper{}
		i.OnGetItem().Return(&Job{
			ID:      "job",
			SubJobs: createSubJobList(3),
		})
		i.OnGetID().Return("job")

		var resp []cache.ItemSyncResponse
		var err error
		for iter := 0; iter < 2; iter++ {
			resp, err = f(ctx, []cache.ItemWrapper{i, i, i})
			assert.NoError(t, err)
			assert.Len(t, resp, 3)
		}

		assert.Len(t, resp, 3)
	})

	t.Run("chunksize > length", func(t *testing.T) {
		ctx := context.Background()
		c := NewCustomBatchClient(mocks2.NewMockAwsBatchClient(), "", "",
			utils.NewRateLimiter("", 10, 20),
			utils.NewRateLimiter("", 10, 20))

		for i := 0; i < 10; i++ {
			_, err := c.SubmitJob(ctx, &batch.SubmitJobInput{
				JobName: refStr(fmt.Sprintf("job-%v", i)),
			})

			assert.NoError(t, err)
		}

		f := syncBatches(ctx, c, EventHandler{Updated: func(ctx context.Context, event Event) {
		}}, 10)

		i := &mocks3.ItemWrapper{}
		i.OnGetItem().Return(&Job{
			ID:      "job",
			SubJobs: createSubJobList(3),
		})
		i.OnGetID().Return("job")

		var resp []cache.ItemSyncResponse
		var err error
		for iter := 0; iter < 2; iter++ {
			resp, err = f(ctx, []cache.ItemWrapper{i, i, i})
			assert.NoError(t, err)
			assert.Len(t, resp, 3)
		}

		assert.Len(t, resp, 3)
	})
}

func Test_toRanges(t *testing.T) {
	tests := []struct {
		name         string
		totalSize    int
		chunkSize    int
		wantStartIdx []int
		wantEndIdx   []int
	}{
		{"even", 10, 2, []int{0, 2, 4, 6, 8}, []int{2, 4, 6, 8, 10}},
		{"odd", 10, 3, []int{0, 3, 6, 9}, []int{3, 6, 9, 10}},
		{"chunk>size", 2, 10, []int{0}, []int{2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStartIdx, gotEndIdx := toRanges(tt.totalSize, tt.chunkSize)
			if !reflect.DeepEqual(gotStartIdx, tt.wantStartIdx) {
				t.Errorf("toRanges() gotStartIdx = %v, want %v", gotStartIdx, tt.wantStartIdx)
			}
			if !reflect.DeepEqual(gotEndIdx, tt.wantEndIdx) {
				t.Errorf("toRanges() gotEndIdx = %v, want %v", gotEndIdx, tt.wantEndIdx)
			}
		})
	}
}

func Test_updateJob(t *testing.T) {
	ctx := context.Background()
	withDefaults := func(source *batch.JobDetail) *batch.JobDetail {
		source.JobId = refStr("job-1")
		source.Status = refStr(batch.JobStatusRunning)
		return source
	}

	t.Run("Current attempt", func(t *testing.T) {
		j := Job{}
		updated := updateJob(ctx, withDefaults(&batch.JobDetail{
			Container: &batch.ContainerDetail{
				LogStreamName: refStr("stream://log2"),
			},
		}), &j)

		assert.True(t, updated)
		assert.Len(t, j.Attempts, 1)
	})

	t.Run("Current attempt, 1 already failed", func(t *testing.T) {
		j := Job{}
		updated := updateJob(ctx, withDefaults(&batch.JobDetail{
			Container: &batch.ContainerDetail{
				LogStreamName: refStr("stream://log2"),
			},
			Attempts: []*batch.AttemptDetail{
				{
					Container: &batch.AttemptContainerDetail{
						LogStreamName: refStr("stream://log1"),
					},
				},
			},
		}), &j)

		assert.True(t, updated)
		assert.Len(t, j.Attempts, 2)
	})

	t.Run("No current attempt, 1 already failed", func(t *testing.T) {
		j := Job{}
		updated := updateJob(ctx, withDefaults(&batch.JobDetail{
			Container: &batch.ContainerDetail{
				LogStreamName: refStr("stream://log1"),
			},
			Attempts: []*batch.AttemptDetail{
				{
					Container: &batch.AttemptContainerDetail{
						LogStreamName: refStr("stream://log1"),
					},
				},
			},
		}), &j)

		assert.True(t, updated)
		assert.Len(t, j.Attempts, 1)
	})
}
