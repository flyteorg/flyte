package containerwatcher

import (
	"context"

	"github.com/flyteorg/flytestdlib/logger"
)

type NoopWatcher struct {
}

func (n NoopWatcher) WaitToStart(ctx context.Context) error {
	logger.Warn(ctx, "noop container watcher setup. assuming container started.")
	return nil
}

func (n NoopWatcher) WaitToExit(ctx context.Context) error {
	logger.Warn(ctx, "noop container watcher setup. assuming container exited.")
	return nil
}
