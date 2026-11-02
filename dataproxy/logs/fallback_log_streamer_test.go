package logs

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/dataproxy/config"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy"
)

type fakeLogStreamer struct {
	err   error
	calls int
}

func (f *fakeLogStreamer) TailLogs(_ context.Context, _ *core.LogContext, _ *connect.ServerStream[dataproxy.TailLogsResponse]) error {
	f.calls++
	return f.err
}

func TestFallbackLogStreamer_LiveSucceeds(t *testing.T) {
	live := &fakeLogStreamer{}
	persisted := &fakeLogStreamer{}
	streamer := NewFallbackLogStreamer(live, persisted)

	err := streamer.TailLogs(context.Background(), &core.LogContext{}, nil)

	require.NoError(t, err)
	assert.Equal(t, 1, live.calls)
	assert.Equal(t, 0, persisted.calls)
}

func TestFallbackLogStreamer_LiveNotFound(t *testing.T) {
	live := &fakeLogStreamer{err: connect.NewError(connect.CodeNotFound, errors.New("pod gone"))}
	persisted := &fakeLogStreamer{}
	streamer := NewFallbackLogStreamer(live, persisted)

	err := streamer.TailLogs(context.Background(), &core.LogContext{}, nil)

	require.NoError(t, err)
	assert.Equal(t, 1, live.calls)
	assert.Equal(t, 1, persisted.calls)
}

func TestFallbackLogStreamer_OtherError(t *testing.T) {
	wantErr := errors.New("boom")
	live := &fakeLogStreamer{err: wantErr}
	persisted := &fakeLogStreamer{}
	streamer := NewFallbackLogStreamer(live, persisted)

	err := streamer.TailLogs(context.Background(), &core.LogContext{}, nil)

	assert.ErrorIs(t, err, wantErr)
	assert.Equal(t, 1, live.calls)
	assert.Equal(t, 0, persisted.calls)
}

func TestFallbackLogStreamer_NoPersistedBackend(t *testing.T) {
	live := &fakeLogStreamer{err: connect.NewError(connect.CodeNotFound, errors.New("pod gone"))}
	streamer := NewFallbackLogStreamer(live, nil)

	err := streamer.TailLogs(context.Background(), &core.LogContext{}, nil)

	require.Error(t, err)
	assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
	assert.Equal(t, 1, live.calls)
}

func TestNewLogStreamerFromConfig(t *testing.T) {
	live := &fakeLogStreamer{}

	t.Run("no persisted backend returns live as-is", func(t *testing.T) {
		got, err := NewLogStreamerFromConfig(config.LogsConfig{}, live)
		require.NoError(t, err)
		assert.Same(t, live, got)
	})

	t.Run("loki wraps live in a fallback", func(t *testing.T) {
		cfg := config.LogsConfig{Persisted: config.PersistedLogsConfig{Type: config.PersistedLogsTypeLoki, Loki: config.DefaultLokiConfig()}}
		cfg.Persisted.Loki.URL = "http://loki:3100"
		got, err := NewLogStreamerFromConfig(cfg, live)
		require.NoError(t, err)
		fb, ok := got.(*FallbackLogStreamer)
		require.True(t, ok, "expected *FallbackLogStreamer, got %T", got)
		assert.Same(t, live, fb.live)
		assert.IsType(t, &LokiLogStreamer{}, fb.persisted)
	})

	t.Run("loki without live uses loki alone", func(t *testing.T) {
		cfg := config.LogsConfig{Persisted: config.PersistedLogsConfig{Type: config.PersistedLogsTypeLoki, Loki: config.DefaultLokiConfig()}}
		cfg.Persisted.Loki.URL = "http://loki:3100"
		got, err := NewLogStreamerFromConfig(cfg, nil)
		require.NoError(t, err)
		assert.IsType(t, &LokiLogStreamer{}, got)
	})

	t.Run("invalid loki config errors", func(t *testing.T) {
		cfg := config.LogsConfig{Persisted: config.PersistedLogsConfig{Type: config.PersistedLogsTypeLoki}}
		_, err := NewLogStreamerFromConfig(cfg, live)
		assert.Error(t, err)
	})

	t.Run("unknown type errors", func(t *testing.T) {
		cfg := config.LogsConfig{Persisted: config.PersistedLogsConfig{Type: "elastic"}}
		_, err := NewLogStreamerFromConfig(cfg, live)
		assert.Error(t, err)
	})

	t.Run("nothing configured returns nil", func(t *testing.T) {
		got, err := NewLogStreamerFromConfig(config.LogsConfig{}, nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
