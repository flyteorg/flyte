package containerwatcher

import (
	"context"
	"fmt"
)

var ErrTimeout = fmt.Errorf("timeout while waiting")

type Watcher interface {
	WaitToStart(ctx context.Context) error
	WaitToExit(ctx context.Context) error
}

type WatcherType = string

const (
	// Uses Kube 1.17 feature - https://kubernetes.io/docs/tasks/configure-pod-container/share-process-namespace/
	// To look for pid in the shared namespace.
	WatcherTypeSharedProcessNS WatcherType = "shared-process-ns"
	// Uses Kube 1.28 feature - https://kubernetes.io/docs/concepts/workloads/pods/sidecar-containers/
	// Watching SIGTERM when main container exit
	WatcherTypeSignal WatcherType = "signal"
	// Dummy watcher. Exits immediately, assuming success
	WatcherTypeNoop WatcherType = "noop"
)

var AllWatcherTypes = []WatcherType{
	WatcherTypeSharedProcessNS,
	WatcherTypeSignal,
	WatcherTypeNoop,
}
