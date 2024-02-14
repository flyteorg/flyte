package testutils

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

// Convenience method to wrap verbose boilerplate for initializing a PluginOverrides MatchingAttributes.
func GetPluginOverridesAttributes(vals map[string][]string) *admin.MatchingAttributes {
	overrides := make([]*admin.PluginOverride, 0, len(vals))
	for taskType, pluginIDs := range vals {
		overrides = append(overrides, &admin.PluginOverride{
			TaskType: taskType,
			PluginId: pluginIDs,
		})
	}
	return &admin.MatchingAttributes{
		Target: &admin.MatchingAttributes_PluginOverrides{
			PluginOverrides: &admin.PluginOverrides{
				Overrides: overrides,
			},
		},
	}
}

func AssertProtoEqual(t testing.TB, expected, actual proto.Message, msgAndArgs ...any) bool {
	if assert.True(t, proto.Equal(expected, actual), msgAndArgs...) {
		return true
	}

	t.Logf("Expected: %v", expected)
	t.Logf("Actual  : %v", actual)

	return false
}
