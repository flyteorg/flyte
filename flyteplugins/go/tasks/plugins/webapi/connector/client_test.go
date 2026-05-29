package connector

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/flytestdlib/config"
)

func TestInitializeClients(t *testing.T) {
	cfg := defaultConfig
	cfg.ConnectorDeployments = map[string]*Deployment{
		"x": {
			Endpoint: "x",
		},
		"y": {
			Endpoint: "y",
		},
	}
	ctx := context.Background()
	err := SetConfig(&cfg)
	assert.NoError(t, err)
	cs := getConnectorClientSets(ctx)
	_, ok := cs.asyncConnectorClients["y"]
	assert.True(t, ok)
	_, ok = cs.asyncConnectorClients["x"]
	assert.True(t, ok)
}

func TestDefaultGRPCServiceConfig(t *testing.T) {
	// Must be valid JSON — grpc.WithDefaultServiceConfig silently ignores a
	// malformed config, so a typo here would disable LB/retry without any error.
	assert.True(t, json.Valid([]byte(DefaultGRPCServiceConfig)), "DefaultGRPCServiceConfig must be valid JSON")

	var parsed struct {
		LoadBalancingConfig []map[string]any `json:"loadBalancingConfig"`
		MethodConfig        []struct {
			Name []struct {
				Service string `json:"service"`
			} `json:"name"`
			RetryPolicy struct {
				MaxAttempts          int      `json:"maxAttempts"`
				RetryableStatusCodes []string `json:"retryableStatusCodes"`
			} `json:"retryPolicy"`
		} `json:"methodConfig"`
		RetryThrottling map[string]any `json:"retryThrottling"`
	}
	require.NoError(t, json.Unmarshal([]byte(DefaultGRPCServiceConfig), &parsed))

	require.Len(t, parsed.LoadBalancingConfig, 1)
	_, hasRoundRobin := parsed.LoadBalancingConfig[0]["round_robin"]
	assert.True(t, hasRoundRobin, "expected round_robin load balancing")

	require.Len(t, parsed.MethodConfig, 1)
	mc := parsed.MethodConfig[0]
	require.Len(t, mc.Name, 1)
	assert.Equal(t, "flyteidl2.connector.AsyncConnectorService", mc.Name[0].Service)
	assert.Greater(t, mc.RetryPolicy.MaxAttempts, 1)
	assert.Equal(t, []string{"UNAVAILABLE"}, mc.RetryPolicy.RetryableStatusCodes)
	assert.NotEmpty(t, parsed.RetryThrottling, "expected retryThrottling to bound retry storms")
}

func TestGetGrpcConnection(t *testing.T) {
	ctx := context.Background()

	// Empty DefaultServiceConfig must fall back to DefaultGRPCServiceConfig
	// (round-robin + retry) rather than no service config at all.
	conn, err := getGrpcConnection(ctx, &Deployment{Endpoint: "x", Insecure: true})
	require.NoError(t, err)
	require.NotNil(t, conn)
	assert.NoError(t, conn.Close())

	// A deployment-specific DefaultServiceConfig still takes precedence.
	custom := `{"loadBalancingConfig": [{"round_robin":{}}]}`
	conn, err = getGrpcConnection(ctx, &Deployment{Endpoint: "y", Insecure: true, DefaultServiceConfig: custom})
	require.NoError(t, err)
	require.NotNil(t, conn)
	assert.NoError(t, conn.Close())

	// Keepalive is opt-in: a deployment may enable it (gateway-fronted connectors).
	conn, err = getGrpcConnection(ctx, &Deployment{
		Endpoint: "z", Insecure: true,
		Keepalive: &KeepaliveConfig{
			Time:                config.Duration{Duration: 30 * time.Second},
			Timeout:             config.Duration{Duration: 10 * time.Second},
			PermitWithoutStream: true,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, conn)
	assert.NoError(t, conn.Close())
}

// ensure the constant referenced from config.go and the LB choice stay in sync
func TestDefaultGRPCServiceConfigMentionsRoundRobin(t *testing.T) {
	assert.True(t, strings.Contains(DefaultGRPCServiceConfig, "round_robin"))
}
