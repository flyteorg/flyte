package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
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
