package validation

import (
	"strconv"
	"strings"

	"github.com/lyft/flyteadmin/pkg/common"
	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/manager/impl/shared"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytepropeller/pkg/compiler/validators"
	"google.golang.org/grpc/codes"
)

var entityToResourceType = map[common.Entity]core.ResourceType{
	common.Task:       core.ResourceType_TASK,
	common.Workflow:   core.ResourceType_WORKFLOW,
	common.LaunchPlan: core.ResourceType_LAUNCH_PLAN,
}

func ValidateEmptyStringField(field, fieldName string) error {
	if field == "" {
		return shared.GetMissingArgumentError(fieldName)
	}
	return nil
}

// Validates that a string field does not exceed a certain character count
func ValidateMaxLengthStringField(field string, fieldName string, limit int) error {
	if len(field) > limit {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument, "%s cannot exceed %d characters", fieldName, limit)
	}
	return nil
}

// Validates that all required fields for an identifier are present.
func ValidateIdentifier(id *core.Identifier, expectedType common.Entity) error {
	if id == nil {
		return shared.GetMissingArgumentError(shared.ID)
	}
	if entityToResourceType[expectedType] != id.ResourceType {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"unexpected resource type %s for identifier [%+v], expected %s instead",
			strings.ToLower(id.ResourceType.String()), id, strings.ToLower(entityToResourceType[expectedType].String()))
	}
	if err := ValidateEmptyStringField(id.Project, shared.Project); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(id.Domain, shared.Domain); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(id.Name, shared.Name); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(id.Version, shared.Version); err != nil {
		return err
	}
	return nil
}

func ValidateVersion(version string) error {
	if err := ValidateEmptyStringField(version, shared.Version); err != nil {
		return err
	}
	return nil
}

func ValidateResourceListRequest(request admin.ResourceListRequest) error {
	if request.Id == nil {
		return shared.GetMissingArgumentError(shared.ID)
	}
	if err := ValidateEmptyStringField(request.Id.Project, shared.Project); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(request.Id.Domain, shared.Domain); err != nil {
		return err
	}
	if request.Limit <= 0 {
		return shared.GetInvalidArgumentError(shared.Limit)
	}
	return nil
}

func ValidateActiveLaunchPlanRequest(request admin.ActiveLaunchPlanRequest) error {
	if err := ValidateEmptyStringField(request.Id.Project, shared.Project); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(request.Id.Domain, shared.Domain); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(request.Id.Name, shared.Name); err != nil {
		return err
	}
	return nil
}

func ValidateActiveLaunchPlanListRequest(request admin.ActiveLaunchPlanListRequest) error {
	if err := ValidateEmptyStringField(request.Project, shared.Project); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(request.Domain, shared.Domain); err != nil {
		return err
	}
	if request.Limit <= 0 {
		return shared.GetInvalidArgumentError(shared.Limit)
	}
	return nil
}

func ValidateNamedEntityIdentifierListRequest(request admin.NamedEntityIdentifierListRequest) error {
	if err := ValidateEmptyStringField(request.Project, shared.Project); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(request.Domain, shared.Domain); err != nil {
		return err
	}
	if request.Limit <= 0 {
		return shared.GetInvalidArgumentError(shared.Limit)
	}
	return nil
}

func validateLiteralMap(inputMap *core.LiteralMap, fieldName string) error {
	if inputMap != nil && len(inputMap.Literals) > 0 {
		for name, fixedInput := range inputMap.Literals {
			if name == "" {
				return errors.NewFlyteAdminErrorf(codes.InvalidArgument, "missing key in %s", fieldName)
			}
			if fixedInput == nil || fixedInput.GetValue() == nil {
				return errors.NewFlyteAdminErrorf(codes.InvalidArgument, "missing valid literal in %s %s", fieldName, name)
			}
		}
	}
	return nil
}

func validateParameterMap(inputMap *core.ParameterMap, fieldName string) error {
	if inputMap != nil && len(inputMap.Parameters) > 0 {
		for name, defaultInput := range inputMap.Parameters {
			if name == "" {
				return errors.NewFlyteAdminErrorf(codes.InvalidArgument, "missing key in %s", fieldName)
			}
			if defaultInput.GetVar() == nil || defaultInput.GetVar().GetType() == nil {
				return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
					"The Variable component of the Parameter %s in %s either is missing, or has a missing Type",
					name, fieldName)
			}
			if defaultInput.GetDefault() == nil && !defaultInput.GetRequired() {
				return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
					"Invalid variable %s in %s - variable has neither default, nor is required. "+
						"One must be specified", name, fieldName)
			}
			defaultValue := defaultInput.GetDefault()
			if defaultValue != nil {
				inputType := validators.LiteralTypeForLiteral(defaultValue)
				if !validators.AreTypesCastable(inputType, defaultInput.GetVar().GetType()) {
					return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
						"Type mismatch for Parameter %s in %s has type %s, expected %s", name, fieldName,
						defaultInput.GetVar().GetType().String(), inputType.String())
				}
			}
		}
	}
	return nil
}

// Offsets are encoded as string tokens to enable future api pagination changes. In addition to validating that an
// offset is a valid integer, we assert that it is non-negative.
func ValidateToken(token string) (int, error) {
	if token == "" {
		return 0, nil
	}
	offset, err := strconv.Atoi(token)
	if err != nil {
		return 0, err
	}
	if offset < 0 {
		return 0, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "Invalid token value: %s", token)
	}
	return offset, nil
}

func ValidateLimit(limit uint32) error {
	if limit <= 0 {
		return shared.GetMissingArgumentError(shared.Limit)
	}
	return nil
}
