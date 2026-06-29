package storage

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/v2/flytestdlib/config"
)

func Test_createHTTPClient(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		client := createHTTPClient(HTTPClientConfig{})

		transport, casted := client.Transport.(*http.Transport)
		assert.True(t, casted)
		defaultTransport := http.DefaultTransport.(*http.Transport)
		assert.NotSame(t, defaultTransport, transport)
		assert.Equal(t, defaultTransport.MaxIdleConns, transport.MaxIdleConns)
		assert.Equal(t, defaultTransport.MaxIdleConnsPerHost, transport.MaxIdleConnsPerHost)
		assert.Equal(t, defaultTransport.MaxConnsPerHost, transport.MaxConnsPerHost)
		assert.Equal(t, defaultTransport.IdleConnTimeout, transport.IdleConnTimeout)
	})

	t.Run("Transport settings", func(t *testing.T) {
		client := createHTTPClient(HTTPClientConfig{
			MaxIdleConns:        512,
			MaxIdleConnsPerHost: 256,
			MaxConnsPerHost:     128,
			IdleConnTimeout:     config.Duration{Duration: 30 * time.Second},
		})

		transport, casted := client.Transport.(*http.Transport)
		assert.True(t, casted)
		assert.Equal(t, 512, transport.MaxIdleConns)
		assert.Equal(t, 256, transport.MaxIdleConnsPerHost)
		assert.Equal(t, 128, transport.MaxConnsPerHost)
		assert.Equal(t, 30*time.Second, transport.IdleConnTimeout)
	})

	t.Run("Some headers", func(t *testing.T) {
		m := map[string][]string{
			"Header1": {"val1", "val2"},
		}

		client := createHTTPClient(HTTPClientConfig{
			Headers:             m,
			MaxIdleConnsPerHost: 256,
		})

		assert.NotNil(t, client.Transport)
		proxyTransport, casted := client.Transport.(*proxyTransport)
		assert.True(t, casted)
		assert.Equal(t, m, proxyTransport.defaultHeaders)
		transport, casted := proxyTransport.RoundTripper.(*http.Transport)
		assert.True(t, casted)
		assert.Equal(t, 256, transport.MaxIdleConnsPerHost)
	})

	t.Run("Set empty timeout", func(t *testing.T) {
		client := createHTTPClient(HTTPClientConfig{
			Timeout: config.Duration{},
		})

		assert.Zero(t, client.Timeout)
	})

	t.Run("Set timeout", func(t *testing.T) {
		client := createHTTPClient(HTTPClientConfig{
			Timeout: config.Duration{Duration: 2 * time.Second},
		})

		assert.Equal(t, 2*time.Second, client.Timeout)
	})
}

func Test_applyDefaultHeaders(t *testing.T) {
	input := map[string][]string{
		"Header1": {"val1", "val2"},
	}

	r := &http.Request{}
	applyDefaultHeaders(r, input)

	assert.Equal(t, http.Header(input), r.Header)
}
