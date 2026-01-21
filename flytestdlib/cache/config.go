package cache

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/flyteorg/flyte/v2/flytestdlib/config"
)

//go:generate enumer --type=Type -json -yaml -trimprefix=Type
//go:generate pflags Config --default-var=defaultConfig

type Type uint8

const (
	TypeInMemoryFixedSize Type = iota
	TypeRedis
)

var (
	defaultConfig = &Config{
		Type: 0,
		InMemoryFixedSize: InMemoryFixedSizeConfig{
			Size: resource.NewScaledQuantity(100, resource.Mega),
			DefaultExpiration: config.Duration{
				Duration: 24 * time.Hour,
			},
		},
		Redis: RedisConfig{
			DefaultExpiration: config.Duration{
				Duration: 24 * time.Hour,
			},
		},
	}

	configSection = config.MustRegisterSection("cache", defaultConfig)
)

type Config struct {
	// Type of cache to use
	Type Type `json:"type" pflag:",type, Type of cache to use"`

	// Config for in-memory cache
	InMemoryFixedSize InMemoryFixedSizeConfig `json:"inMemoryFixedSize" pflag:"-,Config for in-memory cache"`

	// Config for Redis cache
	Redis RedisConfig `json:"redis" pflag:"-,Config for Redis cache"`
}

// InMemoryFixedSizeConfig is a copy of ristretto.Config that can be used in config files (removed func references)
type InMemoryFixedSizeConfig struct {
	Size              *resource.Quantity `json:"size" pflag:"-,Cache size (in bytes) to allocate. Note the memory will be allocated immediately and will never grow or shrink (minimizing GCs)."`
	DefaultExpiration config.Duration    `json:"defaultExpiration" pflag:",Default expiration time for items"`
}

type RedisConfig struct {
	Options           RedisOptions    `json:"options" pflag:"-,Redis options."`
	DefaultExpiration config.Duration `json:"defaultExpiration" pflag:",Default expiration time for items."`
}

// RedisOptions is a copy of redis.Options that can be used in config files (removed func references)
type RedisOptions struct {
	// The network type, either tcp or unix.
	// Default is tcp.
	Network string
	// host:port address.
	Addr string

	// ClientName will execute the `CLIENT SETNAME ClientName` command for each conn.
	ClientName string

	// Protocol 2 or 3. Use the version to negotiate RESP version with redis-server.
	// Default is 3.
	Protocol int
	// Use the specified Username to authenticate the current connection
	// with one of the connections defined in the ACL list when connecting
	// to a Redis 6.0 instance, or greater, that is using the Redis ACL system.
	Username string
	// Optional password. Must match the password specified in the
	// requirepass server configuration option (if connecting to a Redis 5.0 instance, or lower),
	// or the User Password when connecting to a Redis 6.0 instance, or greater,
	// that is using the Redis ACL system.
	Password string

	// PasswordSecretName is the name of the secret that contains the password.
	PasswordSecretName string

	// Database to be selected after connecting to the server.
	DB int

	// Maximum number of retries before giving up.
	// Default is 3 retries; -1 (not 0) disables retries.
	MaxRetries int
	// Minimum backoff between each retry.
	// Default is 8 milliseconds; -1 disables backoff.
	MinRetryBackoff config.Duration
	// Maximum backoff between each retry.
	// Default is 512 milliseconds; -1 disables backoff.
	MaxRetryBackoff config.Duration

	// Dial timeout for establishing new connections.
	// Default is 5 seconds.
	DialTimeout config.Duration
	// Timeout for socket reads. If reached, commands will fail
	// with a timeout instead of blocking. Supported values:
	//   - `0` - default timeout (3 seconds).
	//   - `-1` - no timeout (block indefinitely).
	//   - `-2` - disables SetReadDeadline calls completely.
	ReadTimeout config.Duration
	// Timeout for socket writes. If reached, commands will fail
	// with a timeout instead of blocking.  Supported values:
	//   - `0` - default timeout (3 seconds).
	//   - `-1` - no timeout (block indefinitely).
	//   - `-2` - disables SetWriteDeadline calls completely.
	WriteTimeout config.Duration
	// ContextTimeoutEnabled controls whether the client respects context timeouts and deadlines.
	// See https://redis.uptrace.dev/guide/go-redis-debugging.html#timeouts
	ContextTimeoutEnabled bool

	// Type of connection pool.
	// true for FIFO pool, false for LIFO pool.
	// Note that FIFO has slightly higher overhead compared to LIFO,
	// but it helps closing idle connections faster reducing the pool size.
	PoolFIFO bool
	// Base number of socket connections.
	// Default is 10 connections per every available CPU as reported by runtime.GOMAXPROCS.
	// If there is not enough connections in the pool, new connections will be allocated in excess of PoolSize,
	// you can limit it through MaxActiveConns
	PoolSize int
	// Amount of time client waits for connection if all connections
	// are busy before returning an error.
	// Default is ReadTimeout + 1 second.
	PoolTimeout config.Duration
	// Minimum number of idle connections which is useful when establishing
	// new connection is slow.
	// Default is 0. the idle connections are not closed by default.
	MinIdleConns int
	// Maximum number of idle connections.
	// Default is 0. the idle connections are not closed by default.
	MaxIdleConns int
	// Maximum number of connections allocated by the pool at a given time.
	// When zero, there is no limit on the number of connections in the pool.
	MaxActiveConns int
	// ConnMaxIdleTime is the maximum amount of time a connection may be idle.
	// Should be less than server's timeout.
	//
	// Expired connections may be closed lazily before reuse.
	// If d <= 0, connections are not closed due to a connection's idle time.
	//
	// Default is 30 minutes. -1 disables idle timeout check.
	ConnMaxIdleTime config.Duration
	// ConnMaxLifetime is the maximum amount of time a connection may be reused.
	//
	// Expired connections may be closed lazily before reuse.
	// If <= 0, connections are not closed due to a connection's age.
	//
	// Default is to not close idle connections.
	ConnMaxLifetime config.Duration

	// TLS Config to use. When set, TLS will be negotiated.
	TLSConfig *tls.Config

	// // Disable set-lib on connect. Default is false.
	DisableIndentity bool
}

func (r *RedisOptions) GetOptions(ctx context.Context, secretManager SecretManager) (*redis.Options, error) {
	if len(r.PasswordSecretName) > 0 {
		password, err := secretManager.Get(ctx, r.PasswordSecretName)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get password from secret manager for secret %s", r.PasswordSecretName)
		}

		r.Password = password
	}

	return &redis.Options{
		Network:               r.Network,
		Addr:                  r.Addr,
		ClientName:            r.ClientName,
		Password:              r.Password,
		DB:                    r.DB,
		MaxRetries:            r.MaxRetries,
		MinRetryBackoff:       r.MinRetryBackoff.Duration,
		MaxRetryBackoff:       r.MaxRetryBackoff.Duration,
		DialTimeout:           r.DialTimeout.Duration,
		ReadTimeout:           r.ReadTimeout.Duration,
		WriteTimeout:          r.WriteTimeout.Duration,
		ContextTimeoutEnabled: r.ContextTimeoutEnabled,
		PoolFIFO:              r.PoolFIFO,
		PoolSize:              r.PoolSize,
		PoolTimeout:           r.PoolTimeout.Duration,
		MinIdleConns:          r.MinIdleConns,
		MaxIdleConns:          r.MaxIdleConns,
		MaxActiveConns:        r.MaxActiveConns,
		ConnMaxIdleTime:       r.ConnMaxIdleTime.Duration,
		ConnMaxLifetime:       r.ConnMaxLifetime.Duration,
		TLSConfig:             r.TLSConfig,
		DisableIndentity:      r.DisableIndentity,
		Username:              r.Username,
	}, nil
}

func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}

func MustRegisterSubsection(name string, cfg config.Config) config.Section {
	return configSection.MustRegisterSection(name, cfg)
}
