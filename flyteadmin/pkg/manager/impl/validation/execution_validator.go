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

func ValidateExecutionRequest(ctx context.Context, request *admin.ExecutionCreateRequest,
	db repositoryInterfaces.Repository, config runtimeInterfaces.ApplicationConfiguration) error {
	if err := ValidateEmptyStringField(request.GetProject(), shared.Project); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(request.GetDomain(), shared.Domain); err != nil {
		return err
	}
	if request.GetName() != "" {
		if err := CheckValidExecutionID(strings.ToLower(request.GetName()), shared.Name); err != nil {
			return err
		}
	}
	if len(request.GetName()) > allowedExecutionNameLength {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"name for ExecutionCreateRequest [%+v] exceeded allowed length %d", request, allowedExecutionNameLength)
	}
	if err := ValidateProjectAndDomain(ctx, db, config, request.GetProject(), request.GetDomain()); err != nil {
		return err
	}

	if request.GetSpec() == nil {
		return shared.GetMissingArgumentError(shared.Spec)
	}
	// TODO(katrogan): Change the name of Spec.LaunchPlan to something more generic to permit reference Tasks.
	// https://github.com/flyteorg/flyte/issues/262
	if err := ValidateIdentifierFieldsSet(request.GetSpec().GetLaunchPlan()); err != nil {
		return err
	}
	if _, ok := acceptedReferenceLaunchTypes[request.GetSpec().GetLaunchPlan().GetResourceType()]; !ok {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"Invalid reference entity resource type [%v], only [%+v] allowed",
			request.GetSpec().GetLaunchPlan().GetResourceType(), acceptedReferenceLaunchTypes)
	}
	if err := validateLiteralMap(request.GetInputs(), shared.Inputs); err != nil {
		return err
	}
	if request.GetSpec().GetNotifications() != nil {
		if err := validateNotifications(request.GetSpec().GetNotifications().GetNotifications()); err != nil {
			return err
		}
	}
	if err := validateLabels(request.GetSpec().GetLabels()); err != nil {
		return err
	}
	return nil
}

func CheckAndFetchInputsForExecution(
	userInputs *core.LiteralMap, fixedInputs *core.LiteralMap, expectedInputs *core.ParameterMap) (*core.LiteralMap, error) {

	executionInputMap := map[string]*core.Literal{}
	expectedInputMap := map[string]*core.Parameter{}

	if expectedInputs != nil && len(expectedInputs.GetParameters()) > 0 {
		expectedInputMap = expectedInputs.GetParameters()
	}

	if userInputs != nil && len(userInputs.GetLiterals()) > 0 {
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
			err := validators.ValidateLiteralType(inputType)
			if err != nil {
				return nil, errors.NewInvalidLiteralTypeError(name, err)
			}
			if !validators.AreTypesCastable(inputType, expectedInput.GetVar().GetType()) {
				return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "invalid %s input wrong type. Expected %s, but got %s", name, expectedInput.GetVar().GetType(), inputType)
			}
		}
	}

	if fixedInputs != nil && len(fixedInputs.GetLiterals()) > 0 {
		for name, fixedInput := range fixedInputs.GetLiterals() {
			if _, ok := executionInputMap[name]; ok {
				return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "%s %s cannot be overridden", shared.FixedInputs, name)
			}
			executionInputMap[name] = fixedInput
		}
	}

	return &core.LiteralMap{
		Literals: executionInputMap,
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

func ValidateCreateWorkflowEventRequest(request *admin.WorkflowExecutionEventRequest, maxOutputSizeInBytes int64) error {
	if request.GetEvent() == nil {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"Workflow event handler was called without event")
	} else if request.GetEvent().GetExecutionId() == nil {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"Workflow event handler request event doesn't have an execution id - %v", request.GetEvent())
	}
	if err := ValidateOutputData(request.GetEvent().GetOutputData(), maxOutputSizeInBytes); err != nil {
		return err
	}
	return nil
}

func ValidateWorkflowExecutionIdentifier(identifier *core.WorkflowExecutionIdentifier) error {
	if identifier == nil {
		return shared.GetMissingArgumentError(shared.ID)
	}
	if err := ValidateEmptyStringField(identifier.GetProject(), shared.Project); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(identifier.GetDomain(), shared.Domain); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(identifier.GetName(), shared.Name); err != nil {
		return err
	}
	return nil
}
