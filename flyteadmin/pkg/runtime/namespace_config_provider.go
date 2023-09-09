package runtime

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/logger"
)

const (
	namespaceMappingKey = "namespace_mapping"
	defaultTemplate     = "{{ project }}-{{ domain }}"
)

var namespaceMappingConfig = config.MustRegisterSection(namespaceMappingKey, &interfaces.NamespaceMappingConfig{
	Template: defaultTemplate,
})

type NamespaceMappingConfigurationProvider struct{}

func (p *NamespaceMappingConfigurationProvider) GetNamespaceTemplate() string {
	var template string
	if namespaceMappingConfig != nil && namespaceMappingConfig.GetConfig() != nil {
		template = namespaceMappingConfig.GetConfig().(*interfaces.NamespaceMappingConfig).Template
		if len(namespaceMappingConfig.GetConfig().(*interfaces.NamespaceMappingConfig).Mapping) > 0 {
			logger.Errorf(context.TODO(), "Using `mapping` in namespace configs is deprecated. "+
				"Please use a custom string template like `{{ project }}-{{ domain }}` instead")
		}
	}
	if len(template) == 0 {
		logger.Infof(context.TODO(), "No namespace template specified in config. Using [%+s] by default", defaultTemplate)
		template = defaultTemplate
	}
	return template
}

func NewNamespaceMappingConfigurationProvider() interfaces.NamespaceMappingConfiguration {
	return &NamespaceMappingConfigurationProvider{}
}
