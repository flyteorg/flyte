package validation

import (
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"google.golang.org/grpc/codes"
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

	// Anything but the default state is only permitted for workflow resources.
	if request.Metadata.State != admin.NamedEntityState_NAMED_ENTITY_ACTIVE &&
		request.ResourceType != core.ResourceType_WORKFLOW {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"Only workflow name entities can have their state updated")
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
