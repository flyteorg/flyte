package core

import (
	"context"
	"testing"

	idlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"

	"github.com/flyteorg/flytestdlib/bitarray"

	"github.com/stretchr/testify/assert"
)

func TestInitializeExternalResources(t *testing.T) {
	ctx := context.TODO()
	subTaskCount := 10
	cachedCount := 4

	indexesToCache := InvertBitSet(bitarray.NewBitSet(uint(subTaskCount)), uint(subTaskCount))
	for i := 0; i < cachedCount; i++ {
		indexesToCache.Clear(uint(i))
	}

	tr := &mocks.TaskReader{}
	tr.OnRead(ctx).Return(&idlCore.TaskTemplate{
		Metadata: &idlCore.TaskMetadata{
			Discoverable: true,
		},
	}, nil)

	tID := &mocks.TaskExecutionID{}
	tID.OnGetGeneratedName().Return("notfound")

	tMeta := &mocks.TaskExecutionMetadata{}
	tMeta.OnGetTaskExecutionID().Return(tID)

	tCtx := &mocks.TaskExecutionContext{}
	tCtx.OnTaskReader().Return(tr)
	tCtx.OnTaskExecutionMetadata().Return(tMeta)

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
		assert.Equal(t, uint32(i), externalResource.Index)
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
