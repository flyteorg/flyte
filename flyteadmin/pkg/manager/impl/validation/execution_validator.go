package validation

import (
	"context"
	"regexp"
	"strings"

	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/shared"
	repositoryInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/validators"
)

// Maximum value length of a Kubernetes label
const allowedExecutionNameLength = 63

var executionIDRegex = regexp.MustCompile(`^[a-z][a-z\-0-9]*$`)

var acceptedReferenceLaunchTypes = map[core.ResourceType]interface{}{
	core.ResourceType_LAUNCH_PLAN: nil,
	core.ResourceType_TASK:        nil,
}

func ValidateExecutionRequest(ctx context.Context, request admin.ExecutionCreateRequest,
	db repositoryInterfaces.Repository, config runtimeInterfaces.ApplicationConfiguration) error {
	if err := ValidateEmptyStringField(request.Project, shared.Project); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(request.Domain, shared.Domain); err != nil {
		return err
	}
	if request.Name != "" {
		if err := CheckValidExecutionID(strings.ToLower(request.Name), shared.Name); err != nil {
			return err
		}
	}
	if len(request.Name) > allowedExecutionNameLength {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"name for ExecutionCreateRequest [%+v] exceeded allowed length %d", request, allowedExecutionNameLength)
	}
	if err := ValidateProjectAndDomain(ctx, db, config, request.Project, request.Domain); err != nil {
		return err
	}

	if request.Spec == nil {
		return shared.GetMissingArgumentError(shared.Spec)
	}
	// TODO(katrogan): Change the name of Spec.LaunchPlan to something more generic to permit reference Tasks.
	// https://github.com/flyteorg/flyte/issues/262
	if err := ValidateIdentifierFieldsSet(request.Spec.LaunchPlan); err != nil {
		return err
	}
	if _, ok := acceptedReferenceLaunchTypes[request.Spec.LaunchPlan.ResourceType]; !ok {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"Invalid reference entity resource type [%v], only [%+v] allowed",
			request.Spec.LaunchPlan.ResourceType, acceptedReferenceLaunchTypes)
	}
	if err := validateLiteralMap(request.Inputs, shared.Inputs); err != nil {
		return err
	}
	if request.Spec.GetNotifications() != nil {
		if err := validateNotifications(request.Spec.GetNotifications().Notifications); err != nil {
			return err
		}
	}
	return nil
}

func CheckAndFetchInputsForExecution(
	userInputsData *core.InputData, userInputs *core.LiteralMap,
	fixedInputsData *core.InputData, fixedInputs *core.LiteralMap,
	expectedInputs *core.ParameterMap) (*core.InputData, error) {

	executionInputMap := map[string]*core.Literal{}
	expectedInputMap := map[string]*core.Parameter{}

	if expectedInputs != nil && len(expectedInputs.GetParameters()) > 0 {
		expectedInputMap = expectedInputs.GetParameters()
	}

	if literals := userInputsData.GetInputs().GetLiterals(); len(literals) > 0 {
		for name, value := range literals {
			if _, ok := expectedInputMap[name]; !ok {
				return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "invalid input %s", name)
			}

			executionInputMap[name] = value
		}
		// DEPRECATED: Remove this block once the deprecated field is removed from the API.
	} else if userInputs != nil && len(userInputs.GetLiterals()) > 0 {
		for name, value := range userInputs.GetLiterals() {
			if _, ok := expectedInputMap[name]; !ok {
				return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "invalid input %s", name)
			}

			executionInputMap[name] = value
		}
	}

	for name, expectedInput := range expectedInputMap {
		if _, ok := executionInputMap[name]; !ok {
			if expectedInput.GetRequired() {
				return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "%s %s missing", shared.ExpectedInputs, name)
			}
			executionInputMap[name] = expectedInput.GetDefault()
		} else {
			inputType := validators.LiteralTypeForLiteral(executionInputMap[name])
			if !validators.AreTypesCastable(inputType, expectedInput.GetVar().GetType()) {
				return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "invalid %s input wrong type. Expected %s, but got %s", name, expectedInput.GetVar().GetType(), inputType)
			}
		}
	}

	if literals := fixedInputsData.GetInputs().GetLiterals(); len(literals) > 0 {
		for name, fixedInput := range literals {
			if _, ok := executionInputMap[name]; ok {
				return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "%s %s cannot be overridden", shared.FixedInputs, name)
			}
			executionInputMap[name] = fixedInput
		}
		// DEPRECATED: Remove this block once the deprecated field is removed from the API.
	} else if fixedInputs != nil && len(fixedInputs.GetLiterals()) > 0 {
		for name, fixedInput := range fixedInputs.GetLiterals() {
			if _, ok := executionInputMap[name]; ok {
				return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "%s %s cannot be overridden", shared.FixedInputs, name)
			}
			executionInputMap[name] = fixedInput
		}
	}

	return &core.InputData{
		Inputs: &core.LiteralMap{
			Literals: executionInputMap,
		},
	}, nil
}

func CheckValidExecutionID(executionID, fieldName string) error {
	if len(executionID) > allowedExecutionNameLength {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"size of %s exceeded length %d : %s", fieldName, allowedExecutionNameLength, executionID)
	}
	matched := executionIDRegex.MatchString(executionID)

	if !matched {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument, "invalid %s format: %s, does not match regex '%s'", fieldName, executionID, executionIDRegex.String())
	}

	return nil
}

func ValidateCreateWorkflowEventRequest(request admin.WorkflowExecutionEventRequest, maxOutputSizeInBytes int64) error {
	if request.Event == nil {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"Workflow event handler was called without event")
	} else if request.Event.ExecutionId == nil {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"Workflow event handler request event doesn't have an execution id - %v", request.Event)
	}

	outputData := request.GetEvent().GetOutputData()
	if outputData == nil {
		outputData = &core.OutputData{
			Outputs: request.GetEvent().GetDeprecatedOutputData(),
		}
	}

	if err := ValidateOutputData(outputData, maxOutputSizeInBytes); err != nil {
		return err
	}
	return nil
}

func ValidateWorkflowExecutionIdentifier(identifier *core.WorkflowExecutionIdentifier) error {
	if identifier == nil {
		return shared.GetMissingArgumentError(shared.ID)
	}
	if err := ValidateEmptyStringField(identifier.Project, shared.Project); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(identifier.Domain, shared.Domain); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(identifier.Name, shared.Name); err != nil {
		return err
	}
	return nil
}
