package plugin

import (
	"context"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

type GetMutableAttributesInput struct {
	ClusterAssignment *admin.ClusterAssignment
}

type ProjectConfigurationPlugin interface {
	GetMutableAttributes(ctx context.Context, input *GetMutableAttributesInput) (sets.Set[admin.MatchableResource], error)
}

type DefaultProjectConfigurationPlugin struct {
	mutableAttributes sets.Set[admin.MatchableResource]
}

func (m *DefaultProjectConfigurationPlugin) GetMutableAttributes(ctx context.Context, input *GetMutableAttributesInput) (sets.Set[admin.MatchableResource], error) {
	return m.mutableAttributes, nil
}

func NewDefaultProjectConfigurationPlugin() *DefaultProjectConfigurationPlugin {
	// In OSS, all attributes are mutable
	mutableAttributes := sets.Set[admin.MatchableResource]{}
	for _, v := range admin.MatchableResource_value {
		mutableAttributes.Insert(admin.MatchableResource(v))
	}
	return &DefaultProjectConfigurationPlugin{
		mutableAttributes: mutableAttributes,
	}
}
