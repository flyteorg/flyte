package config

import (
	"testing"

	"github.com/flyteorg/flyte/v2/flytestdlib/serviceclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunServiceAuthConfigDecoding(t *testing.T) {
	actual := DataProxyConfig{}
	require.NoError(t, decode_DataProxyConfig(map[string]any{
		"runService": map[string]any{
			"url": "https://flyte.example.com",
			"auth": map[string]any{
				"type":         "OAuth2ClientCredentials",
				"issuerUrl":    "https://identity.example.com",
				"clientId":     "dataproxy-service",
				"clientSecret": "secret",
			},
		},
	}, &actual))

	assert.Equal(t, "https://flyte.example.com", actual.RunService.URL)
	assert.Equal(t, serviceclient.AuthTypeOAuth2ClientCredentials, actual.RunService.Auth.Type)
	assert.Equal(t, "https://identity.example.com", actual.RunService.Auth.IssuerURL)
	assert.Equal(t, "dataproxy-service", actual.RunService.Auth.ClientID)
}
