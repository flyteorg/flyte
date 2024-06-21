package plugin

import (
	"time"

	pluginsConfig "github.com/flyteorg/flyte/flyteplugins/go/tasks/config"
	"github.com/flyteorg/flyte/flytestdlib/config"
)

//go:generate pflags Config --default-var=defaultConfig

var (
	defaultConfig = &Config{
		CallbackURI:               "http://host.k3d.internal:15605",
		Endpoint:                  "0.0.0.0:15605",
		EnvDetectOrphanInterval:   config.Duration{Duration: time.Second * 60},
		EnvGCInterval:             config.Duration{Duration: time.Second * 5},
		EnvRepairInterval:         config.Duration{Duration: time.Second * 10},
		GracePeriodStatusNotFound: config.Duration{Duration: time.Second * 90},
		HeartbeatBufferSize:       512,
		NonceLength:               12,
		TaskStatusBufferSize:      512,
		AdditionalWorkerArgs:      []string{},
	}

	configSection = pluginsConfig.MustRegisterSubSection("fasttask", defaultConfig)
)

type Config struct {
	CallbackURI               string          `json:"callback-uri" pflag:",Fasttask gRPC service URI that fasttask workers will connect to."`
	Endpoint                  string          `json:"endpoint" pflag:",Fasttask gRPC service endpoint."`
	EnvDetectOrphanInterval   config.Duration `json:"env-detect-orphan-interval" pflag:",Frequency that orphaned environments detection is performed."`
	EnvGCInterval             config.Duration `json:"env-gc-interval" pflag:",Frequency that environments are GCed in case of TTL expirations."`
	EnvRepairInterval         config.Duration `json:"env-repair-interval" pflag:",Frequency that environments are repaired in case of external modifications (ex. pod deletion)."`
	GracePeriodStatusNotFound config.Duration `json:"grace-period-status-not-found" pflag:",The grace period for a task status to be reported before the task is considered failed."`
	HeartbeatBufferSize       int             `json:"heartbeat-buffer-size" pflag:",The size of the heartbeat buffer for each worker."`
	NonceLength               int             `json:"nonce-length" pflag:",The length of the nonce value to uniquely link a fasttask replica to the environment instance, ensuring fast turnover of environments regardless of cache freshness."`
	TaskStatusBufferSize      int             `json:"task-status-buffer-size" pflag:",The size of the task status buffer for each task."`
	AdditionalWorkerArgs      []string        `json:"additional-worker-args" pflag:",Additional arguments to pass to the fasttask worker binary."`
}

func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}

// This method should be used for unit testing only
func setConfig(cfg *Config) error { //nolint:unused
	return configSection.SetConfig(cfg)
}
