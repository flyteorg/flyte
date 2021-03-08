package webapi

import (
	"context"
	"fmt"
	"testing"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/webapi"
	"github.com/flyteorg/flytestdlib/promutils"

	mocks2 "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/internal/webapi/mocks"
	"github.com/flyteorg/flytestdlib/cache"
	cacheMocks "github.com/flyteorg/flytestdlib/cache/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewResourceCache(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		c, err := NewResourceCache(context.Background(), "Cache1", &mocks.Client{}, webapi.CachingConfig{
			Size: 10,
		}, promutils.NewTestScope())
		assert.NoError(t, err)
		assert.NotNil(t, c)
	})

	t.Run("Error", func(t *testing.T) {
		_, err := NewResourceCache(context.Background(), "Cache1", &mocks.Client{}, webapi.CachingConfig{},
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

		state := State{
			Phase: PhaseSucceeded,
		}

		cacheItem := CacheItem{
			State: state,
		}

		iw := &cacheMocks.ItemWrapper{}
		iw.OnGetItem().Return(cacheItem)
		iw.OnGetID().Return("some-id")

		newCacheItem, err := q.SyncResource(ctx, []cache.ItemWrapper{iw})
		assert.NoError(t, err)
		assert.Equal(t, cache.Unchanged, newCacheItem[0].Action)
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

		state := State{
			ResourceMeta: "123456",
			Phase:        PhaseResourcesCreated,
		}

		cacheItem := CacheItem{
			State: state,
		}

		mockClient.OnGet(ctx, newPluginContext("123456", nil, "", nil)).Return("newID", nil)
		mockClient.OnStatusMatch(mock.Anything, "newID", mock.Anything).Return(core.PhaseInfoSuccess(nil), nil)

		iw := &cacheMocks.ItemWrapper{}
		iw.OnGetItem().Return(cacheItem)
		iw.OnGetID().Return("some-id")

		newCacheItem, err := q.SyncResource(ctx, []cache.ItemWrapper{iw})
		assert.NoError(t, err)
		assert.Equal(t, cache.Update, newCacheItem[0].Action)
	})

	t.Run("Failing to retrieve latest", func(t *testing.T) {
		mockCache := &cacheMocks.AutoRefresh{}
		mockClient := &mocks.Client{}
		mockSecretManager := &mocks2.SecretManager{}
		mockSecretManager.OnGetMatch(mock.Anything, mock.Anything).Return("fake key", nil)

		q := ResourceCache{
			AutoRefresh: mockCache,
			client:      mockClient,
			cfg: webapi.CachingConfig{
				MaxSystemFailures: 5,
			},
		}

		state := State{
			ResourceMeta: "123456",
			Phase:        PhaseResourcesCreated,
		}

		cacheItem := CacheItem{
			State: state,
		}

		mockClient.OnGet(ctx, newPluginContext("123456", nil, "", nil)).Return("newID", fmt.Errorf("failed to retrieve resource"))

		iw := &cacheMocks.ItemWrapper{}
		iw.OnGetItem().Return(cacheItem)
		iw.OnGetID().Return("some-id")

		newCacheItem, err := q.SyncResource(ctx, []cache.ItemWrapper{iw})
		newExecutionState := newCacheItem[0].Item.(CacheItem)
		assert.NoError(t, err)
		assert.Equal(t, cache.Update, newCacheItem[0].Action)
		assert.Equal(t, PhaseResourcesCreated, newExecutionState.Phase)
	})
}

func TestToPluginPhase(t *testing.T) {
	tests := []struct {
		args    core.Phase
		want    Phase
		wantErr bool
	}{
		{core.PhaseNotReady, PhaseNotStarted, false},
		{core.PhaseUndefined, PhaseNotStarted, false},
		{core.PhaseInitializing, PhaseResourcesCreated, false},
		{core.PhaseWaitingForResources, PhaseResourcesCreated, false},
		{core.PhaseQueued, PhaseResourcesCreated, false},
		{core.PhaseRunning, PhaseResourcesCreated, false},
		{core.PhaseSuccess, PhaseSucceeded, false},
		{core.PhasePermanentFailure, PhaseUserFailure, false},
		{core.PhaseRetryableFailure, PhaseUserFailure, false},
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
