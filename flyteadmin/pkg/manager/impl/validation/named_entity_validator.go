package validation

import (
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"google.golang.org/grpc/codes"
	"k8s.io/apimachinery/pkg/util/sets"
)

var archivableResourceTypes = sets.NewInt32(int32(core.ResourceType_WORKFLOW), int32(core.ResourceType_TASK))

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

	// Only tasks and workflow resources can be modified from the default state.
	if request.Metadata.State != admin.NamedEntityState_NAMED_ENTITY_ACTIVE &&
		!archivableResourceTypes.Has(int32(request.ResourceType)) {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"Resource [%s] cannot have its state updated", request.ResourceType.String())
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
