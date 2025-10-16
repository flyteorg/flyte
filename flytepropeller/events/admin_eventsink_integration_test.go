//go:build integration
// +build integration

// Add this tag to your project settings if you want to pick it up.
// Run with: go test -tags=integration -v

package events

import (
	"context"
	"fmt"
	netUrl "net/url"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/fastcheck"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

var (
	u, _               = netUrl.Parse("localhost:8089")
	adminServiceConfig = admin.Config{
		Endpoint:              config.URL{URL: *u},
		UseInsecureConnection: true,
		PerRetryTimeout:       config.Duration{1 * time.Second},
		MaxRetries:            1,
	}
)

// To run this test, and see if the deadline working, pick an existing successful execution from your admin database
//
//	select * from executions;
//
// Then delete all the events from it.
//
//	delete from execution_events where execution_name = 'ikuy55mn0y';
//
// Then run this
//
//	begin work; lock table executions in ACCESS EXCLUSIVE mode; SELECT pg_sleep(20); commit work;
//
// This will lock your table so that admin can't read it, causing the grpc call to timeout.
// On timeout, you should get a deadline exceeded error.  Otherwise, you should get an error to the effect of
// "Invalid phase change from SUCCEEDED to RUNNING" or something like that.
// Lastly be sure to port forward admin, or change url above to the dns name if running in-cluster
func TestAdminEventSinkTimeout(t *testing.T) {
	ctx := context.Background()
	fmt.Println(u.Scheme)

	adminClient := admin.InitializeAdminClient(ctx, adminServiceConfig)

	eventSinkConfig := &Config{
		Rate:     1,
		Capacity: 1,
	}

	scope := promutils.NewTestScope()
	filter, err := fastcheck.NewOppoBloomFilter(1000, scope.NewSubScope("integration_filter"))
	assert.NoError(t, err)

	eventSink, err := NewAdminEventSink(ctx, adminClient, eventSinkConfig, filter, scope)
	assert.NoError(t, err)
	defer eventSink.Close()

	wfEvent := &event.WorkflowExecutionEvent{
		Phase:      core.WorkflowExecution_RUNNING,
		OccurredAt: ptypes.TimestampNow(),
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Project: "flyteexamples",
			Domain:  "development",
			Name:    "ikuy55mn0y",
		},
		ProducerId:   "testproducer",
		OutputResult: &event.WorkflowExecutionEvent_OutputUri{OutputUri: "s3://blah/blah/blah"},
	}

	err = eventSink.Sink(ctx, wfEvent)
	assert.NoError(t, err)

	// Wait for async processing
	time.Sleep(2 * time.Second)
}

// TestAdminEventSinkAsyncIntegration tests async event processing against a real admin service
// Requires admin to be running on localhost:8089
// Run with: go test -tags=integration -v -run TestAdminEventSinkAsyncIntegration
func TestAdminEventSinkAsyncIntegration(t *testing.T) {
	ctx := context.Background()

	adminClient := admin.InitializeAdminClient(ctx, adminServiceConfig)

	eventSinkConfig := &Config{
		Rate:     10, // 10 events per second
		Capacity: 10,
	}

	scope := promutils.NewTestScope()
	filter, err := fastcheck.NewOppoBloomFilter(1000, scope.NewSubScope("integration_filter"))
	assert.NoError(t, err)

	eventSink, err := NewAdminEventSink(ctx, adminClient, eventSinkConfig, filter, scope)
	assert.NoError(t, err)
	defer eventSink.Close()

	// Send multiple events asynchronously
	numEvents := 5
	for i := 0; i < numEvents; i++ {
		wfEvent := &event.WorkflowExecutionEvent{
			Phase:      core.WorkflowExecution_RUNNING,
			OccurredAt: ptypes.TimestampNow(),
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "flyteexamples",
				Domain:  "development",
				Name:    fmt.Sprintf("async-integration-test-%d", i),
			},
			ProducerId:   "integration-test",
			OutputResult: &event.WorkflowExecutionEvent_OutputUri{OutputUri: "s3://test/output"},
		}

		err = eventSink.Sink(ctx, wfEvent)
		assert.NoError(t, err, "Should enqueue event %d without blocking", i)
	}

	// Wait for all events to be processed
	time.Sleep(3 * time.Second)

	// Try to send duplicate - should be rejected as already sent
	duplicateEvent := &event.WorkflowExecutionEvent{
		Phase:      core.WorkflowExecution_RUNNING,
		OccurredAt: ptypes.TimestampNow(),
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Project: "flyteexamples",
			Domain:  "development",
			Name:    "async-integration-test-0",
		},
		ProducerId:   "integration-test",
		OutputResult: &event.WorkflowExecutionEvent_OutputUri{OutputUri: "s3://test/output"},
	}

	err = eventSink.Sink(ctx, duplicateEvent)
	assert.Error(t, err, "Should reject duplicate event")
}
