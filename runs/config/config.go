package config

import (
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

	// Auth configuration for the AuthMetadataService
	Auth AuthConfig `json:"auth"`
}

// AuthorizationServerType defines the type of authorization server.
type AuthorizationServerType int

const (
	// AuthorizationServerTypeSelf indicates the service acts as its own authorization server.
	AuthorizationServerTypeSelf AuthorizationServerType = iota
	// AuthorizationServerTypeExternal indicates an external authorization server is used.
	AuthorizationServerTypeExternal
)

// AuthConfig holds authentication configuration matching flyteadmin's auth config shape.
type AuthConfig struct {
	// AuthorizedURIs is the set of URIs clients can access the service on.
	// The first entry is used as the public URL for building OAuth2 metadata URLs.
	AuthorizedURIs []config.URL `json:"authorizedUris"`

	// GrpcAuthorizationHeader is the authorization metadata key returned to clients.
	GrpcAuthorizationHeader string `json:"grpcAuthorizationHeader"`

	// AppAuth defines app-level OAuth2 settings.
	AppAuth OAuth2Options `json:"appAuth"`

	// HTTPProxyURL allows accessing external OAuth2 servers through an HTTP proxy.
	HTTPProxyURL config.URL `json:"httpProxyURL"`

	// TokenEndpointProxyConfig proxies token endpoint calls through admin.
	TokenEndpointProxyConfig TokenEndpointProxyConfig `json:"tokenEndpointProxyConfig"`
}

// OAuth2Options holds OAuth2 authorization server options.
type OAuth2Options struct {
	// AuthServerType determines whether to use a self-hosted or external auth server.
	AuthServerType AuthorizationServerType `json:"authServerType"`

	// SelfAuthServer configures the self-hosted authorization server.
	SelfAuthServer AuthorizationServer `json:"selfAuthServer"`

	// ExternalAuthServer configures the external authorization server.
	ExternalAuthServer ExternalAuthorizationServer `json:"externalAuthServer"`

	// ThirdParty configures third-party (public client) settings.
	ThirdParty ThirdPartyConfigOptions `json:"thirdParty"`
}

// AuthorizationServer configures a self-hosted authorization server.
type AuthorizationServer struct {
	// Issuer is the issuer URL. If empty, the first AuthorizedURI is used.
	Issuer string `json:"issuer"`
}

// ExternalAuthorizationServer configures an external authorization server.
type ExternalAuthorizationServer struct {
	// BaseURL is the base URL of the external authorization server.
	BaseURL config.URL `json:"baseURL"`

	// MetadataEndpointURL overrides the default .well-known/oauth-authorization-server endpoint.
	MetadataEndpointURL config.URL `json:"metadataEndpointURL"`

	// RetryAttempts is the number of retry attempts for fetching metadata.
	RetryAttempts int `json:"retryAttempts"`

	// RetryDelay is the delay between retry attempts.
	RetryDelay config.Duration `json:"retryDelay"`
}

// ThirdPartyConfigOptions holds third-party OAuth2 client settings.
type ThirdPartyConfigOptions struct {
	// FlyteClientConfig holds public client configuration.
	FlyteClientConfig FlyteClientConfig `json:"flyteClient"`
}

// FlyteClientConfig holds the public client configuration.
type FlyteClientConfig struct {
	// ClientID is the public client ID.
	ClientID string `json:"clientId"`

	// RedirectURI is the redirect URI for the client.
	RedirectURI string `json:"redirectUri"`

	// Scopes are the OAuth2 scopes to request.
	Scopes []string `json:"scopes"`

	// Audience is the intended audience for OAuth2 tokens.
	Audience string `json:"audience"`
}

// TokenEndpointProxyConfig configures proxying of token endpoint calls.
type TokenEndpointProxyConfig struct {
	// Enabled enables token endpoint proxying.
	Enabled bool `json:"enabled"`

	// PublicURL is the public URL to use for rewriting the token endpoint.
	PublicURL config.URL `json:"publicURL"`

	// PathPrefix is appended to the public URL when rewriting.
	PathPrefix string `json:"pathPrefix"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port int    `json:"port" pflag:",Port to bind the HTTP server"`
	Host string `json:"host" pflag:",Host to bind the HTTP server"`
}

// GetConfig returns the parsed runs configuration
func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}
