package validation

import (
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

func ValidateConfigurationGetRequest(request admin.ConfigurationGetRequest) error {
	if request.Id == nil {
		return shared.GetMissingArgumentError(shared.ID)
	}
	if request.Id.Workflow != "" && (request.Id.Project == "" || request.Id.Domain == "") {
		return shared.GetInvalidArgumentError(shared.ID)
	}
	if request.Id.Domain != "" && request.Id.Project == "" {
		return shared.GetInvalidArgumentError(shared.ID)
	}
	return nil
}

func ValidateConfigurationUpdateRequest(request admin.ConfigurationUpdateRequest) error {
	if request.Id == nil {
		return shared.GetMissingArgumentError(shared.ID)
	}
	if request.Id.Workflow != "" && (request.Id.Project == "" || request.Id.Domain == "") {
		return shared.GetInvalidArgumentError(shared.ID)
	}
	if request.Id.Domain != "" && request.Id.Project == "" {
		return shared.GetInvalidArgumentError(shared.ID)
	}
	if request.Configuration == nil {
		return shared.GetMissingArgumentError(shared.Configuration)
	}
	return ValidateVersion(request.VersionToUpdate)
}
