package presto

import (
	"context"
	"testing"

	prestoMocks "github.com/flyteorg/flyteplugins/go/tasks/plugins/presto/client/mocks"

	"github.com/flyteorg/flytestdlib/cache"
	cacheMocks "github.com/flyteorg/flytestdlib/cache/mocks"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/presto/client"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/presto/config"

	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
		iw.OnGetItem().Return(cacheItem)
		iw.OnGetID().Return("some-id")

		newCacheItem, err := p.SyncPrestoQuery(ctx, []cache.ItemWrapper{iw})
		assert.NoError(t, err)
		assert.Equal(t, cache.Unchanged, newCacheItem[0].Action)
		assert.Equal(t, cacheItem, newCacheItem[0].Item)
	})

	t.Run("move to success", func(t *testing.T) {
		mockCache := &cacheMocks.AutoRefresh{}
		mockPresto := &prestoMocks.PrestoClient{}
		mockSecretManager := &mocks.SecretManager{}
		mockSecretManager.OnGetMatch(mock.Anything, mock.Anything).Return("fake key", nil)

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
		mockPresto.OnGetCommandStatusMatch(mock.Anything, mock.MatchedBy(func(commandId string) bool {
			return commandId == state.CommandID
		}), mock.Anything).Return(client.PrestoStatusFinished, nil)

		iw := &cacheMocks.ItemWrapper{}
		iw.OnGetItem().Return(cacheItem)
		iw.OnGetID().Return("some-id")

		newCacheItem, err := p.SyncPrestoQuery(ctx, []cache.ItemWrapper{iw})
		newExecutionState := newCacheItem[0].Item.(ExecutionStateCacheItem)
		assert.NoError(t, err)
		assert.Equal(t, cache.Update, newCacheItem[0].Action)
		assert.Equal(t, PhaseQuerySucceeded, newExecutionState.CurrentPhase)
	})
}
