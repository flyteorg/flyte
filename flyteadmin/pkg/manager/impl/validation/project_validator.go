package validation

import (
	"context"

	repositoryInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/shared"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"google.golang.org/grpc/codes"
	"k8s.io/apimachinery/pkg/util/validation"
)

const projectID = "project_id"
const projectName = "project_name"
const projectDescription = "project_description"
const maxNameLength = 64
const maxDescriptionLength = 300
const maxLabelArrayLength = 16

func ValidateProjectRegisterRequest(request admin.ProjectRegisterRequest) error {
	if request.Project == nil {
		return shared.GetMissingArgumentError(shared.Project)
	}
	project := *request.Project
	if err := ValidateEmptyStringField(project.Name, projectName); err != nil {
		return err
	}
	return ValidateProject(project)
}

func ValidateProject(project admin.Project) error {
	if err := ValidateEmptyStringField(project.Id, projectID); err != nil {
		return err
	}
	if err := validateLabels(project.Labels); err != nil {
		return err
	}
	if errs := validation.IsDNS1123Label(project.Id); len(errs) > 0 {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument, "invalid project id [%s]: %v", project.Id, errs)
	}
	if err := ValidateMaxLengthStringField(project.Name, projectName, maxNameLength); err != nil {
		return err
	}
	if err := ValidateMaxLengthStringField(project.Description, projectDescription, maxDescriptionLength); err != nil {
		return err
	}
	if project.Domains != nil {
		return errors.NewFlyteAdminError(codes.InvalidArgument,
			"Domains are currently only set system wide. Please retry without domains included in your request.")
	}
	return nil
}

// Validates that a specified project and domain combination has been registered and exists in the db.
func ValidateProjectAndDomain(
	ctx context.Context, db repositoryInterfaces.Repository, config runtimeInterfaces.ApplicationConfiguration, projectID, domainID string) error {
	project, err := db.ProjectRepo().Get(ctx, projectID)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"failed to validate that project [%s] and domain [%s] are registered, err: [%+v]",
			projectID, domainID, err)
	}
	if *project.State != int32(admin.Project_ACTIVE) {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"project [%s] is not active", projectID)
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

func ValidateProjectForUpdate(
	ctx context.Context, db repositoryInterfaces.Repository, projectID string) error {

	project, err := db.ProjectRepo().Get(ctx, projectID)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"failed to validate that project [%s] is registered, err: [%+v]",
			projectID, err)
	}
	if *project.State != int32(admin.Project_ACTIVE) {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"project [%s] is not active", projectID)
	}
	return nil
}

// ValidateProjectExists doesn't check that the project is active. This is used to get Project level attributes, which you should
// be able to do even for inactive projects.
func ValidateProjectExists(
	ctx context.Context, db repositoryInterfaces.Repository, projectID string) error {

	_, err := db.ProjectRepo().Get(ctx, projectID)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"failed to validate that project [%s] exists, err: [%+v]",
			projectID, err)
	}
	return nil
}
