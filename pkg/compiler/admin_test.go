package compiler

import (
	"testing"

	"github.com/flyteorg/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
)

var launchPlanIdentifier = core.Identifier{
	ResourceType: core.ResourceType_LAUNCH_PLAN,
	Project:      "project",
	Domain:       "domain",
	Name:         "name",
	Version:      "version",
}

var inputs = core.ParameterMap{
	Parameters: map[string]*core.Parameter{
		"foo": {
			Var: &core.Variable{
				Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
			},
			Behavior: &core.Parameter_Default{
				Default: coreutils.MustMakeLiteral("foo-value"),
			},
		},
	},
}
var outputs = core.VariableMap{
	Variables: map[string]*core.Variable{
		"foo": {
			Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
		},
	},
}

func getDummyLaunchPlan() admin.LaunchPlan {
	launchPlanClosure := admin.LaunchPlanClosure{
		ExpectedInputs:  &inputs,
		ExpectedOutputs: &outputs,
	}
	return admin.LaunchPlan{
		Id:      &launchPlanIdentifier,
		Spec:    nil,
		Closure: &launchPlanClosure,
	}
}

func TestGetId(t *testing.T) {
	launchPlan := getDummyLaunchPlan()
	provider := NewLaunchPlanInterfaceProvider(launchPlan)
	assert.Equal(t, &core.Identifier{ResourceType: 3, Project: "project", Domain: "domain", Name: "name", Version: "version"}, provider.GetID())
}

func TestGetExpectedInputs(t *testing.T) {
	launchPlan := getDummyLaunchPlan()
	provider := NewLaunchPlanInterfaceProvider(launchPlan)
	assert.Contains(t, (*provider.GetExpectedInputs()).Parameters, "foo")
	assert.NotNil(t, (*provider.GetExpectedInputs()).Parameters["foo"].Var.Type.GetSimple())
	assert.EqualValues(t, "STRING", (*provider.GetExpectedInputs()).Parameters["foo"].Var.Type.GetSimple().String())
	assert.NotNil(t, (*provider.GetExpectedInputs()).Parameters["foo"].GetDefault())
}

func TestGetExpectedOutputs(t *testing.T) {
	launchPlan := getDummyLaunchPlan()
	provider := NewLaunchPlanInterfaceProvider(launchPlan)
	assert.EqualValues(t, outputs.Variables["foo"].GetType().GetType(),
		provider.GetExpectedOutputs().Variables["foo"].GetType().GetType())
}
