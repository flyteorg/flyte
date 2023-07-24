package implementations

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/async/notifications/interfaces"

	"github.com/NYTimes/gizmo/pubsub"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/golang/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
)

type publisherSystemMetrics struct {
	Scope        promutils.Scope
	PublishTotal prometheus.Counter
	PublishError prometheus.Counter
}

// TODO: Add a counter that encompasses the publisher stats grouped by project and domain.
type Publisher struct {
	pub           pubsub.Publisher
	systemMetrics publisherSystemMetrics
}

// The key is the notification type as defined as an enum.
func (p *Publisher) Publish(ctx context.Context, notificationType string, msg proto.Message) error {
	p.systemMetrics.PublishTotal.Inc()
	logger.Debugf(ctx, "Publishing the following message [%s]", msg.String())
	err := p.pub.Publish(ctx, notificationType, msg)
	if err != nil {
		p.systemMetrics.PublishError.Inc()
		logger.Errorf(ctx, "Failed to publish a message with key [%s] and message [%s] and error: %v", notificationType, msg.String(), err)
	}
	return err
}

func newPublisherSystemMetrics(scope promutils.Scope) publisherSystemMetrics {
	return publisherSystemMetrics{
		Scope:        scope,
		PublishTotal: scope.MustNewCounter("publish_total", "overall count of publish messages"),
		PublishError: scope.MustNewCounter("publish_errors", "count of publish errors"),
	}
}

func NewPublisher(pub pubsub.Publisher, scope promutils.Scope) interfaces.Publisher {
	return &Publisher{
		pub:           pub,
		systemMetrics: newPublisherSystemMetrics(scope.NewSubScope("publisher")),
	}
}
