package runtime

import (
	"context"

	"github.com/lyft/flyteadmin/pkg/common"
	"github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flytestdlib/config"
	"github.com/lyft/flytestdlib/logger"
)

const (
	namespaceMappingKey   = "namespace_mapping"
	domainVariable        = "domain"
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
		return common.Domain
	case projectDomainVariable:
		return common.ProjectDomain
	default:
		logger.Warningf(context.Background(), "Unsupported value for namespace_mapping in config, defaulting to <project>-<domain>")
		return common.ProjectDomain
	}
}

func NewNamespaceMappingConfigurationProvider() interfaces.NamespaceMappingConfiguration {
	return &NamespaceMappingConfigurationProvider{}
}
