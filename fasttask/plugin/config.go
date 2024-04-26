package plugin

import (
	"time"

	"k8s.io/apimachinery/pkg/api/resource"

	pluginsConfig "github.com/flyteorg/flyte/flyteplugins/go/tasks/config"
	"github.com/flyteorg/flyte/flytestdlib/config"
)

//go:generate pflags Config --default-var=defaultConfig

var (
	defaultCPU = resource.MustParse("500m")
	defaultMemory = resource.MustParse("128Mi")
	defaultConfig = &Config{
		CallbackURI:                   "http://host.k3d.internal:15605",
		Endpoint:                      "0.0.0.0:15605",
		EnvDetectOrphanInterval:       config.Duration{Duration: time.Second * 60},
		EnvGCInterval:                 config.Duration{Duration: time.Second * 5},
		EnvRepairInterval:             config.Duration{Duration: time.Second * 10},
		GracePeriodStatusNotFound:     config.Duration{Duration: time.Second * 90},
		GracePeriodWorkersUnavailable: config.Duration{Duration: time.Second * 30},
		HeartbeatBufferSize:	       512,
		Image:                         "flyteorg/fasttask:latest",
		InitContainerCPU:              defaultCPU,
		InitContainerMemory:           defaultMemory,
		NonceLength:                   12,
		TaskStatusBufferSize:	       512,
	}

	configSection = pluginsConfig.MustRegisterSubSection("fasttask", defaultConfig)
)

type Config struct {
	CallbackURI                   string            `json:"callback-uri" pflag:",Fasttask gRPC service URI that fasttask workers will connect to."`
	Endpoint                      string            `json:"endpoint" pflag:",Fasttask gRPC service endpoint."`
	EnvDetectOrphanInterval       config.Duration   `json:"env-detect-orphan-interval" pflag:",Frequency that orphaned environments detection is performed."`
	EnvGCInterval                 config.Duration   `json:"env-gc-interval" pflag:",Frequency that environments are GCed in case of TTL expirations."`
	EnvRepairInterval             config.Duration   `json:"env-repair-interval" pflag:",Frequency that environments are repaired in case of external modifications (ex. pod deletion)."`
	GracePeriodStatusNotFound     config.Duration   `json:"grace-period-status-not-found" pflag:",The grace period for a task status to be reported before the task is considered failed."`
	GracePeriodWorkersUnavailable config.Duration   `json:"grace-period-workers-unavailable" pflag:",The grace period for a worker to become available before the task is considered failed."`
	HeartbeatBufferSize           int               `json:"heartbeat-buffer-size" pflag:",The size of the heartbeat buffer for each worker."`
	Image                         string            `json:"image" pflag:",Fasttask image to wrap the task execution with."`
	InitContainerCPU              resource.Quantity `json:"init-container-cpu" pflag:",The default cpu request / limit for the init container used to inject the fasttask worker binary."`
	InitContainerMemory           resource.Quantity `json:"init-container-memory" pflag:",The default memory request / limit for the init container used to inject the fasttask worker binary."`
	NonceLength                   int               `json:"nonce-length" pflag:",The length of the nonce value to uniquely link a fasttask replica to the environment instance, ensuring fast turnover of environments regardless of cache freshness."`
	TaskStatusBufferSize          int               `json:"task-status-buffer-size" pflag:",The size of the task status buffer for each task."`
}

func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}

// This method should be used for unit testing only
func setConfig(cfg *Config) error {
	return configSection.SetConfig(cfg)
}
