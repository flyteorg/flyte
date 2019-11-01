package validation

import (
	"github.com/lyft/flyteadmin/pkg/manager/impl/shared"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
)

func ValidateProjectDomainAttributesUpdateRequest(request admin.ProjectDomainAttributesUpdateRequest) error {
	if request.Attributes == nil {
		return shared.GetMissingArgumentError(shared.ProjectDomain)
	}
	if err := ValidateEmptyStringField(request.Attributes.Project, shared.Project); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(request.Attributes.Domain, shared.Domain); err != nil {
		return err
	}
	// Resource attributes are not a required field and therefore are not checked in validation.
	return nil
}
