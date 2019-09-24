package runtime

import (
	"context"

	"github.com/lyft/flyteadmin/pkg/runtime/interfaces"

	"github.com/lyft/flytestdlib/logger"

	"github.com/lyft/flytestdlib/config"
)

const clustersKey = "clusters"

var clusterConfig = config.MustRegisterSection(clustersKey, &interfaces.Clusters{})

// Implementation of an interfaces.ClusterConfiguration
type ClusterConfigurationProvider struct{}

func (p *ClusterConfigurationProvider) GetClusterConfigs() []interfaces.ClusterConfig {
	if clusterConfig != nil {
		clusters := clusterConfig.GetConfig().(*interfaces.Clusters)
		return clusters.ClusterConfigs
	}
	logger.Warningf(context.Background(), "Failed to find clusters in config. Returning an empty slice")
	return make([]interfaces.ClusterConfig, 0)
}

func (p *ClusterConfigurationProvider) GetCurrentCluster() *interfaces.ClusterConfig {
	if clusterConfig != nil {
		clusters := clusterConfig.GetConfig().(*interfaces.Clusters)
		currentCluster := clusters.CurrentCluster
		for _, cluster := range clusters.ClusterConfigs {
			if cluster.Name == currentCluster {
				return &cluster
			}
		}
	}
	logger.Warningf(context.Background(), "Failed to find current cluster in config.")
	return nil
}

func NewClusterConfigurationProvider() interfaces.ClusterConfiguration {
	return &ClusterConfigurationProvider{}
}
