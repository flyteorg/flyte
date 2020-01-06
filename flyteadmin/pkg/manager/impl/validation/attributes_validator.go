package validation

import (
	"fmt"

	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/manager/impl/shared"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"google.golang.org/grpc/codes"
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
	}
	return defaultMatchableResource, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
		"Unrecognized matching attributes type for request %s", identifier)
}

func ValidateProjectAttributesUpdateRequest(request admin.ProjectAttributesUpdateRequest) (
	admin.MatchableResource, error) {
	if request.Attributes == nil {
		return defaultMatchableResource, shared.GetMissingArgumentError(shared.Attributes)
	}
	if err := ValidateEmptyStringField(request.Attributes.Project, shared.Project); err != nil {
		return defaultMatchableResource, err
	}

	return validateMatchingAttributes(request.Attributes.MatchingAttributes, request.Attributes.Project)
}

func ValidateProjectAttributesGetRequest(request admin.ProjectAttributesGetRequest) error {
	if err := ValidateEmptyStringField(request.Project, shared.Project); err != nil {
		return err
	}

	return nil
}

func ValidateProjectAttributesDeleteRequest(request admin.ProjectAttributesDeleteRequest) error {
	if err := ValidateEmptyStringField(request.Project, shared.Project); err != nil {
		return err
	}

	return nil
}

func ValidateProjectDomainAttributesUpdateRequest(request admin.ProjectDomainAttributesUpdateRequest) (
	admin.MatchableResource, error) {
	if request.Attributes == nil {
		return defaultMatchableResource, shared.GetMissingArgumentError(shared.Attributes)
	}
	if err := ValidateEmptyStringField(request.Attributes.Project, shared.Project); err != nil {
		return defaultMatchableResource, err
	}
	if err := ValidateEmptyStringField(request.Attributes.Domain, shared.Domain); err != nil {
		return defaultMatchableResource, err
	}

	return validateMatchingAttributes(request.Attributes.MatchingAttributes,
		fmt.Sprintf("%s-%s", request.Attributes.Project, request.Attributes.Domain))
}

func ValidateProjectDomainAttributesGetRequest(request admin.ProjectDomainAttributesGetRequest) error {
	if err := ValidateEmptyStringField(request.Project, shared.Project); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(request.Domain, shared.Domain); err != nil {
		return err
	}

	return nil
}

func ValidateProjectDomainAttributesDeleteRequest(request admin.ProjectDomainAttributesDeleteRequest) error {
	if err := ValidateEmptyStringField(request.Project, shared.Project); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(request.Domain, shared.Domain); err != nil {
		return err
	}

	return nil
}

func ValidateWorkflowAttributesUpdateRequest(request admin.WorkflowAttributesUpdateRequest) (
	admin.MatchableResource, error) {
	if request.Attributes == nil {
		return defaultMatchableResource, shared.GetMissingArgumentError(shared.Attributes)
	}
	if err := ValidateEmptyStringField(request.Attributes.Project, shared.Project); err != nil {
		return defaultMatchableResource, err
	}
	if err := ValidateEmptyStringField(request.Attributes.Domain, shared.Domain); err != nil {
		return defaultMatchableResource, err
	}
	if err := ValidateEmptyStringField(request.Attributes.Workflow, shared.Name); err != nil {
		return defaultMatchableResource, err
	}

	return validateMatchingAttributes(request.Attributes.MatchingAttributes,
		fmt.Sprintf("%s-%s-%s", request.Attributes.Project, request.Attributes.Domain, request.Attributes.Workflow))
}

func ValidateWorkflowAttributesGetRequest(request admin.WorkflowAttributesGetRequest) error {
	if err := ValidateEmptyStringField(request.Project, shared.Project); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(request.Domain, shared.Domain); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(request.Workflow, shared.Name); err != nil {
		return err
	}

	return nil
}

func ValidateWorkflowAttributesDeleteRequest(request admin.WorkflowAttributesDeleteRequest) error {
	if err := ValidateEmptyStringField(request.Project, shared.Project); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(request.Domain, shared.Domain); err != nil {
		return err
	}
	if err := ValidateEmptyStringField(request.Workflow, shared.Name); err != nil {
		return err
	}

	return nil
}
