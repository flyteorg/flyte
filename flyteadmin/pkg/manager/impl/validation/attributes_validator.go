package validation

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/shared"
	repositoryInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

var defaultMatchableResource = admin.MatchableResource(-1)

func validateMatchingAttributes(attributes *admin.MatchingAttributes, identifier string) (admin.MatchableResource, error) {
	if attributes == nil {
		return defaultMatchableResource, shared.GetMissingArgumentError(shared.MatchingAttributes)
	}
	if attributes.GetTaskResourceAttributes() != nil {
		return admin.MatchableResource_TASK_RESOURCE, nil
	} else if attributes.GetClusterResourceAttributes() != nil {
		return admin.MatchableResource_CLUSTER_RESOURCE, nil
	} else if attributes.GetExecutionQueueAttributes() != nil {
		return admin.MatchableResource_EXECUTION_QUEUE, nil
	} else if attributes.GetExecutionClusterLabel() != nil {
		return admin.MatchableResource_EXECUTION_CLUSTER_LABEL, nil
	} else if attributes.GetPluginOverrides() != nil {
		return admin.MatchableResource_PLUGIN_OVERRIDE, nil
	} else if attributes.GetWorkflowExecutionConfig() != nil {
		return admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG, nil
	} else if attributes.GetClusterAssignment() != nil {
		return admin.MatchableResource_CLUSTER_ASSIGNMENT, nil
	}
	return defaultMatchableResource, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
		"Unrecognized matching attributes type for request %s", identifier)
}

func ValidateProjectDomainAttributesUpdateRequest(ctx context.Context,
	db repositoryInterfaces.Repository, config runtimeInterfaces.ApplicationConfiguration,
	request *admin.ProjectDomainAttributesUpdateRequest) (
	admin.MatchableResource, error) {
	if request.GetAttributes() == nil {
		return defaultMatchableResource, shared.GetMissingArgumentError(shared.Attributes)
	}
	if err := ValidateProjectAndDomain(ctx, db, config, request.GetAttributes().GetProject(), request.GetAttributes().GetDomain()); err != nil {
		return defaultMatchableResource, err
	}

	return validateMatchingAttributes(request.GetAttributes().GetMatchingAttributes(),
		fmt.Sprintf("%s-%s", request.GetAttributes().GetProject(), request.GetAttributes().GetDomain()))
}

func ValidateProjectAttributesUpdateRequest(ctx context.Context,
	db repositoryInterfaces.Repository,
	request *admin.ProjectAttributesUpdateRequest) (
	admin.MatchableResource, error) {

	if request.GetAttributes() == nil {
		return defaultMatchableResource, shared.GetMissingArgumentError(shared.Attributes)
	}
	if err := ValidateProjectForUpdate(ctx, db, request.GetAttributes().GetProject()); err != nil {
		return defaultMatchableResource, err
	}

	return validateMatchingAttributes(request.GetAttributes().GetMatchingAttributes(), request.GetAttributes().GetProject())
}

func ValidateProjectDomainAttributesGetRequest(ctx context.Context, db repositoryInterfaces.Repository,
	config runtimeInterfaces.ApplicationConfiguration, request *admin.ProjectDomainAttributesGetRequest) error {
	if err := ValidateProjectAndDomain(ctx, db, config, request.GetProject(), request.GetDomain()); err != nil {
		return err
	}

	return nil
}

func ValidateProjectDomainAttributesDeleteRequest(ctx context.Context, db repositoryInterfaces.Repository,
	config runtimeInterfaces.ApplicationConfiguration, request *admin.ProjectDomainAttributesDeleteRequest) error {
	if err := ValidateProjectAndDomain(ctx, db, config, request.GetProject(), request.GetDomain()); err != nil {
		return err
	}

	return nil
}

func ValidateWorkflowAttributesUpdateRequest(ctx context.Context, db repositoryInterfaces.Repository,
	config runtimeInterfaces.ApplicationConfiguration, request *admin.WorkflowAttributesUpdateRequest) (
	admin.MatchableResource, error) {
	if request.GetAttributes() == nil {
		return defaultMatchableResource, shared.GetMissingArgumentError(shared.Attributes)
	}
	if err := ValidateProjectAndDomain(ctx, db, config, request.GetAttributes().GetProject(), request.GetAttributes().GetDomain()); err != nil {
		return defaultMatchableResource, err
	}
	if err := ValidateEmptyStringField(request.GetAttributes().GetWorkflow(), shared.Name); err != nil {
		return defaultMatchableResource, err
	}

	return validateMatchingAttributes(request.GetAttributes().GetMatchingAttributes(),
		fmt.Sprintf("%s-%s-%s", request.GetAttributes().GetProject(), request.GetAttributes().GetDomain(), request.GetAttributes().GetWorkflow()))
}

func ValidateWorkflowAttributesGetRequest(ctx context.Context, db repositoryInterfaces.Repository,
	config runtimeInterfaces.ApplicationConfiguration, request *admin.WorkflowAttributesGetRequest) error {
	if err := ValidateProjectAndDomain(ctx, db, config, request.GetProject(), request.GetDomain()); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(request.GetWorkflow(), shared.Name); err != nil {
		return err
	}

	return nil
}

func ValidateWorkflowAttributesDeleteRequest(ctx context.Context, db repositoryInterfaces.Repository,
	config runtimeInterfaces.ApplicationConfiguration, request *admin.WorkflowAttributesDeleteRequest) error {
	if err := ValidateProjectAndDomain(ctx, db, config, request.GetProject(), request.GetDomain()); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(request.GetWorkflow(), shared.Name); err != nil {
		return err
	}

	return nil
}

func ValidateListAllMatchableAttributesRequest(request *admin.ListMatchableAttributesRequest) error {
	if _, ok := admin.MatchableResource_name[int32(request.GetResourceType())]; !ok {
		return shared.GetInvalidArgumentError(shared.ResourceType)
	}
	return nil
}
