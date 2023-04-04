package implementations

import (
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	event2 "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
)

func TestNodeExecutionEventWriter(t *testing.T) {
	db := mocks.NewMockRepository()

	event := admin.NodeExecutionEventRequest{
		RequestId: "request_id",
		Event: &event2.NodeExecutionEvent{
			Id: &core.NodeExecutionIdentifier{
				NodeId: "node_id",
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "project",
					Domain:  "domain",
					Name:    "exec_name",
				},
			},
		},
	}

	nodeExecEventRepo := mocks.NodeExecutionEventRepoInterface{}
	nodeExecEventRepo.On("Create", event).Return(nil)
	db.(*mocks.MockRepository).NodeExecutionEventRepoIface = &nodeExecEventRepo
	writer := NewNodeExecutionEventWriter(db, 100)
	// Assert we can write an event using the buffered channel without holding up this process.
	writer.Write(event)
	go func() { writer.Run() }()
	close(writer.(*nodeExecutionEventWriter).events)
}
