package clusterresourceattribute

import (
	"testing"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/stretchr/testify/assert"
)

func TestFileConfig(t *testing.T) {
	clusterAttrFileConfig := AttrFileConfig{
		Project: "dummyProject",
		Domain:  "dummyDomain",
		ClusterResourceAttributes: &admin.ClusterResourceAttributes{
			Attributes: map[string]string{"foo": "bar"},
		},
	}
	matchingAttr := &admin.MatchingAttributes{
		Target: &admin.MatchingAttributes_ClusterResourceAttributes{
			ClusterResourceAttributes: clusterAttrFileConfig.ClusterResourceAttributes,
		},
	}
	t.Run("decorate", func(t *testing.T) {
		assert.Equal(t, matchingAttr, clusterAttrFileConfig.Decorate())
	})

	t.Run("decorate", func(t *testing.T) {
		clusterAttrFileConfigNew := AttrFileConfig{
			Project: "dummyProject",
			Domain:  "dummyDomain",
		}
		clusterAttrFileConfigNew.UnDecorate(matchingAttr)
		assert.Equal(t, clusterAttrFileConfig, clusterAttrFileConfigNew)
	})
	t.Run("get project domain workflow", func(t *testing.T) {
		clusterAttrFileConfigNew := AttrFileConfig{
			Project:  "dummyProject",
			Domain:   "dummyDomain",
			Workflow: "workflow",
		}
		assert.Equal(t, "dummyProject", clusterAttrFileConfigNew.GetProject())
		assert.Equal(t, "dummyDomain", clusterAttrFileConfigNew.GetDomain())
		assert.Equal(t, "workflow", clusterAttrFileConfigNew.GetWorkflow())
	})
}
