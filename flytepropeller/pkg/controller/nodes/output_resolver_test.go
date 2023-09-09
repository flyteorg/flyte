package nodes

import (
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestCreateAliasMap(t *testing.T) {
	{
		aliases := []v1alpha1.Alias{
			{Alias: core.Alias{Var: "x", Alias: "y"}},
		}
		m := CreateAliasMap(aliases)
		assert.Equal(t, map[string]string{
			"y": "x",
		}, m)
	}
	{
		var aliases []v1alpha1.Alias
		m := CreateAliasMap(aliases)
		assert.Equal(t, map[string]string{}, m)
	}
	{
		m := CreateAliasMap(nil)
		assert.Equal(t, map[string]string{}, m)
	}
}
