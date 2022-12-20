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
	// Uses KubeAPI to determine if the container is completed
	WatcherTypeKubeAPI WatcherType = "kube-api"
	// Uses a success file to determine if the container has completed.
	// CAUTION: Does not work if the container exits because of OOM, etc
	WatcherTypeFile WatcherType = "file"
	// Uses Kube 1.17 feature - https://kubernetes.io/docs/tasks/configure-pod-container/share-process-namespace/
	// To look for pid in the shared namespace.
	WatcherTypeSharedProcessNS WatcherType = "shared-process-ns"
	// Dummy watcher. Exits immediately, assuming success
	WatcherTypeNoop WatcherType = "noop"
)

var AllWatcherTypes = []WatcherType{
	WatcherTypeKubeAPI,
	WatcherTypeSharedProcessNS,
	WatcherTypeFile,
	WatcherTypeNoop,
}
