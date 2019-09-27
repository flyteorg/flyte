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

func (p *ClusterConfigurationProvider) GetClusterSelectionStrategy() interfaces.ClusterSelectionStrategy {
	if clusterConfig != nil {
		return clusterConfig.GetConfig().(*interfaces.Clusters).ClusterSelection
	}
	return interfaces.ClusterSelectionRandom
}

func (p *ClusterConfigurationProvider) GetClusterConfigs() []interfaces.ClusterConfig {
	if clusterConfig != nil {
		clusters := clusterConfig.GetConfig().(*interfaces.Clusters)
		return clusters.ClusterConfigs
	}
	logger.Warningf(context.Background(), "Failed to find clusters in config. Returning an empty slice")
	return make([]interfaces.ClusterConfig, 0)
}

func NewClusterConfigurationProvider() interfaces.ClusterConfiguration {
	clusterConfigProvider := ClusterConfigurationProvider{}
	clusterNameMap := make(map[string]bool)
	for _, config := range clusterConfigProvider.GetClusterConfigs() {
		if clusterNameMap[config.Name] {
			panic("Duplicate cluster names in runtime config")
		}
		clusterNameMap[config.Name] = true
	}
	return &clusterConfigProvider
}
