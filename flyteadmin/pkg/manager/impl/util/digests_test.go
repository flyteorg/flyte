package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes/duration"
	_struct "github.com/golang/protobuf/ptypes/struct"
	"github.com/stretchr/testify/assert"
)

var testLaunchPlanDigest = []byte{
	0x13, 0x50, 0xe9, 0xa, 0xb9, 0x13, 0xe2, 0x2a, 0x47, 0x2e, 0xf2, 0x37, 0xc, 0xbd, 0x9a, 0x4c, 0xa8, 0x52, 0x7e,
	0x87, 0x4a, 0xf6, 0x5e, 0x95, 0x22, 0x6a, 0x4f, 0xa1, 0xee, 0xf4, 0xc8, 0xd5}

var taskIdentifier = core.Identifier{
	ResourceType: core.ResourceType_TASK,
	Project:      "project",
	Domain:       "domain",
	Name:         "name",
	Version:      "version",
}

var compiledTask = &core.CompiledTask{
	Template: &core.TaskTemplate{
		Id:   &taskIdentifier,
		Type: "foo type",
		Metadata: &core.TaskMetadata{
			Timeout: &duration.Duration{
				Seconds: 60,
			},
		},
		Custom: &_struct.Struct{},
	},
}

var compiledTaskDigest = []byte{
	0x85, 0xb2, 0x84, 0xaa, 0x87, 0x26, 0xa1, 0x3e, 0xc2, 0x20, 0x53, 0x69, 0x82, 0x81, 0xb1, 0x3f, 0xd8, 0xa8, 0xa5,
	0xa, 0x22, 0x80, 0xb1, 0x8, 0x44, 0x53, 0xf3, 0xca, 0x60, 0x4, 0xf7, 0x6f}

var compiledWorkflowDigest = []byte{0xeb, 0x66, 0x44, 0xe8, 0x1c, 0xa8, 0x51, 0x7d, 0x3f, 0x33, 0xf0, 0x77, 0x95, 0x24, 0x84, 0xc2, 0xbe, 0x79, 0xcd, 0xa, 0x63, 0x42, 0x2c, 0xf8, 0xfd, 0x35, 0x37, 0xb6, 0xc7, 0x89, 0xc5, 0x83}

func getLaunchPlan() *admin.LaunchPlan {
	return &admin.LaunchPlan{
		Closure: &admin.LaunchPlanClosure{
			ExpectedInputs: &core.ParameterMap{
				Parameters: map[string]*core.Parameter{
					"foo": {},
					"bar": {},
				},
			},
			ExpectedOutputs: &core.VariableMap{
				Variables: map[string]*core.Variable{
					"baz": {},
				},
			},
		},
		Spec: &admin.LaunchPlanSpec{
			WorkflowId: &core.Identifier{
				ResourceType: core.ResourceType_WORKFLOW,
				Project:      "project",
				Domain:       "domain",
				Name:         "workflow name",
				Version:      "version",
			},
			EntityMetadata: &admin.LaunchPlanMetadata{
				Schedule: &admin.Schedule{
					ScheduleExpression: &admin.Schedule_CronExpression{
						CronExpression: "* * * * ",
					},
				},
			},
		},
	}
}

func getCompiledWorkflow() (*core.CompiledWorkflowClosure, error) {
	var compiledWorkflow core.CompiledWorkflowClosure
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	workflowJSONFile := filepath.Join(pwd, "testdata", "workflow.json")
	workflowJSON, err := ioutil.ReadFile(workflowJSONFile)
	if err != nil {
		return nil, err
	}
	err = jsonpb.UnmarshalString(string(workflowJSON), &compiledWorkflow)
	if err != nil {
		return nil, err
	}
	return &compiledWorkflow, nil
}

func TestGetLaunchPlanDigest(t *testing.T) {
	launchPlanDigest, err := GetLaunchPlanDigest(context.Background(), getLaunchPlan())
	assert.Equal(t, testLaunchPlanDigest, launchPlanDigest)
	assert.Nil(t, err)
}

func TestGetLaunchPlanDigest_Unequal(t *testing.T) {
	launchPlanWithDifferentInputs := getLaunchPlan()
	launchPlanWithDifferentInputs.Closure.ExpectedInputs.Parameters["unexpected"] = &core.Parameter{}
	launchPlanDigest, err := GetLaunchPlanDigest(context.Background(), launchPlanWithDifferentInputs)
	assert.NotEqual(t, testLaunchPlanDigest, launchPlanDigest)
	assert.Nil(t, err)

	launchPlanWithDifferentOutputs := getLaunchPlan()
	launchPlanWithDifferentOutputs.Closure.ExpectedOutputs.Variables["unexpected"] = &core.Variable{}
	launchPlanDigest, err = GetLaunchPlanDigest(context.Background(), launchPlanWithDifferentOutputs)
	assert.NotEqual(t, testLaunchPlanDigest, launchPlanDigest)
	assert.Nil(t, err)
}

func TestGetTaskDigest(t *testing.T) {
	taskDigest, err := GetTaskDigest(context.Background(), compiledTask)
	assert.Equal(t, compiledTaskDigest, taskDigest)
	assert.Nil(t, err)
}

func TestGetTaskDigest_Unequal(t *testing.T) {
	compiledTaskRequest := &core.CompiledTask{
		Template: &core.TaskTemplate{
			Id:   &taskIdentifier,
			Type: "foo type",
		},
	}
	changedTaskDigest, err := GetTaskDigest(context.Background(), compiledTaskRequest)
	assert.Nil(t, err)
	assert.NotEqual(t, compiledTaskDigest, changedTaskDigest)
}

func TestGetWorkflowDigest(t *testing.T) {
	compiledWorkflow, err := getCompiledWorkflow()
	assert.Nil(t, err)
	workflowDigest, err := GetWorkflowDigest(context.Background(), compiledWorkflow)
	assert.Equal(t, compiledWorkflowDigest, workflowDigest)
	assert.Nil(t, err)
}

func TestGetWorkflowDigest_Unequal(t *testing.T) {
	workflowWithDifferentNodes, err := getCompiledWorkflow()
	assert.Nil(t, err)
	workflowWithDifferentNodes.Primary.Template.Nodes = append(
		workflowWithDifferentNodes.Primary.Template.Nodes, &core.Node{
			Id: "unexpected",
		})
	workflowDigest, err := GetWorkflowDigest(context.Background(), workflowWithDifferentNodes)
	assert.NotEqual(t, compiledWorkflowDigest, workflowDigest)
	assert.Nil(t, err)

	workflowWithDifferentInputs, err := getCompiledWorkflow()
	assert.Nil(t, err)
	workflowWithDifferentInputs.Primary.Template.Interface.Inputs.Variables["unexpected"] = &core.Variable{}
	workflowDigest, err = GetWorkflowDigest(context.Background(), workflowWithDifferentInputs)
	assert.NotEqual(t, compiledWorkflowDigest, workflowDigest)
	assert.Nil(t, err)

	workflowWithDifferentOutputs, err := getCompiledWorkflow()
	assert.Nil(t, err)
	workflowWithDifferentOutputs.Primary.Template.Interface.Outputs.Variables["unexpected"] = &core.Variable{}
	workflowDigest, err = GetWorkflowDigest(context.Background(), workflowWithDifferentOutputs)
	assert.NotEqual(t, compiledWorkflowDigest, workflowDigest)
	assert.Nil(t, err)
}
