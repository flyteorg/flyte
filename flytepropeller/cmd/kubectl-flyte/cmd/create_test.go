package cmd

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghodss/yaml"

	"github.com/flyteorg/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/compiler"
	"github.com/flyteorg/flytepropeller/pkg/compiler/common"
	"github.com/flyteorg/flytepropeller/pkg/compiler/transformers/k8s"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

var update = flag.Bool("update", false, "Update .golden files")

func init() {
}

func createEmptyVariableMap() *core.VariableMap {
	res := &core.VariableMap{
		Variables: map[string]*core.Variable{},
	}
	return res
}

func createVariableMap(variableMap map[string]*core.Variable) *core.VariableMap {
	res := &core.VariableMap{
		Variables: variableMap,
	}
	return res
}

func TestCreate(t *testing.T) {
	t.Run("Generate simple workflow", generateSimpleWorkflow)
	t.Run("Generate workflow with inputs", generateWorkflowWithInputs)
	t.Run("Compile", testCompile)
}

func generateSimpleWorkflow(t *testing.T) {
	if !*update {
		t.SkipNow()
	}

	t.Log("Generating golden files.")
	closure := core.WorkflowClosure{
		Workflow: &core.WorkflowTemplate{
			Id: &core.Identifier{Name: "workflow-id-123"},
			Interface: &core.TypedInterface{
				Inputs: createEmptyVariableMap(),
			},
			Nodes: []*core.Node{
				{
					Id: "node-1",
					Target: &core.Node_TaskNode{
						TaskNode: &core.TaskNode{
							Reference: &core.TaskNode_ReferenceId{
								ReferenceId: &core.Identifier{Name: "task-1"},
							},
						},
					},
				},
				{
					Id: "node-2",
					Target: &core.Node_TaskNode{
						TaskNode: &core.TaskNode{
							Reference: &core.TaskNode_ReferenceId{
								ReferenceId: &core.Identifier{Name: "task-2"},
							},
						},
					},
				},
			},
		},
		Tasks: []*core.TaskTemplate{
			{
				Id: &core.Identifier{Name: "task-1"},
				Interface: &core.TypedInterface{
					Inputs: createEmptyVariableMap(),
				},
				Target: &core.TaskTemplate_Container{
					Container: &core.Container{
						Image:   "myflyteimage:latest",
						Command: []string{"execute-task"},
						Args:    []string{"testArg"},
					},
				},
			},
			{
				Id: &core.Identifier{Name: "task-2"},
				Interface: &core.TypedInterface{
					Inputs: createEmptyVariableMap(),
				},
				Target: &core.TaskTemplate_Container{
					Container: &core.Container{
						Image:   "myflyteimage:latest",
						Command: []string{"execute-task"},
						Args:    []string{"testArg"},
					},
				},
			},
		},
	}

	marshaller := &jsonpb.Marshaler{}
	s, err := marshaller.MarshalToString(&closure)
	assert.NoError(t, err)
	assert.NoError(t, ioutil.WriteFile(filepath.Join("testdata", "workflow.json.golden"), []byte(s), os.ModePerm))

	m := map[string]interface{}{}
	err = json.Unmarshal([]byte(s), &m)
	assert.NoError(t, err)

	b, err := yaml.Marshal(m)
	assert.NoError(t, err)
	assert.NoError(t, ioutil.WriteFile(filepath.Join("testdata", "workflow.yaml.golden"), b, os.ModePerm))

	raw, err := proto.Marshal(&closure)
	assert.NoError(t, err)
	assert.NoError(t, ioutil.WriteFile(filepath.Join("testdata", "workflow.pb.golden"), raw, os.ModePerm))
}

func generateWorkflowWithInputs(t *testing.T) {
	if !*update {
		t.SkipNow()
	}

	t.Log("Generating golden files.")
	closure := core.WorkflowClosure{
		Workflow: &core.WorkflowTemplate{
			Id: &core.Identifier{Name: "workflow-with-inputs"},
			Interface: &core.TypedInterface{
				Inputs: createVariableMap(map[string]*core.Variable{
					"x": {
						Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}},
					},
					"y": {
						Type: &core.LiteralType{
							Type: &core.LiteralType_CollectionType{
								CollectionType: &core.LiteralType{
									Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
								},
							},
						}},
				}),
			},
			Nodes: []*core.Node{
				{
					Id: "node-1",
					Inputs: []*core.Binding{
						{Var: "x", Binding: &core.BindingData{Value: &core.BindingData_Promise{Promise: &core.OutputReference{Var: "x"}}}},
						{Var: "y", Binding: &core.BindingData{Value: &core.BindingData_Promise{Promise: &core.OutputReference{Var: "y"}}}},
					},
					Target: &core.Node_TaskNode{
						TaskNode: &core.TaskNode{
							Reference: &core.TaskNode_ReferenceId{
								ReferenceId: &core.Identifier{Name: "task-1"},
							},
						},
					},
				},
				{
					Id: "node-2",
					Target: &core.Node_TaskNode{
						TaskNode: &core.TaskNode{
							Reference: &core.TaskNode_ReferenceId{
								ReferenceId: &core.Identifier{Name: "task-2"},
							},
						},
					},
				},
			},
		},
		Tasks: []*core.TaskTemplate{
			{
				Id: &core.Identifier{Name: "task-1"},
				Interface: &core.TypedInterface{
					Inputs: createVariableMap(map[string]*core.Variable{
						"x": {
							Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}},
						},
						"y": {
							Type: &core.LiteralType{
								Type: &core.LiteralType_CollectionType{
									CollectionType: &core.LiteralType{
										Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
									},
								},
							}},
					}),
				},
				Target: &core.TaskTemplate_Container{
					Container: &core.Container{
						Image:   "myflyteimage:latest",
						Command: []string{"execute-task"},
						Args:    []string{"testArg"},
						Resources: &core.Resources{
							Requests: []*core.Resources_ResourceEntry{
								{Name: core.Resources_CPU, Value: "2"},
								{Name: core.Resources_MEMORY, Value: "2048Mi"},
							},
						},
					},
				},
			},
			{
				Id: &core.Identifier{Name: "task-2"},
				Interface: &core.TypedInterface{
					Inputs: createEmptyVariableMap(),
				},
				Target: &core.TaskTemplate_Container{
					Container: &core.Container{
						Image:   "myflyteimage:latest",
						Command: []string{"execute-task"},
						Args:    []string{"testArg"},
					},
				},
			},
		},
	}

	marshalGolden(t, &closure, "workflow_w_inputs")
	sampleInputs := core.LiteralMap{
		Literals: map[string]*core.Literal{
			"x": coreutils.MustMakeLiteral(2),
			"y": coreutils.MustMakeLiteral([]interface{}{"val1", "val2", "val3"}),
		},
	}

	marshalGolden(t, &sampleInputs, "inputs")
}

func marshalGolden(t *testing.T, message proto.Message, filename string) {
	marshaller := &jsonpb.Marshaler{}
	s, err := marshaller.MarshalToString(message)
	assert.NoError(t, err)
	assert.NoError(t, ioutil.WriteFile(filepath.Join("testdata", filename+".json.golden"), []byte(s), os.ModePerm))

	m := map[string]interface{}{}
	err = json.Unmarshal([]byte(s), &m)
	assert.NoError(t, err)

	b, err := yaml.Marshal(m)
	assert.NoError(t, err)
	assert.NoError(t, ioutil.WriteFile(filepath.Join("testdata", filename+".yaml.golden"), b, os.ModePerm))

	raw, err := proto.Marshal(message)
	assert.NoError(t, err)
	assert.NoError(t, ioutil.WriteFile(filepath.Join("testdata", filename+".pb.golden"), raw, os.ModePerm))
}

func testCompile(t *testing.T) {
	f := func(t *testing.T, filePath, format string) {
		raw, err := ioutil.ReadFile(filepath.Join("testdata", filePath))
		assert.NoError(t, err)
		wf := &core.WorkflowClosure{}
		err = unmarshal(raw, format, wf)
		assert.NoError(t, err)
		assert.NotNil(t, wf)
		assert.Equal(t, 2, len(wf.Tasks))
		if len(wf.Tasks) == 2 {
			c := wf.Tasks[0].GetContainer()
			assert.NotNil(t, c)
			compiledTasks, err := compileTasks(wf.Tasks)
			assert.NoError(t, err)
			compiledWf, err := compiler.CompileWorkflow(wf.Workflow, []*core.WorkflowTemplate{}, compiledTasks, []common.InterfaceProvider{})
			assert.NoError(t, err)
			_, err = k8s.BuildFlyteWorkflow(compiledWf, nil, nil, "")
			assert.NoError(t, err)
		}
	}

	t.Run("yaml", func(t *testing.T) {
		f(t, "workflow.yaml.golden", formatYaml)
	})

	t.Run("json", func(t *testing.T) {
		f(t, "workflow.json.golden", formatJSON)
	})

	t.Run("proto", func(t *testing.T) {
		f(t, "workflow.pb.golden", formatProto)
	})
}
