package core

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/flytestdlib/bitarray"
	idlCore "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

func TestInitializeExternalResources(t *testing.T) {
	ctx := context.TODO()
	subTaskCount := 10
	cachedCount := 4

	indexesToCache := InvertBitSet(bitarray.NewBitSet(uint(subTaskCount)), uint(subTaskCount)) // #nosec G115
	for i := 0; i < cachedCount; i++ {
		indexesToCache.Clear(uint(i)) // #nosec G115
	}

	tr := &mocks.TaskReader{}
	tr.EXPECT().Read(ctx).Return(&idlCore.TaskTemplate{
		Metadata: &idlCore.TaskMetadata{
			Discoverable: true,
		},
	}, nil)

	tID := &mocks.TaskExecutionID{}
	tID.EXPECT().GetGeneratedName().Return("notfound")

	tMeta := &mocks.TaskExecutionMetadata{}
	tMeta.EXPECT().GetTaskExecutionID().Return(tID)

	tCtx := &mocks.TaskExecutionContext{}
	tCtx.EXPECT().TaskReader().Return(tr)
	tCtx.EXPECT().TaskExecutionMetadata().Return(tMeta)

	state := State{
		OriginalArraySize:  int64(subTaskCount),
		ExecutionArraySize: subTaskCount - cachedCount,
		IndexesToCache:     indexesToCache,
	}

	externalResources, err := InitializeExternalResources(ctx, tCtx, &state,
		func(_ core.TaskExecutionContext, i int) string {
			return ""
		},
	)

	assert.Nil(t, err)
	assert.Equal(t, subTaskCount, len(externalResources))
	for i, externalResource := range externalResources {
		assert.Equal(t, uint32(i), externalResource.Index) // #nosec G115
		assert.Equal(t, 0, len(externalResource.Logs))
		assert.Equal(t, uint32(0), externalResource.RetryAttempt)
		if i < cachedCount {
			assert.Equal(t, core.PhaseSuccess, externalResource.Phase)
			assert.Equal(t, idlCore.CatalogCacheStatus_CACHE_HIT, externalResource.CacheStatus)
		} else {
			assert.Equal(t, core.PhaseUndefined, externalResource.Phase)
			assert.Equal(t, idlCore.CatalogCacheStatus_CACHE_MISS, externalResource.CacheStatus)
		}
	}
}
