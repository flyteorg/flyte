package task

import (
	"context"
	"testing"
	"time"

	pluginCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/stretchr/testify/assert"
)

func TestBufferedEventRecorder(t *testing.T) {
	ctx := context.TODO()

	bev := newBufferedEventRecorder()
	assert.NotNil(t, bev)
	assert.Equal(t, bev.GetAll(ctx), []pluginCore.PhaseInfo{})

	ev1 := pluginCore.PhaseInfoInitializing(time.Now(), 0, "starting", nil)
	assert.NoError(t, bev.RecordRaw(ctx, ev1))
	assert.Equal(t, bev.GetAll(ctx), []pluginCore.PhaseInfo{ev1})

	ev2 := pluginCore.PhaseInfoSuccess(nil)
	assert.NoError(t, bev.RecordRaw(ctx, ev2))
	assert.Equal(t, bev.GetAll(ctx), []pluginCore.PhaseInfo{ev1, ev2})
}
