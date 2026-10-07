package clustered

import (
	pluginsConfig "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/config"
)

//go:generate pflags Config --default-var=defaultConfig

var (
	defaultConfig = Config{
		Kueue: KueueConfig{
			QueueName:          "user-queue",
			EvictAsSystemRetry: true,
		},
	}

	configSection = pluginsConfig.MustRegisterSubSection("clustered", &defaultConfig)
)

// Config is the configuration for the clustered plugin (plugins.clustered).
type Config struct {
	// StartupTimeout is how long each worker waits for its peers (DNS and rendezvous) before
	// it gives up. It is passed to the workers as FLYTE_CLUSTERED_STARTUP_TIMEOUT in whole
	// seconds. Zero leaves the variable unset, so the launcher keeps its own default. On a
	// cluster with an admission gate it should exceed the gate's pods-ready timeout, so the
	// gate, not the workers, gives up first on a gang that cannot assemble.
	StartupTimeout config.Duration `json:"startup-timeout" pflag:",Per-worker wait for peers; 0 keeps the launcher's."`

	Kueue KueueConfig `json:"kueue" pflag:",Gang admission through Kueue."`
}

// KueueConfig configures gang admission through Kueue.
type KueueConfig struct {
	// Enabled creates every JobSet suspended and labelled with QueueName, so Kueue admits
	// the whole gang at once. When false the JobSet is built exactly as without Kueue.
	Enabled bool `json:"enabled" pflag:",Create JobSets suspended and labelled for Kueue."`

	// QueueName is the Kueue LocalQueue every JobSet is submitted to. Tasks cannot choose
	// another one: the plugin owns the kueue.x-k8s.io/queue-name label and overrides any
	// value set through a pod template or execution labels.
	QueueName string `json:"queue-name" pflag:",Kueue LocalQueue that every clustered JobSet is submitted to."`

	// EvictAsSystemRetry reports a gang evicted after it had fully started as a
	// system-retryable failure, so it does not use up the task's own retries. When false it
	// is reported as a user-retryable failure.
	EvictAsSystemRetry bool `json:"evict-as-system-retry" pflag:",Report evictions of a running gang as system retries."`

	// AdmissionTimeout fails a JobSet that has been held suspended without ever starting
	// for this long, as a system-retryable failure. Zero disables it. It applies to any
	// suspended JobSet, including one held by a Kueue that this plugin did not label.
	AdmissionTimeout config.Duration `json:"admission-timeout" pflag:",Retry a JobSet suspended this long; 0 disables."`
}

func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}

func SetConfig(cfg *Config) error {
	return configSection.SetConfig(cfg)
}
