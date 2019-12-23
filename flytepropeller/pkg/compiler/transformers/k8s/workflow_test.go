package k8s

import (
	"testing"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytepropeller/pkg/compiler/common"
	"github.com/lyft/flytepropeller/pkg/compiler/errors"
	"github.com/lyft/flytepropeller/pkg/utils"
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
	assert.Equal(t, "wf-1", wf.Labels[WorkflowNameLabel])
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

	intLiteral, err := utils.MakePrimitiveLiteral(123)
	assert.NoError(t, err)
	stringLiteral, err := utils.MakePrimitiveLiteral("hello")
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

func TestGenerateName(t *testing.T) {
	t.Run("Invalid params", func(t *testing.T) {
		_, _, _, err := generateName(nil, nil)
		assert.Error(t, err)
	})

	t.Run("wfID full", func(t *testing.T) {
		name, generateName, _, err := generateName(&core.Identifier{
			Name:    "myworkflow",
			Project: "myproject",
			Domain:  "development",
		}, nil)

		assert.NoError(t, err)
		assert.Empty(t, name)
		assert.Equal(t, "myproject-development-myworkflow-", generateName)
	})

	t.Run("wfID missing project domain", func(t *testing.T) {
		name, generateName, _, err := generateName(&core.Identifier{
			Name: "myworkflow",
		}, nil)

		assert.NoError(t, err)
		assert.Empty(t, name)
		assert.Equal(t, "myworkflow-", generateName)
	})

	t.Run("wfID too long", func(t *testing.T) {
		name, generateName, _, err := generateName(&core.Identifier{
			Name:    "workflowsomethingsomethingsomething",
			Project: "myproject",
			Domain:  "development",
		}, nil)

		assert.NoError(t, err)
		assert.Empty(t, name)
		assert.Equal(t, "myproject-development-workflowso-", generateName)
	})

	t.Run("execID full", func(t *testing.T) {
		name, generateName, _, err := generateName(nil, &core.WorkflowExecutionIdentifier{
			Name:    "myexecution",
			Project: "myproject",
			Domain:  "development",
		})

		assert.NoError(t, err)
		assert.Empty(t, generateName)
		assert.Equal(t, "myexecution", name)
	})

	t.Run("execID missing project domain", func(t *testing.T) {
		name, generateName, _, err := generateName(nil, &core.WorkflowExecutionIdentifier{
			Name: "myexecution",
		})

		assert.NoError(t, err)
		assert.Empty(t, generateName)
		assert.Equal(t, "myexecution", name)
	})
}
