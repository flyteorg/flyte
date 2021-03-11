package runtime

import (
	"fmt"

	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/config"
	"github.com/golang/protobuf/ptypes"
)

const qualityOfServiceKey = "qualityOfService"

var qualityOfServiceConfig = config.MustRegisterSection(qualityOfServiceKey, &interfaces.QualityOfServiceConfig{})

// Implementation of an interfaces.QualityOfServiceConfiguration
type QualityOfServiceConfigProvider struct {
}

func (p *QualityOfServiceConfigProvider) GetTierExecutionValues() map[core.QualityOfService_Tier]core.QualityOfServiceSpec {
	tierExecutionValues := make(map[core.QualityOfService_Tier]core.QualityOfServiceSpec)
	if qualityOfServiceConfig != nil {
		values := qualityOfServiceConfig.GetConfig().(*interfaces.QualityOfServiceConfig).TierExecutionValues
		for tierName, spec := range values {
			tierExecutionValues[core.QualityOfService_Tier(core.QualityOfService_Tier_value[tierName])] =
				core.QualityOfServiceSpec{
					QueueingBudget: ptypes.DurationProto(spec.QueueingBudget.Duration),
				}
		}
	}
	return tierExecutionValues
}

func (p *QualityOfServiceConfigProvider) GetDefaultTiers() map[interfaces.DomainName]core.QualityOfService_Tier {
	defaultTiers := make(map[interfaces.DomainName]core.QualityOfService_Tier)
	if qualityOfServiceConfig != nil {
		tiers := qualityOfServiceConfig.GetConfig().(*interfaces.QualityOfServiceConfig).DefaultTiers
		for domainName, tierName := range tiers {
			defaultTiers[domainName] = core.QualityOfService_Tier(core.QualityOfService_Tier_value[tierName])
		}
	}
	return defaultTiers
}

func validateConfigValues() {
	if qualityOfServiceConfig != nil {
		values := qualityOfServiceConfig.GetConfig().(*interfaces.QualityOfServiceConfig).TierExecutionValues
		for tierName, spec := range values {
			_, err := ptypes.Duration(ptypes.DurationProto(spec.QueueingBudget.Duration))
			if err != nil {
				panic(fmt.Sprintf("Invalid duration [%+v] specified for %s", spec.QueueingBudget.Duration, tierName))
			}
		}
	}
}

func NewQualityOfServiceConfigProvider() interfaces.QualityOfServiceConfiguration {
	validateConfigValues()
	return &QualityOfServiceConfigProvider{}
}
