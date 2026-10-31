package config

import (
	"testing"

	"github.com/flyteorg/flyte/v2/flytestdlib/serviceclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParentLabelQueriesConfig(t *testing.T) {
	assert.False(t, defaultConfig.ParentLabelQueries)
	flags := (Config{}).GetPFlagSet("actions.")
	value, err := flags.GetBool("actions.parentLabelQueries")
	require.NoError(t, err)
	assert.False(t, value)
	require.NoError(t, flags.Set("actions.parentLabelQueries", "true"))
	value, err = flags.GetBool("actions.parentLabelQueries")
	require.NoError(t, err)
	assert.True(t, value)
	actual := Config{}
	require.NoError(t, decode_Config(map[string]any{"parentLabelQueries": true}, &actual))
	assert.True(t, actual.ParentLabelQueries)
}

func TestServiceAuthConfigDecoding(t *testing.T) {
	actual := Config{}
	require.NoError(t, decode_Config(map[string]any{
		"runService": map[string]any{
			"url": "https://flyte.example.com",
			"auth": map[string]any{
				"type":         "OAuth2ClientCredentials",
				"issuerUrl":    "https://identity.example.com",
				"clientId":     "actions-service",
				"clientSecret": "secret",
			},
		},
	}, &actual))

	assert.Equal(t, "https://flyte.example.com", actual.RunService.URL)
	assert.Equal(t, serviceclient.AuthTypeOAuth2ClientCredentials, actual.RunService.Auth.Type)
	assert.Equal(t, "https://identity.example.com", actual.RunService.Auth.IssuerURL)
	assert.Equal(t, "actions-service", actual.RunService.Auth.ClientID)
}
