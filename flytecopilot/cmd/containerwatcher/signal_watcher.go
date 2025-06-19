package containerwatcher

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/flyteorg/flyte/flytestdlib/logger"
)

type SignalWatcher struct {
}

func (n SignalWatcher) WaitToStart(ctx context.Context) error {
	logger.Warn(ctx, "WaitToStart is not needed for signal watcher.")
	return nil
}

func (n SignalWatcher) WaitToExit(ctx context.Context) error {
	logger.Infof(ctx, "Signal Watcher waiting for termination signal")
	defer logger.Infof(ctx, "Signal Watcher exiting on termination signal")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Listen for SIGTERM
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM)
	defer signal.Stop(sigs)

	// Wait for SIGTERM signal or cancel context
	select {
	case sig := <-sigs:
		logger.Infof(ctx, "Received signal: %v", sig)
		return nil
	case <-ctx.Done():
		logger.Infof(ctx, "Context canceled")
		return nil
	}
}
