package runtime

import (
	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"

	"github.com/flyteorg/flytestdlib/config"
)

const clusterPoolsKey = "clusterPools"

var clusterPoolsConfig = config.MustRegisterSection(clusterPoolsKey, &interfaces.ClusterPoolAssignmentConfig{
	ClusterPoolAssignments: make(interfaces.ClusterPoolAssignments),
})

// Implementation of an interfaces.ClusterPoolAssignmentConfiguration
type ClusterPoolAssignmentConfigurationProvider struct{}

func (p *ClusterPoolAssignmentConfigurationProvider) GetClusterPoolAssignments() interfaces.ClusterPoolAssignments {
	return clusterPoolsConfig.GetConfig().(*interfaces.ClusterPoolAssignmentConfig).ClusterPoolAssignments
}

func NewClusterPoolAssignmentConfigurationProvider() interfaces.ClusterPoolAssignmentConfiguration {
	return &ClusterPoolAssignmentConfigurationProvider{}
}
