package notifications

import (
	"github.com/NYTimes/gizmo/pubsub"
	gizmoAWS "github.com/NYTimes/gizmo/pubsub/aws"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	"golang.org/x/net/context"
	"testing"

	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

func TestGetEmailer(t *testing.T) {
	defer func() { r := recover(); assert.NotNil(t, r) }()
	cfg := runtimeInterfaces.NotificationsConfig{
		NotificationsEmailerConfig: runtimeInterfaces.NotificationsEmailerConfig{
			EmailerConfig: runtimeInterfaces.EmailServerConfig{
				ServiceName: "unsupported",
			},
		},
	}

	GetEmailer(cfg, promutils.NewTestScope())

	// shouldn't reach here
	t.Errorf("did not panic")
}

func TestAWSNotificationsPublisher(t *testing.T) {
	cfg := runtimeInterfaces.NotificationsConfig{
		Type:   "aws",
		Region: "us-west-2",
		NotificationsPublisherConfig: runtimeInterfaces.NotificationsPublisherConfig{
			TopicName: "arn:aws:sns:us-west-2:590375264460:webhook",
		},
	}
	var publisher pubsub.Publisher
	var err error
	snsConfig := gizmoAWS.SNSConfig{
		Topic: cfg.NotificationsPublisherConfig.TopicName,
	}
	if cfg.AWSConfig.Region != "" {
		snsConfig.Region = cfg.AWSConfig.Region
	} else {
		snsConfig.Region = cfg.Region
	}
	publisher, err = gizmoAWS.NewPublisher(snsConfig)
	assert.Nil(t, err)

	var executionID = core.WorkflowExecutionIdentifier{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
	}

	var workflowRequest = &admin.WorkflowExecutionEventRequest{
		Event: &event.WorkflowExecutionEvent{
			Phase: core.WorkflowExecution_SUCCEEDED,
			OutputResult: &event.WorkflowExecutionEvent_OutputUri{
				OutputUri: "somestring",
			},
			ExecutionId: &executionID,
		},
	}
	err = publisher.Publish(context.Background(), "test", workflowRequest)
	assert.Nil(t, err)
}
