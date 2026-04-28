package config

import (
	"time"

	"github.com/flyteorg/flyte/v2/flytestdlib/config"
)

// AppConfig holds configuration for the control plane AppService.
type AppConfig struct {
	// InternalAppServiceURL is the base URL of the InternalAppService (data plane).
	// In unified mode this is overridden by the shared mux BaseURL.
	InternalAppServiceURL string `json:"internalAppServiceUrl" pflag:",URL of the internal app service"`

	// CacheTTL is the TTL for the in-memory app status cache.
	// Defaults to 30s. Set to 0 to disable caching.
	CacheTTL time.Duration `json:"cacheTtl" pflag:",TTL for app status cache"`
}


const appConfigSectionKey = "apps"

var defaultAppConfig = &AppConfig{
	InternalAppServiceURL: "http://localhost:8091",
	CacheTTL:              30 * time.Second,
}

var appConfigSection = config.MustRegisterSection(appConfigSectionKey, defaultAppConfig)

// GetAppConfig returns the current AppConfig.
func GetAppConfig() *AppConfig {
	return appConfigSection.GetConfig().(*AppConfig)
}

const internalAppConfigSectionKey = "internalApps"

//go:generate pflags InternalAppConfig --default-var=defaultInternalAppConfig

// InternalAppConfig holds configuration for the data plane InternalAppService.
type InternalAppConfig struct {
	// Enabled controls whether the InternalAppService is started.
	Enabled bool `json:"enabled" pflag:",Enable app deployment controller"`

	// BaseDomain is the base domain used to generate public URLs for apps.
	// Apps are exposed at "{name}-{project}-{domain}.{base_domain}".
	BaseDomain string `json:"baseDomain" pflag:",Base domain for app public URLs"`

	// Scheme is the URL scheme used for public app URLs ("http" or "https").
	// Defaults to "https" if unset.
	Scheme string `json:"scheme" pflag:",URL scheme for app public URLs (http or https)"`

	// DefaultRequestTimeout is the request timeout applied to apps that don't specify one.
	DefaultRequestTimeout time.Duration `json:"defaultRequestTimeout" pflag:",Default request timeout for apps"`

	// MaxRequestTimeout is the hard cap on request timeout (Knative max is 3600s).
	MaxRequestTimeout time.Duration `json:"maxRequestTimeout" pflag:",Maximum allowed request timeout for apps"`

	// IngressAppsPort is the port appended to the public app URL (e.g. 30081).
	// Set to 0 to omit the port when behind a standard 80/443 proxy.
	IngressAppsPort int `json:"ingressAppsPort" pflag:",Port for app subdomain URLs (0 = omit)"`

	// DefaultEnvVars is a map of environment variables injected into every KService
	// pod at deploy time, in addition to any env vars specified in the app spec.
	// Use this to inject cluster-internal endpoints (e.g. _U_EP_OVERRIDE) that app
	// processes need to connect back to the Flyte manager.
	DefaultEnvVars map[string]string `json:"defaultEnvVars" pflag:"-,Default env vars injected into every app pod"`

	// WatchBufferSize is the buffer size for each subscriber's event channel.
	// A larger value reduces the chance of dropped events under burst load.
	WatchBufferSize int `json:"watchBufferSize" pflag:",Buffer size for watch subscriber channels"`
}

var defaultInternalAppConfig = &InternalAppConfig{
	DefaultRequestTimeout: 300 * time.Second,
	MaxRequestTimeout:     3600 * time.Second,
	WatchBufferSize:       100,
}

var internalAppConfigSection = config.MustRegisterSection(internalAppConfigSectionKey, defaultInternalAppConfig)

// GetInternalAppConfig returns the current InternalAppConfig.
func GetInternalAppConfig() *InternalAppConfig {
	return internalAppConfigSection.GetConfig().(*InternalAppConfig)
}
