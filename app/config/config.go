package config

import "time"

// AppConfig holds configuration for the App deployment controller.
type AppConfig struct {
	// Enabled controls whether the app deployment controller is started.
	Enabled bool `json:"enabled" pflag:",Enable app deployment controller"`

	// BaseDomain is the base domain used to generate public URLs for apps.
	// Apps are exposed at "{name}-{project}-{domain}.{base_domain}".
	BaseDomain string `json:"baseDomain" pflag:",Base domain for app public URLs"`

	// DefaultRequestTimeout is the request timeout applied to apps that don't specify one.
	DefaultRequestTimeout time.Duration `json:"defaultRequestTimeout" pflag:",Default request timeout for apps"`

	// MaxRequestTimeout is the hard cap on request timeout (Knative max is 3600s).
	MaxRequestTimeout time.Duration `json:"maxRequestTimeout" pflag:",Maximum allowed request timeout for apps"`
}
