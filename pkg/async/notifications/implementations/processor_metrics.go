package implementations

import (
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/prometheus/client_golang/prometheus"
)

type processorSystemMetrics struct {
	Scope                 promutils.Scope
	MessageTotal          prometheus.Counter
	MessageDoneError      prometheus.Counter
	MessageDecodingError  prometheus.Counter
	MessageDataError      prometheus.Counter
	MessageProcessorError prometheus.Counter
	MessageSuccess        prometheus.Counter
	ChannelClosedError    prometheus.Counter
	StopError             prometheus.Counter
}

func newProcessorSystemMetrics(scope promutils.Scope) processorSystemMetrics {
	return processorSystemMetrics{
		Scope:                 scope,
		MessageTotal:          scope.MustNewCounter("message_total", "overall count of messages processed"),
		MessageDecodingError:  scope.MustNewCounter("message_decoding_error", "count of messages with decoding errors"),
		MessageDataError:      scope.MustNewCounter("message_data_error", "count of message data processing errors experience when preparing the message to be notified."),
		MessageDoneError:      scope.MustNewCounter("message_done_error", "count of message errors when marking it as done with underlying processor"),
		MessageProcessorError: scope.MustNewCounter("message_processing_error", "count of errors when interacting with notification processor"),
		MessageSuccess:        scope.MustNewCounter("message_ok", "count of messages successfully processed by underlying notification mechanism"),
		ChannelClosedError:    scope.MustNewCounter("channel_closed_error", "count of channel closing errors"),
		StopError:             scope.MustNewCounter("stop_error", "count of errors in Stop() method"),
	}
}
