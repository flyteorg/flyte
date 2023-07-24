package resources

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/util"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/validation"
	repo_interface "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/transformers"

	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
)

type ResourceManager struct {
	db     repo_interface.Repository
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

func (m *ResourceManager) UpdateProjectAttributes(ctx context.Context, request admin.ProjectAttributesUpdateRequest) (
	*admin.ProjectAttributesUpdateResponse, error) {

	var resource admin.MatchableResource
	var err error

	if resource, err = validation.ValidateProjectAttributesUpdateRequest(ctx, m.db, request); err != nil {
		return nil, err
	}
	model, err := transformers.ProjectAttributesToResourceModel(*request.Attributes, resource)
	if err != nil {
		return nil, err
	}

	if request.Attributes.GetMatchingAttributes().GetPluginOverrides() != nil {
		return m.createOrMergeUpdateProjectAttributes(ctx, request, model, admin.MatchableResource_PLUGIN_OVERRIDE)
	}

	err = m.db.ResourceRepo().CreateOrUpdate(ctx, model)
	if err != nil {
		return nil, err
	}

	return &admin.ProjectAttributesUpdateResponse{}, nil
}

func (m *ResourceManager) GetProjectAttributesBase(ctx context.Context, request admin.ProjectAttributesGetRequest) (
	*admin.ProjectAttributesGetResponse, error) {

	if err := validation.ValidateProjectExists(ctx, m.db, request.Project); err != nil {
		return nil, err
	}

	projectAttributesModel, err := m.db.ResourceRepo().GetProjectLevel(
		ctx, repo_interface.ResourceID{Project: request.Project, Domain: "", ResourceType: request.ResourceType.String()})
	if err != nil {
		return nil, err
	}

	ma, err := transformers.FromResourceModelToMatchableAttributes(projectAttributesModel)
	if err != nil {
		return nil, err
	}

	return &admin.ProjectAttributesGetResponse{
		Attributes: &admin.ProjectAttributes{
			Project:            request.Project,
			MatchingAttributes: ma.Attributes,
		},
	}, nil
}

// GetProjectAttributes combines the call to the database to get the Project level settings with
// Admin server level configuration.
// Note this merge is only done for WorkflowExecutionConfig
// This code should be removed pending implementation of a complete settings implementation.
func (m *ResourceManager) GetProjectAttributes(ctx context.Context, request admin.ProjectAttributesGetRequest) (
	*admin.ProjectAttributesGetResponse, error) {

	getResponse, err := m.GetProjectAttributesBase(ctx, request)
	configLevelDefaults := m.config.GetTopLevelConfig().GetAsWorkflowExecutionConfig()
	if err != nil {
		ec, ok := err.(errors.FlyteAdminError)
		if ok && ec.Code() == codes.NotFound && request.ResourceType == admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG {
			// TODO: Will likely be removed after overarching settings project is done
			return &admin.ProjectAttributesGetResponse{
				Attributes: &admin.ProjectAttributes{
					Project: request.Project,
					MatchingAttributes: &admin.MatchingAttributes{
						Target: &admin.MatchingAttributes_WorkflowExecutionConfig{
							WorkflowExecutionConfig: &configLevelDefaults,
						},
					},
				},
			}, nil
		}
		return nil, err

	}
	// If found, then merge result with the default values for the platform
	// TODO: Remove this logic once the overarching settings project is done. Those endpoints should take
	//   default configuration into account.
	responseAttributes := getResponse.Attributes.GetMatchingAttributes().GetWorkflowExecutionConfig()
	if responseAttributes != nil {
		logger.Warningf(ctx, "Merging response %s with defaults %s", responseAttributes, configLevelDefaults)
		tmp := util.MergeIntoExecConfig(*responseAttributes, &configLevelDefaults)
		responseAttributes = &tmp
		return &admin.ProjectAttributesGetResponse{
			Attributes: &admin.ProjectAttributes{
				Project: request.Project,
				MatchingAttributes: &admin.MatchingAttributes{
					Target: &admin.MatchingAttributes_WorkflowExecutionConfig{
						WorkflowExecutionConfig: responseAttributes,
					},
				},
			},
		}, nil
	}

	return getResponse, nil
}

func (m *ResourceManager) DeleteProjectAttributes(ctx context.Context, request admin.ProjectAttributesDeleteRequest) (
	*admin.ProjectAttributesDeleteResponse, error) {

	if err := validation.ValidateProjectForUpdate(ctx, m.db, request.Project); err != nil {
		return nil, err
	}
	if err := m.db.ResourceRepo().Delete(
		ctx, repo_interface.ResourceID{Project: request.Project, ResourceType: request.ResourceType.String()}); err != nil {
		return nil, err
	}
	logger.Infof(ctx, "Deleted project attributes for: %s-%s (%s)", request.Project, request.ResourceType.String())
	return &admin.ProjectAttributesDeleteResponse{}, nil
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
	updatedModel, err := transformers.MergeUpdatePluginAttributes(
		ctx, existing, resourceType, &resourceID, request.Attributes.MatchingAttributes)
	if err != nil {
		return nil, err
	}
	err = m.db.ResourceRepo().CreateOrUpdate(ctx, updatedModel)
	if err != nil {
		return nil, err
	}
	return &admin.ProjectDomainAttributesUpdateResponse{}, nil
}

func (m *ResourceManager) createOrMergeUpdateProjectAttributes(
	ctx context.Context, request admin.ProjectAttributesUpdateRequest, model models.Resource,
	resourceType admin.MatchableResource) (*admin.ProjectAttributesUpdateResponse, error) {

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
			return &admin.ProjectAttributesUpdateResponse{}, nil
		}
		return nil, err
	}
	updatedModel, err := transformers.MergeUpdatePluginAttributes(
		ctx, existing, resourceType, &resourceID, request.Attributes.MatchingAttributes)
	if err != nil {
		return nil, err
	}
	err = m.db.ResourceRepo().CreateOrUpdate(ctx, updatedModel)
	if err != nil {
		return nil, err
	}
	return &admin.ProjectAttributesUpdateResponse{}, nil
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

func NewResourceManager(db repo_interface.Repository, config runtimeInterfaces.ApplicationConfiguration) interfaces.ResourceInterface {
	return &ResourceManager{
		db:     db,
		config: config,
	}
}
