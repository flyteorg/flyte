package impl

import (
	"testing"

	"github.com/flyteorg/flyteidl/clients/go/coreutils"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/compiler/common"
	"github.com/golang/protobuf/proto"
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

func getProviderForTest(t *testing.T) common.InterfaceProvider {
	launchPlanStatus := admin.LaunchPlanClosure{
		ExpectedInputs:  &inputs,
		ExpectedOutputs: &outputs,
	}
	bytes, _ := proto.Marshal(&launchPlanStatus)
	provider, err := NewLaunchPlanInterfaceProvider(
		models.LaunchPlan{
			Closure: bytes,
		}, launchPlanIdentifier)
	if err != nil {
		t.Fatalf("Failed to initialize LaunchPlanInterfaceProvider for test with err %v", err)
	}
	return provider
}

func TestGetId(t *testing.T) {
	provider := getProviderForTest(t)
	assert.Equal(t, &core.Identifier{ResourceType: 3, Project: "project", Domain: "domain", Name: "name", Version: "version"}, provider.GetID())
}

func TestGetExpectedInputs(t *testing.T) {
	provider := getProviderForTest(t)
	assert.Contains(t, (*provider.GetExpectedInputs()).Parameters, "foo")
	assert.NotNil(t, (*provider.GetExpectedInputs()).Parameters["foo"].Var.Type.GetSimple())
	assert.EqualValues(t, "STRING", (*provider.GetExpectedInputs()).Parameters["foo"].Var.Type.GetSimple().String())
	assert.NotNil(t, (*provider.GetExpectedInputs()).Parameters["foo"].GetDefault())
}

func TestGetExpectedOutputs(t *testing.T) {
	provider := getProviderForTest(t)
	assert.EqualValues(t, outputs.Variables["foo"].GetType().GetType(),
		provider.GetExpectedOutputs().Variables["foo"].GetType().GetType())
}
