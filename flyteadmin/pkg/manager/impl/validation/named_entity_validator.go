package validation

import (
	"github.com/lyft/flyteadmin/pkg/manager/impl/shared"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
)

func ValidateNamedEntityGetRequest(request admin.NamedEntityGetRequest) error {
	if err := ValidateResourceType(request.ResourceType); err != nil {
		return err
	}
	if err := ValidateNamedEntityIdentifier(request.Id); err != nil {
		return err
	}
	return nil
}

func ValidateNamedEntityUpdateRequest(request admin.NamedEntityUpdateRequest) error {
	if err := ValidateResourceType(request.ResourceType); err != nil {
		return err
	}
	if err := ValidateNamedEntityIdentifier(request.Id); err != nil {
		return err
	}
	if request.Metadata == nil {
		return shared.GetMissingArgumentError(shared.Metadata)
	}
	return nil
}

func ValidateNamedEntityListRequest(request admin.NamedEntityListRequest) error {
	if err := ValidateEmptyStringField(request.Project, shared.Project); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(request.Domain, shared.Domain); err != nil {
		return err
	}
	if err := ValidateResourceType(request.ResourceType); err != nil {
		return err
	}
	if err := ValidateLimit(request.Limit); err != nil {
		return err
	}
	return nil
}
