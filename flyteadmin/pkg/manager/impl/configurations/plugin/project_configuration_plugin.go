package plugin

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

type GetAttributeIsMutable struct {
	ClusterAssignment *admin.ClusterAssignment
	ConfigurationID   *admin.ConfigurationID
}

type ProjectConfigurationPlugin interface {
	GetAttributeIsMutable(ctx context.Context, input *GetAttributeIsMutable) (map[admin.MatchableResource]*admin.AttributeIsMutable, error)
}

type DefaultProjectConfigurationPlugin struct {
	attributeIsMutable map[admin.MatchableResource]*admin.AttributeIsMutable
}

func (m *DefaultProjectConfigurationPlugin) GetAttributeIsMutable(ctx context.Context, input *GetAttributeIsMutable) (map[admin.MatchableResource]*admin.AttributeIsMutable, error) {
	return m.attributeIsMutable, nil
}

func NewDefaultProjectConfigurationPlugin() *DefaultProjectConfigurationPlugin {
	// In OSS, all attributes are mutable
	attributeIsMutable := map[admin.MatchableResource]*admin.AttributeIsMutable{}
	for _, v := range admin.MatchableResource_value {
		attributeIsMutable[admin.MatchableResource(v)] = &admin.AttributeIsMutable{Value: true}
	}
	return &DefaultProjectConfigurationPlugin{
		attributeIsMutable: attributeIsMutable,
	}
}
