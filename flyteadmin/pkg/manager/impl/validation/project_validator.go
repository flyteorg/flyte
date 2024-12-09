package validation

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"k8s.io/apimachinery/pkg/util/validation"

	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/shared"
	repositoryInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

const projectID = "project_id"
const projectName = "project_name"
const projectDescription = "project_description"
const maxNameLength = 64
const maxDescriptionLength = 300
const maxLabelArrayLength = 16

const orgErrorMsg = " and org '%s'"

func getOrgForErrorMsg(org string) string {
	if len(org) == 0 {
		return ""
	}
	return fmt.Sprintf(orgErrorMsg, org)
}

func ValidateProjectRegisterRequest(request *admin.ProjectRegisterRequest) error {
	if request.Project == nil {
		return shared.GetMissingArgumentError(shared.Project)
	}
	project := request.Project
	if err := ValidateEmptyStringField(project.Name, projectName); err != nil {
		return err
	}
	return ValidateProject(project)
}

func ValidateProjectGetRequest(request *admin.ProjectGetRequest) error {
	if err := ValidateEmptyStringField(request.Id, projectID); err != nil {
		return err
	}
	return nil
}

func ValidateProject(project *admin.Project) error {
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
	ctx context.Context, db repositoryInterfaces.Repository, config runtimeInterfaces.ApplicationConfiguration, projectID, domainID, org string) error {
	if err := ValidateProjectExistsAndActive(ctx, db, projectID, org); err != nil {
		return err
	}
	if err := ValidateDomainExists(ctx, config, domainID); err != nil {
		return err
	}
	return nil
}

func ValidateProjectExistsAndActive(
	ctx context.Context, db repositoryInterfaces.Repository, projectID, org string) error {
	project, err := db.ProjectRepo().Get(ctx, projectID, org)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"failed to validate that project [%s]%s is registered, err: [%+v]",
			projectID, getOrgForErrorMsg(org), err)
	}

	if *project.State != int32(admin.Project_ACTIVE) {
		return errors.NewInactiveProjectError(ctx, projectID)
	}
	return nil
}

// ValidateProjectAndDomainForRegistration validates that a specified project and domain combination has been registered and exists in the db.
// This method is used for task and workflow creation or registration, where the project may be in a system archived state.
func ValidateProjectAndDomainForRegistration(
	ctx context.Context, db repositoryInterfaces.Repository, config runtimeInterfaces.ApplicationConfiguration, projectID, domainID, org string) error {
	if err := ValidateProjectExistsAndActiveOrSystemArchived(ctx, db, projectID, org); err != nil {
		return err
	}
	if err := ValidateDomainExists(ctx, config, domainID); err != nil {
		return err
	}
	return nil
}

// ValidateProjectExistsAndActiveOrSystemArchived validates that a specified project exists and is either active or system archived.
func ValidateProjectExistsAndActiveOrSystemArchived(
	ctx context.Context, db repositoryInterfaces.Repository, projectID, org string) error {
	project, err := db.ProjectRepo().Get(ctx, projectID, org)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"failed to validate that project [%s]%s is registered, err: [%+v]",
			projectID, getOrgForErrorMsg(org), err)
	}

	if *project.State != int32(admin.Project_ACTIVE) && *project.State != int32(admin.Project_SYSTEM_ARCHIVED) {
		return errors.NewInactiveProjectError(ctx, projectID)
	}
	return nil
}

// ValidateProjectExists doesn't check that the project is active. This is used to get Project level attributes, which you should
// be able to do even for inactive projects.
func ValidateProjectExists(
	ctx context.Context, db repositoryInterfaces.Repository, projectID, org string) error {

	_, err := db.ProjectRepo().Get(ctx, projectID, org)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"failed to validate that project [%s]%s exists, err: [%+v]",
			projectID, getOrgForErrorMsg(org), err)
	}
	return nil
}

func ValidateDomainExists(ctx context.Context, config runtimeInterfaces.ApplicationConfiguration, domainID string) error {
	var validDomain bool
	domains := config.GetDomainsConfig()
	for _, domain := range *domains {
		if domain.ID == domainID {
			validDomain = true
			break
		}
	}
	if !validDomain {
		logger.Errorf(ctx, "domain [%s] is unrecognized by system", domainID)
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument, "domain [%s] is unrecognized by system", domainID)
	}
	return nil
}
