package runtime

import (
	"time"

	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/config"
)

const clusterResourceKey = "cluster_resources"

var clusterResourceConfig = config.MustRegisterSection(clusterResourceKey, &interfaces.ClusterResourceConfig{
	TemplateData: make(map[string]interfaces.DataSource),
	RefreshInterval: config.Duration{
		Duration: time.Minute,
	},
	CustomData: make(map[interfaces.DomainName]interfaces.TemplateData),
})

// Implementation of an interfaces.ClusterResourceConfiguration
type ClusterResourceConfigurationProvider struct{}

func (p *ClusterResourceConfigurationProvider) GetTemplatePath() string {
	return clusterResourceConfig.GetConfig().(*interfaces.ClusterResourceConfig).TemplatePath
}

func (p *ClusterResourceConfigurationProvider) GetTemplateData() interfaces.TemplateData {
	return clusterResourceConfig.GetConfig().(*interfaces.ClusterResourceConfig).TemplateData
}

func (p *ClusterResourceConfigurationProvider) GetRefreshInterval() time.Duration {
	return clusterResourceConfig.GetConfig().(*interfaces.ClusterResourceConfig).RefreshInterval.Duration
}

func (p *ClusterResourceConfigurationProvider) GetCustomTemplateData() map[interfaces.DomainName]interfaces.TemplateData {
	return clusterResourceConfig.GetConfig().(*interfaces.ClusterResourceConfig).CustomData
}

func (p *ClusterResourceConfigurationProvider) IsStandaloneDeployment() bool {
	return clusterResourceConfig.GetConfig().(*interfaces.ClusterResourceConfig).StandaloneDeployment
}

func NewClusterResourceConfigurationProvider() interfaces.ClusterResourceConfiguration {
	return &ClusterResourceConfigurationProvider{}
}
