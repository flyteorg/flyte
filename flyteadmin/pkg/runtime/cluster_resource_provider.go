package runtime

import (
	"context"
	"time"

	"github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flytestdlib/config"
	"github.com/lyft/flytestdlib/logger"
)

const clusterResourceKey = "cluster_resources"

var clusterResourceConfig = config.MustRegisterSection(clusterResourceKey, &interfaces.ClusterResourceConfig{})

// Implementation of an interfaces.ClusterResourceConfiguration
type ClusterResourceConfigurationProvider struct{}

func (p *ClusterResourceConfigurationProvider) GetTemplatePath() string {
	if clusterResourceConfig != nil && clusterResourceConfig.GetConfig() != nil {
		return clusterResourceConfig.GetConfig().(*interfaces.ClusterResourceConfig).TemplatePath
	}
	logger.Warningf(context.Background(),
		"Failed to find cluster resource values in config. Returning an empty string for template path")
	return ""
}

func (p *ClusterResourceConfigurationProvider) GetTemplateData() interfaces.TemplateData {
	if clusterResourceConfig != nil && clusterResourceConfig.GetConfig() != nil {
		return clusterResourceConfig.GetConfig().(*interfaces.ClusterResourceConfig).TemplateData
	}
	logger.Warningf(context.Background(),
		"Failed to find cluster resource values in config. Returning an empty map for template data")
	return make(interfaces.TemplateData)
}

func (p *ClusterResourceConfigurationProvider) GetRefreshInterval() time.Duration {
	if clusterResourceConfig != nil && clusterResourceConfig.GetConfig() != nil {
		return clusterResourceConfig.GetConfig().(*interfaces.ClusterResourceConfig).RefreshInterval.Duration
	}
	logger.Warningf(context.Background(),
		"Failed to find cluster resource values in config. Returning 1 minute for refresh interval")
	return time.Minute
}

func (p *ClusterResourceConfigurationProvider) GetCustomTemplateData() map[interfaces.DomainName]interfaces.TemplateData {
	if clusterResourceConfig != nil && clusterResourceConfig.GetConfig() != nil {
		return clusterResourceConfig.GetConfig().(*interfaces.ClusterResourceConfig).CustomData
	}
	logger.Warningf(context.Background(),
		"Failed to find cucluster resource values in config. Returning an empty map for custom template data")
	return make(map[interfaces.DomainName]interfaces.TemplateData)
}

func NewClusterResourceConfigurationProvider() interfaces.ClusterResourceConfiguration {
	return &ClusterResourceConfigurationProvider{}
}
