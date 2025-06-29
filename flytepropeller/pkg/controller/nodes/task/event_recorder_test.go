package task

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	pluginCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
)

func TestBufferedEventRecorder(t *testing.T) {
	ctx := context.TODO()

	bev := newBufferedEventRecorder()
	assert.NotNil(t, bev)
	assert.Equal(t, bev.GetAll(ctx), []pluginCore.PhaseInfo{})

	ev1 := pluginCore.PhaseInfoInitializing(0, "starting", nil)
	assert.NoError(t, bev.RecordRaw(ctx, ev1))
	assert.Equal(t, bev.GetAll(ctx), []pluginCore.PhaseInfo{ev1})

	ev2 := pluginCore.PhaseInfoSuccess(nil)
	assert.NoError(t, bev.RecordRaw(ctx, ev2))
	assert.Equal(t, bev.GetAll(ctx), []pluginCore.PhaseInfo{ev1, ev2})
}
