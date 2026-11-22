package app

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApp_MetricsEndpointAndScope(t *testing.T) {
	var capturedScope interface{}
	setupDone := make(chan struct{})

	testApp := &App{
		Name:  "test-metrics-app",
		Short: "Testing the metrics framework plumbing",
		Setup: func(ctx context.Context, sc *SetupContext) error {
			sc.Port = 8099
			capturedScope = sc.Scope
			close(setupDone)
			return nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = testApp.serve(ctx)
	}()

	select {
	case <-setupDone:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for app setup")
	}

	assert.NotNil(t, capturedScope, "SetupContext.Scope should be initialized by the framework before Setup is called")

	resp, err := http.Get("http://localhost:8099/metrics")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected /metrics to return a 200 OK status")
}
