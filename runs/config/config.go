package config

import (
	"time"

	"github.com/flyteorg/flyte/v2/flytestdlib/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/database"
)

const configSectionKey = "runs"

//go:generate pflags Config --default-var=defaultConfig

var defaultConfig = &Config{
	Server: ServerConfig{
		Port: 8090,
		Host: "0.0.0.0",
	},
	WatchBufferSize:   100,
	ActionsServiceURL: "http://localhost:8090",
	StoragePrefix:     "file:///tmp/flyte/data",
	SeedProjects:      []string{"flytesnacks"},
	Domains: []DomainConfig{
		{ID: "development", Name: "Development"},
		{ID: "production", Name: "Production"},
		{ID: "staging", Name: "Staging"},
	},
	TriggerScheduler: TriggerSchedulerConfig{
		Enabled:               true,
		ResyncInterval:        30 * time.Second,
		MaxCatchupRunsPerLoop: 100,
		ExecutionQPS:          10.0,
		ExecutionBurst:        20,
	},
}

var configSection = config.MustRegisterSection(configSectionKey, defaultConfig)

// Config holds the configuration for the Runs service
type Config struct {
	// HTTP server configuration
	Server ServerConfig `json:"server"`

	// Database configuration
	Database database.DbConfig `json:"database"`

	// Watch/streaming settings
	WatchBufferSize int `json:"watchBufferSize" pflag:",Buffer size for watch streams"`

	// Actions service URL for enqueuing actions
	ActionsServiceURL string `json:"actionsServiceUrl" pflag:",URL of the actions service"`

	// StoragePrefix is the base URI for storing run data (inputs, outputs)
	// e.g. "s3://my-bucket" or "gs://my-bucket" or "file:///tmp/flyte/data"
	StoragePrefix string `json:"storagePrefix" pflag:",Base URI prefix for storing run inputs and outputs"`

	// SeedProjects are created at startup. If unset, the service seeds the default project list.
	SeedProjects []string `json:"seedProjects" pflag:",Projects to create by default at startup"`

	// Domains are injected into project responses (not stored per project row).
	Domains []DomainConfig `json:"domains"`

	// TriggerScheduler configures the cron-based trigger scheduler worker.
	TriggerScheduler TriggerSchedulerConfig `json:"triggerScheduler"`

	// AuthMetadata configures the OAuth2 authorization-server metadata endpoint
	// (the GetOAuth2Metadata RPC and /.well-known/oauth-authorization-server).
	AuthMetadata AuthMetadataConfig `json:"authMetadata"`
}

// AuthMetadataConfig controls how the runs service serves OAuth2 authorization
// server metadata. When ExternalAuthServerBaseURL is set, the service proxies
// the external authorization server's metadata document (e.g. Okta) so that
// clients discovering auth at this deployment are pointed at the external IdP
// and obtain externally-issued tokens. When empty, GetOAuth2Metadata returns
// Unimplemented (HTTP 501 for the well-known handler).
type AuthMetadataConfig struct {
	// ExternalAuthServerBaseURL is the base URL of the external OAuth2
	// authorization server to proxy metadata from
	// (e.g. "https://signin.example.com/oauth2/default"). Empty disables the
	// endpoint (GetOAuth2Metadata returns Unimplemented).
	ExternalAuthServerBaseURL string `json:"externalAuthServerBaseUrl" pflag:",Base URL of the external OAuth2 authorization server to proxy metadata from"`

	// ExternalMetadataURL optionally overrides the metadata path resolved
	// against ExternalAuthServerBaseURL. Defaults to
	// ".well-known/oauth-authorization-server".
	ExternalMetadataURL string `json:"externalMetadataUrl" pflag:",Override for the external metadata path"`

	// RetryAttempts is how many times to try fetching external metadata (default 5).
	RetryAttempts int `json:"retryAttempts" pflag:",Attempts to fetch external metadata"`

	// RetryDelay is the delay between fetch attempts (default 1s).
	RetryDelay time.Duration `json:"retryDelay" pflag:",Delay between external metadata fetch attempts"`

	// AuthorizationMetadataKey is the header/metadata key clients should place
	// tokens in, returned by GetPublicClientConfig (default "authorization").
	AuthorizationMetadataKey string `json:"authorizationMetadataKey" pflag:",Header key clients should use for tokens"`

	// FlyteClient is the public (CLI/SDK) OAuth2 client configuration returned
	// by GetPublicClientConfig.
	FlyteClient FlyteClientConfig `json:"flyteClient"`
}

// FlyteClientConfig mirrors flyteadmin's appAuth.thirdPartyConfig.flyteClient:
// the public OAuth2 client (flytectl/pyflyte) settings advertised to SDKs via
// GetPublicClientConfig.
type FlyteClientConfig struct {
	// ClientID is the public client id used by CLI/SDK login flows.
	ClientID string `json:"clientId" pflag:",Public OAuth2 client id advertised to SDKs"`

	// RedirectURI is the callback the public client listens on during login.
	RedirectURI string `json:"redirectUri" pflag:",Redirect URI for the public client login flow"`

	// Scopes are the OAuth2 scopes the public client should request.
	Scopes []string `json:"scopes" pflag:",Scopes the public client should request"`

	// Audience is the intended audience for requested tokens (sent when the IdP
	// requires it, e.g. Auth0/Okta custom authorization servers).
	Audience string `json:"audience" pflag:",Audience for requested tokens"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port int    `json:"port" pflag:",Port to bind the HTTP server"`
	Host string `json:"host" pflag:",Host to bind the HTTP server"`
}

// DomainConfig defines a system domain to inject into project responses.
type DomainConfig struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// TriggerSchedulerConfig controls the cron-based scheduler worker.
type TriggerSchedulerConfig struct {
	// Enabled turns the scheduler worker on or off.
	Enabled bool `json:"enabled" pflag:",Enable the trigger scheduler worker"`

	// ResyncInterval is how often the scheduler re-reads active triggers from the DB.
	ResyncInterval time.Duration `json:"resyncInterval" pflag:",How often to resync active triggers from the database"`

	// MaxCatchupRunsPerLoop caps how many catchup runs are fired per resync loop.
	MaxCatchupRunsPerLoop int `json:"maxCatchupRunsPerLoop" pflag:",Maximum catchup runs fired per resync loop"`

	// ExecutionQPS is the token-bucket rate for CreateRun calls (tokens/second).
	ExecutionQPS float64 `json:"executionQps" pflag:",Rate limit for CreateRun calls (requests per second)"`

	// ExecutionBurst is the token-bucket burst size.
	ExecutionBurst int `json:"executionBurst" pflag:",Burst size for CreateRun rate limiter"`
}

// GetConfig returns the parsed runs configuration
func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}
