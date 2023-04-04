package interfaces

import (
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/config"
)

type TierName = string

// Just incrementally start using mockery, replace with -all when working on https://github.com/flyteorg/flyte/issues/149
//go:generate mockery -name QualityOfServiceConfiguration -output=mocks -case=underscore

type QualityOfServiceSpec struct {
	QueueingBudget config.Duration `json:"queueingBudget"`
}

type QualityOfServiceConfig struct {
	TierExecutionValues map[TierName]QualityOfServiceSpec `json:"tierExecutionValues"`
	DefaultTiers        map[DomainName]TierName           `json:"defaultTiers"`
}

type QualityOfServiceConfiguration interface {
	GetTierExecutionValues() map[core.QualityOfService_Tier]core.QualityOfServiceSpec
	GetDefaultTiers() map[DomainName]core.QualityOfService_Tier
}
