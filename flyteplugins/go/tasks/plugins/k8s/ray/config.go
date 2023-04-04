package ray

import (
	pluginsConfig "github.com/flyteorg/flyteplugins/go/tasks/config"
	pluginmachinery "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/k8s"
)

//go:generate pflags Config --default-var=defaultConfig

var (
	defaultConfig = Config{
		ShutdownAfterJobFinishes: true,
		TTLSecondsAfterFinished:  3600,
		ServiceType:              "NodePort",
		IncludeDashboard:         true,
		DashboardHost:            "0.0.0.0",
		NodeIPAddress:            "$MY_POD_IP",
	}

	configSection = pluginsConfig.MustRegisterSubSection("ray", &defaultConfig)
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

	// NodeIPAddress the IP address of the head node. By default, this is pod ip address.
	NodeIPAddress string `json:"nodeIPAddress,omitempty"`

	// Remote Ray Cluster Config
	RemoteClusterConfig pluginmachinery.ClusterConfig `json:"remoteClusterConfig" pflag:"Configuration of remote K8s cluster for ray jobs"`
}

func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}

func SetConfig(cfg *Config) error {
	return configSection.SetConfig(cfg)
}
