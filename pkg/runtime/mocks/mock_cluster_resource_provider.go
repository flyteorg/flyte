package mocks

import (
	"time"

	"github.com/lyft/flyteadmin/pkg/runtime/interfaces"
)

type MockClusterResourceConfiguration struct {
	TemplatePath       string
	TemplateData       interfaces.TemplateData
	RefreshInterval    time.Duration
	CustomTemplateData map[interfaces.DomainName]interfaces.TemplateData
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

func NewMockClusterResourceConfiguration() interfaces.ClusterResourceConfiguration {
	return &MockClusterResourceConfiguration{}
}
