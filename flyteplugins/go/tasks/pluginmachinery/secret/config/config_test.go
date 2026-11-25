package config

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	stdconfig "github.com/flyteorg/flyte/v2/flytestdlib/config"
)

// Ensure HostnameReplacements resolves to non-nil empty list.
func TestConfig_DefaultImageBuilder(t *testing.T) {
	assert.Empty(t, DefaultConfig.ImageBuilderConfig.HostnameReplacement)
}

func TestConfig_LoadSimpleJSON(t *testing.T) {
	expectedJSON := `{
		"metrics-prefix": "test-prefix",
		"certDir": "/test/cert/dir",
		"localCert": true,
		"listenPort": 8080,
		"serviceName": "test-service",
		"servicePort": 8081,
		"secretName": "test-secret",
		"secretManagerType": "K8s"
	}`

	var config Config
	err := json.Unmarshal([]byte(expectedJSON), &config)
	assert.Nil(t, err)

	assert.Empty(t, config.ImageBuilderConfig)
}

func TestConfig_ImageBuilderConfig(t *testing.T) {

	t.Run("With verification enabled", func(t *testing.T) {
		expectedJSON := `{
		"metrics-prefix": "test-prefix",
		"certDir": "/test/cert/dir",
		"localCert": true,
		"listenPort": 8080,
		"serviceName": "test-service",
		"servicePort": 8081,
		"secretName": "test-secret",
		"secretManagerType": "K8s",
		"imageBuilderConfig": {
			"hostnameReplacement": {
				"existing": "test.existing.hostname",
				"replacement": "test.replacement.hostname"
			},
			"labelSelector": {
				"matchLabels": {
					"test-key": "test-value"
				}
			}
			}
		}`

		var config Config
		err := json.Unmarshal([]byte(expectedJSON), &config)
		assert.Nil(t, err)

		assert.Equal(t, "test.existing.hostname", config.ImageBuilderConfig.HostnameReplacement.Existing)
		assert.Equal(t, "test.replacement.hostname", config.ImageBuilderConfig.HostnameReplacement.Replacement)
		assert.Equal(t, false, config.ImageBuilderConfig.HostnameReplacement.DisableVerification)
		assert.Equal(t, "test-value", config.ImageBuilderConfig.LabelSelector.MatchLabels["test-key"])
	})

	t.Run("With verification disabled", func(t *testing.T) {
		expectedJSON := `{
		"metrics-prefix": "test-prefix",
		"certDir": "/test/cert/dir",
		"localCert": true,
		"listenPort": 8080,
		"serviceName": "test-service",
		"servicePort": 8081,
		"secretName": "test-secret",
		"secretManagerType": "K8s",
		"imageBuilderConfig": {
			"hostnameReplacement": {
				"existing": "test.existing.hostname",
				"replacement": "test.replacement.hostname",
				"disableVerification": true
			},
			"labelSelector": {
				"matchLabels": {
					"test-key": "test-value"
				}
			}
			}
		}`

		var config Config
		err := json.Unmarshal([]byte(expectedJSON), &config)
		assert.Nil(t, err)

		assert.Equal(t, "test.existing.hostname", config.ImageBuilderConfig.HostnameReplacement.Existing)
		assert.Equal(t, "test.replacement.hostname", config.ImageBuilderConfig.HostnameReplacement.Replacement)
		assert.Equal(t, true, config.ImageBuilderConfig.HostnameReplacement.DisableVerification)
		assert.Equal(t, "test-value", config.ImageBuilderConfig.LabelSelector.MatchLabels["test-key"])
	})

}

func TestGetConfigConcurrentAndReload(t *testing.T) {
	prev := configSection.GetConfig().(*Config)
	t.Cleanup(func() {
		require.NoError(t, configSection.SetConfig(prev))
	})

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				cfg := GetConfig()
				k8s := cfg.EmbeddedSecretManagerConfig.K8sConfig.KubeClientConfig
				if k8s.QPS == 0 || k8s.Burst == 0 || k8s.Timeout.Duration == 0 {
					t.Errorf("expected resolved KubeClientConfig fallback, got %+v", k8s)
					return
				}
			}
		})
	}

	for i := range 100 {
		qps := int32(50 + i%2*50)
		require.NoError(t, configSection.SetConfig(&Config{
			KubeClientConfig: KubeClientConfig{
				QPS:     qps,
				Burst:   25,
				Timeout: stdconfig.Duration{Duration: 15 * time.Second},
			},
		}))
	}

	cancel()
	wg.Wait()

	require.NoError(t, configSection.SetConfig(&Config{
		KubeClientConfig: KubeClientConfig{
			QPS:     77,
			Burst:   33,
			Timeout: stdconfig.Duration{Duration: 12 * time.Second},
		},
	}))
	got := GetConfig().EmbeddedSecretManagerConfig.K8sConfig.KubeClientConfig
	assert.Equal(t, int32(77), got.QPS)
	assert.Equal(t, 33, got.Burst)
	assert.Equal(t, 12*time.Second, got.Timeout.Duration)
}
