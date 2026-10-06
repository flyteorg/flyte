package redis

import (
	"context"
	"errors"
	"testing"

	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testSecretManager struct {
	password string
	err      error
}

func (s testSecretManager) Get(context.Context, string) (string, error) {
	return s.password, s.err
}

func TestGetUniversalOptionsPasswordSecret(t *testing.T) {
	cfg := Config{Password: "original", PasswordSecretName: "redis-password", Protocol: 2}
	opts, err := cfg.GetUniversalOptions(context.Background(), WithSecretManager(testSecretManager{password: "resolved"}))
	require.NoError(t, err)
	assert.Equal(t, "resolved", opts.Password)
	assert.Equal(t, 2, opts.Protocol)
	assert.Equal(t, "original", cfg.Password)

	_, err = cfg.GetUniversalOptions(context.Background())
	require.ErrorContains(t, err, "requires a secret manager")

	secretErr := errors.New("secret unavailable")
	_, err = cfg.GetUniversalOptions(context.Background(), WithSecretManager(testSecretManager{err: secretErr}))
	require.ErrorIs(t, err, secretErr)
}

func TestNewClientModes(t *testing.T) {
	for _, tc := range []struct {
		name    string
		cfg     Config
		cluster bool
	}{
		{name: "legacy standalone", cfg: Config{Addr: "localhost:6379"}},
		{name: "standalone unix", cfg: Config{Addr: "/tmp/redis.sock", Network: "unix"}},
		{name: "cluster seeds", cfg: Config{Addrs: []string{"localhost:6379", "localhost:6380"}}, cluster: true},
		{name: "single cluster endpoint", cfg: Config{Addr: "localhost:6379", IsClusterMode: true}, cluster: true},
		{name: "sentinel", cfg: Config{Addrs: []string{"localhost:26379", "localhost:26380"}, MasterName: "primary"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client, err := tc.cfg.NewClient(context.Background())
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, client.Close()) })
			if tc.cluster {
				require.IsType(t, &goredis.ClusterClient{}, client)
			} else {
				require.IsType(t, &goredis.Client{}, client)
				if tc.cfg.MasterName == "" && tc.cfg.Network != "" {
					assert.Equal(t, tc.cfg.Network, client.(*goredis.Client).Options().Network)
				}
			}
		})
	}
}

func TestUniversalOptions(t *testing.T) {
	cfg := Config{
		Addr: "ignored:6379", Addrs: []string{"sentinel:26379"}, MasterName: "primary",
		Username: "redis-user", Password: "redis-password", DB: 2,
		SentinelUsername: "sentinel-user", SentinelPasswordSecretName: "sentinel-secret",
		UseTLS: true, Protocol: 2, PoolSize: 12,
	}
	opts, err := cfg.GetUniversalOptions(context.Background(), WithSecretManager(testSecretManager{password: "sentinel-password"}))
	require.NoError(t, err)
	assert.Equal(t, cfg.Addrs, opts.Addrs)
	failover := opts.Failover()
	assert.Equal(t, "primary", failover.MasterName)
	assert.Equal(t, "sentinel-user", failover.SentinelUsername)
	assert.Equal(t, "sentinel-password", failover.SentinelPassword)
	assert.Equal(t, "redis-password", failover.Password)
	assert.Equal(t, 2, failover.DB)
	assert.Equal(t, 2, failover.Protocol)
	assert.Equal(t, 12, failover.PoolSize)
	require.NotNil(t, failover.TLSConfig)
	assert.Empty(t, cfg.SentinelPassword)

	_, err = cfg.GetUniversalOptions(context.Background())
	require.ErrorContains(t, err, "requires a secret manager")
	secretErr := errors.New("unavailable")
	_, err = cfg.GetUniversalOptions(context.Background(), WithSecretManager(testSecretManager{err: secretErr}))
	require.ErrorIs(t, err, secretErr)

	for _, cfg := range []Config{
		{IsClusterMode: true, DB: 1},
		{Addrs: []string{"one:6379", "two:6379"}, DB: 1},
		{IsClusterMode: true, Network: "unix"},
	} {
		_, err := cfg.NewClient(context.Background())
		require.Error(t, err)
	}
}

func TestOptionalSecretManager(t *testing.T) {
	cfg := Config{Addr: "localhost:6379", Password: "direct-password"}
	opts, err := cfg.GetUniversalOptions(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "direct-password", opts.Password)

	opts, err = cfg.GetUniversalOptions(context.Background(), WithSecretManager(nil))
	require.NoError(t, err)
	assert.Equal(t, "direct-password", opts.Password)

	cfg.PasswordSecretName = "redis-password"
	_, err = cfg.NewClient(context.Background())
	require.ErrorContains(t, err, "requires a secret manager")
	client, err := cfg.NewClient(context.Background(), WithSecretManager(testSecretManager{password: "resolved"}))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	assert.Equal(t, "resolved", client.(*goredis.Client).Options().Password)
}
