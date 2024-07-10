package interfaces

import (
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/watch"
)

//go:generate mockery -name=WatchInterface -output=../mocks -case=underscore

// WatchInterface watches flyteadmin events
type WatchInterface interface {
	WatchExecutionStatusUpdates(*watch.WatchExecutionStatusUpdatesRequest, service.WatchService_WatchExecutionStatusUpdatesServer) error
}
