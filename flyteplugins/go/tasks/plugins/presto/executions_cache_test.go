package presto

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/presto/client"
	prestoMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/presto/client/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/presto/config"
	"github.com/flyteorg/flyte/flytestdlib/cache"
	cacheMocks "github.com/flyteorg/flyte/flytestdlib/cache/mocks"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

func TestPrestoExecutionsCache_SyncQuboleQuery(t *testing.T) {
	ctx := context.Background()

	t.Run("terminal state return unchanged", func(t *testing.T) {
		mockCache := &cacheMocks.AutoRefresh{}
		mockPresto := &prestoMocks.PrestoClient{}
		testScope := promutils.NewTestScope()

		p := ExecutionsCache{
			AutoRefresh:  mockCache,
			prestoClient: mockPresto,
			scope:        testScope,
			cfg:          config.GetPrestoConfig(),
		}

		state := ExecutionState{
			CurrentPhase: PhaseQuerySucceeded,
		}
		cacheItem := ExecutionStateCacheItem{
			ExecutionState: state,
			Identifier:     "some-id",
		}

		iw := &cacheMocks.ItemWrapper{}
		iw.EXPECT().GetItem().Return(cacheItem)
		iw.EXPECT().GetID().Return("some-id")

		newCacheItem, err := p.SyncPrestoQuery(ctx, []cache.ItemWrapper{iw})
		assert.NoError(t, err)
		assert.Equal(t, cache.Unchanged, newCacheItem[0].Action)
		assert.Equal(t, cacheItem, newCacheItem[0].Item)
	})

	t.Run("move to success", func(t *testing.T) {
		mockCache := &cacheMocks.AutoRefresh{}
		mockPresto := &prestoMocks.PrestoClient{}
		mockSecretManager := &mocks.SecretManager{}
		mockSecretManager.EXPECT().Get(mock.Anything, mock.Anything).Return("fake key", nil)

		testScope := promutils.NewTestScope()

		p := ExecutionsCache{
			AutoRefresh:  mockCache,
			prestoClient: mockPresto,
			scope:        testScope,
			cfg:          config.GetPrestoConfig(),
		}

		state := ExecutionState{
			CommandID:    "123456",
			CurrentPhase: PhaseSubmitted,
		}
		cacheItem := ExecutionStateCacheItem{
			ExecutionState: state,
			Identifier:     "some-id",
		}
		mockPresto.EXPECT().GetCommandStatus(mock.Anything, mock.MatchedBy(func(commandId string) bool {
			return commandId == state.CommandID
		})).Return(client.PrestoStatusFinished, nil)

		iw := &cacheMocks.ItemWrapper{}
		iw.EXPECT().GetItem().Return(cacheItem)
		iw.EXPECT().GetID().Return("some-id")

		newCacheItem, err := p.SyncPrestoQuery(ctx, []cache.ItemWrapper{iw})
		newExecutionState := newCacheItem[0].Item.(ExecutionStateCacheItem)
		assert.NoError(t, err)
		assert.Equal(t, cache.Update, newCacheItem[0].Action)
		assert.Equal(t, PhaseQuerySucceeded, newExecutionState.CurrentPhase)
	})
}
