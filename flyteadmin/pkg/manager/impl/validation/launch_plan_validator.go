package validation

import (
	"context"

	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/shared"
	repositoryInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/validators"
)

func ValidateLaunchPlan(ctx context.Context,
	request *admin.LaunchPlanCreateRequest, db repositoryInterfaces.Repository,
	config runtimeInterfaces.ApplicationConfiguration, workflowInterface *core.TypedInterface) error {
	if err := ValidateIdentifier(request.GetId(), common.LaunchPlan); err != nil {
		return err
	}
	if err := ValidateProjectAndDomain(ctx, db, config, request.GetId().GetProject(), request.GetId().GetDomain()); err != nil {
		return err
	}
	if request.GetSpec() == nil {
		return shared.GetMissingArgumentError(shared.Spec)
	}

	if err := ValidateIdentifier(request.GetSpec().GetWorkflowId(), common.Workflow); err != nil {
		return err
	}
	if err := validateLabels(request.GetSpec().GetLabels()); err != nil {
		return err
	}

	if err := validateLiteralMap(request.GetSpec().GetFixedInputs(), shared.FixedInputs); err != nil {
		return err
	}
	if config.GetTopLevelConfig().FeatureGates.EnableArtifacts {
		if err := validateParameterMapAllowArtifacts(request.GetSpec().GetDefaultInputs(), shared.DefaultInputs); err != nil {
			return err
		}
	} else {
		if err := validateParameterMapDisableArtifacts(request.GetSpec().GetDefaultInputs(), shared.DefaultInputs); err != nil {
			return err
		}
	}
	expectedInputs, err := checkAndFetchExpectedInputForLaunchPlan(workflowInterface.GetInputs(), request.GetSpec().GetFixedInputs(), request.GetSpec().GetDefaultInputs())
	if err != nil {
		return err
	}
	if err := validateSchedule(request, expectedInputs); err != nil {
		return err
	}

	// Augment default inputs with the unbound workflow inputs.
	request.Spec.DefaultInputs = expectedInputs
	if request.GetSpec().GetEntityMetadata() != nil {
		if err := validateNotifications(request.GetSpec().GetEntityMetadata().GetNotifications()); err != nil {
			return err
		}
		if request.GetSpec().GetEntityMetadata().GetLaunchConditions() != nil {
			return errors.NewFlyteAdminErrorf(
				codes.InvalidArgument,
				"Launch condition must be empty, found %v", request.GetSpec().GetEntityMetadata().GetLaunchConditions())
		}
	}
	return nil
}

func validateSchedule(request *admin.LaunchPlanCreateRequest, expectedInputs *core.ParameterMap) error {
	schedule := request.GetSpec().GetEntityMetadata().GetSchedule()
	if schedule.GetCronExpression() != "" || schedule.GetRate() != nil || schedule.GetCronSchedule() != nil {
		for key, value := range expectedInputs.GetParameters() {
			if value.GetRequired() && key != schedule.GetKickoffTimeInputArg() {
				return errors.NewFlyteAdminErrorf(
					codes.InvalidArgument,
					"Cannot create a launch plan with a schedule if there is an unbound required input. [%v] is required", key)
			}
		}
		if schedule.GetKickoffTimeInputArg() != "" {
			if param, ok := expectedInputs.GetParameters()[schedule.GetKickoffTimeInputArg()]; !ok {
				return errors.NewFlyteAdminErrorf(
					codes.InvalidArgument,
					"Cannot create a schedule with a KickoffTimeInputArg that does not point to a free input. [%v] is not free or does not exist.", schedule.GetKickoffTimeInputArg())
			} else if param.GetVar().GetType().GetSimple() != core.SimpleType_DATETIME {
				return errors.NewFlyteAdminErrorf(
					codes.InvalidArgument,
					"KickoffTimeInputArg must reference a datetime input. [%v] is a [%v]", schedule.GetKickoffTimeInputArg(), param.GetVar().GetType())
			}
		}
	}
	return nil
}

func checkAndFetchExpectedInputForLaunchPlan(
	workflowVariableMap *core.VariableMap, fixedInputs *core.LiteralMap, defaultInputs *core.ParameterMap) (*core.ParameterMap, error) {
	expectedInputMap := map[string]*core.Parameter{}
	var workflowExpectedInputMap map[string]*core.Variable
	var defaultInputMap map[string]*core.Parameter
	var fixedInputMap map[string]*core.Literal

	if defaultInputs != nil && len(defaultInputs.GetParameters()) > 0 {
		defaultInputMap = defaultInputs.GetParameters()
	}

	if fixedInputs != nil && len(fixedInputs.GetLiterals()) > 0 {
		fixedInputMap = fixedInputs.GetLiterals()
	}

	// If there are no inputs that the workflow requires, there should be none at launch plan as well
	if workflowVariableMap == nil || len(workflowVariableMap.GetVariables()) == 0 {
		if len(defaultInputMap) > 0 {
			return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
				"invalid launch plan default inputs, expected none but found %d", len(defaultInputMap))
		}
		if len(fixedInputMap) > 0 {
			return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
				"invalid launch plan fixed inputs, expected none but found %d", len(fixedInputMap))
		}
		return &core.ParameterMap{
			Parameters: expectedInputMap,
		}, nil
	}

	workflowExpectedInputMap = workflowVariableMap.GetVariables()
	for name, defaultInput := range defaultInputMap {
		value, ok := workflowExpectedInputMap[name]
		if !ok {
			return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "unexpected default_input %s", name)
		} else if !validators.AreTypesCastable(defaultInput.GetVar().GetType(), value.GetType()) {
			return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
				"invalid default_input wrong type %s, expected %v, got %v instead",
				name, defaultInput.GetVar().GetType().String(), value.GetType().String())
		}
	}

	for name, fixedInput := range fixedInputMap {
		value, ok := workflowExpectedInputMap[name]
		if !ok {
			return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "unexpected fixed_input %s", name)
		}
		inputType := validators.LiteralTypeForLiteral(fixedInput)
		err := validators.ValidateLiteralType(inputType)
		if err != nil {
			return nil, errors.NewInvalidLiteralTypeError(name, err)
		}
		if !validators.AreTypesCastable(inputType, value.GetType()) {
			return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
				"invalid fixed_input wrong type %s, expected %v, got %v instead", name, value.GetType(), inputType)
		}
	}

	for name, workflowExpectedInput := range workflowExpectedInputMap {
		if value, ok := defaultInputMap[name]; ok {
			// If the launch plan has a default value - then use this value
			expectedInputMap[name] = value
		} else if _, ok = fixedInputMap[name]; !ok {
			// If there is no mention of the input in LaunchPlan, then copy from the workflow
			expectedInputMap[name] = &core.Parameter{
				Var: &core.Variable{
					Type:        workflowExpectedInput.GetType(),
					Description: workflowExpectedInput.GetDescription(),
				},
				Behavior: &core.Parameter_Required{
					Required: true,
				},
			}
		}
	}
	return &core.ParameterMap{
		Parameters: expectedInputMap,
	}, nil
}
