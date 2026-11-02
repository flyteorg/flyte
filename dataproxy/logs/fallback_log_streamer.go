package logs

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/dataproxy/config"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy"
)

// FallbackLogStreamer uses persisted when live reports the pod is gone.
type FallbackLogStreamer struct {
	live      LogStreamer
	persisted LogStreamer
}

var _ LogStreamer = (*FallbackLogStreamer)(nil)

func NewFallbackLogStreamer(live, persisted LogStreamer) *FallbackLogStreamer {
	return &FallbackLogStreamer{live: live, persisted: persisted}
}

func (s *FallbackLogStreamer) TailLogs(ctx context.Context, logContext *core.LogContext, stream *connect.ServerStream[dataproxy.TailLogsResponse]) error {
	err := s.live.TailLogs(ctx, logContext, stream)

	if err == nil {
		return nil
	}

	if connect.CodeOf(err) != connect.CodeNotFound || s.persisted == nil {
		return err
	}

	return s.persisted.TailLogs(ctx, logContext, stream)
}

// NewLogStreamerFromConfig returns nil when neither a live nor a persisted
// backend is available.
func NewLogStreamerFromConfig(cfg config.LogsConfig, live LogStreamer) (LogStreamer, error) {
	var persisted LogStreamer
	switch cfg.Persisted.Type {
	case config.PersistedLogsTypeNone:
	case config.PersistedLogsTypeLoki:
		loki, err := NewLokiLogStreamer(cfg.Persisted.Loki)
		if err != nil {
			return nil, fmt.Errorf("configuring persisted log backend: %w", err)
		}
		persisted = loki
	default:
		return nil, fmt.Errorf("unknown persisted log backend type %q (supported: %q)", cfg.Persisted.Type, config.PersistedLogsTypeLoki)
	}

	switch {
	case live == nil && persisted == nil:
		return nil, nil
	case persisted == nil:
		return live, nil
	case live == nil:
		return persisted, nil
	default:
		return NewFallbackLogStreamer(live, persisted), nil
	}
}
