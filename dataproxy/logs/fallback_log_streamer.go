package logs

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/dataproxy/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy"
)

type FallbackLogStreamer struct {
	streamers []LogStreamer
}

var _ LogStreamer = (*FallbackLogStreamer)(nil)

func NewFallbackLogStreamer(streamers ...LogStreamer) *FallbackLogStreamer {
	return &FallbackLogStreamer{streamers: streamers}
}

func (s *FallbackLogStreamer) TailLogs(ctx context.Context, logContext *core.LogContext, stream *connect.ServerStream[dataproxy.TailLogsResponse]) error {
	var err error
	for _, streamer := range s.streamers {
		err = streamer.TailLogs(ctx, logContext, stream)
		if err == nil || connect.CodeOf(err) != connect.CodeNotFound {
			return err
		}
	}
	return err
}

// NewLogStreamerFromConfig returns nil when no streamer is available. k8s is nil
// when the dataproxy has no Kubernetes access.
func NewLogStreamerFromConfig(ctx context.Context, cfg config.LogsConfig, k8s LogStreamer) (LogStreamer, error) {
	streamers := make([]LogStreamer, 0, len(cfg.Streamers))
	seen := make(map[string]bool, len(cfg.Streamers))
	for _, name := range cfg.Streamers {
		if seen[name] {
			return nil, fmt.Errorf("log streamer %q is listed more than once", name)
		}
		seen[name] = true

		switch name {
		case config.LogStreamerK8s:
			if k8s == nil {
				logger.Warnf(ctx, "log streamer %q skipped: dataproxy has no Kubernetes access", name)
				continue
			}
			streamers = append(streamers, k8s)
		case config.LogStreamerLoki:
			loki, err := NewLokiLogStreamer(cfg.Loki)
			if err != nil {
				return nil, fmt.Errorf("configuring log streamer %q: %w", name, err)
			}
			streamers = append(streamers, loki)
		default:
			return nil, fmt.Errorf("unknown log streamer %q (supported: %q, %q)", name, config.LogStreamerK8s, config.LogStreamerLoki)
		}
	}

	switch len(streamers) {
	case 0:
		return nil, nil
	case 1:
		return streamers[0], nil
	default:
		return NewFallbackLogStreamer(streamers...), nil
	}
}
