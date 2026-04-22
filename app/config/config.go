package config

import "time"

// AppConfig holds configuration for the control plane AppService.
type AppConfig struct {
	// InternalAppServiceURL is the base URL of the InternalAppService (data plane).
	// In unified mode this is overridden by the shared mux BaseURL.
	InternalAppServiceURL string `json:"internalAppServiceUrl" pflag:",URL of the internal app service"`

	// CacheTTL is the TTL for the in-memory app status cache.
	// Defaults to 30s. Set to 0 to disable caching.
	CacheTTL time.Duration `json:"cacheTtl" pflag:",TTL for app status cache"`
}

// DefaultAppConfig returns the default control plane AppConfig.
func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		InternalAppServiceURL: "http://localhost:8091",
		CacheTTL:              30 * time.Second,
	}
}

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

	// IngressEntryPoint is the Traefik entry point name that app routes are
	// attached to (default: "apps").
	IngressEntryPoint string `json:"ingressEntryPoint" pflag:",Traefik entry point name for app ingress routes"`

	// IngressAppsDomain is the domain suffix for subdomain-based app URLs.
	// Apps are exposed at {name}-{project}-{domain}.{IngressAppsDomain}.
	// Use "localhost" for local devbox (resolves without /etc/hosts on macOS/Linux).
	// Windows users can override to e.g. "localtest.me".
	IngressAppsDomain string `json:"ingressAppsDomain" pflag:",Domain suffix for app subdomain URLs"`

	// IngressAppsPort is the port appended to the public app URL (e.g. 30081).
	// Set to 0 to omit the port when behind a standard 80/443 proxy.
	IngressAppsPort int `json:"ingressAppsPort" pflag:",Port for app subdomain URLs (0 = omit)"`

	// DefaultEnvVars is a map of environment variables injected into every KService
	// pod at deploy time, in addition to any env vars specified in the app spec.
	// Use this to inject cluster-internal endpoints (e.g. _U_EP_OVERRIDE) that app
	// processes need to connect back to the Flyte manager.
	DefaultEnvVars map[string]string `json:"defaultEnvVars" pflag:"-,Default env vars injected into every app pod"`
}
