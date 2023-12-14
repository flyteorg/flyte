package ray

import (
	"context"

	v1 "k8s.io/api/core/v1"

	pluginsConfig "github.com/flyteorg/flyte/flyteplugins/go/tasks/config"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/logs"
	pluginmachinery "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/tasklog"
	"github.com/flyteorg/flyte/flytestdlib/config"
)

//go:generate pflags Config --default-var=defaultConfig

var (
	defaultConfig = Config{
		ShutdownAfterJobFinishes: true,
		TTLSecondsAfterFinished:  3600,
		ServiceType:              "NodePort",
		IncludeDashboard:         true,
		DashboardHost:            "0.0.0.0",
		EnableUsageStats:         false,
		Defaults: DefaultConfig{
			HeadNode: NodeConfig{
				StartParameters: map[string]string{
					// Disable usage reporting by default: https://docs.ray.io/en/latest/cluster/usage-stats.html
					DisableUsageStatsStartParameter: "true",
				},
				IPAddress: "$MY_POD_IP",
			},
			WorkerNode: NodeConfig{
				StartParameters: map[string]string{
					// Disable usage reporting by default: https://docs.ray.io/en/latest/cluster/usage-stats.html
					DisableUsageStatsStartParameter: "true",
				},
				IPAddress: "$MY_POD_IP",
			},
		},
	}

	configSection = pluginsConfig.MustRegisterSubSectionWithUpdates("ray", &defaultConfig,
		func(ctx context.Context, newValue config.Config) {
			if newValue == nil {
				return
			}

			if len(newValue.(*Config).Defaults.HeadNode.IPAddress) == 0 {
				newValue.(*Config).Defaults.HeadNode.IPAddress = newValue.(*Config).DeprecatedNodeIPAddress
			}

			if len(newValue.(*Config).Defaults.WorkerNode.IPAddress) == 0 {
				newValue.(*Config).Defaults.WorkerNode.IPAddress = newValue.(*Config).DeprecatedNodeIPAddress
			}
		})
)

// Config is config for 'ray' plugin
type Config struct {
	// ShutdownAfterJobFinishes will determine whether to delete the ray cluster once rayJob succeed or failed
	ShutdownAfterJobFinishes bool `json:"shutdownAfterJobFinishes,omitempty"`

	// TTLSecondsAfterFinished is the TTL to clean up RayCluster.
	// It's only working when ShutdownAfterJobFinishes set to true.
	TTLSecondsAfterFinished int32 `json:"ttlSecondsAfterFinished,omitempty"`

	// Kubernetes Service Type, valid values are 'ClusterIP', 'NodePort' and 'LoadBalancer'
	ServiceType string `json:"serviceType,omitempty"`

	// IncludeDashboard is used to start a Ray Dashboard if set to true
	IncludeDashboard bool `json:"includeDashboard,omitempty"`

	// DashboardHost the host to bind the dashboard server to, either localhost (127.0.0.1)
	// or 0.0.0.0 (available from all interfaces). By default, this is localhost.
	DashboardHost string `json:"dashboardHost,omitempty"`

	// DeprecatedNodeIPAddress the IP address of the head node. By default, this is pod ip address.
	DeprecatedNodeIPAddress string `json:"nodeIPAddress,omitempty" pflag:"-,DEPRECATED. Please use DefaultConfig.[HeadNode|WorkerNode].IPAddress"`

	// Remote Ray Cluster Config
	RemoteClusterConfig  pluginmachinery.ClusterConfig `json:"remoteClusterConfig" pflag:"Configuration of remote K8s cluster for ray jobs"`
	Logs                 logs.LogConfig                `json:"logs" pflag:"-,Log configuration for ray jobs"`
	LogsSidecar          *v1.Container                 `json:"logsSidecar" pflag:"-,Sidecar to inject into head pods for capturing ray job logs"`
	DashboardURLTemplate *tasklog.TemplateLogPlugin    `json:"dashboardURLTemplate" pflag:"-,Template for URL of Ray dashboard running on a head node."`
	Defaults             DefaultConfig                 `json:"defaults" pflag:"-,Default configuration for ray jobs"`
	EnableUsageStats     bool                          `json:"enableUsageStats" pflag:",Enable usage stats for ray jobs. These stats are submitted to usage-stats.ray.io per https://docs.ray.io/en/latest/cluster/usage-stats.html"`
}

type DefaultConfig struct {
	HeadNode   NodeConfig `json:"headNode,omitempty" pflag:"-,Default configuration for head node of ray jobs"`
	WorkerNode NodeConfig `json:"workerNode,omitempty" pflag:"-,Default configuration for worker node of ray jobs"`
}

type NodeConfig struct {
	StartParameters map[string]string `json:"startParameters,omitempty" pflag:"-,Start parameters for the node"`
	IPAddress       string            `json:"ipAddress,omitempty" pflag:"-,IP address of the node"`
}

func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}

func SetConfig(cfg *Config) error {
	return configSection.SetConfig(cfg)
}
