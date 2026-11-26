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

func TestConfig_EnvVarMountMode(t *testing.T) {
	t.Run("defaults to FlyteManagedReference", func(t *testing.T) {
		assert.Equal(t, EnvVarMountModeFlyteManagedReference, DefaultConfig.EmbeddedSecretManagerConfig.EnvVarMountMode)
	})

	t.Run("loads from JSON", func(t *testing.T) {
		for _, mode := range []EnvVarMountMode{EnvVarMountModeReference, EnvVarMountModeInline, EnvVarMountModeFlyteManagedReference} {
			var config Config
			err := json.Unmarshal([]byte(`{"embeddedSecretManagerConfig": {"envVarMountMode": "`+string(mode)+`"}}`), &config)
			assert.NoError(t, err)
			assert.Equal(t, mode, config.EmbeddedSecretManagerConfig.EnvVarMountMode)
		}
	})

	t.Run("loads case-insensitively", func(t *testing.T) {
		for input, want := range map[string]EnvVarMountMode{
			"inline":                EnvVarMountModeInline,
			"INLINE":                EnvVarMountModeInline,
			"reference":             EnvVarMountModeReference,
			"REFERENCE":             EnvVarMountModeReference,
			"flytemanagedreference": EnvVarMountModeFlyteManagedReference,
		} {
			var config Config
			err := json.Unmarshal([]byte(`{"embeddedSecretManagerConfig": {"envVarMountMode": "`+input+`"}}`), &config)
			assert.NoError(t, err, input)
			assert.Equal(t, want, config.EmbeddedSecretManagerConfig.EnvVarMountMode, input)
		}
	})

	t.Run("loads an empty value as FlyteManagedReference", func(t *testing.T) {
		var config Config
		err := json.Unmarshal([]byte(`{"embeddedSecretManagerConfig": {"envVarMountMode": ""}}`), &config)
		assert.NoError(t, err)
		assert.Equal(t, EnvVarMountModeFlyteManagedReference, config.EmbeddedSecretManagerConfig.EnvVarMountMode)
	})

	t.Run("loads invalid values as Inline", func(t *testing.T) {
		for _, input := range []string{`"inlined"`, `1`} {
			var config Config
			err := json.Unmarshal([]byte(`{"embeddedSecretManagerConfig": {"envVarMountMode": `+input+`}}`), &config)
			assert.NoError(t, err, input)
			assert.Equal(t, EnvVarMountModeInline, config.EmbeddedSecretManagerConfig.EnvVarMountMode, input)
		}
	})

	t.Run("ParseEnvVarMountMode", func(t *testing.T) {
		for input, want := range map[string]EnvVarMountMode{
			"":                      EnvVarMountModeFlyteManagedReference,
			"Reference":             EnvVarMountModeReference,
			"reference":             EnvVarMountModeReference,
			"Inline":                EnvVarMountModeInline,
			"INLINE":                EnvVarMountModeInline,
			"FlyteManagedReference": EnvVarMountModeFlyteManagedReference,
		} {
			got, err := ParseEnvVarMountMode(input)
			assert.NoError(t, err, input)
			assert.Equal(t, want, got, input)
		}
	})

	t.Run("ParseEnvVarMountMode falls back to Inline", func(t *testing.T) {
		for _, input := range []string{"inlined", "1"} {
			got, err := ParseEnvVarMountMode(input)
			assert.ErrorContains(t, err, "invalid envVarMountMode", input)
			assert.Equal(t, EnvVarMountModeInline, got, input)
		}
	})
}
