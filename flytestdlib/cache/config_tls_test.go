package cache

import (
	"context"
	"crypto/tls"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisOptions_GetOptions_TLS(t *testing.T) {
	ctx := context.Background()

	t.Run("UseTLS negotiates TLS with the system pool", func(t *testing.T) {
		opts, err := (&RedisOptions{Addr: "host:6379", UseTLS: true}).GetOptions(ctx, nil)
		require.NoError(t, err)
		require.NotNil(t, opts.TLSConfig)
		assert.False(t, opts.TLSConfig.InsecureSkipVerify)
		assert.Equal(t, uint16(tls.VersionTLS12), opts.TLSConfig.MinVersion)
	})

	t.Run("TLSInsecureSkipVerify is honored", func(t *testing.T) {
		opts, err := (&RedisOptions{Addr: "host:6379", UseTLS: true, TLSInsecureSkipVerify: true}).GetOptions(ctx, nil)
		require.NoError(t, err)
		require.NotNil(t, opts.TLSConfig)
		assert.True(t, opts.TLSConfig.InsecureSkipVerify)
	})

	t.Run("no TLS by default", func(t *testing.T) {
		opts, err := (&RedisOptions{Addr: "host:6379"}).GetOptions(ctx, nil)
		require.NoError(t, err)
		assert.Nil(t, opts.TLSConfig)
	})

	t.Run("explicit TLSConfig takes precedence over UseTLS", func(t *testing.T) {
		explicit := &tls.Config{MinVersion: tls.VersionTLS13}
		opts, err := (&RedisOptions{Addr: "host:6379", UseTLS: true, TLSConfig: explicit}).GetOptions(ctx, nil)
		require.NoError(t, err)
		assert.Same(t, explicit, opts.TLSConfig)
	})
}
