// Initializes an Admin Client that exposes all implemented services by FlyteAdmin server. The library supports different
// authentication flows (see AuthType). It initializes the grpc connection once and reuses it. The gRPC connection is
// sticky (it hogs one server and keeps the connection alive). For better load balancing against the server, place a
// proxy service in between instead.
package admin

import (
	"context"
	"path/filepath"
	"time"

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
	// Chooses Client Secret OAuth2 protocol (ref: https://tools.ietf.org/html/rfc6749#section-4.4)
	AuthTypeClientSecret AuthType = iota
	// Chooses Proof Key Code Exchange OAuth2 extension protocol (ref: https://tools.ietf.org/html/rfc7636)
	AuthTypePkce
	// Chooses an external authentication process
	AuthTypeExternalCommand
)

type Config struct {
	Endpoint              config.URL      `json:"endpoint" pflag:",For admin types, specify where the uri of the service is located."`
	UseInsecureConnection bool            `json:"insecure" pflag:",Use insecure connection."`
	InsecureSkipVerify    bool            `json:"insecureSkipVerify" pflag:",InsecureSkipVerify controls whether a client verifies the server's certificate chain and host name. Caution : shouldn't be use for production usecases'"`
	MaxBackoffDelay       config.Duration `json:"maxBackoffDelay" pflag:",Max delay for grpc backoff"`
	PerRetryTimeout       config.Duration `json:"perRetryTimeout" pflag:",gRPC per retry timeout"`
	MaxRetries            int             `json:"maxRetries" pflag:",Max number of gRPC retries"`
	AuthType              AuthType        `json:"authType" pflag:"-,Type of OAuth2 flow used for communicating with admin."`
	// Deprecated: settings will be discovered dynamically
	DeprecatedUseAuth    bool     `json:"useAuth" pflag:",Deprecated: Auth will be enabled/disabled based on admin's dynamically discovered information."`
	ClientID             string   `json:"clientId" pflag:",Client ID"`
	ClientSecretLocation string   `json:"clientSecretLocation" pflag:",File containing the client secret"`
	Scopes               []string `json:"scopes" pflag:",List of scopes to request"`

	// There are two ways to get the token URL. If the authorization server url is provided, the client will try to use RFC 8414 to
	// try to get the token URL. Or it can be specified directly through TokenURL config.
	// Deprecated: This will now be discovered through admin's anonymously accessible metadata.
	DeprecatedAuthorizationServerURL string `json:"authorizationServerUrl" pflag:",This is the URL to your IdP's authorization server. It'll default to Endpoint"`
	// If not provided, it'll be discovered through admin's anonymously accessible metadata endpoint.
	TokenURL string `json:"tokenUrl" pflag:",OPTIONAL: Your IdP's token endpoint. It'll be discovered from flyte admin's OAuth Metadata endpoint if not provided."`

	// See the implementation of the 'grpcAuthorizationHeader' option in Flyte Admin for more information. But
	// basically we want to be able to use a different string to pass the token from this client to the the Admin service
	// because things might be running in a service mesh (like Envoy) that already uses the default 'authorization' header
	// Deprecated: It will automatically be discovered through an anonymously accessible auth metadata service.
	DeprecatedAuthorizationHeader string `json:"authorizationHeader" pflag:",Custom metadata header to pass JWT"`

	PkceConfig pkce.Config `json:"pkceConfig" pflag:",Config for Pkce authentication flow."`

	Command []string `json:"command" pflag:",Command for external authentication token generation"`
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
			BrowserSessionTimeout:   config.Duration{Duration: 15 * time.Second},
		},
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
