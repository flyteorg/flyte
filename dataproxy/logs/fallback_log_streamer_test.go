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

func notFound() error {
	return connect.NewError(connect.CodeNotFound, errors.New("pod gone"))
}

func TestFallbackLogStreamer_NotFoundTriesNextInOrder(t *testing.T) {
	first := &fakeLogStreamer{err: notFound()}
	second := &fakeLogStreamer{err: notFound()}
	third := &fakeLogStreamer{}
	fourth := &fakeLogStreamer{}
	streamer := NewFallbackLogStreamer(first, second, third, fourth)

	err := streamer.TailLogs(context.Background(), &core.LogContext{}, nil)

	require.NoError(t, err)
	assert.Equal(t, 1, first.calls)
	assert.Equal(t, 1, second.calls)
	assert.Equal(t, 1, third.calls)
	assert.Equal(t, 0, fourth.calls)
}

func TestFallbackLogStreamer_OtherErrorStops(t *testing.T) {
	wantErr := errors.New("boom")
	first := &fakeLogStreamer{err: wantErr}
	second := &fakeLogStreamer{}
	streamer := NewFallbackLogStreamer(first, second)

	err := streamer.TailLogs(context.Background(), &core.LogContext{}, nil)

	assert.ErrorIs(t, err, wantErr)
	assert.Equal(t, 1, first.calls)
	assert.Equal(t, 0, second.calls)
}

func TestFallbackLogStreamer_AllNotFound(t *testing.T) {
	first := &fakeLogStreamer{err: notFound()}
	second := &fakeLogStreamer{err: notFound()}
	streamer := NewFallbackLogStreamer(first, second)

	err := streamer.TailLogs(context.Background(), &core.LogContext{}, nil)

	require.Error(t, err)
	assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
	assert.Equal(t, 1, first.calls)
	assert.Equal(t, 1, second.calls)
}

func TestNewLogStreamerFromConfig(t *testing.T) {
	ctx := context.Background()
	k8s := &fakeLogStreamer{}
	newCfg := func(streamers ...string) config.LogsConfig {
		cfg := config.LogsConfig{Streamers: streamers, Loki: config.DefaultLokiConfig()}
		cfg.Loki.URL = "http://loki:3100"
		return cfg
	}

	t.Run("k8s alone is returned as-is", func(t *testing.T) {
		got, err := NewLogStreamerFromConfig(ctx, newCfg(config.LogStreamerK8s), k8s)
		require.NoError(t, err)
		assert.Same(t, k8s, got)
	})

	t.Run("loki alone ignores k8s", func(t *testing.T) {
		got, err := NewLogStreamerFromConfig(ctx, newCfg(config.LogStreamerLoki), k8s)
		require.NoError(t, err)
		assert.IsType(t, &LokiLogStreamer{}, got)
	})

	t.Run("several streamers keep the configured order", func(t *testing.T) {
		got, err := NewLogStreamerFromConfig(ctx, newCfg(config.LogStreamerK8s, config.LogStreamerLoki), k8s)
		require.NoError(t, err)
		fb, ok := got.(*FallbackLogStreamer)
		require.True(t, ok, "expected *FallbackLogStreamer, got %T", got)
		require.Len(t, fb.streamers, 2)
		assert.Same(t, k8s, fb.streamers[0])
		assert.IsType(t, &LokiLogStreamer{}, fb.streamers[1])

		got, err = NewLogStreamerFromConfig(ctx, newCfg(config.LogStreamerLoki, config.LogStreamerK8s), k8s)
		require.NoError(t, err)
		fb, ok = got.(*FallbackLogStreamer)
		require.True(t, ok, "expected *FallbackLogStreamer, got %T", got)
		require.Len(t, fb.streamers, 2)
		assert.IsType(t, &LokiLogStreamer{}, fb.streamers[0])
		assert.Same(t, k8s, fb.streamers[1])
	})

	t.Run("k8s is skipped without Kubernetes access", func(t *testing.T) {
		got, err := NewLogStreamerFromConfig(ctx, newCfg(config.LogStreamerK8s, config.LogStreamerLoki), nil)
		require.NoError(t, err)
		assert.IsType(t, &LokiLogStreamer{}, got)
	})

	t.Run("loki is not validated unless listed", func(t *testing.T) {
		got, err := NewLogStreamerFromConfig(ctx, config.LogsConfig{Streamers: []string{config.LogStreamerK8s}}, k8s)
		require.NoError(t, err)
		assert.Same(t, k8s, got)
	})

	t.Run("invalid loki config errors", func(t *testing.T) {
		_, err := NewLogStreamerFromConfig(ctx, config.LogsConfig{Streamers: []string{config.LogStreamerLoki}}, k8s)
		assert.Error(t, err)
	})

	t.Run("unknown streamer errors", func(t *testing.T) {
		_, err := NewLogStreamerFromConfig(ctx, newCfg("elastic"), k8s)
		assert.ErrorContains(t, err, "elastic")
	})

	t.Run("duplicate streamer errors", func(t *testing.T) {
		_, err := NewLogStreamerFromConfig(ctx, newCfg(config.LogStreamerLoki, config.LogStreamerLoki), k8s)
		assert.ErrorContains(t, err, "more than once")
	})

	t.Run("nothing available returns nil", func(t *testing.T) {
		got, err := NewLogStreamerFromConfig(ctx, newCfg(), k8s)
		require.NoError(t, err)
		assert.Nil(t, got)

		got, err = NewLogStreamerFromConfig(ctx, newCfg(config.LogStreamerK8s), nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
