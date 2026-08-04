package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

// startTestServer brings the invalidation server up on a free port and returns its base URL.
func startTestServer(t *testing.T, mutator *secret.SecretsPodMutator) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	require.NoError(t, listener.Close())

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- StartCacheInvalidationServer(ctx, port, mutator) }()
	t.Cleanup(func() {
		cancel()
		select {
		case err := <-done:
			assert.NoError(t, err)
		case <-time.After(5 * time.Second):
			t.Error("cache invalidation server did not shut down")
		}
	})

	url := fmt.Sprintf("http://127.0.0.1:%d", port)
	require.Eventually(t, func() bool {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 100*time.Millisecond)
		if err != nil {
			return false
		}
		return conn.Close() == nil
	}, 5*time.Second, 20*time.Millisecond, "server never started listening")

	return url
}

func post(t *testing.T, url string, body []byte) *http.Response {
	t.Helper()
	resp, err := http.Post(url+InvalidateSecretPath, "application/json", bytes.NewReader(body)) // nolint: noctx
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

func TestCacheInvalidationServer(t *testing.T) {
	// A mutator with only the global injector: it caches nothing, so InvalidateCache is a no-op
	// and this exercises the server's routing/decoding rather than the cache itself.
	mutator, err := secret.NewSecretsMutator(context.Background(),
		&config.Config{SecretManagerTypes: []config.SecretManagerType{config.SecretManagerTypeGlobal}},
		"flyte", promutils.NewTestScope())
	require.NoError(t, err)
	url := startTestServer(t, mutator)

	t.Run("accepts a valid request", func(t *testing.T) {
		body, err := json.Marshal(InvalidateRequest{
			Org: "flyte", Domain: "development", Project: "flytesnacks", Name: "my-secret",
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, post(t, url, body).StatusCode)
	})

	// Name is the one field that must be present: without it the request identifies no secret
	// and would clear whatever happens to be cached at the bare scope keys.
	t.Run("rejects a missing name", func(t *testing.T) {
		body, err := json.Marshal(InvalidateRequest{Org: "flyte", Domain: "development"})
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, post(t, url, body).StatusCode)
	})

	t.Run("rejects a malformed body", func(t *testing.T) {
		assert.Equal(t, http.StatusBadRequest, post(t, url, []byte("{not json")).StatusCode)
	})

	t.Run("rejects a GET", func(t *testing.T) {
		resp, err := http.Get(url + InvalidateSecretPath) // nolint: noctx
		require.NoError(t, err)
		defer resp.Body.Close() // nolint: errcheck
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})
}
