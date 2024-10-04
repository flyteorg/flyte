package validation

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/testutils"
	"github.com/flyteorg/flyte/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flytestdlib/utils"
)

var execConfig = testutils.GetApplicationConfigWithDefaultDomains()

const failedToValidateLiteralType = "Failed to validate literal type"

func TestValidateExecEmptyProject(t *testing.T) {
	request := testutils.GetExecutionRequest()
	request.Project = ""
	err := ValidateExecutionRequest(context.Background(), request, testutils.GetRepoWithDefaultProject(), execConfig)
	assert.EqualError(t, err, "missing project")
}

func TestValidateExecEmptyDomain(t *testing.T) {
	request := testutils.GetExecutionRequest()
	request.Domain = ""
	err := ValidateExecutionRequest(context.Background(), request, testutils.GetRepoWithDefaultProject(), execConfig)
	assert.EqualError(t, err, "missing domain")
}

func TestValidateExecEmptyName(t *testing.T) {
	request := testutils.GetExecutionRequest()
	request.Name = ""
	err := ValidateExecutionRequest(context.Background(), request, testutils.GetRepoWithDefaultProject(), execConfig)
	assert.Nil(t, err)
}

func TestValidateExecInvalidName(t *testing.T) {
	request := testutils.GetExecutionRequest()
	request.Name = "12345"
	err := ValidateExecutionRequest(context.Background(), request, testutils.GetRepoWithDefaultProject(), execConfig)
	assert.EqualError(t, err, "invalid name format: 12345, does not match regex '^[a-z][a-z\\-0-9]*$'")

	request.Name = "e2345"
	err = ValidateExecutionRequest(context.Background(), request, testutils.GetRepoWithDefaultProject(), execConfig)
	assert.Nil(t, err)

	request.Name = "abc-123"
	err = ValidateExecutionRequest(context.Background(), request, testutils.GetRepoWithDefaultProject(), execConfig)
	assert.Nil(t, err)
}

func TestValidateExecEmptySpec(t *testing.T) {
	request := testutils.GetExecutionRequest()
	request.Spec = nil
	err := ValidateExecutionRequest(context.Background(), request, testutils.GetRepoWithDefaultProject(), execConfig)
	assert.EqualError(t, err, "missing spec")
}

func TestValidateExecInvalidProjectAndDomain(t *testing.T) {
	request := testutils.GetExecutionRequest()
	err := ValidateExecutionRequest(context.Background(), request, testutils.GetRepoWithDefaultProjectAndErr(errors.New("foo")), execConfig)
	assert.EqualError(t, err, "failed to validate that project [project] and domain [domain] are registered, err: [foo]")
}

func TestValidateExecInvalidLabels(t *testing.T) {
	request := testutils.GetExecutionRequest()
	request.Spec.Labels = &admin.Labels{
		Values: map[string]string{
			"foo": "#bar",
		},
	}
	err := ValidateExecutionRequest(context.Background(), request, testutils.GetRepoWithDefaultProject(), execConfig)
	assert.ErrorContains(t, err, "invalid label value [#bar]:")

	request.Spec.Labels = &admin.Labels{
		Values: map[string]string{
			"#foo": "bar",
		},
	}
	err = ValidateExecutionRequest(context.Background(), request, testutils.GetRepoWithDefaultProject(), execConfig)
	assert.ErrorContains(t, err, "invalid label key [#foo]:")
}

func TestGetExecutionInputs(t *testing.T) {
	executionRequest := testutils.GetExecutionRequest()
	lpRequest := testutils.GetLaunchPlanRequest()

	actualInputs, err := CheckAndFetchInputsForExecution(
		executionRequest.Inputs,
		lpRequest.Spec.FixedInputs,
		lpRequest.Spec.DefaultInputs,
	)
	expectedMap := &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"foo": coreutils.MustMakeLiteral("foo-value-1"),
			"bar": coreutils.MustMakeLiteral("bar-value"),
		},
	}
	assert.Nil(t, err)
	assert.NotNil(t, actualInputs)
	assert.EqualValues(t, expectedMap, actualInputs)
}

func TestGetExecutionWithOffloadedInputs(t *testing.T) {
	execLiteral := &core.Literal{
		Value: &core.Literal_OffloadedMetadata{
			OffloadedMetadata: &core.LiteralOffloadedMetadata{
				Uri:       "s3://bucket/key",
				SizeBytes: 100,
				InferredType: &core.LiteralType{
					Type: &core.LiteralType_Simple{
						Simple: core.SimpleType_STRING,
					},
				},
			},
		},
	}
	executionRequest := testutils.GetExecutionRequestWithOffloadedInputs("foo", execLiteral)
	lpRequest := testutils.GetLaunchPlanRequest()

	actualInputs, err := CheckAndFetchInputsForExecution(
		executionRequest.Inputs,
		lpRequest.Spec.FixedInputs,
		lpRequest.Spec.DefaultInputs,
	)
	expectedMap := core.LiteralMap{
		Literals: map[string]*core.Literal{
			"foo": execLiteral,
			"bar": coreutils.MustMakeLiteral("bar-value"),
		},
	}
	assert.Nil(t, err)
	assert.NotNil(t, actualInputs)
	assert.EqualValues(t, expectedMap.GetLiterals()["foo"], actualInputs.Literals["foo"])
	assert.EqualValues(t, expectedMap.GetLiterals()["bar"], actualInputs.Literals["bar"])
}

func TestValidateExecInputsWrongType(t *testing.T) {
	executionRequest := testutils.GetExecutionRequest()
	lpRequest := testutils.GetLaunchPlanRequest()
	executionRequest.Inputs = &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"foo": coreutils.MustMakeLiteral(1),
		},
	}
	_, err := CheckAndFetchInputsForExecution(
		executionRequest.Inputs,
		lpRequest.Spec.FixedInputs,
		lpRequest.Spec.DefaultInputs,
	)
	utils.AssertEqualWithSanitizedRegex(t, "invalid foo input wrong type. Expected simple:STRING, but got simple:INTEGER", err.Error())
}

func TestValidateExecInputsExtraInputs(t *testing.T) {
	executionRequest := testutils.GetExecutionRequest()
	lpRequest := testutils.GetLaunchPlanRequest()
	executionRequest.Inputs = &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"foo":       coreutils.MustMakeLiteral("foo-value-1"),
			"foo-extra": coreutils.MustMakeLiteral("foo-value-1"),
		},
	}
	_, err := CheckAndFetchInputsForExecution(
		executionRequest.Inputs,
		lpRequest.Spec.FixedInputs,
		lpRequest.Spec.DefaultInputs,
	)
	assert.EqualError(t, err, "invalid input foo-extra")
}

func TestValidateExecInputsOverrideFixed(t *testing.T) {
	executionRequest := testutils.GetExecutionRequest()
	lpRequest := testutils.GetLaunchPlanRequest()
	executionRequest.Inputs = &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"foo": coreutils.MustMakeLiteral("foo-value-1"),
			"bar": coreutils.MustMakeLiteral("bar-value"),
		},
	}
	_, err := CheckAndFetchInputsForExecution(
		executionRequest.Inputs,
		lpRequest.Spec.FixedInputs,
		lpRequest.Spec.DefaultInputs,
	)
	assert.EqualError(t, err, "invalid input bar")
}

func TestValidateExecEmptyInputs(t *testing.T) {
	executionRequest := testutils.GetExecutionRequest()
	lpRequest := testutils.GetLaunchPlanRequest()
	executionRequest.Inputs = nil
	actualInputs, err := CheckAndFetchInputsForExecution(
		executionRequest.Inputs,
		lpRequest.Spec.FixedInputs,
		lpRequest.Spec.DefaultInputs,
	)
	expectedMap := &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"foo": coreutils.MustMakeLiteral("foo-value"),
			"bar": coreutils.MustMakeLiteral("bar-value"),
		},
	}
	assert.Nil(t, err)
	assert.NotNil(t, actualInputs)
	assert.EqualValues(t, expectedMap, actualInputs)
}

func TestValidateExecUnknownIDLInputs(t *testing.T) {
	unsupportedLiteral := &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{},
		},
	}
	defaultInputs := &core.ParameterMap{
		Parameters: map[string]*core.Parameter{
			"foo": {
				Var: &core.Variable{
					// 1000 means an unsupported type
					Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: 1000}},
				},
				Behavior: &core.Parameter_Default{
					Default: unsupportedLiteral,
				},
			},
		},
	}
	userInputs := &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"foo": unsupportedLiteral, // This will lead to a nil inputType
		},
	}

	_, err := CheckAndFetchInputsForExecution(
		userInputs,
		nil,
		defaultInputs,
	)
	assert.NotNil(t, err)

	// Expected error message
	assert.Contains(t, err.Error(), failedToValidateLiteralType)
}

func TestValidExecutionId(t *testing.T) {
	err := CheckValidExecutionID("abcde123", "a")
	assert.Nil(t, err)
}

func TestValidExecutionIdInvalidLength(t *testing.T) {
	err := CheckValidExecutionID("abcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabc", "a")
	assert.NotNil(t, err)
	assert.EqualError(t, err, "size of a exceeded length 63 : abcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabc")
}

func TestValidExecutionIdInvalidChars(t *testing.T) {
	err := CheckValidExecutionID("a_sdd", "a")
	assert.NotNil(t, err)
	assert.EqualError(t, err, "invalid a format: a_sdd, does not match regex '^[a-z][a-z\\-0-9]*$'")
	err = CheckValidExecutionID("asd@", "a")
	assert.NotNil(t, err)
	assert.EqualError(t, err, "invalid a format: asd@, does not match regex '^[a-z][a-z\\-0-9]*$'")
}

func TestValidateCreateWorkflowEventRequest(t *testing.T) {
	request := &admin.WorkflowExecutionEventRequest{
		RequestId: "1",
	}
	err := ValidateCreateWorkflowEventRequest(request, maxOutputSizeInBytes)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "Workflow event handler was called without event")

	request = &admin.WorkflowExecutionEventRequest{
		RequestId: "1",
		Event: &event.WorkflowExecutionEvent{
			Phase:        core.WorkflowExecution_FAILED,
			OutputResult: &event.WorkflowExecutionEvent_Error{},
		},
	}
	err = ValidateCreateWorkflowEventRequest(request, maxOutputSizeInBytes)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Workflow event handler request event doesn't have an execution id")
}

func TestValidateWorkflowExecutionIdentifier(t *testing.T) {
	identifier := &core.WorkflowExecutionIdentifier{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
	}
	assert.Nil(t, ValidateWorkflowExecutionIdentifier(identifier))
}

func TestValidateWorkflowExecutionIdentifier_Error(t *testing.T) {
	assert.NotNil(t, ValidateWorkflowExecutionIdentifier(nil))

	assert.NotNil(t, ValidateWorkflowExecutionIdentifier(&core.WorkflowExecutionIdentifier{
		Project: "project",
		Domain:  "domain",
	}))

	assert.NotNil(t, ValidateWorkflowExecutionIdentifier(&core.WorkflowExecutionIdentifier{
		Project: "project",
		Name:    "name",
	}))

	assert.NotNil(t, ValidateWorkflowExecutionIdentifier(&core.WorkflowExecutionIdentifier{
		Domain: "domain",
		Name:   "name",
	}))
}
