package resources

import (
	"context"

	"github.com/lyft/flyteadmin/pkg/repositories/models"

	"github.com/gogo/protobuf/proto"
	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flytestdlib/contextutils"
	"google.golang.org/grpc/codes"

	"github.com/lyft/flytestdlib/logger"

	"github.com/lyft/flyteadmin/pkg/manager/impl/validation"
	repo_interface "github.com/lyft/flyteadmin/pkg/repositories/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories/transformers"

	"github.com/lyft/flyteadmin/pkg/manager/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"

	runtimeInterfaces "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
)

type ResourceManager struct {
	db     repositories.RepositoryInterface
	config runtimeInterfaces.ApplicationConfiguration
}

func (m *ResourceManager) GetResource(ctx context.Context, request interfaces.ResourceRequest) (*interfaces.ResourceResponse, error) {
	resource, err := m.db.ResourceRepo().Get(ctx, repo_interface.ResourceID{
		ResourceType: request.ResourceType.String(),
		Project:      request.Project,
		Domain:       request.Domain,
		Workflow:     request.Workflow,
		LaunchPlan:   request.LaunchPlan,
	})
	if err != nil {
		return nil, err
	}

	var attributes admin.MatchingAttributes
	err = proto.Unmarshal(resource.Attributes, &attributes)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(
			codes.Internal, "Failed to decode resource attribute with err: %v", err)
	}
	return &interfaces.ResourceResponse{
		ResourceType: resource.ResourceType,
		Project:      resource.Project,
		Domain:       resource.Domain,
		Workflow:     resource.Workflow,
		LaunchPlan:   resource.LaunchPlan,
		Attributes:   &attributes,
	}, nil
}

func (m *ResourceManager) createOrMergeUpdateWorkflowAttributes(
	ctx context.Context, request admin.WorkflowAttributesUpdateRequest, model models.Resource,
	resourceType admin.MatchableResource) (*admin.WorkflowAttributesUpdateResponse, error) {
	resourceID := repo_interface.ResourceID{
		Project:      model.Project,
		Domain:       model.Domain,
		Workflow:     model.Workflow,
		LaunchPlan:   model.LaunchPlan,
		ResourceType: model.ResourceType,
	}
	existing, err := m.db.ResourceRepo().GetRaw(ctx, resourceID)
	if err != nil {
		ec, ok := err.(errors.FlyteAdminError)
		if ok && ec.Code() == codes.NotFound {
			// Proceed with the default CreateOrUpdate call since there's no existing model to update.
			err = m.db.ResourceRepo().CreateOrUpdate(ctx, model)
			if err != nil {
				return nil, err
			}
			return &admin.WorkflowAttributesUpdateResponse{}, nil
		}
		return nil, err
	}
	updatedModel, err := transformers.MergeUpdateWorkflowAttributes(
		ctx, existing, resourceType, &resourceID, request.Attributes)
	if err != nil {
		return nil, err
	}
	err = m.db.ResourceRepo().CreateOrUpdate(ctx, updatedModel)
	if err != nil {
		return nil, err
	}
	return &admin.WorkflowAttributesUpdateResponse{}, nil
}

func (m *ResourceManager) UpdateWorkflowAttributes(
	ctx context.Context, request admin.WorkflowAttributesUpdateRequest) (
	*admin.WorkflowAttributesUpdateResponse, error) {
	var resource admin.MatchableResource
	var err error
	if resource, err = validation.ValidateWorkflowAttributesUpdateRequest(ctx, m.db, m.config, request); err != nil {
		return nil, err
	}

	model, err := transformers.WorkflowAttributesToResourceModel(*request.Attributes, resource)
	if err != nil {
		return nil, err
	}
	if request.Attributes.GetMatchingAttributes().GetPluginOverrides() != nil {
		return m.createOrMergeUpdateWorkflowAttributes(ctx, request, model, admin.MatchableResource_PLUGIN_OVERRIDE)
	}
	err = m.db.ResourceRepo().CreateOrUpdate(ctx, model)
	if err != nil {
		return nil, err
	}

	return &admin.WorkflowAttributesUpdateResponse{}, nil
}

func (m *ResourceManager) GetWorkflowAttributes(
	ctx context.Context, request admin.WorkflowAttributesGetRequest) (
	*admin.WorkflowAttributesGetResponse, error) {
	if err := validation.ValidateWorkflowAttributesGetRequest(ctx, m.db, m.config, request); err != nil {
		return nil, err
	}
	workflowAttributesModel, err := m.db.ResourceRepo().Get(
		ctx, repo_interface.ResourceID{Project: request.Project, Domain: request.Domain, Workflow: request.Workflow, ResourceType: request.ResourceType.String()})
	if err != nil {
		return nil, err
	}
	workflowAttributes, err := transformers.FromResourceModelToWorkflowAttributes(workflowAttributesModel)
	if err != nil {
		return nil, err
	}
	return &admin.WorkflowAttributesGetResponse{
		Attributes: &workflowAttributes,
	}, nil
}

func (m *ResourceManager) DeleteWorkflowAttributes(ctx context.Context,
	request admin.WorkflowAttributesDeleteRequest) (*admin.WorkflowAttributesDeleteResponse, error) {
	if err := validation.ValidateWorkflowAttributesDeleteRequest(ctx, m.db, m.config, request); err != nil {
		return nil, err
	}
	if err := m.db.ResourceRepo().Delete(
		ctx, repo_interface.ResourceID{Project: request.Project, Domain: request.Domain, Workflow: request.Workflow, ResourceType: request.ResourceType.String()}); err != nil {
		return nil, err
	}
	logger.Infof(ctx, "Deleted workflow attributes for: %s-%s-%s (%s)", request.Project,
		request.Domain, request.Workflow, request.ResourceType.String())
	return &admin.WorkflowAttributesDeleteResponse{}, nil
}

func (m *ResourceManager) createOrMergeUpdateProjectDomainAttributes(
	ctx context.Context, request admin.ProjectDomainAttributesUpdateRequest, model models.Resource,
	resourceType admin.MatchableResource) (*admin.ProjectDomainAttributesUpdateResponse, error) {
	resourceID := repo_interface.ResourceID{
		Project:      model.Project,
		Domain:       model.Domain,
		Workflow:     model.Workflow,
		LaunchPlan:   model.LaunchPlan,
		ResourceType: model.ResourceType,
	}
	existing, err := m.db.ResourceRepo().GetRaw(ctx, resourceID)
	if err != nil {
		ec, ok := err.(errors.FlyteAdminError)
		if ok && ec.Code() == codes.NotFound {
			// Proceed with the default CreateOrUpdate call since there's no existing model to update.
			err = m.db.ResourceRepo().CreateOrUpdate(ctx, model)
			if err != nil {
				return nil, err
			}
			return &admin.ProjectDomainAttributesUpdateResponse{}, nil
		}
		return nil, err
	}
	updatedModel, err := transformers.MergeUpdateProjectDomainAttributes(
		ctx, existing, resourceType, &resourceID, request.Attributes)
	if err != nil {
		return nil, err
	}
	err = m.db.ResourceRepo().CreateOrUpdate(ctx, updatedModel)
	if err != nil {
		return nil, err
	}
	return &admin.ProjectDomainAttributesUpdateResponse{}, nil
}

func (m *ResourceManager) UpdateProjectDomainAttributes(
	ctx context.Context, request admin.ProjectDomainAttributesUpdateRequest) (
	*admin.ProjectDomainAttributesUpdateResponse, error) {
	var resource admin.MatchableResource
	var err error
	if resource, err = validation.ValidateProjectDomainAttributesUpdateRequest(ctx, m.db, m.config, request); err != nil {
		return nil, err
	}
	ctx = contextutils.WithProjectDomain(ctx, request.Attributes.Project, request.Attributes.Domain)

	model, err := transformers.ProjectDomainAttributesToResourceModel(*request.Attributes, resource)
	if err != nil {
		return nil, err
	}
	if request.Attributes.GetMatchingAttributes().GetPluginOverrides() != nil {
		return m.createOrMergeUpdateProjectDomainAttributes(ctx, request, model, admin.MatchableResource_PLUGIN_OVERRIDE)
	}
	err = m.db.ResourceRepo().CreateOrUpdate(ctx, model)
	if err != nil {
		return nil, err
	}
	return &admin.ProjectDomainAttributesUpdateResponse{}, nil
}

func (m *ResourceManager) GetProjectDomainAttributes(
	ctx context.Context, request admin.ProjectDomainAttributesGetRequest) (
	*admin.ProjectDomainAttributesGetResponse, error) {
	if err := validation.ValidateProjectDomainAttributesGetRequest(ctx, m.db, m.config, request); err != nil {
		return nil, err
	}
	projectAttributesModel, err := m.db.ResourceRepo().Get(
		ctx, repo_interface.ResourceID{Project: request.Project, Domain: request.Domain, ResourceType: request.ResourceType.String()})
	if err != nil {
		return nil, err
	}
	projectAttributes, err := transformers.FromResourceModelToProjectDomainAttributes(projectAttributesModel)
	if err != nil {
		return nil, err
	}
	return &admin.ProjectDomainAttributesGetResponse{
		Attributes: &projectAttributes,
	}, nil
}

func (m *ResourceManager) DeleteProjectDomainAttributes(ctx context.Context,
	request admin.ProjectDomainAttributesDeleteRequest) (*admin.ProjectDomainAttributesDeleteResponse, error) {
	if err := validation.ValidateProjectDomainAttributesDeleteRequest(ctx, m.db, m.config, request); err != nil {
		return nil, err
	}
	if err := m.db.ResourceRepo().Delete(
		ctx, repo_interface.ResourceID{Project: request.Project, Domain: request.Domain, ResourceType: request.ResourceType.String()}); err != nil {
		return nil, err
	}
	logger.Infof(ctx, "Deleted project-domain attributes for: %s-%s (%s)", request.Project,
		request.Domain, request.ResourceType.String())
	return &admin.ProjectDomainAttributesDeleteResponse{}, nil
}

func (m *ResourceManager) ListAll(ctx context.Context, request admin.ListMatchableAttributesRequest) (
	*admin.ListMatchableAttributesResponse, error) {
	if err := validation.ValidateListAllMatchableAttributesRequest(request); err != nil {
		return nil, err
	}
	resources, err := m.db.ResourceRepo().ListAll(ctx, request.ResourceType.String())
	if err != nil {
		return nil, err
	}
	if resources == nil {
		// That's fine - there don't necessarily need to exist overrides in the database
		return &admin.ListMatchableAttributesResponse{}, nil
	}
	configurations, err := transformers.FromResourceModelsToMatchableAttributes(resources)
	if err != nil {
		return nil, err
	}
	return &admin.ListMatchableAttributesResponse{
		Configurations: configurations,
	}, nil
}

func NewResourceManager(db repositories.RepositoryInterface, config runtimeInterfaces.ApplicationConfiguration) interfaces.ResourceInterface {
	return &ResourceManager{
		db:     db,
		config: config,
	}
}
