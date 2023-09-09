package transformers

import (
	"testing"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
)

func TestCreateExecutionEventModel(t *testing.T) {
	requestID := "foo"
	phase := core.WorkflowExecution_RUNNING

	timestamp := time.Now().UTC()
	occurredAt, _ := ptypes.TimestampProto(timestamp)
	executionEvent, err := CreateExecutionEventModel(
		admin.WorkflowExecutionEventRequest{
			RequestId: requestID,
			Event: &event.WorkflowExecutionEvent{
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
				Phase:      phase,
				OccurredAt: occurredAt,
			},
		})

	assert.Nil(t, err)
	assert.Equal(t, requestID, executionEvent.RequestID)
	assert.Equal(t, models.ExecutionKey{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
	}, executionEvent.ExecutionKey)
	assert.Equal(t, timestamp, executionEvent.OccurredAt)
	assert.Equal(t, phase.String(), executionEvent.Phase)
}
