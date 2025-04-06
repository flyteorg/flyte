//go:build integration
// +build integration

// Add this tag to your project settings if you want to pick it up.

package events

import (
	"context"
	"fmt"
	netUrl "net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/ptypes"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flytestdlib/config"
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

	eventSink, err := NewAdminEventSink(ctx, adminClient, eventSinkConfig)

	wfEvent := &event.WorkflowExecutionEvent{
		Phase:      core.WorkflowExecution_RUNNING,
		OccurredAt: timestamppb.Now(),
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Project: "flyteexamples",
			Domain:  "development",
			Name:    "ikuy55mn0y",
		},
		ProducerId:   "testproducer",
		OutputResult: &event.WorkflowExecutionEvent_OutputUri{"s3://blah/blah/blah"},
	}

	err = eventSink.Sink(ctx, wfEvent)
	assert.NoError(t, err)
}
