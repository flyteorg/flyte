package validation

import (
	"context"

	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/manager/impl/shared"
	"github.com/lyft/flyteadmin/pkg/repositories"
	runtimeInterfaces "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"google.golang.org/grpc/codes"
	"k8s.io/apimachinery/pkg/util/validation"
)

const projectID = "project_id"
const projectName = "project_name"
const projectDescription = "project_description"
const maxDescriptionLength = 300

func ValidateProjectRegisterRequest(request admin.ProjectRegisterRequest) error {
	if request.Project == nil {
		return shared.GetMissingArgumentError(shared.Project)
	}
	if err := ValidateEmptyStringField(request.Project.Id, projectID); err != nil {
		return err
	}
	if errs := validation.IsDNS1123Label(request.Project.Id); len(errs) > 0 {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument, "invalid project id [%s]: %v", request.Project.Id, errs)
	}
	if err := ValidateEmptyStringField(request.Project.Name, projectName); err != nil {
		return err
	}
	if err := ValidateMaxLengthStringField(request.Project.Description, projectDescription, maxDescriptionLength); err != nil {
		return err
	}
	if request.Project.Domains != nil {
		return errors.NewFlyteAdminError(codes.InvalidArgument,
			"Domains are currently only set system wide. Please retry without domains included in your request.")
	}
	return nil
}

// Validates that a specified project and domain combination has been registered and exists in the db.
func ValidateProjectAndDomain(
	ctx context.Context, db repositories.RepositoryInterface, config runtimeInterfaces.ApplicationConfiguration, projectID, domainID string) error {
	_, err := db.ProjectRepo().Get(ctx, projectID)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"failed to validate that project [%s] and domain [%s] are registered, err: [%+v]",
			projectID, domainID, err)
	}
	var validDomain bool
	domains := config.GetDomainsConfig()
	for _, domain := range *domains {
		if domain.ID == domainID {
			validDomain = true
			break
		}
	}
	if !validDomain {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument, "domain [%s] is unrecognized by system", domainID)
	}
	return nil
}
