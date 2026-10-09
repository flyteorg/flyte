package redis

import (
	"context"
	"crypto/tls"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/flyteorg/flyte/v2/flytestdlib/config"
)

// SecretManager resolves named Redis password secrets.
type SecretManager interface {
	Get(ctx context.Context, key string) (string, error)
}

// Option supplies a runtime dependency when resolving Redis configuration.
type Option func(*configOptions)

type configOptions struct {
	secretManager  SecretManager
	tracerProvider trace.TracerProvider
	meterProvider  metric.MeterProvider
}

// WithSecretManager supplies the manager used to resolve Redis and Sentinel password secrets.
func WithSecretManager(secretManager SecretManager) Option {
	return func(options *configOptions) {
		options.secretManager = secretManager
	}
}

// WithTracerProvider enables Redis OpenTelemetry tracing with the supplied provider.
func WithTracerProvider(provider trace.TracerProvider) Option {
	return func(options *configOptions) { options.tracerProvider = provider }
}

// WithMeterProvider enables Redis OpenTelemetry metrics with the supplied provider.
func WithMeterProvider(provider metric.MeterProvider) Option {
	return func(options *configOptions) { options.meterProvider = provider }
}

func resolveOptions(opts []Option) configOptions {
	var options configOptions
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// Config contains the Redis client settings shared by cache and storage.
// Function-valued go-redis options are omitted so it can be used in config files.
type Config struct {
	// The network type, either tcp or unix.
	// Default is tcp.
	Network string
	// host:port address.
	Addr string `json:"addr"`

	// Addrs contains cluster seed addresses or Sentinel addresses. Takes precedence over Addr.
	// Multiple addresses select cluster mode unless MasterName is set.
	Addrs []string `json:"addrs,omitempty"`
	// IsClusterMode enables cluster mode even with a single seed address.
	IsClusterMode bool `json:"isClusterMode,omitempty"`
	// MasterName selects Sentinel mode and identifies the monitored Redis master.
	MasterName string `json:"masterName,omitempty"`
	// Sentinel credentials are separate from Redis server credentials.
	SentinelUsername           string `json:"sentinelUsername,omitempty"`
	SentinelPassword           string `json:"sentinelPassword,omitempty"`
	SentinelPasswordSecretName string `json:"sentinelPasswordSecretName,omitempty"`
	// Cluster routing and redirection settings. ReadOnly also selects Sentinel replicas.
	MaxRedirects   int  `json:"maxRedirects,omitempty"`
	ReadOnly       bool `json:"readOnly,omitempty"`
	RouteByLatency bool `json:"routeByLatency,omitempty"`
	RouteRandomly  bool `json:"routeRandomly,omitempty"`

	// ClientName will execute the `CLIENT SETNAME ClientName` command for each conn.
	ClientName string

	// Protocol 2 or 3. Use the version to negotiate RESP version with redis-server.
	// Default is 3.
	Protocol int
	// Use the specified Username to authenticate the current connection
	// with one of the connections defined in the ACL list when connecting
	// to a Redis 6.0 instance, or greater, that is using the Redis ACL system.
	Username string `json:"username"`
	// Optional password. Must match the password specified in the
	// requirepass server configuration option (if connecting to a Redis 5.0 instance, or lower),
	// or the User Password when connecting to a Redis 6.0 instance, or greater,
	// that is using the Redis ACL system.
	Password string `json:"password"`

	// PasswordPath points to a file containing the Redis password. Its contents are
	// trimmed of surrounding whitespace, as with Postgres. Takes precedence over
	// Password and PasswordSecretName.
	PasswordPath string `json:"passwordPath" pflag:",Points to the file containing the Redis password."`

	// PasswordSecretName is the name of the secret that contains the password.
	PasswordSecretName string

	// Database to be selected after connecting to the server.
	DB int `json:"db"`

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

	// TLS Config to use. When set, TLS will be negotiated. Not settable from a
	// config file; use UseTLS for config-driven TLS.
	TLSConfig *tls.Config

	// UseTLS negotiates TLS using the system certificate pool. Prefer this over
	// TLSConfig when configuring from YAML/JSON, where TLSConfig cannot be set.
	// Ignored when TLSConfig is already provided.
	UseTLS bool

	// TLSInsecureSkipVerify disables server certificate verification. Only set
	// this for testing against self-signed certificates.
	TLSInsecureSkipVerify bool

	// // Disable set-lib on connect. Default is false.
	DisableIndentity bool
}

// Addresses returns configured endpoints, falling back to the legacy Addr field.
func (r Config) Addresses() []string {
	if len(r.Addrs) > 0 {
		return append([]string(nil), r.Addrs...)
	}
	if r.Addr != "" {
		return []string{r.Addr}
	}
	return nil
}

// GetUniversalOptions resolves settings for standalone, cluster, or Sentinel clients.
// Use WithSecretManager when a password secret is configured.
func (r Config) GetUniversalOptions(ctx context.Context, opts ...Option) (*redis.UniversalOptions, error) {
	return r.getUniversalOptions(ctx, resolveOptions(opts))
}

func (r Config) getUniversalOptions(ctx context.Context, options configOptions) (*redis.UniversalOptions, error) {
	secretManager := options.secretManager
	if r.PasswordPath != "" {
		password, err := os.ReadFile(r.PasswordPath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read Redis password from path %s", r.PasswordPath)
		}
		r.Password = strings.TrimSpace(string(password))
	} else if len(r.PasswordSecretName) > 0 {
		if secretManager == nil {
			return nil, errors.Errorf("password secret %s requires a secret manager", r.PasswordSecretName)
		}
		password, err := secretManager.Get(ctx, r.PasswordSecretName)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get password from secret manager for secret %s", r.PasswordSecretName)
		}

		r.Password = password
	}

	tlsConfig := r.TLSConfig
	if tlsConfig == nil && r.UseTLS {
		tlsConfig = &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: r.TLSInsecureSkipVerify, //nolint:gosec // gated behind explicit UseTLS/TLSInsecureSkipVerify config
		}
	}

	var err error
	sentinelPassword := r.SentinelPassword
	if r.SentinelPasswordSecretName != "" {
		if secretManager == nil {
			return nil, errors.Errorf("password secret %s requires a secret manager", r.SentinelPasswordSecretName)
		}
		sentinelPassword, err = secretManager.Get(ctx, r.SentinelPasswordSecretName)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get Sentinel password from secret manager for secret %s", r.SentinelPasswordSecretName)
		}
	}
	addrs := r.Addresses()
	cluster := r.MasterName == "" && (r.IsClusterMode || len(addrs) > 1)
	if cluster && r.DB != 0 {
		return nil, errors.New("Redis cluster mode only supports database 0")
	}
	if (cluster || r.MasterName != "") && r.Network != "" && r.Network != "tcp" {
		return nil, errors.New("Redis cluster and Sentinel modes require the tcp network")
	}
	return &redis.UniversalOptions{
		Addrs:                 addrs,
		IsClusterMode:         r.IsClusterMode,
		MasterName:            r.MasterName,
		SentinelUsername:      r.SentinelUsername,
		SentinelPassword:      sentinelPassword,
		MaxRedirects:          r.MaxRedirects,
		ReadOnly:              r.ReadOnly,
		RouteByLatency:        r.RouteByLatency,
		RouteRandomly:         r.RouteRandomly,
		ClientName:            r.ClientName,
		Protocol:              r.Protocol,
		Username:              r.Username,
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
		TLSConfig:             tlsConfig,
		DisableIndentity:      r.DisableIndentity,
	}, nil
}

// NewClient creates the client selected by the configured addresses and MasterName.
// Use WithSecretManager when a password secret is configured.
func (r Config) NewClient(ctx context.Context, opts ...Option) (redis.UniversalClient, error) {
	runtimeOptions := resolveOptions(opts)
	options, err := r.getUniversalOptions(ctx, runtimeOptions)
	if err != nil {
		return nil, err
	}
	var client redis.UniversalClient
	if r.MasterName == "" && !r.IsClusterMode && len(options.Addrs) <= 1 {
		simple := options.Simple()
		simple.Network = r.Network
		client = redis.NewClient(simple)
	} else {
		client = redis.NewUniversalClient(options)
	}
	if runtimeOptions.tracerProvider != nil {
		if err := redisotel.InstrumentTracing(client,
			redisotel.WithTracerProvider(runtimeOptions.tracerProvider),
			redisotel.WithDBStatement(false),
		); err != nil {
			_ = client.Close()
			return nil, errors.Wrap(err, "failed to instrument Redis tracing")
		}
	}
	if runtimeOptions.meterProvider != nil {
		if err := redisotel.InstrumentMetrics(client, redisotel.WithMeterProvider(runtimeOptions.meterProvider)); err != nil {
			_ = client.Close()
			return nil, errors.Wrap(err, "failed to instrument Redis metrics")
		}
	}
	return client, nil
}
