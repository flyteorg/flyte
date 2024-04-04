package validation

import (
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

func ValidateConfigurationGetRequest(request admin.ConfigurationGetRequest) error {
	if request.Id == nil {
		return shared.GetMissingArgumentError(shared.ID)
	}
	if err := ValidateEmptyStringField(request.Id.Project, shared.Project); err != nil {
		return err
	}
	return ValidateEmptyStringField(request.Id.Domain, shared.Domain)
}

func ValidateConfigurationUpdateRequest(request admin.ConfigurationUpdateRequest) error {
	if request.Id == nil {
		return shared.GetMissingArgumentError(shared.ID)
	}
	if err := ValidateEmptyStringField(request.Id.Project, shared.Project); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(request.Id.Domain, shared.Domain); err != nil {
		return err
	}
	if request.Configuration == nil {
		return shared.GetMissingArgumentError(shared.Configuration)
	}
	// version is optional
	return nil
}
