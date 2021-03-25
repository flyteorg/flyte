package runtime

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/logger"
)

const (
	namespaceMappingKey   = "namespace_mapping"
	domainVariable        = "domain"
	projectVariable       = "project"
	projectDomainVariable = "project-domain"
)

var namespaceMappingConfig = config.MustRegisterSection(namespaceMappingKey, &interfaces.NamespaceMappingConfig{})

type NamespaceMappingConfigurationProvider struct{}

func (p *NamespaceMappingConfigurationProvider) GetNamespaceMappingConfig() common.NamespaceMapping {
	var mapping string
	if namespaceMappingConfig != nil && namespaceMappingConfig.GetConfig() != nil {
		mapping = namespaceMappingConfig.GetConfig().(*interfaces.NamespaceMappingConfig).Mapping
	}

	switch mapping {
	case domainVariable:
		return common.NamespaceMappingDomain
	case projectVariable:
		return common.NamespaceMappingProject
	case projectDomainVariable:
		return common.NamespaceMappingProjectDomain
	default:
		logger.Warningf(context.Background(), "Unsupported value for namespace_mapping in config, defaulting to <project>-<domain>")
		return common.NamespaceMappingProjectDomain
	}
}

func NewNamespaceMappingConfigurationProvider() interfaces.NamespaceMappingConfiguration {
	return &NamespaceMappingConfigurationProvider{}
}
