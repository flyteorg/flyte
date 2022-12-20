package runtime

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flytestdlib/config"
)

const clustersKey = "clusters"

var clusterConfig = config.MustRegisterSection(clustersKey, &interfaces.Clusters{})

// Implementation of an interfaces.ClusterConfiguration
type ClusterConfigurationProvider struct{}

func (p *ClusterConfigurationProvider) GetLabelClusterMap() map[string][]interfaces.ClusterEntity {
	if clusterConfig != nil {
		clusters := clusterConfig.GetConfig().(*interfaces.Clusters)
		return clusters.LabelClusterMap
	}
	logger.Warningf(context.Background(), "Failed to find clusters in config. Returning an empty slice")
	return make(map[string][]interfaces.ClusterEntity)
}

func (p *ClusterConfigurationProvider) GetClusterConfigs() []interfaces.ClusterConfig {
	if clusterConfig != nil {
		clusters := clusterConfig.GetConfig().(*interfaces.Clusters)
		return clusters.ClusterConfigs
	}
	logger.Warningf(context.Background(), "Failed to find clusters in config. Returning an empty slice")
	return make([]interfaces.ClusterConfig, 0)
}

func (p *ClusterConfigurationProvider) GetDefaultExecutionLabel() string {
	if clusterConfig != nil {
		clusters := clusterConfig.GetConfig().(*interfaces.Clusters)
		return clusters.DefaultExecutionLabel
	}
	logger.Debug(context.Background(), "Failed to find default execution label in config. Will use random cluster if no execution label matches.")
	return ""
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
