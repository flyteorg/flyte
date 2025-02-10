package plugin

import (
	"time"

	pluginsConfig "github.com/flyteorg/flyte/flyteplugins/go/tasks/config"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/logs"
	"github.com/flyteorg/flyte/flytestdlib/config"
)

//go:generate pflags Config --default-var=defaultConfig

var (
	defaultConfig = &Config{
		AdditionalWorkerArgs:      []string{},
		CallbackURI:               "http://host.k3d.internal:15605",
		DefaultTTL:                config.Duration{Duration: time.Second * 90},
		Endpoint:                  "0.0.0.0:15605",
		EnvDetectOrphanInterval:   config.Duration{Duration: time.Second * 60},
		EnvGCInterval:             config.Duration{Duration: time.Second * 5},
		EnvRepairInterval:         config.Duration{Duration: time.Second * 10},
		GracePeriodStatusNotFound: config.Duration{Duration: time.Second * 90},
		HeartbeatBufferSize:       512,
		Logs: logs.LogConfig{
			IsKubernetesEnabled: true,
		},
		NonceLength:          12,
		TaskStatusBufferSize: 512,
		WorkerLogLevel:       logLevelWarn,
	}

	configSection = pluginsConfig.MustRegisterSubSection("fasttask", defaultConfig)
)

type logLevel = string

const (
	logLevelDebug logLevel = "debug"
	logLevelInfo  logLevel = "info"
	logLevelWarn  logLevel = "warn"
	logLevelError logLevel = "error"
)

var logLevels = []logLevel{logLevelDebug, logLevelInfo, logLevelWarn, logLevelError}

type Config struct {
	AdditionalWorkerArgs      []string        `json:"additional-worker-args" pflag:",Additional arguments to pass to the fasttask worker binary."`
	CallbackURI               string          `json:"callback-uri" pflag:",Fasttask gRPC service URI that fasttask workers will connect to."`
	DefaultTTL                config.Duration `json:"default-ttl" pflag:",Default TTL for environments."`
	Endpoint                  string          `json:"endpoint" pflag:",Fasttask gRPC service endpoint."`
	EnvDetectOrphanInterval   config.Duration `json:"env-detect-orphan-interval" pflag:",Frequency that orphaned environments detection is performed."`
	EnvGCInterval             config.Duration `json:"env-gc-interval" pflag:",Frequency that environments are GCed in case of TTL expirations."`
	EnvRepairInterval         config.Duration `json:"env-repair-interval" pflag:",Frequency that environments are repaired in case of external modifications (ex. pod deletion)."`
	GracePeriodStatusNotFound config.Duration `json:"grace-period-status-not-found" pflag:",The grace period for a task status to be reported before the task is considered failed."`
	HeartbeatBufferSize       int             `json:"heartbeat-buffer-size" pflag:",The size of the heartbeat buffer for each worker."`
	Logs                      logs.LogConfig  `json:"logs" pflag:",Log configuration for fasttasks"`
	NonceLength               int             `json:"nonce-length" pflag:",The length of the nonce value to uniquely link a fasttask replica to the environment instance, ensuring fast turnover of environments regardless of cache freshness."`
	TaskStatusBufferSize      int             `json:"task-status-buffer-size" pflag:",The size of the task status buffer for each task."`
	WorkerLogLevel            logLevel        `json:"worker-log-level" pflag:",The log level for the fasttask worker."`
}

func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}

// This method should be used for unit testing only
func setConfig(cfg *Config) error { //nolint:unused
	return configSection.SetConfig(cfg)
}
