package v1alpha1_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytestdlib/storage"
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
	expectedSecurityContext := v1alpha1.SecurityContext{
		core.SecurityContext{
			RunAs: &core.Identity{
				IamRole:           "abc-def",
				K8SServiceAccount: "service-account",
				Oauth2Client: &core.OAuth2Client{
					ClientId: "client-id",
					ClientSecret: &core.Secret{
						Group:        "group",
						GroupVersion: "group-version",
						Key:          "key",
					},
				},
				ExecutionIdentity: "execution-identity",
			},
		},
	}
	securityContext := w.GetSecurityContext()
	assert.True(t, proto.Equal(&securityContext, &expectedSecurityContext))
}

func TestWrappedValuesInFlyteWorkflow(t *testing.T) {
	j, err := ReadYamlFileAsJSON("testdata/workflowspec.yaml")
	assert.NoError(t, err)
	w := &v1alpha1.FlyteWorkflow{}
	err = json.Unmarshal(j, w)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	wCopy := w.DeepCopy()

	wSecurityContext := w.GetSecurityContext()
	wCopySecurityContext := wCopy.GetSecurityContext()
	assert.True(t, proto.Equal(&wSecurityContext, &wCopySecurityContext))

}

func TestWorkflow(t *testing.T) {
	// Instantiate a new workflow, but be as comprehensive as possible.
	// This is a bit of a pain, but it's the only way to ensure that the defaults are set correctly.
	w := &v1alpha1.FlyteWorkflow{
		WorkflowSpec: &v1alpha1.WorkflowSpec{
			ID:    v1alpha1.WorkflowID("some-id"),
			Nodes: map[v1alpha1.NodeID]*v1alpha1.NodeSpec{},
			Connections: v1alpha1.Connections{
				Downstream: map[v1alpha1.NodeID][]v1alpha1.NodeID{},
				Upstream:   map[v1alpha1.NodeID][]v1alpha1.NodeID{},
			},
		},
		WorkflowMeta: &v1alpha1.WorkflowMeta{
			EventVersion: 42,
		},
		Inputs: &v1alpha1.Inputs{
			&core.LiteralMap{
				Literals: map[string]*core.Literal{
					"input1": {
						Value: &core.Literal_Scalar{
							Scalar: &core.Scalar{
								Value: &core.Scalar_Primitive{
									Primitive: &core.Primitive{
										Value: &core.Primitive_Integer{
											Integer: 1,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		ExecutionID: v1alpha1.WorkflowExecutionIdentifier{
			&core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
				Org:     "org",
			},
		},
		Tasks: map[v1alpha1.TaskID]*v1alpha1.TaskSpec{
			"task1": &v1alpha1.TaskSpec{
				&core.TaskTemplate{
					Id: &core.Identifier{
						ResourceType: core.ResourceType_TASK,
						Project:      "project",
						Domain:       "domain",
						Name:         "name",
						Version:      "version",
					},
					Interface: &core.TypedInterface{
						Inputs: &core.VariableMap{
							Variables: map[string]*core.Variable{
								"input1": {
									Type: &core.LiteralType{
										Type: &core.LiteralType_Simple{
											Simple: core.SimpleType_INTEGER,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		ActiveDeadlineSeconds: proto.Int64(42),
		NodeDefaults: v1alpha1.NodeDefaults{
			Interruptible: true,
		},
		AcceptedAt: &metav1.Time{
			Time: time.Now(),
		},
		SecurityContext: v1alpha1.SecurityContext{
			core.SecurityContext{
				RunAs: &core.Identity{
					IamRole:           "abc-def",
					K8SServiceAccount: "service-account",
					Oauth2Client: &core.OAuth2Client{
						ClientId: "client-id",
						ClientSecret: &core.Secret{
							Group:        "group",
							GroupVersion: "group-version",
							Key:          "key",
						},
					},
					ExecutionIdentity: "execution-identity",
				},
			},
		},
		Status: v1alpha1.WorkflowStatus{
			Phase: v1alpha1.WorkflowPhase(41),
			StartedAt: &metav1.Time{
				Time: time.Now(),
			},
			StoppedAt: &metav1.Time{
				Time: time.Now(),
			},
			LastUpdatedAt: &metav1.Time{
				Time: time.Now(),
			},
			Message:         "message",
			DataDir:         storage.DataReference("data-dir"),
			OutputReference: storage.DataReference("output-reference"),
			NodeStatus:      map[string]*v1alpha1.NodeStatus{},
			FailedAttempts:  32,
			Error: &v1alpha1.ExecutionError{
				&core.ExecutionError{
					Code:    "code",
					Message: "message",
				},
			},
		},
		RawOutputDataConfig: v1alpha1.RawOutputDataConfig{
			&admin.RawOutputDataConfig{
				OutputLocationPrefix: "output-location-prefix",
			},
		},
		ExecutionConfig: v1alpha1.ExecutionConfig{
			TaskPluginImpls: map[string]v1alpha1.TaskPluginOverride{},
			MaxParallelism:  42,
			RecoveryExecution: v1alpha1.WorkflowExecutionIdentifier{
				&core.WorkflowExecutionIdentifier{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
					Org:     "org",
				},
			},
		},
		DataReferenceConstructor: storage.URLPathConstructor{},
		WorkflowClosureReference: storage.DataReference("data-reference"),
	}

	wCopy := w.DeepCopy()

	// Test that the pointers are different.
	assert.NotSame(t, w, wCopy)

	assert.True(t, w.WorkflowSpec != wCopy.WorkflowSpec)
	assert.Equal(t, w.WorkflowSpec, wCopy.WorkflowSpec)
	assert.True(t, w.WorkflowMeta != wCopy.WorkflowMeta)
	assert.Equal(t, w.WorkflowMeta, wCopy.WorkflowMeta)
	assert.True(t, w.Inputs != wCopy.Inputs)
	assert.Equal(t, w.Inputs, wCopy.Inputs)

	wSecurityContext := w.GetSecurityContext()
	wCopySecurityContext := wCopy.GetSecurityContext()
	assert.True(t, proto.Equal(&wSecurityContext, &wCopySecurityContext))
}

func TestWorkflowIsInterruptible(t *testing.T) {
	w := &v1alpha1.FlyteWorkflow{}

	// no execution spec or metadata defined -> interruptible defaults to false
	assert.False(t, w.IsInterruptible())

	// marked as interruptible via execution config (e.g. for a single execution)
	execConfigInterruptible := true
	w.ExecutionConfig.Interruptible = &execConfigInterruptible
	assert.True(t, w.IsInterruptible())

	// marked as not interruptible via execution config, overwriting node defaults
	execConfigInterruptible = false
	w.NodeDefaults.Interruptible = true
	assert.False(t, w.IsInterruptible())

	// marked as interruptible via execution config, overwriting node defaults
	execConfigInterruptible = true
	w.NodeDefaults.Interruptible = false
	assert.True(t, w.IsInterruptible())

	// interruptible flag retrieved from node defaults (e.g. workflow definition), no execution config override
	w.ExecutionConfig.Interruptible = nil
	w.NodeDefaults.Interruptible = true
	assert.True(t, w.IsInterruptible())
}

func TestWrappedInputsDeepCopy(t *testing.T) {
	// 1. Setup proto
	litMap := core.LiteralMap{
		Literals: map[string]*core.Literal{
			"p1": {
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Primitive{
							Primitive: &core.Primitive{
								Value: &core.Primitive_Integer{
									Integer: 1,
								},
							},
						},
					},
				},
			},
		},
	}

	// 2. Define wrapper
	inputs := v1alpha1.Inputs{
		&litMap,
	}

	// 3. Deep copy
	inputsCopy := inputs.DeepCopy()

	// 4. Assert that pointers are different
	assert.True(t, inputs.LiteralMap != inputsCopy.LiteralMap)

	// 5. Assert that the content is the same
	assert.True(t, proto.Equal(inputs.LiteralMap, inputsCopy.LiteralMap))
}
