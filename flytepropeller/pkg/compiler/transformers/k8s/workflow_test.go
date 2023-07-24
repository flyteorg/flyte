package k8s

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/flyteorg/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/compiler/common"
	"github.com/flyteorg/flytepropeller/pkg/compiler/errors"
	"github.com/golang/protobuf/jsonpb"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/sets"
)

func createSampleMockWorkflow() *mockWorkflow {
	return &mockWorkflow{
		tasks: common.TaskIndex{
			"task_1": &mockTask{
				task: &core.TaskTemplate{
					Id: &core.Identifier{Name: "task_1"},
				},
			},
		},
		nodes: common.NodeIndex{
			"node_1": &mockNode{
				upstream: []string{common.StartNodeID},
				inputs:   []*core.Binding{},
				task:     &mockTask{},
				id:       "node_1",
				Node:     createNodeWithTask(),
			},
			common.StartNodeID: &mockNode{
				id:   common.StartNodeID,
				Node: &core.Node{},
			},
		},
		//failureNode: &mockNode{
		//	id: "node_1",
		//},
		downstream: common.StringAdjacencyList{
			common.StartNodeID: sets.NewString("node_1"),
		},
		upstream: common.StringAdjacencyList{
			"node_1": sets.NewString(common.StartNodeID),
		},
		CompiledWorkflow: &core.CompiledWorkflow{
			Template: &core.WorkflowTemplate{
				Id: &core.Identifier{Name: "wf_1"},
				Nodes: []*core.Node{
					createNodeWithTask(),
					{
						Id: common.StartNodeID,
					},
				},
				MetadataDefaults: &core.WorkflowMetadataDefaults{
					Interruptible: true,
				},
			},
			Connections: &core.ConnectionSet{
				Downstream: map[string]*core.ConnectionSet_IdList{
					common.StartNodeID: {
						Ids: []string{"node_1"},
					},
				},
			},
		},
	}
}

func TestWorkflowIDAsString(t *testing.T) {
	assert.Equal(t, "project:domain:name", WorkflowIDAsString(&core.Identifier{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
		Version: "v1",
	}))

	assert.Equal(t, ":domain:name", WorkflowIDAsString(&core.Identifier{
		Domain:  "domain",
		Name:    "name",
		Version: "v1",
	}))

	assert.Equal(t, "project::name", WorkflowIDAsString(&core.Identifier{
		Project: "project",
		Name:    "name",
		Version: "v1",
	}))

	assert.Equal(t, "project:domain:name", WorkflowIDAsString(&core.Identifier{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
	}))
}

func TestBuildFlyteWorkflow(t *testing.T) {
	w := createSampleMockWorkflow()

	errors.SetConfig(errors.Config{IncludeSource: true})
	wf, err := BuildFlyteWorkflow(
		&core.CompiledWorkflowClosure{
			Primary: w.GetCoreWorkflow(),
			Tasks: []*core.CompiledTask{
				{
					Template: &core.TaskTemplate{
						Id: &core.Identifier{Name: "ref_1"},
					},
				},
			},
		},
		nil, nil, "")
	assert.Equal(t, true, wf.NodeDefaults.Interruptible)
	assert.True(t, *wf.WorkflowSpec.Nodes["n_1"].Interruptible)
	assert.Nil(t, wf.WorkflowSpec.Nodes[common.StartNodeID].Interruptible)
	assert.Equal(t, "wf-1", wf.Labels[WorkflowNameLabel])
	assert.Equal(t, "4", wf.Labels[ShardKeyLabel])
	assert.NoError(t, err)
	assert.NotNil(t, wf)
	errors.SetConfig(errors.Config{})
}

func TestBuildFlyteWorkflow_withInputs(t *testing.T) {
	w := createSampleMockWorkflow()

	startNode := w.GetNodes()[common.StartNodeID].(*mockNode)
	vars := []*core.Variable{
		{
			Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}},
		},
		{
			Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
		},
	}

	w.Template.Interface = &core.TypedInterface{
		Inputs: &core.VariableMap{
			Variables: map[string]*core.Variable{
				"x": vars[0],
				"y": vars[1],
			},
		},
	}

	startNode.iface = &core.TypedInterface{
		Outputs: &core.VariableMap{
			Variables: map[string]*core.Variable{
				"x": vars[0],
				"y": vars[1],
			},
		},
	}

	intLiteral, err := coreutils.MakePrimitiveLiteral(123)
	assert.NoError(t, err)
	stringLiteral, err := coreutils.MakePrimitiveLiteral("hello")
	assert.NoError(t, err)
	inputs := &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"x": intLiteral,
			"y": stringLiteral,
		},
	}

	errors.SetConfig(errors.Config{IncludeSource: true})
	wf, err := BuildFlyteWorkflow(
		&core.CompiledWorkflowClosure{
			Primary: w.GetCoreWorkflow(),
			Tasks: []*core.CompiledTask{
				{
					Template: &core.TaskTemplate{
						Id: &core.Identifier{Name: "ref_1"},
					},
				},
			},
		},
		inputs, nil, "")
	assert.NoError(t, err)
	assert.NotNil(t, wf)
	errors.SetConfig(errors.Config{})

	assert.Equal(t, 2, len(wf.Inputs.Literals))
	assert.Equal(t, int64(123), wf.Inputs.Literals["x"].GetScalar().GetPrimitive().GetInteger())
}

func TestBuildFlyteWorkflow_withUnionInputs(t *testing.T) {
	w := createSampleMockWorkflow()

	startNode := w.GetNodes()[common.StartNodeID].(*mockNode)
	strType := core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}, Structure: &core.TypeStructure{Tag: "str"}}
	floatType := core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_FLOAT}, Structure: &core.TypeStructure{Tag: "float"}}
	vars := []*core.Variable{
		{
			Type: &core.LiteralType{Type: &core.LiteralType_UnionType{UnionType: &core.UnionType{Variants: []*core.LiteralType{&strType, &floatType}}}},
		},
		{
			Type: &core.LiteralType{Type: &core.LiteralType_UnionType{UnionType: &core.UnionType{Variants: []*core.LiteralType{&strType, &floatType}}}},
		},
	}

	w.Template.Interface = &core.TypedInterface{
		Inputs: &core.VariableMap{
			Variables: map[string]*core.Variable{
				"x": vars[0],
				"y": vars[1],
			},
		},
	}

	startNode.iface = &core.TypedInterface{
		Outputs: &core.VariableMap{
			Variables: map[string]*core.Variable{
				"x": vars[0],
				"y": vars[1],
			},
		},
	}

	stringLiteral, err := coreutils.MakePrimitiveLiteral("hello")
	assert.NoError(t, err)
	floatLiteral, err := coreutils.MakePrimitiveLiteral(1.0)
	assert.NoError(t, err)
	inputs := &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"x": {Value: &core.Literal_Scalar{Scalar: &core.Scalar{Value: &core.Scalar_Union{Union: &core.Union{Value: floatLiteral, Type: &floatType}}}}},
			"y": {Value: &core.Literal_Scalar{Scalar: &core.Scalar{Value: &core.Scalar_Union{Union: &core.Union{Value: stringLiteral, Type: &strType}}}}},
		},
	}

	errors.SetConfig(errors.Config{IncludeSource: true})
	wf, err := BuildFlyteWorkflow(
		&core.CompiledWorkflowClosure{
			Primary: w.GetCoreWorkflow(),
			Tasks: []*core.CompiledTask{
				{
					Template: &core.TaskTemplate{
						Id: &core.Identifier{Name: "ref_1"},
					},
				},
			},
		},
		inputs, nil, "")
	assert.NoError(t, err)
	assert.NotNil(t, wf)
	errors.SetConfig(errors.Config{})

	assert.Equal(t, 2, len(wf.Inputs.Literals))
	assert.Equal(t, 1.0, wf.Inputs.Literals["x"].GetScalar().GetUnion().GetValue().GetScalar().GetPrimitive().GetFloatValue())
	assert.Equal(t, "hello", wf.Inputs.Literals["y"].GetScalar().GetUnion().GetValue().GetScalar().GetPrimitive().GetStringValue())
}

func TestGenerateName(t *testing.T) {
	t.Run("Invalid params", func(t *testing.T) {
		_, _, _, _, _, err := generateName(nil, nil)
		assert.Error(t, err)
	})

	t.Run("wfID full", func(t *testing.T) {
		name, generateName, _, project, domain, err := generateName(&core.Identifier{
			Name:    "myworkflow",
			Project: "myproject",
			Domain:  "development",
		}, nil)

		assert.NoError(t, err)
		assert.Empty(t, name)
		assert.Equal(t, "myproject-development-myworkflow-", generateName)
		assert.Equal(t, "myproject", project)
		assert.Equal(t, "development", domain)
	})

	t.Run("wfID missing project domain", func(t *testing.T) {
		name, generateName, _, project, domain, err := generateName(&core.Identifier{
			Name: "myworkflow",
		}, nil)

		assert.NoError(t, err)
		assert.Empty(t, name)
		assert.Equal(t, "myworkflow-", generateName)
		assert.Equal(t, "", project)
		assert.Equal(t, "", domain)
	})

	t.Run("wfID too long", func(t *testing.T) {
		name, generateName, _, project, domain, err := generateName(&core.Identifier{
			Name:    "workflowsomethingsomethingsomething",
			Project: "myproject",
			Domain:  "development",
		}, nil)

		assert.NoError(t, err)
		assert.Empty(t, name)
		assert.Equal(t, "myproject-development-workflowso-", generateName)
		assert.Equal(t, "myproject", project)
		assert.Equal(t, "development", domain)
	})

	t.Run("execID full", func(t *testing.T) {
		name, generateName, _, project, domain, err := generateName(nil, &core.WorkflowExecutionIdentifier{
			Name:    "myexecution",
			Project: "myproject",
			Domain:  "development",
		})

		assert.NoError(t, err)
		assert.Empty(t, generateName)
		assert.Equal(t, "myexecution", name)
		assert.Equal(t, "myproject", project)
		assert.Equal(t, "development", domain)
	})

	t.Run("execID missing project domain", func(t *testing.T) {
		name, generateName, _, project, domain, err := generateName(nil, &core.WorkflowExecutionIdentifier{
			Name: "myexecution",
		})

		assert.NoError(t, err)
		assert.Empty(t, generateName)
		assert.Equal(t, "myexecution", name)
		assert.Equal(t, "", project)
		assert.Equal(t, "", domain)
	})
}

func TestBuildFlyteWorkflow_withBranch(t *testing.T) {
	c, err := ioutil.ReadFile("testdata/compiled_closure_branch_nested.json")
	assert.NoError(t, err)

	r := bytes.NewReader(c)

	w := &core.CompiledWorkflowClosure{}
	assert.NoError(t, jsonpb.Unmarshal(r, w))

	assert.Len(t, w.Primary.Connections.Downstream, 2)
	ids := w.Primary.Connections.Downstream["start-node"]
	assert.Len(t, ids.Ids, 1)
	assert.Equal(t, ids.Ids[0], "n0")

	wf, err := BuildFlyteWorkflow(
		w,
		&core.LiteralMap{
			Literals: map[string]*core.Literal{
				"my_input": coreutils.MustMakePrimitiveLiteral(1.0),
			},
		}, &core.WorkflowExecutionIdentifier{
			Project: "p",
			Domain:  "d",
			Name:    "n",
		}, "ns")

	assert.NoError(t, err)
	assert.NotNil(t, wf)
	assert.Len(t, wf.Nodes, 8)

	s := sets.NewString("start-node", "end-node", "n0", "n0-n0", "n0-n1", "n0-n2", "n0-n0-n0", "n0-n0-n0-n0", "n0-n0-n0-n1")

	for _, n := range wf.Nodes {
		assert.True(t, s.Has(n.ID), "nodeId: %s for node: %s not found", n.ID, n.Name)
	}
}
