package config

import (
	"time"

	stdconfig "github.com/flyteorg/flyte/v2/flytestdlib/config"
)

const configSectionKey = "executor"

//go:generate pflags Config --default-var=defaultConfig

var (
	defaultConfig = &Config{
		MetricsBindAddress:     ":10254",
		HealthProbeBindAddress: ":8081",
		LeaderElect:            false,
		MetricsSecure:          true,
		WebhookCertName:        "tls.crt",
		WebhookCertKey:         "tls.key",
		MetricsCertName:        "tls.crt",
		MetricsCertKey:         "tls.key",
		EnableHTTP2:            false,
		EventsServiceURL:       "http://localhost:8090",
		CacheServiceURL:        "http://localhost:8094",
		Cluster:                "",
		MaxSystemFailures:      3,
		GC: GCConfig{
			Interval: stdconfig.Duration{Duration: 30 * time.Minute},
			MaxTTL:   stdconfig.Duration{Duration: 1 * time.Hour},
		},
	}

	configSection = stdconfig.MustRegisterSection(configSectionKey, defaultConfig)
)

// Config holds the configuration for the executor controller manager
type Config struct {
	// MetricsBindAddress is the address the metrics endpoint binds to.
	// Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable.
	MetricsBindAddress string `json:"metricsBindAddress" pflag:",Address the metrics endpoint binds to"`

	// HealthProbeBindAddress is the address the probe endpoint binds to.
	HealthProbeBindAddress string `json:"healthProbeBindAddress" pflag:",Address the probe endpoint binds to"`

	// LeaderElect enables leader election for controller manager.
	LeaderElect bool `json:"leaderElect" pflag:",Enable leader election for controller manager"`

	// MetricsSecure controls whether the metrics endpoint is served via HTTPS.
	MetricsSecure bool `json:"metricsSecure" pflag:",Serve metrics endpoint via HTTPS"`

	// WebhookCertPath is the directory containing the webhook certificate.
	WebhookCertPath string `json:"webhookCertPath" pflag:",Directory containing the webhook certificate"`

	// WebhookCertName is the name of the webhook certificate file.
	WebhookCertName string `json:"webhookCertName" pflag:",Name of the webhook certificate file"`

	// WebhookCertKey is the name of the webhook key file.
	WebhookCertKey string `json:"webhookCertKey" pflag:",Name of the webhook key file"`

	// MetricsCertPath is the directory containing the metrics server certificate.
	MetricsCertPath string `json:"metricsCertPath" pflag:",Directory containing the metrics server certificate"`

	// MetricsCertName is the name of the metrics server certificate file.
	MetricsCertName string `json:"metricsCertName" pflag:",Name of the metrics server certificate file"`

	// MetricsCertKey is the name of the metrics server key file.
	MetricsCertKey string `json:"metricsCertKey" pflag:",Name of the metrics server key file"`

	// EnableHTTP2 enables HTTP/2 for the metrics and webhook servers.
	EnableHTTP2 bool `json:"enableHTTP2" pflag:",Enable HTTP/2 for metrics and webhook servers"`

	// EventsServiceURL is the URL of the event Service for reporting action state updates.
	EventsServiceURL string `json:"EventsServiceURL" pflag:",URL of the Event Service for action event updates"`

	// CacheServiceURL is the URL of the cache service for catalog operations.
	CacheServiceURL string `json:"cacheServiceURL" pflag:",URL of the cache service for task cache operations"`

	// Cluster is the cluster identifier attached to action events.
	Cluster string `json:"cluster" pflag:",Cluster identifier for action events"`

	// MaxSystemFailures bounds consecutive system-level failures (Plugin.Handle Go
	// errors and plugin-reported system-retryable failures) before a TaskAction is
	// converted to a permanent failure.
	MaxSystemFailures uint32 `json:"maxSystemFailures" pflag:",Max consecutive system-level failures before forcing permanent failure"`

	// GC configures the garbage collector for terminal TaskActions.
	GC GCConfig `json:"gc" pflag:",Garbage collector configuration for terminal TaskActions"`
}

// GCConfig holds the configuration for the TaskAction garbage collector.
type GCConfig struct {
	// Interval is how often the garbage collector runs. 0 disables GC.
	Interval stdconfig.Duration `json:"interval" pflag:",How often the garbage collector runs. 0 disables GC."`

	// MaxTTL is the time-to-live for terminal TaskActions before deletion.
	MaxTTL stdconfig.Duration `json:"maxTTL" pflag:",Time-to-live for terminal TaskActions before deletion."`
}

// GetConfig returns the parsed executor configuration
func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}
