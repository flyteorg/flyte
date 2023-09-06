package task

import (
	"context"

	pluginCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
)

type bufferedEventRecorder struct {
	ev []pluginCore.PhaseInfo
}

func (b *bufferedEventRecorder) RecordRaw(ctx context.Context, ev pluginCore.PhaseInfo) error {
	b.ev = append(b.ev, ev)
	return nil
}

func (b *bufferedEventRecorder) GetAll(ctx context.Context) []pluginCore.PhaseInfo {
	return b.ev
}

func newBufferedEventRecorder() *bufferedEventRecorder {
	return &bufferedEventRecorder{
		ev: make([]pluginCore.PhaseInfo, 0, 1),
	}
}
