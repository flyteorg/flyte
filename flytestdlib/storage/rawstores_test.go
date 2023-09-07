package storage

import (
	"net/http"
	"testing"
	"time"

	"github.com/flyteorg/flytestdlib/config"

	"github.com/stretchr/testify/assert"
)

func Test_createHTTPClient(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		client := createHTTPClient(HTTPClientConfig{})
		assert.Nil(t, client.Transport)
	})

	t.Run("Some headers", func(t *testing.T) {
		m := map[string][]string{
			"Header1": {"val1", "val2"},
		}

		client := createHTTPClient(HTTPClientConfig{
			Headers: m,
		})

		assert.NotNil(t, client.Transport)
		proxyTransport, casted := client.Transport.(*proxyTransport)
		assert.True(t, casted)
		assert.Equal(t, m, proxyTransport.defaultHeaders)
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
