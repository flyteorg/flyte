package validation

import (
	"context"

	"github.com/lyft/flyteadmin/pkg/manager/impl/shared"

	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/repositories"
	runtimeInterfaces "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"google.golang.org/grpc/codes"
	"k8s.io/apimachinery/pkg/util/validation"
)

const projectID = "project_id"
const projectName = "project_name"
const projectDescription = "project_description"
const labels = "labels"
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
	if err := validateProjectLabels(project); err != nil {
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

func validateProjectLabels(project admin.Project) error {
	if project.Labels == nil || len(project.Labels.Values) == 0 {
		return nil
	}
	if err := ValidateMaxMapLengthField(project.Labels.Values, labels, maxLabelArrayLength); err != nil {
		return err
	}
	if err := validateProjectLabelsAlphanumeric(project.Labels); err != nil {
		return err
	}
	return nil
}

// Validates that a specified project and domain combination has been registered and exists in the db.
func ValidateProjectAndDomain(
	ctx context.Context, db repositories.RepositoryInterface, config runtimeInterfaces.ApplicationConfiguration, projectID, domainID string) error {
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

// Given an admin.Project, checks if the project has labels and if it does, checks if the labels are K8s compliant,
// i.e. alphanumeric + - and _
func validateProjectLabelsAlphanumeric(labels *admin.Labels) error {
	for key, value := range labels.Values {
		if errs := validation.IsDNS1123Label(key); len(errs) > 0 {
			return errors.NewFlyteAdminErrorf(codes.InvalidArgument, "invalid label key [%s]: %v", key, errs)
		}
		if errs := validation.IsDNS1123Label(value); len(errs) > 0 {
			return errors.NewFlyteAdminErrorf(codes.InvalidArgument, "invalid label value [%s]: %v", value, errs)
		}
	}
	return nil
}
