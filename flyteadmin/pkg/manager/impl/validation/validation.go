package validation

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"k8s.io/apimachinery/pkg/util/validation"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/compiler/validators"
	"google.golang.org/grpc/codes"
)

var entityToResourceType = map[common.Entity]core.ResourceType{
	common.Task:       core.ResourceType_TASK,
	common.Workflow:   core.ResourceType_WORKFLOW,
	common.LaunchPlan: core.ResourceType_LAUNCH_PLAN,
}

// See https://www.rfc-editor.org/rfc/rfc3986#section-2.2
var uriReservedChars = "!*'();:@&=+$,/?#[]"

func ValidateEmptyStringField(field, fieldName string) error {
	if field == "" {
		return shared.GetMissingArgumentError(fieldName)
	}
	return nil
}

// ValidateMaxLengthStringField Validates that a string field does not exceed a certain character count
func ValidateMaxLengthStringField(field string, fieldName string, limit int) error {
	if len(field) > limit {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument, "%s cannot exceed %d characters", fieldName, limit)
	}
	return nil
}

// ValidateMaxMapLengthField Validates that a map field does not exceed a certain amount of entries
func ValidateMaxMapLengthField(m map[string]string, fieldName string, limit int) error {
	if len(m) > limit {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument, "%s map cannot exceed %d entries", fieldName, limit)
	}
	return nil
}

func validateLabels(labels *admin.Labels) error {
	if labels == nil || len(labels.Values) == 0 {
		return nil
	}
	if err := ValidateMaxMapLengthField(labels.Values, "labels", maxLabelArrayLength); err != nil {
		return err
	}
	if err := validateLabelsAlphanumeric(labels); err != nil {
		return err
	}
	return nil
}

// Given an admin.Labels, checks if the labels exist or not and if it does, checks if the labels are K8s compliant,
// i.e. alphanumeric + - and _
func validateLabelsAlphanumeric(labels *admin.Labels) error {
	for key, value := range labels.Values {
		if errs := validation.IsQualifiedName(key); len(errs) > 0 {
			return errors.NewFlyteAdminErrorf(codes.InvalidArgument, "invalid label key [%s]: %v", key, errs)
		}
		if errs := validation.IsValidLabelValue(value); len(errs) > 0 {
			return errors.NewFlyteAdminErrorf(codes.InvalidArgument, "invalid label value [%s]: %v", value, errs)
		}
	}
	return nil
}

func ValidateIdentifierFieldsSet(id *core.Identifier) error {
	if id == nil {
		return shared.GetMissingArgumentError(shared.ID)
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

// ValidateIdentifier Validates that all required fields for an identifier are present.
func ValidateIdentifier(id *core.Identifier, expectedType common.Entity) error {
	if id == nil {
		return shared.GetMissingArgumentError(shared.ID)
	}
	if entityToResourceType[expectedType] != id.ResourceType {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"unexpected resource type %s for identifier [%+v], expected %s instead",
			strings.ToLower(id.ResourceType.String()), id, strings.ToLower(entityToResourceType[expectedType].String()))
	}
	return ValidateIdentifierFieldsSet(id)
}

// ValidateNamedEntityIdentifier Validates that all required fields for an identifier are present.
func ValidateNamedEntityIdentifier(id *admin.NamedEntityIdentifier) error {
	if id == nil {
		return shared.GetMissingArgumentError(shared.ID)
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
	return nil
}

func ValidateResourceType(resourceType core.ResourceType) error {
	if resourceType == core.ResourceType_UNSPECIFIED || core.ResourceType.String(resourceType) == "" {
		return shared.GetInvalidArgumentError(shared.ResourceType)
	}
	return nil
}

func ValidateVersion(version string) error {
	if err := ValidateEmptyStringField(version, shared.Version); err != nil {
		return err
	}
	sanitizedVersion := url.QueryEscape(version)
	if !strings.EqualFold(sanitizedVersion, version) {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument, "version [%s] must be url safe, cannot contains chars [%s]", version, uriReservedChars)
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
	if err := ValidateLimit(request.Limit); err != nil {
		return err
	}
	return nil
}

func ValidateDescriptionEntityListRequest(request admin.DescriptionEntityListRequest) error {
	if request.Id == nil {
		return shared.GetMissingArgumentError(shared.ID)
	}
	if err := ValidateEmptyStringField(request.Id.Project, shared.Project); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(request.Id.Domain, shared.Domain); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(request.Id.Name, shared.Name); err != nil {
		return err
	}
	if err := ValidateLimit(request.Limit); err != nil {
		return err
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
	if err := ValidateLimit(request.Limit); err != nil {
		return err
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
	if err := ValidateLimit(request.Limit); err != nil {
		return err
	}
	return nil
}

func ValidateDescriptionEntityGetRequest(request admin.ObjectGetRequest) error {
	if err := ValidateResourceType(request.Id.ResourceType); err != nil {
		return err
	}
	if err := ValidateIdentifierFieldsSet(request.Id); err != nil {
		return err
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
			if isDateTime(fixedInput) {
				// Make datetime specific validations
				return ValidateDatetime(fixedInput)
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
				if defaultInput.GetVar().GetType().GetSimple() == core.SimpleType_DATETIME {
					// Make datetime specific validations
					return ValidateDatetime(defaultValue)
				}
			}
		}
	}
	return nil
}

// ValidateToken Offsets are encoded as string tokens to enable future api pagination changes. In addition to validating that an
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
	if limit == 0 {
		return shared.GetInvalidArgumentError(shared.Limit)
	}
	return nil
}

func ValidateOutputData(outputData *core.LiteralMap, maxSizeInBytes int64) error {
	if outputData == nil {
		return nil
	}

	outputSizeInBytes := int64(proto.Size(outputData))
	if outputSizeInBytes <= maxSizeInBytes {
		return nil
	}
	return errors.NewFlyteAdminErrorf(codes.ResourceExhausted, "Output data size exceeds platform configured threshold (%+v > %v)", outputSizeInBytes, maxSizeInBytes)
}

func ValidateDatetime(literal *core.Literal) error {
	if literal == nil {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument, "Found invalid nil datetime")
	}

	timestamp := literal.GetScalar().GetPrimitive().GetDatetime()

	err := timestamp.CheckValid()
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument, err.Error())
	}
	return nil
}

func isDateTime(input *core.Literal) bool {
	return input.GetScalar() != nil && input.GetScalar().GetPrimitive() != nil && input.GetScalar().GetPrimitive().GetDatetime() != nil
}
