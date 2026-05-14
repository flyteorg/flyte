package webapi

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/internal/webapi/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/webapi"
	"github.com/flyteorg/flyte/v2/flytestdlib/autorefreshcache"
	cacheMocks "github.com/flyteorg/flyte/v2/flytestdlib/autorefreshcache/mocks"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

func TestNewResourceCache(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		c, err := NewResourceCache(context.Background(), "Cache1", &mocks.Client{}, webapi.CachingConfig{
			Size: 10,
		}, webapi.RateLimiterConfig{QPS: 1, Burst: 1}, promutils.NewTestScope())
		assert.NoError(t, err)
		assert.NotNil(t, c)
	})

	t.Run("Error", func(t *testing.T) {
		_, err := NewResourceCache(context.Background(), "Cache1", &mocks.Client{}, webapi.CachingConfig{},
			webapi.RateLimiterConfig{},
			promutils.NewTestScope())
		assert.Error(t, err)
	})
}

func TestResourceCache_SyncResource(t *testing.T) {
	ctx := context.Background()

	t.Run("Terminal state return unchanged", func(t *testing.T) {
		mockCache := &cacheMocks.AutoRefresh{}
		mockClient := &mocks.Client{}

		q := ResourceCache{
			AutoRefresh: mockCache,
			client:      mockClient,
			cfg: webapi.CachingConfig{
				MaxSystemFailures: 5,
			},
		}

		state := webapi.State{
			Phase: webapi.PhaseSucceeded,
		}

		cacheItem := CacheItem{
			State: state,
		}

		iw := &cacheMocks.ItemWrapper{}
		iw.EXPECT().GetItem().Return(cacheItem)
		iw.EXPECT().GetID().Return("some-id")

		newCacheItem, err := q.SyncResource(ctx, []autorefreshcache.ItemWrapper{iw})
		assert.NoError(t, err)
		assert.Equal(t, autorefreshcache.Unchanged, newCacheItem[0].Action)
		assert.Equal(t, cacheItem, newCacheItem[0].Item)
	})

	t.Run("Retry limit exceeded", func(t *testing.T) {
		mockCache := &cacheMocks.AutoRefresh{}
		mockClient := &mocks.Client{}

		q := ResourceCache{
			AutoRefresh: mockCache,
			client:      mockClient,
			cfg: webapi.CachingConfig{
				MaxSystemFailures: 2,
			},
		}

		cacheItem := CacheItem{
			State: webapi.State{
				SyncFailureCount: 5,
				ErrorMessage:     "some error",
			},
		}

		iw := &cacheMocks.ItemWrapper{}
		iw.EXPECT().GetItem().Return(cacheItem)
		iw.EXPECT().GetID().Return("some-id")

		newCacheItem, err := q.SyncResource(ctx, []autorefreshcache.ItemWrapper{iw})
		assert.NoError(t, err)
		assert.Equal(t, autorefreshcache.Update, newCacheItem[0].Action)
		cacheItem.State.Phase = webapi.PhaseSystemFailure
		assert.Equal(t, cacheItem, newCacheItem[0].Item)
	})

	t.Run("move to success", func(t *testing.T) {
		mockCache := &cacheMocks.AutoRefresh{}
		mockClient := &mocks.Client{}
		q := ResourceCache{
			AutoRefresh: mockCache,
			client:      mockClient,
			cfg: webapi.CachingConfig{
				MaxSystemFailures: 5,
			},
		}

		state := webapi.State{
			ResourceMeta: "123456",
			Phase:        webapi.PhaseResourcesCreated,
		}

		cacheItem := CacheItem{
			State: state,
		}

		mockClient.EXPECT().Get(ctx, newPluginContext("123456", nil, "", nil)).Return("newID", nil)
		mockClient.EXPECT().Status(mock.Anything, mock.Anything).Return(core.PhaseInfoSuccess(nil), nil)

		iw := &cacheMocks.ItemWrapper{}
		iw.EXPECT().GetItem().Return(cacheItem)
		iw.EXPECT().GetID().Return("some-id")

		newCacheItem, err := q.SyncResource(ctx, []autorefreshcache.ItemWrapper{iw})
		assert.NoError(t, err)
		assert.Equal(t, autorefreshcache.Update, newCacheItem[0].Action)
	})

	t.Run("Failing to retrieve latest", func(t *testing.T) {
		mockCache := &cacheMocks.AutoRefresh{}
		mockClient := &mocks.Client{}

		q := ResourceCache{
			AutoRefresh: mockCache,
			client:      mockClient,
			cfg: webapi.CachingConfig{
				MaxSystemFailures: 5,
			},
		}

		state := webapi.State{
			ResourceMeta: "123456",
			Phase:        webapi.PhaseResourcesCreated,
		}

		cacheItem := CacheItem{
			State: state,
		}

		mockClient.EXPECT().Get(ctx, newPluginContext("123456", nil, "", nil)).Return("newID", fmt.Errorf("failed to retrieve resource"))

		iw := &cacheMocks.ItemWrapper{}
		iw.EXPECT().GetItem().Return(cacheItem)
		iw.EXPECT().GetID().Return("some-id")

		newCacheItem, err := q.SyncResource(ctx, []autorefreshcache.ItemWrapper{iw})
		newExecutionState := newCacheItem[0].Item.(CacheItem)
		assert.NoError(t, err)
		assert.Equal(t, autorefreshcache.Update, newCacheItem[0].Action)
		assert.Equal(t, webapi.PhaseResourcesCreated, newExecutionState.Phase)
	})
}

func TestToPluginPhase(t *testing.T) {
	tests := []struct {
		args    core.Phase
		want    webapi.Phase
		wantErr bool
	}{
		{core.PhaseNotReady, webapi.PhaseNotStarted, false},
		{core.PhaseUndefined, webapi.PhaseNotStarted, false},
		{core.PhaseInitializing, webapi.PhaseResourcesCreated, false},
		{core.PhaseWaitingForResources, webapi.PhaseResourcesCreated, false},
		{core.PhaseQueued, webapi.PhaseResourcesCreated, false},
		{core.PhaseRunning, webapi.PhaseResourcesCreated, false},
		{core.PhaseSuccess, webapi.PhaseSucceeded, false},
		{core.PhasePermanentFailure, webapi.PhaseUserFailure, false},
		{core.PhaseRetryableFailure, webapi.PhaseUserFailure, false},
	}
	for _, tt := range tests {
		t.Run(tt.args.String(), func(t *testing.T) {
			got, err := ToPluginPhase(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToPluginPhase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ToPluginPhase() got = %v, want %v", got, tt.want)
			}
		})
	}
}
