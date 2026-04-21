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

	// IngressEnabled controls whether a Traefik IngressRoute and Middleware are
	// created for each deployed app so it is reachable at
	// /<project>/<domain>/<app> through the Traefik ingress controller.
	// Enable this for sandbox/local setups where Traefik is the ingress controller.
	IngressEnabled bool `json:"ingressEnabled" pflag:",Create Traefik IngressRoute for each app"`

	// IngressEntryPoint is the Traefik entry point name that app routes are
	// attached to (default: "web").
	IngressEntryPoint string `json:"ingressEntryPoint" pflag:",Traefik entry point name for app ingress routes"`

	// IngressBaseURL is the externally reachable base URL of the ingress controller
	// (e.g. "http://localhost:30080"). When set, the public URL surfaced for each app
	// is built as {IngressBaseURL}/{project}/{domain}/{app} instead of the Knative
	// host-based URL. Only used when IngressEnabled is true.
	IngressBaseURL string `json:"ingressBaseUrl" pflag:",Base URL of the ingress controller for app public URLs"`

	// DefaultEnvVars is a list of environment variables injected into every KService
	// pod at deploy time, in addition to any env vars specified in the app spec.
	// Use this to inject cluster-internal endpoints (e.g. _U_EP_OVERRIDE) that app
	// processes need to connect back to the Flyte manager.
	DefaultEnvVars []EnvVar `json:"defaultEnvVars" pflag:"-,Default env vars injected into every app pod"`
}

// EnvVar is a name/value environment variable pair for app pod injection.
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
