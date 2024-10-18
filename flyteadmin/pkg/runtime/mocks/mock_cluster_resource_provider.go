package mocks

import (
	"time"

	"github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
)

type MockClusterResourceConfiguration struct {
	TemplatePath           string
	TemplateData           interfaces.TemplateData
	RefreshInterval        time.Duration
	CustomTemplateData     map[interfaces.DomainName]interfaces.TemplateData
	StandaloneDeployment   bool
	SelfServe              bool
	ClusterName            string
	UnionProjectSyncConfig interfaces.UnionProjectSyncConfig
}

func (c MockClusterResourceConfiguration) GetTemplatePath() string {
	return c.TemplatePath
}
func (c MockClusterResourceConfiguration) GetTemplateData() interfaces.TemplateData {
	return c.TemplateData
}

func (c MockClusterResourceConfiguration) GetRefreshInterval() time.Duration {
	return c.RefreshInterval
}

func (c MockClusterResourceConfiguration) GetCustomTemplateData() map[interfaces.DomainName]interfaces.TemplateData {
	return c.CustomTemplateData
}

func (c MockClusterResourceConfiguration) IsStandaloneDeployment() bool {
	return c.StandaloneDeployment
}

func (c MockClusterResourceConfiguration) GetArchiveProjectConfig() interfaces.UnionProjectSyncConfig {
	return interfaces.UnionProjectSyncConfig{}
}

func (c MockClusterResourceConfiguration) GetUnionProjectSyncConfig() interfaces.UnionProjectSyncConfig {
	return c.UnionProjectSyncConfig
}

func (c MockClusterResourceConfiguration) IsSelfServe() bool {
	return c.SelfServe
}

func (c MockClusterResourceConfiguration) GetClusterName() string {
	return c.ClusterName
}

func NewMockClusterResourceConfiguration() interfaces.ClusterResourceConfiguration {
	return &MockClusterResourceConfiguration{}
}
