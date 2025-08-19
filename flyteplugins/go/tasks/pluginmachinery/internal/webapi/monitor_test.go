package webapi

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/util/workqueue"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	core2 "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	internalMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/internal/webapi/mocks"
	"github.com/flyteorg/flyte/flytestdlib/autorefreshcache"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

func Test_monitor(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	tCtx := &mocks.TaskExecutionContext{}
	ctxMeta := &mocks.TaskExecutionMetadata{}
	execID := &mocks.TaskExecutionID{}
	execID.OnGetGeneratedName().Return("generated_name")
	execID.OnGetID().Return(core.TaskExecutionIdentifier{})
	ctxMeta.OnGetTaskExecutionID().Return(execID)
	tCtx.OnTaskExecutionMetadata().Return(ctxMeta)

	client := &internalMocks.Client{}
	client.OnStatusMatch(ctx, mock.Anything).Return(core2.PhaseInfoSuccess(nil), nil)

	wg := sync.WaitGroup{}
	wg.Add(8)
	cacheObj, err := autorefreshcache.NewAutoRefreshCache(rand.String(5), func(ctx context.Context, batch autorefreshcache.Batch) (updatedBatch []autorefreshcache.ItemSyncResponse, err error) {
		wg.Done()
		t.Logf("Syncing Item [%+v]", batch[0])
		return []autorefreshcache.ItemSyncResponse{
			{
				ID:     batch[0].GetID(),
				Item:   batch[0].GetItem(),
				Action: autorefreshcache.Update,
			},
		}, nil
	}, workqueue.DefaultControllerRateLimiter(), time.Second, 1, 10, promutils.NewTestScope())
	assert.NoError(t, err)

	assert.NoError(t, cacheObj.Start(ctx))

	// Insert a dummy item to make sure the sync loop keeps getting invoked
	_, err = cacheObj.GetOrCreate("generated_name2", CacheItem{Resource: "fake_resource2"})
	assert.NoError(t, err)

	_, err = cacheObj.GetOrCreate("generated_name", CacheItem{Resource: "fake_resource"})
	assert.NoError(t, err)

	s := &State{}
	newState, phaseInfo, err := monitor(ctx, tCtx, client, cacheObj, s)
	assert.NoError(t, err)
	assert.NotNil(t, newState)
	assert.NotNil(t, phaseInfo)
	assert.Equal(t, core2.PhaseSuccess.String(), phaseInfo.Phase().String())

	// Make sure the item is still in the cache as is...
	cachedItem, err := cacheObj.GetOrCreate("generated_name", CacheItem{Resource: "shouldnt_insert"})
	assert.NoError(t, err)
	assert.Equal(t, "fake_resource", cachedItem.(CacheItem).Resource.(string))

	// Wait for sync to run to actually delete the resource
	wg.Wait()
	cancel()
	cachedItem, err = cacheObj.GetOrCreate("generated_name", CacheItem{Resource: "new_resource"})
	assert.NoError(t, err)
	assert.Equal(t, "new_resource", cachedItem.(CacheItem).Resource.(string))
}
