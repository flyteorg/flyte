// Initializes an Admin Client that exposes all implemented services by FlyteAdmin server. The library supports different
// authentication flows (see AuthType). It initializes the grpc connection once and reuses it. A grpc load balancing policy
// can be configured as well.
package admin

import (
	"context"
	"path/filepath"
	"time"

	"github.com/flyteorg/flyteidl/clients/go/admin/deviceflow"
	"github.com/flyteorg/flyteidl/clients/go/admin/pkce"
	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/logger"
)

//go:generate pflags Config --default-var=defaultConfig

const (
	configSectionKey = "admin"
	DefaultClientID  = "flytepropeller"
)

var DefaultClientSecretLocation = filepath.Join(string(filepath.Separator), "etc", "secrets", "client_secret")

//go:generate enumer --type=AuthType -json -yaml -trimprefix=AuthType
type AuthType uint8

const (
	// AuthTypeClientSecret Chooses Client Secret OAuth2 protocol (ref: https://tools.ietf.org/html/rfc6749#section-4.4)
	AuthTypeClientSecret AuthType = iota
	// AuthTypePkce Chooses Proof Key Code Exchange OAuth2 extension protocol (ref: https://tools.ietf.org/html/rfc7636)
	AuthTypePkce
	// AuthTypeExternalCommand Chooses an external authentication process
	AuthTypeExternalCommand
	// AuthTypeDeviceFlow Uses device flow to authenticate in a constrained environment with no access to browser
	AuthTypeDeviceFlow
)

type Config struct {
	Endpoint              config.URL      `json:"endpoint" pflag:",For admin types, specify where the uri of the service is located."`
	UseInsecureConnection bool            `json:"insecure" pflag:",Use insecure connection."`
	InsecureSkipVerify    bool            `json:"insecureSkipVerify" pflag:",InsecureSkipVerify controls whether a client verifies the server's certificate chain and host name. Caution : shouldn't be use for production usecases'"`
	CACertFilePath        string          `json:"caCertFilePath" pflag:",Use specified certificate file to verify the admin server peer."`
	MaxBackoffDelay       config.Duration `json:"maxBackoffDelay" pflag:",Max delay for grpc backoff"`
	PerRetryTimeout       config.Duration `json:"perRetryTimeout" pflag:",gRPC per retry timeout"`
	MaxRetries            int             `json:"maxRetries" pflag:",Max number of gRPC retries"`
	AuthType              AuthType        `json:"authType" pflag:",Type of OAuth2 flow used for communicating with admin.ClientSecret,Pkce,ExternalCommand are valid values"`
	TokenRefreshWindow    config.Duration `json:"tokenRefreshWindow" pflag:",Max duration between token refresh attempt and token expiry."`
	// Deprecated: settings will be discovered dynamically
	DeprecatedUseAuth    bool     `json:"useAuth" pflag:",Deprecated: Auth will be enabled/disabled based on admin's dynamically discovered information."`
	ClientID             string   `json:"clientId" pflag:",Client ID"`
	ClientSecretLocation string   `json:"clientSecretLocation" pflag:",File containing the client secret"`
	ClientSecretEnvVar   string   `json:"clientSecretEnvVar" pflag:",Environment variable containing the client secret"`
	Scopes               []string `json:"scopes" pflag:",List of scopes to request"`
	UseAudienceFromAdmin bool     `json:"useAudienceFromAdmin" pflag:",Use Audience configured from admins public endpoint config."`
	Audience             string   `json:"audience" pflag:",Audience to use when initiating OAuth2 authorization requests."`

	// There are two ways to get the token URL. If the authorization server url is provided, the client will try to use RFC 8414 to
	// try to get the token URL. Or it can be specified directly through TokenURL config.
	// Deprecated: This will now be discovered through admin's anonymously accessible metadata.
	DeprecatedAuthorizationServerURL string `json:"authorizationServerUrl" pflag:",This is the URL to your IdP's authorization server. It'll default to Endpoint"`
	// If not provided, it'll be discovered through admin's anonymously accessible metadata endpoint.
	TokenURL string `json:"tokenUrl" pflag:",OPTIONAL: Your IdP's token endpoint. It'll be discovered from flyte admin's OAuth Metadata endpoint if not provided."`

	// See the implementation of the 'grpcAuthorizationHeader' option in Flyte Admin for more information. But
	// basically we want to be able to use a different string to pass the token from this client to the the Admin service
	// because things might be running in a service mesh (like Envoy) that already uses the default 'authorization' header
	AuthorizationHeader string `json:"authorizationHeader" pflag:",Custom metadata header to pass JWT"`

	PkceConfig pkce.Config `json:"pkceConfig" pflag:",Config for Pkce authentication flow."`

	DeviceFlowConfig deviceflow.Config `json:"deviceFlowConfig" pflag:",Config for Device authentication flow."`

	Command []string `json:"command" pflag:",Command for external authentication token generation"`

	// Set the gRPC service config formatted as a json string https://github.com/grpc/grpc/blob/master/doc/service_config.md
	// eg. {"loadBalancingConfig": [{"round_robin":{}}], "methodConfig": [{"name":[{"service": "foo", "method": "bar"}, {"service": "baz"}], "timeout": "1.000000001s"}]}
	// find the full schema here https://github.com/grpc/grpc-proto/blob/master/grpc/service_config/service_config.proto#L625
	// Note that required packages may need to be preloaded to support certain service config. For example "google.golang.org/grpc/balancer/roundrobin" should be preloaded to have round-robin policy supported.
	DefaultServiceConfig string `json:"defaultServiceConfig" pdflag:",Set the default service config for the admin gRPC client"`

	// HTTPProxyURL allows operators to access external OAuth2 servers using an external HTTP Proxy
	HTTPProxyURL config.URL `json:"httpProxyURL" pflag:",OPTIONAL: HTTP Proxy to be used for OAuth requests."`
}

var (
	defaultConfig = Config{
		MaxBackoffDelay:      config.Duration{Duration: 8 * time.Second},
		PerRetryTimeout:      config.Duration{Duration: 15 * time.Second},
		MaxRetries:           4,
		ClientID:             DefaultClientID,
		AuthType:             AuthTypeClientSecret,
		ClientSecretLocation: DefaultClientSecretLocation,
		PkceConfig: pkce.Config{
			TokenRefreshGracePeriod: config.Duration{Duration: 5 * time.Minute},
			BrowserSessionTimeout:   config.Duration{Duration: 2 * time.Minute},
		},
		DeviceFlowConfig: deviceflow.Config{
			TokenRefreshGracePeriod: config.Duration{Duration: 5 * time.Minute},
			Timeout:                 config.Duration{Duration: 10 * time.Minute},
			PollInterval:            config.Duration{Duration: 5 * time.Second},
		},
		TokenRefreshWindow: config.Duration{Duration: 0},
	}

	configSection = config.MustRegisterSectionWithUpdates(configSectionKey, &defaultConfig, func(ctx context.Context, newValue config.Config) {
		if newValue.(*Config).MaxRetries < 0 {
			logger.Panicf(ctx, "Admin configuration given with negative gRPC retry value.")
		}
	})
)

func GetConfig(ctx context.Context) *Config {
	if c, ok := configSection.GetConfig().(*Config); ok {
		return c
	}

	logger.Warnf(ctx, "Failed to retrieve config section [%v].", configSectionKey)
	return nil
}

func SetConfig(cfg *Config) error {
	return configSection.SetConfig(cfg)
}
