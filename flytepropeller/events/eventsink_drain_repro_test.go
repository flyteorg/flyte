package events

import (
	"context"
	"testing"

	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flytestdlib/fastcheck"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

func taskEventFor(name string) *event.TaskExecutionEvent {
	return &event.TaskExecutionEvent{
		Phase:        core.TaskExecution_SUCCEEDED,
		OccurredAt:   ptypes.TimestampNow(),
		TaskId:       &core.Identifier{ResourceType: core.ResourceType_TASK, Name: "task-id"},
		RetryAttempt: 1,
		ParentNodeExecutionId: &core.NodeExecutionIdentifier{
			NodeId: "node-id",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "p",
				Domain:  "d",
				Name:    name,
			},
		},
		Logs: []*core.TaskLog{{Uri: "logs.txt"}},
	}
}

// A workflow that keeps re-emitting an event Admin has already durably recorded must not be able
// to drain the process-wide event-sink rate limiter and starve other executions.
//
// The sink is built exactly as flytepropeller builds it in NewAdminEventSink: a real
// OppoBloomFilter and a real token-bucket limiter. Only the Admin client is a stub, and it
// answers exactly as a real Admin does for an event it already holds: codes.AlreadyExists.
func TestAdminEventSink_AlreadyExistsDrainsRateLimiter(t *testing.T) {
	ctx := context.Background()
	scope := promutils.NewTestScope()

	filter, err := fastcheck.NewOppoBloomFilter(50000, scope.NewSubScope("filter"))
	assert.NoError(t, err)

	// Capacity 10: a real deployment's shared budget, scaled down so the drain is observable.
	adminClient := &mocks.AdminServiceClient{}
	sink, err := NewAdminEventSink(ctx, adminClient, &Config{Rate: 1, Capacity: 10}, filter)
	assert.NoError(t, err)

	stuck := taskEventFor("stuck-execution")
	healthy := taskEventFor("healthy-execution")

	adminCalls := 0
	adminClient.On("CreateTaskEvent", ctx,
		mock.MatchedBy(func(req *admin.TaskExecutionEventRequest) bool {
			return req.GetEvent().GetParentNodeExecutionId().GetExecutionId().GetName() == "stuck-execution"
		}),
	).Run(func(mock.Arguments) { adminCalls++ }).
		Return(nil, status.Error(codes.AlreadyExists, "Grpc AlreadyExists error"))

	adminClient.On("CreateTaskEvent", ctx,
		mock.MatchedBy(func(req *admin.TaskExecutionEventRequest) bool {
			return req.GetEvent().GetParentNodeExecutionId().GetExecutionId().GetName() == "healthy-execution"
		}),
	).Return(&admin.TaskExecutionEventResponse{}, nil)

	// Admin already holds this event, so re-emitting it can never succeed. Every attempt should
	// be answered from the local dedup filter, not by spending a token on a doomed round trip.
	for i := 0; i < 10; i++ {
		_ = sink.Sink(ctx, stuck)
	}
	t.Logf("re-emits of an already-recorded event that reached Admin: %d/10", adminCalls)
	assert.Equal(t, 1, adminCalls,
		"an already-recorded event should reach Admin once; afterwards the dedup filter should answer")

	// The unrelated execution must still be able to record its event.
	healthyErr := sink.Sink(ctx, healthy)
	assert.NoError(t, healthyErr,
		"a healthy execution was starved of rate-limiter capacity by an already-recorded event")
}
