package hive

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/hive/client"
	quboleMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/hive/client/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/hive/config"
	"github.com/flyteorg/flyte/flytestdlib/cache"
	cacheMocks "github.com/flyteorg/flyte/flytestdlib/cache/mocks"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

func TestQuboleHiveExecutionsCache_SyncQuboleQuery(t *testing.T) {
	ctx := context.Background()

	t.Run("terminal state return unchanged", func(t *testing.T) {
		mockCache := &cacheMocks.AutoRefresh{}
		mockQubole := &quboleMocks.QuboleClient{}
		testScope := promutils.NewTestScope()

		q := QuboleHiveExecutionsCache{
			AutoRefresh:  mockCache,
			quboleClient: mockQubole,
			scope:        testScope,
			cfg:          config.GetQuboleConfig(),
		}

		state := ExecutionState{
			Phase: PhaseQuerySucceeded,
		}
		cacheItem := ExecutionStateCacheItem{
			ExecutionState: state,
			Identifier:     "some-id",
		}

		iw := &cacheMocks.ItemWrapper{}
		iw.EXPECT().GetItem().Return(cacheItem)
		iw.EXPECT().GetID().Return("some-id")

		newCacheItem, err := q.SyncQuboleQuery(ctx, []cache.ItemWrapper{iw})
		assert.NoError(t, err)
		assert.Equal(t, cache.Unchanged, newCacheItem[0].Action)
		assert.Equal(t, cacheItem, newCacheItem[0].Item)
	})

	t.Run("move to success", func(t *testing.T) {
		mockCache := &cacheMocks.AutoRefresh{}
		mockQubole := &quboleMocks.QuboleClient{}
		mockSecretManager := &mocks.SecretManager{}
		mockSecretManager.EXPECT().Get(mock.Anything, mock.Anything).Return("fake key", nil)

		testScope := promutils.NewTestScope()

		q := QuboleHiveExecutionsCache{
			AutoRefresh:   mockCache,
			quboleClient:  mockQubole,
			scope:         testScope,
			secretManager: mockSecretManager,
			cfg:           config.GetQuboleConfig(),
		}

		state := ExecutionState{
			CommandID: "123456",
			Phase:     PhaseSubmitted,
		}
		cacheItem := ExecutionStateCacheItem{
			ExecutionState: state,
			Identifier:     "some-id",
		}
		mockQubole.EXPECT().GetCommandStatus(mock.Anything, mock.MatchedBy(func(commandId string) bool {
			return commandId == state.CommandID
		}), mock.Anything).Return(client.QuboleStatusDone, nil)

		iw := &cacheMocks.ItemWrapper{}
		iw.EXPECT().GetItem().Return(cacheItem)
		iw.EXPECT().GetID().Return("some-id")

		newCacheItem, err := q.SyncQuboleQuery(ctx, []cache.ItemWrapper{iw})
		newExecutionState := newCacheItem[0].Item.(ExecutionStateCacheItem)
		assert.NoError(t, err)
		assert.Equal(t, cache.Update, newCacheItem[0].Action)
		assert.Equal(t, PhaseWriteOutputFile, newExecutionState.Phase)
	})
}
