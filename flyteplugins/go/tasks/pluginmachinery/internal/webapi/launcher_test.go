package webapi

import (
	"context"
	"fmt"
	"testing"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/webapi"
	mocks2 "github.com/flyteorg/flytestdlib/cache/mocks"
	"github.com/stretchr/testify/assert"
)

func Test_launch(t *testing.T) {
	t.Run("Successful launch", func(t *testing.T) {
		ctx := context.Background()
		tCtx := &mocks.TaskExecutionContext{}
		meta := &mocks.TaskExecutionMetadata{}
		taskID := &mocks.TaskExecutionID{}
		taskID.OnGetGeneratedName().Return("my-id")
		meta.OnGetTaskExecutionID().Return(taskID)
		tCtx.OnTaskExecutionMetadata().Return(meta)

		c := &mocks2.AutoRefresh{}
		s := State{
			ResourceMeta: "abc",
			Phase:        PhaseResourcesCreated,
		}
		c.OnGetOrCreate("my-id", CacheItem{State: s}).Return(CacheItem{State: s}, nil)

		plgn := newPluginWithProperties(webapi.PluginConfig{})
		plgn.OnCreate(ctx, tCtx).Return("abc", nil, nil)
		plgn.OnStatus(ctx, newPluginContext("abc", nil, "", tCtx)).Return(core.PhaseInfoSuccess(nil), nil)
		newS, phaseInfo, err := launch(ctx, plgn, tCtx, c, &s)
		assert.NoError(t, err)
		assert.NotNil(t, newS)
		assert.NotNil(t, phaseInfo)
	})

	t.Run("Already succeeded when launched", func(t *testing.T) {
		ctx := context.Background()
		tCtx := &mocks.TaskExecutionContext{}
		meta := &mocks.TaskExecutionMetadata{}
		taskID := &mocks.TaskExecutionID{}
		taskID.OnGetGeneratedName().Return("my-id")
		meta.OnGetTaskExecutionID().Return(taskID)
		tCtx.OnTaskExecutionMetadata().Return(meta)

		c := &mocks2.AutoRefresh{}
		s := State{
			Phase:        PhaseResourcesCreated,
			ResourceMeta: "abc",
		}

		plgn := newPluginWithProperties(webapi.PluginConfig{})
		plgn.OnCreate(ctx, tCtx).Return("abc", "abc-r", nil)
		plgn.OnStatus(ctx, newPluginContext("abc", "abc-r", "", tCtx)).Return(core.PhaseInfoSuccess(nil), nil)
		newS, phaseInfo, err := launch(ctx, plgn, tCtx, c, &s)
		assert.NoError(t, err)
		assert.NotNil(t, newS)
		assert.NotNil(t, phaseInfo)
		assert.Equal(t, core.PhaseSuccess, phaseInfo.Phase())
	})

	t.Run("Failed to create resource", func(t *testing.T) {
		ctx := context.Background()
		tCtx := &mocks.TaskExecutionContext{}
		meta := &mocks.TaskExecutionMetadata{}
		taskID := &mocks.TaskExecutionID{}
		taskID.OnGetGeneratedName().Return("my-id")
		meta.OnGetTaskExecutionID().Return(taskID)
		tCtx.OnTaskExecutionMetadata().Return(meta)

		c := &mocks2.AutoRefresh{}
		s := State{}
		c.OnGetOrCreate("my-id", CacheItem{State: s}).Return(CacheItem{State: s}, nil)

		plgn := newPluginWithProperties(webapi.PluginConfig{})
		plgn.OnCreate(ctx, tCtx).Return("", nil, fmt.Errorf("error creating"))
		_, _, err := launch(ctx, plgn, tCtx, c, &s)
		assert.Error(t, err)
	})

	t.Run("Failed to cache", func(t *testing.T) {
		ctx := context.Background()
		tCtx := &mocks.TaskExecutionContext{}
		meta := &mocks.TaskExecutionMetadata{}
		taskID := &mocks.TaskExecutionID{}
		taskID.OnGetGeneratedName().Return("my-id")
		meta.OnGetTaskExecutionID().Return(taskID)
		tCtx.OnTaskExecutionMetadata().Return(meta)

		c := &mocks2.AutoRefresh{}
		s := State{
			Phase:        PhaseResourcesCreated,
			ResourceMeta: "my-id",
		}
		c.OnGetOrCreate("my-id", CacheItem{State: s}).Return(CacheItem{State: s}, fmt.Errorf("failed to cache"))

		plgn := newPluginWithProperties(webapi.PluginConfig{})
		plgn.OnCreate(ctx, tCtx).Return("my-id", nil, nil)
		plgn.OnStatus(ctx, newPluginContext("my-id", nil, "", tCtx)).Return(core.PhaseInfoRunning(0, nil), nil)
		_, _, err := launch(ctx, plgn, tCtx, c, &s)
		assert.Error(t, err)
	})
}
