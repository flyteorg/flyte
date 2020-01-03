package resourcematching

import (
	"context"

	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/repositories"
	"github.com/lyft/flyteadmin/pkg/repositories/transformers"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"google.golang.org/grpc/codes"
)

type GetOverrideValuesInput struct {
	Db       repositories.RepositoryInterface
	Project  string
	Domain   string
	Workflow string
	Resource admin.MatchableResource
}

func isNotFoundErr(err error) bool {
	return err.(errors.FlyteAdminError) != nil && err.(errors.FlyteAdminError).Code() == codes.NotFound
}

func GetOverrideValuesToApply(ctx context.Context, input GetOverrideValuesInput) (
	*admin.MatchingAttributes, error) {
	if len(input.Project) == 0 || len(input.Domain) == 0 {
		return nil, errors.NewFlyteAdminErrorf(
			codes.InvalidArgument, "Invalid overrides values request configuration: [%+v]", input)
	}
	if len(input.Workflow) > 0 {
		// Only the workflow input argument is optional
		workflowAttributesModel, err := input.Db.WorkflowAttributesRepo().Get(
			ctx, input.Project, input.Domain, input.Workflow, input.Resource.String())
		if err != nil && !isNotFoundErr(err) {
			// Not found is fine, since not every workflow will necessarily have resource overrides.
			// Any other error should be bubbled back up.
			return nil, err
		} else if err == nil {
			workflowAttributes, err := transformers.FromWorkflowAttributesModel(workflowAttributesModel)
			if err != nil {
				return nil, err
			}
			return workflowAttributes.MatchingAttributes, nil
		}
	}

	projectDomainAttributesModel, err := input.Db.ProjectDomainAttributesRepo().Get(
		ctx, input.Project, input.Domain, input.Resource.String())
	if err != nil && !isNotFoundErr(err) {
		// Not found is fine, since not every project+domain will necessarily have resource overrides.
		// Any other error should be bubbled back up.
		return nil, err
	} else if err == nil {
		projectDomainAttributes, err := transformers.FromProjectDomainAttributesModel(projectDomainAttributesModel)
		if err != nil {
			return nil, err
		}
		return projectDomainAttributes.MatchingAttributes, nil
	}

	projectAttributesModel, err := input.Db.ProjectAttributesRepo().Get(ctx, input.Project, input.Resource.String())
	if err != nil && !isNotFoundErr(err) {
		// Not found is fine, since not every project will necessarily have resource overrides.
		// Any other error should be bubbled back up.
		return nil, err
	} else if err == nil {
		projectAttributes, err := transformers.FromProjectAttributesModel(projectAttributesModel)
		if err != nil {
			return nil, err
		}
		return projectAttributes.MatchingAttributes, nil
	}

	// If we've made it this far then there are no matching overrides.
	return nil, nil
}
