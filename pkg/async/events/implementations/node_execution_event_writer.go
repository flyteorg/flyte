package implementations

import (
	"context"

	repositoryInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"

	"github.com/flyteorg/flyteadmin/pkg/async/events/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/transformers"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/logger"
)

// This event writer acts to asynchronously persist node execution events. As flytepropeller sends node
// events, node execution processing doesn't have to wait on these to be committed.
type nodeExecutionEventWriter struct {
	db     repositoryInterfaces.Repository
	events chan admin.NodeExecutionEventRequest
}

func (w *nodeExecutionEventWriter) Write(event admin.NodeExecutionEventRequest) {
	w.events <- event
}

func (w *nodeExecutionEventWriter) Run() {
	for event := range w.events {
		eventModel, err := transformers.CreateNodeExecutionEventModel(event)
		if err != nil {
			logger.Warnf(context.TODO(), "Failed to transform event [%+v] to database model with err [%+v]", event, err)
			continue
		}
		err = w.db.NodeExecutionEventRepo().Create(context.TODO(), *eventModel)
		if err != nil {
			// It's okay to be lossy here. These events aren't used to fetch execution state but rather as a convenience
			// to replay and understand the event execution timeline.
			logger.Warnf(context.TODO(), "Failed to write event [%+v] to database with err [%+v]", event, err)
		}
	}
}

func NewNodeExecutionEventWriter(db repositoryInterfaces.Repository, bufferSize int) interfaces.NodeExecutionEventWriter {
	return &nodeExecutionEventWriter{
		db:     db,
		events: make(chan admin.NodeExecutionEventRequest, bufferSize),
	}
}
