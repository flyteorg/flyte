package implementations

import (
	"context"

	repositoryInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"

	"github.com/flyteorg/flyteadmin/pkg/async/events/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/transformers"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/logger"
)

// This event writer acts to asynchronously persist workflow execution events. As flytepropeller sends workflow
// events, workflow execution processing doesn't have to wait on these to be committed.
type workflowExecutionEventWriter struct {
	db     repositoryInterfaces.Repository
	events chan admin.WorkflowExecutionEventRequest
}

func (w *workflowExecutionEventWriter) Write(event admin.WorkflowExecutionEventRequest) {
	w.events <- event
}

func (w *workflowExecutionEventWriter) Run() {
	for event := range w.events {
		eventModel, err := transformers.CreateExecutionEventModel(event)
		if err != nil {
			logger.Warnf(context.TODO(), "Failed to transform event [%+v] to database model with err [%+v]", event, err)
			continue
		}
		err = w.db.ExecutionEventRepo().Create(context.TODO(), *eventModel)
		if err != nil {
			// It's okay to be lossy here. These events aren't used to fetch execution state but rather as a convenience
			// to replay and understand the event execution timeline.
			logger.Warnf(context.TODO(), "Failed to write event [%+v] to database with err [%+v]", event, err)
		}
	}
}

func NewWorkflowExecutionEventWriter(db repositoryInterfaces.Repository, bufferSize int) interfaces.WorkflowExecutionEventWriter {
	return &workflowExecutionEventWriter{
		db:     db,
		events: make(chan admin.WorkflowExecutionEventRequest, bufferSize),
	}
}
