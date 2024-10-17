package mocks

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/configurations/plugin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

type GetAttributeIsMutableFunc func(ctx context.Context, input *plugin.GetAttributeIsMutable) (map[admin.MatchableResource]*admin.AttributeIsMutable, error)

type MockProjectConfigurationPlugin struct {
	getAttributeIsMutableFunc GetAttributeIsMutableFunc
}

func (m *MockProjectConfigurationPlugin) SetGetAttributeIsMutableCallback(getAttributeIsMutableFunc GetAttributeIsMutableFunc) {
	m.getAttributeIsMutableFunc = getAttributeIsMutableFunc
}

func (m *MockProjectConfigurationPlugin) GetAttributeIsMutable(ctx context.Context, input *plugin.GetAttributeIsMutable) (map[admin.MatchableResource]*admin.AttributeIsMutable, error) {
	if m.getAttributeIsMutableFunc != nil {
		return m.getAttributeIsMutableFunc(ctx, input)
	}
	return nil, nil
}
