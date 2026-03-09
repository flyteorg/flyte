package config

import (
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
		EventsServiceURL:        "http://localhost:8090",
		Cluster:                "",
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

	// EventsServiceURL is the URL of the State Service for reporting action state updates.
	EventsServiceURL string `json:"stateServiceURL" pflag:",URL of the State Service for action state updates"`

	// Cluster is the cluster identifier attached to action events.
	Cluster string `json:"cluster" pflag:",Cluster identifier for action events"`
}

// GetConfig returns the parsed executor configuration
func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}
