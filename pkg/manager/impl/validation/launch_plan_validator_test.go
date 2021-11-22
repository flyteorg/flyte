package validation

import (
	"context"
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/flyteorg/flyteidl/clients/go/coreutils"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/testutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
)

var lpApplicationConfig = testutils.GetApplicationConfigWithDefaultDomains()

func getWorkflowInterface() *core.TypedInterface {
	return testutils.GetSampleWorkflowSpecForTest().Template.Interface
}

func TestValidateLpEmptyProject(t *testing.T) {
	request := testutils.GetLaunchPlanRequest()
	request.Id.Project = ""
	err := ValidateLaunchPlan(context.Background(), request, testutils.GetRepoWithDefaultProject(), lpApplicationConfig, getWorkflowInterface())
	assert.EqualError(t, err, "missing project")
}

func TestValidateLpEmptyDomain(t *testing.T) {
	request := testutils.GetLaunchPlanRequest()
	request.Id.Domain = ""
	err := ValidateLaunchPlan(context.Background(), request, testutils.GetRepoWithDefaultProject(), lpApplicationConfig, getWorkflowInterface())
	assert.EqualError(t, err, "missing domain")
}

func TestValidateLpEmptyName(t *testing.T) {
	request := testutils.GetLaunchPlanRequest()
	request.Id.Name = ""
	err := ValidateLaunchPlan(context.Background(), request, testutils.GetRepoWithDefaultProject(), lpApplicationConfig, getWorkflowInterface())
	assert.EqualError(t, err, "missing name")
}

func TestValidateLpLabels(t *testing.T) {
	request := testutils.GetLaunchPlanRequest()
	request.Spec.Labels = &admin.Labels{
		Values: map[string]string{
			"foo": "#badlabel",
			"bar": "baz",
		}}
	err := ValidateLaunchPlan(context.Background(), request, testutils.GetRepoWithDefaultProject(), lpApplicationConfig, getWorkflowInterface())
	assert.EqualError(t, err, "invalid label value [#badlabel]: [a valid label must be an empty string or consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyValue',  or 'my_value',  or '12345', regex used for validation is '(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?')]")
}

func TestValidateLpEmptyVersion(t *testing.T) {
	request := testutils.GetLaunchPlanRequest()
	request.Id.Version = ""
	err := ValidateLaunchPlan(context.Background(), request, testutils.GetRepoWithDefaultProject(), lpApplicationConfig, getWorkflowInterface())
	assert.EqualError(t, err, "missing version")
}

func TestValidateLpEmptySpec(t *testing.T) {
	request := testutils.GetLaunchPlanRequest()
	request.Spec = nil
	err := ValidateLaunchPlan(context.Background(), request, testutils.GetRepoWithDefaultProject(), lpApplicationConfig, getWorkflowInterface())
	assert.EqualError(t, err, "missing spec")
}

func TestGetLpExpectedInputs(t *testing.T) {
	request := testutils.GetLaunchPlanRequest()
	actualExpectedMap, err := checkAndFetchExpectedInputForLaunchPlan(
		&core.VariableMap{
			Variables: map[string]*core.Variable{
				"foo": {
					Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
				},
				"bar": {
					Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
				},
			},
		},
		request.GetSpec().GetFixedInputs(), request.GetSpec().GetDefaultInputs(),
	)
	expectedMap := core.ParameterMap{
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
	assert.Nil(t, err)
	assert.NotNil(t, actualExpectedMap)
	assert.EqualValues(t, expectedMap, *actualExpectedMap)
}

func TestValidateLpDefaultInputsWrongType(t *testing.T) {
	request := testutils.GetLaunchPlanRequest()
	request.Spec.DefaultInputs.Parameters["foo"].Var.Type = &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_FLOAT}}
	err := ValidateLaunchPlan(context.Background(), request, testutils.GetRepoWithDefaultProject(), lpApplicationConfig, getWorkflowInterface())
	assert.EqualError(t, err, "Type mismatch for Parameter foo in default_inputs has type simple:FLOAT , expected simple:STRING ")
}

func TestValidateLpDefaultInputsEmptyName(t *testing.T) {
	request := testutils.GetLaunchPlanRequest()
	request.Spec.DefaultInputs.Parameters = map[string]*core.Parameter{
		"": nil,
	}
	err := ValidateLaunchPlan(context.Background(), request, testutils.GetRepoWithDefaultProject(), lpApplicationConfig, getWorkflowInterface())
	assert.EqualError(t, err, "missing key in default_inputs")
}

func TestValidateLpDefaultInputsEmptyType(t *testing.T) {
	request := testutils.GetLaunchPlanRequest()
	request.Spec.DefaultInputs.Parameters["foo"].Var.Type = nil
	err := ValidateLaunchPlan(context.Background(), request, testutils.GetRepoWithDefaultProject(), lpApplicationConfig, getWorkflowInterface())
	assert.EqualError(t, err, "The Variable component of the Parameter foo in default_inputs either is missing, or has a missing Type")
}

func TestValidateLpDefaultInputsEmptyVar(t *testing.T) {
	request := testutils.GetLaunchPlanRequest()
	request.Spec.DefaultInputs.Parameters["foo"].Var = nil
	err := ValidateLaunchPlan(context.Background(), request, testutils.GetRepoWithDefaultProject(), lpApplicationConfig, getWorkflowInterface())
	assert.EqualError(t, err, "The Variable component of the Parameter foo in default_inputs either is missing, or has a missing Type")
}

func TestValidateLpFixedInputsEmptyName(t *testing.T) {
	request := testutils.GetLaunchPlanRequest()
	request.Spec.FixedInputs.Literals = map[string]*core.Literal{
		"": nil,
	}
	err := ValidateLaunchPlan(context.Background(), request, testutils.GetRepoWithDefaultProject(), lpApplicationConfig, getWorkflowInterface())
	assert.EqualError(t, err, "missing key in fixed_inputs")
}

func TestValidateLpFixedInputsEmptyValue(t *testing.T) {
	request := testutils.GetLaunchPlanRequest()
	request.Spec.FixedInputs.Literals = map[string]*core.Literal{
		"a": nil,
	}
	err := ValidateLaunchPlan(context.Background(), request, testutils.GetRepoWithDefaultProject(), lpApplicationConfig, getWorkflowInterface())
	assert.EqualError(t, err, "missing valid literal in fixed_inputs a")
}

func TestGetLpExpectedInvalidDefaultInput(t *testing.T) {
	request := testutils.GetLaunchPlanRequest()
	actualMap, err := checkAndFetchExpectedInputForLaunchPlan(
		&core.VariableMap{
			Variables: map[string]*core.Variable{
				"foo-x": {
					Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
				},
				"bar": {
					Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
				},
			},
		},
		request.GetSpec().GetFixedInputs(), request.GetSpec().GetDefaultInputs(),
	)

	assert.EqualError(t, err, "unexpected default_input foo")
	assert.Nil(t, actualMap)
}

func TestGetLpExpectedInvalidDefaultInputType(t *testing.T) {
	request := testutils.GetLaunchPlanRequest()
	actualMap, err := checkAndFetchExpectedInputForLaunchPlan(
		&core.VariableMap{
			Variables: map[string]*core.Variable{
				"foo": {
					Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_BINARY}},
				},
				"bar": {
					Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
				},
			},
		},
		request.GetSpec().GetFixedInputs(), request.GetSpec().GetDefaultInputs(),
	)

	assert.EqualError(t, err, "invalid default_input wrong type foo, expected simple:STRING , got simple:BINARY  instead")
	assert.Nil(t, actualMap)
}

func TestGetLpExpectedInvalidFixedInputType(t *testing.T) {
	request := testutils.GetLaunchPlanRequest()
	actualMap, err := checkAndFetchExpectedInputForLaunchPlan(
		&core.VariableMap{
			Variables: map[string]*core.Variable{
				"foo": {
					Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
				},
				"bar": {
					Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_BINARY}},
				},
			},
		},
		request.GetSpec().GetFixedInputs(), request.GetSpec().GetDefaultInputs(),
	)

	assert.EqualError(t, err, "invalid fixed_input wrong type bar, expected simple:BINARY , got simple:STRING  instead")
	assert.Nil(t, actualMap)
}

func TestGetLpExpectedInvalidFixedInput(t *testing.T) {
	request := testutils.GetLaunchPlanRequest()
	actualMap, err := checkAndFetchExpectedInputForLaunchPlan(
		&core.VariableMap{
			Variables: map[string]*core.Variable{
				"foo": {
					Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
				},
				"bar-y": {
					Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
				},
			},
		},
		request.GetSpec().GetFixedInputs(), request.GetSpec().GetDefaultInputs(),
	)

	assert.EqualError(t, err, "unexpected fixed_input bar")
	assert.Nil(t, actualMap)
}

func TestGetLpExpectedNoFixedInput(t *testing.T) {
	request := testutils.GetLaunchPlanRequest()
	actualMap, err := checkAndFetchExpectedInputForLaunchPlan(
		&core.VariableMap{
			Variables: map[string]*core.Variable{
				"foo": {
					Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
				},
			},
		},
		nil, request.GetSpec().GetDefaultInputs(),
	)

	expectedMap := core.ParameterMap{
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
	assert.Nil(t, err)
	assert.NotNil(t, actualMap)
	assert.EqualValues(t, expectedMap, *actualMap)
}

func TestGetLpExpectedNoDefaultInput(t *testing.T) {
	request := testutils.GetLaunchPlanRequest()
	actualMap, err := checkAndFetchExpectedInputForLaunchPlan(
		&core.VariableMap{
			Variables: map[string]*core.Variable{
				"bar": {
					Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
				},
			},
		},
		request.GetSpec().GetFixedInputs(), nil,
	)

	expectedMap := core.ParameterMap{
		Parameters: map[string]*core.Parameter{},
	}
	assert.Nil(t, err)
	assert.NotNil(t, actualMap)
	assert.EqualValues(t, expectedMap, *actualMap)
}

func TestValidateSchedule_NoSchedule(t *testing.T) {
	request := testutils.GetLaunchPlanRequest()
	inputMap := &core.ParameterMap{
		Parameters: map[string]*core.Parameter{
			"foo": {
				Var: &core.Variable{
					Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
				},
				Behavior: &core.Parameter_Required{
					Required: true,
				},
			},
		},
	}
	err := validateSchedule(request, inputMap)
	assert.Nil(t, err)
}

func TestValidateSchedule_ArgNotFixed(t *testing.T) {
	request := testutils.GetLaunchPlanRequestWithCronSchedule("* * * * * *")
	inputMap := &core.ParameterMap{
		Parameters: map[string]*core.Parameter{
			"foo": {
				Var: &core.Variable{
					Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
				},
				Behavior: &core.Parameter_Required{
					Required: true,
				},
			},
		},
	}

	err := validateSchedule(request, inputMap)
	assert.NotNil(t, err)
}

func TestValidateSchedule_KickoffTimeArgDoesNotExist(t *testing.T) {
	request := testutils.GetLaunchPlanRequestWithCronSchedule("* * * * * *")
	inputMap := &core.ParameterMap{
		Parameters: map[string]*core.Parameter{},
	}
	request.Spec.EntityMetadata.Schedule.KickoffTimeInputArg = "Does not exist"

	err := validateSchedule(request, inputMap)
	assert.NotNil(t, err)
}

func TestValidateSchedule_KickoffTimeArgPointsAtWrongType(t *testing.T) {
	request := testutils.GetLaunchPlanRequestWithCronSchedule("* * * * * *")
	inputMap := &core.ParameterMap{
		Parameters: map[string]*core.Parameter{
			"foo": {
				Var: &core.Variable{
					Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
				},
				Behavior: &core.Parameter_Required{
					Required: true,
				},
			},
		},
	}
	request.Spec.EntityMetadata.Schedule.KickoffTimeInputArg = "foo"

	err := validateSchedule(request, inputMap)
	assert.NotNil(t, err)
}

func TestValidateSchedule_NoRequired(t *testing.T) {
	request := testutils.GetLaunchPlanRequestWithCronSchedule("* * * * * *")
	inputMap := &core.ParameterMap{
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

	err := validateSchedule(request, inputMap)
	assert.Nil(t, err)
}

func TestValidateSchedule_KickoffTimeBound(t *testing.T) {
	request := testutils.GetLaunchPlanRequestWithCronSchedule("* * * * * *")
	inputMap := &core.ParameterMap{
		Parameters: map[string]*core.Parameter{
			"foo": {
				Var: &core.Variable{
					Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_DATETIME}},
				},
				Behavior: &core.Parameter_Required{
					Required: true,
				},
			},
		},
	}
	request.Spec.EntityMetadata.Schedule.KickoffTimeInputArg = "foo"

	err := validateSchedule(request, inputMap)
	assert.Nil(t, err)
}
