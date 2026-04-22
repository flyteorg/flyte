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
