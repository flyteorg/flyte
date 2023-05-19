package implementations

import (
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/prometheus/client_golang/prometheus"
)

type webhookMetrics struct {
	Scope       promutils.Scope
	SendSuccess prometheus.Counter
	SendError   prometheus.Counter
	SendTotal   prometheus.Counter
}

func newWebhookMetrics(scope promutils.Scope) webhookMetrics {
	return webhookMetrics{
		Scope:       scope,
		SendSuccess: scope.MustNewCounter("send_success", "Number of successful emails sent via Emailer."),
		SendError:   scope.MustNewCounter("send_error", "Number of errors when sending email via Emailer"),
		SendTotal:   scope.MustNewCounter("send_total", "Total number of emails attempted to be sent"),
	}
}
