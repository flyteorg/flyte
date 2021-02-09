package implementations

import (
	"context"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/lyft/flyteadmin/pkg/async/notifications/interfaces"

	"github.com/NYTimes/gizmo/pubsub"
	"github.com/golang/protobuf/proto"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/prometheus/client_golang/prometheus"
)

type eventPublisherSystemMetrics struct {
	Scope          promutils.Scope
	PublishTotal   prometheus.Counter
	PublishSuccess prometheus.Counter
	PublishError   prometheus.Counter
}

// TODO: Add a counter that encompasses the publisher stats grouped by project and domain.
type EventPublisher struct {
	pub           pubsub.Publisher
	systemMetrics eventPublisherSystemMetrics
	events        sets.String
}

var taskExecutionReq admin.TaskExecutionEventRequest
var nodeExecutionReq admin.NodeExecutionEventRequest
var workflowExecutionReq admin.WorkflowExecutionEventRequest

const (
	Task          = "task"
	Node          = "node"
	Workflow      = "workflow"
	AllTypes      = "all"
	AllTypesShort = "*"
)

var supportedEvents = map[string]string{
	Task:     proto.MessageName(&taskExecutionReq),
	Node:     proto.MessageName(&nodeExecutionReq),
	Workflow: proto.MessageName(&workflowExecutionReq),
}

// The key is the notification type as defined as an enum.
func (p *EventPublisher) Publish(ctx context.Context, notificationType string, msg proto.Message) error {
	p.systemMetrics.PublishTotal.Inc()

	if !p.shouldPublishEvent(notificationType) {
		return nil
	}
	logger.Debugf(ctx, "Publishing the following message [%+v]", msg)

	err := p.pub.Publish(ctx, notificationType, msg)
	if err != nil {
		p.systemMetrics.PublishError.Inc()
		logger.Errorf(ctx, "Failed to publish a message with key [%s] and message [%s] and error: %v", notificationType, msg.String(), err)
	} else {
		p.systemMetrics.PublishSuccess.Inc()
	}
	return err
}

func (p *EventPublisher) shouldPublishEvent(notificationType string) bool {
	return p.events.Has(notificationType)
}

func newEventPublisherSystemMetrics(scope promutils.Scope) eventPublisherSystemMetrics {
	return eventPublisherSystemMetrics{
		Scope:          scope,
		PublishTotal:   scope.MustNewCounter("event_publish_total", "overall count of publish messages"),
		PublishSuccess: scope.MustNewCounter("event_publish_success", "success count of publish messages"),
		PublishError:   scope.MustNewCounter("event_publish_errors", "count of publish errors"),
	}
}

func NewEventsPublisher(pub pubsub.Publisher, scope promutils.Scope, eventTypes []string) interfaces.Publisher {
	eventSet := sets.NewString()

	for _, event := range eventTypes {
		if event == AllTypes || event == AllTypesShort {
			for _, e := range supportedEvents {
				eventSet = eventSet.Insert(e)
			}
			break
		}
		if e, found := supportedEvents[event]; found {
			eventSet = eventSet.Insert(e)
		} else {
			logger.Errorf(context.Background(), "Unsupported event type [%s] in the config")
		}
	}

	return &EventPublisher{
		pub:           pub,
		systemMetrics: newEventPublisherSystemMetrics(scope.NewSubScope("events_publisher")),
		events:        eventSet,
	}
}
