package v1alpha1_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/sets"
)

func TestMarshalUnmarshal_Connections(t *testing.T) {
	r, err := ioutil.ReadFile("testdata/connections.json")
	assert.NoError(t, err)
	o := v1alpha1.DeprecatedConnections{}
	err = json.Unmarshal(r, &o)
	assert.NoError(t, err)
	assert.Equal(t, map[v1alpha1.NodeID][]v1alpha1.NodeID{
		"n1": {"n2", "n3"},
		"n2": {"n4"},
		"n3": {"n4"},
		"n4": {"n5"},
	}, o.DownstreamEdges)
	assert.Equal(t, []v1alpha1.NodeID{"n1"}, o.UpstreamEdges["n2"])
	assert.Equal(t, []v1alpha1.NodeID{"n1"}, o.UpstreamEdges["n3"])
	assert.Equal(t, []v1alpha1.NodeID{"n4"}, o.UpstreamEdges["n5"])
	assert.True(t, sets.NewString(o.UpstreamEdges["n4"]...).Equal(sets.NewString("n2", "n3")))
}

func ReadYamlFileAsJSON(path string) ([]byte, error) {
	r, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return yaml.YAMLToJSON(r)
}

func TestWorkflowSpec(t *testing.T) {
	j, err := ReadYamlFileAsJSON("testdata/workflowspec.yaml")
	assert.NoError(t, err)
	w := &v1alpha1.FlyteWorkflow{}
	err = json.Unmarshal(j, w)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.NotNil(t, w.WorkflowSpec)
	assert.Nil(t, w.GetOnFailureNode())
	assert.Equal(t, 7, len(w.GetConnections().Downstream))
	assert.Equal(t, 8, len(w.GetConnections().Upstream))
}
