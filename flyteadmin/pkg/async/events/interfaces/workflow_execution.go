package interfaces

import (
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

//go:generate mockery -name=WorkflowExecutionEventWriter -output=../mocks -case=underscore

type WorkflowExecutionEventWriter interface {
	Run()
	Write(workflowExecutionEvent admin.WorkflowExecutionEventRequest)
}
