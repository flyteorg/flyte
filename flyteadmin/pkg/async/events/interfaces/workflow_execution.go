package interfaces

import (
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

type WorkflowExecutionEventWriter interface {
	Run()
	Write(workflowExecutionEvent *admin.WorkflowExecutionEventRequest)
}
