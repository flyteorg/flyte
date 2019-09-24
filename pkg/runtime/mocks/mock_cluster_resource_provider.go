package mocks

import (
	"time"

	"github.com/lyft/flyteadmin/pkg/runtime/interfaces"
)

type MockClusterResourceConfiguration struct {
	TemplatePath    string
	TemplateData    map[string]interfaces.DataSource
	RefreshInterval time.Duration
}

func (c MockClusterResourceConfiguration) GetTemplatePath() string {
	return c.TemplatePath
}
func (c MockClusterResourceConfiguration) GetTemplateData() map[string]interfaces.DataSource {
	return c.TemplateData
}

func (c MockClusterResourceConfiguration) GetRefreshInterval() time.Duration {
	return c.RefreshInterval
}

func NewMockClusterResourceConfiguration() interfaces.ClusterResourceConfiguration {
	return &MockClusterResourceConfiguration{}
}
