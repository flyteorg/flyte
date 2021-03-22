package webhook

import "github.com/flyteorg/flytestdlib/config"

//go:generate pflags Config --default-var=defaultConfig

var (
	defaultConfig = &Config{
		SecretName:    "flyte-pod-webhook",
		ServiceName:   "flyte-pod-webhook",
		MetricsPrefix: "flyte:",
		CertDir:       "/etc/webhook/certs",
		ListenPort:    9443,
	}

	configSection = config.MustRegisterSection("webhook", defaultConfig)
)

type Config struct {
	MetricsPrefix string `json:"metrics-prefix" pflag:",An optional prefix for all published metrics."`
	CertDir       string `json:"certDir" pflag:",Certificate directory to use to write generated certs. Defaults to /etc/webhook/certs/"`
	ListenPort    int    `json:"listenPort" pflag:",The port to use to listen to webhook calls. Defaults to 9443"`
	ServiceName   string `json:"serviceName" pflag:",The name of the webhook service."`
	SecretName    string `json:"secretName" pflag:",Secret name to write generated certs to."`
}

func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}
