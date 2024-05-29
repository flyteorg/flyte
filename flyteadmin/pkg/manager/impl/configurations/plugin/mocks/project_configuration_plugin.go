package mocks

import (
	"context"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/configurations/plugin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

type GetMutableAttributesFunc func(ctx context.Context, input *plugin.GetMutableAttributesInput) (sets.Set[admin.MatchableResource], error)

type MockProjectConfigurationPlugin struct {
	getMutableAttributesFunc GetMutableAttributesFunc
}

func (m *MockProjectConfigurationPlugin) SetGetMutableAttributesCallback(getMutableAttributesFunc GetMutableAttributesFunc) {
	m.getMutableAttributesFunc = getMutableAttributesFunc
}

func (m *MockProjectConfigurationPlugin) GetMutableAttributes(ctx context.Context, input *plugin.GetMutableAttributesInput) (sets.Set[admin.MatchableResource], error) {
	if m.getMutableAttributesFunc != nil {
		return m.getMutableAttributesFunc(ctx, input)
	}
	return nil, nil
}
