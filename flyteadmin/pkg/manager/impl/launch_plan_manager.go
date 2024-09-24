package impl

import (
	"bytes"
	"context"
	"strconv"

	"github.com/golang/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/artifacts"
	scheduleInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/async/schedule/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/util"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/validation"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/interfaces"
	repoInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/transformers"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/plugins"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

type launchPlanMetrics struct {
	Scope                 promutils.Scope
	FailedScheduleUpdates prometheus.Counter
	FailedTriggerUpdates  prometheus.Counter
	SpecSizeBytes         prometheus.Summary
	ClosureSizeBytes      prometheus.Summary
}

type LaunchPlanManager struct {
	db                 repoInterfaces.Repository
	config             runtimeInterfaces.Configuration
	scheduler          scheduleInterfaces.EventScheduler
	metrics            launchPlanMetrics
	artifactRegistry   *artifacts.ArtifactRegistry
	pluginRegistry     *plugins.Registry
	storageClient      *storage.DataStore
	workflowManager    interfaces.WorkflowInterface
	namedEntityManager interfaces.NamedEntityInterface
}

func getLaunchPlanContext(ctx context.Context, identifier *core.Identifier) context.Context {
	ctx = contextutils.WithProjectDomain(ctx, identifier.Project, identifier.Domain)
	return contextutils.WithLaunchPlanID(ctx, identifier.Name)
}

func (m *LaunchPlanManager) getNamedEntityContext(ctx context.Context, identifier *admin.NamedEntityIdentifier) context.Context {
	ctx = contextutils.WithProjectDomain(ctx, identifier.Project, identifier.Domain)
	return contextutils.WithLaunchPlanID(ctx, identifier.Name)
}

func (m *LaunchPlanManager) CreateLaunchPlan(
	ctx context.Context,
	request admin.LaunchPlanCreateRequest) (*admin.LaunchPlanCreateResponse, error) {
	if err := validation.ValidateIdentifier(request.GetSpec().GetWorkflowId(), common.Workflow); err != nil {
		logger.Debugf(ctx, "Failed to validate provided workflow ID for CreateLaunchPlan with err: %v", err)
		return nil, err
	}
	workflowModel, err := util.GetWorkflowModel(ctx, m.db, *request.Spec.WorkflowId)
	if err != nil {
		logger.Debugf(ctx, "Failed to get workflow with id [%+v] for CreateLaunchPlan with id [%+v] with err %v",
			*request.Spec.WorkflowId, request.Id)
		return nil, err
	}
	var workflowInterface core.TypedInterface
	if workflowModel.TypedInterface != nil && len(workflowModel.TypedInterface) > 0 {
		err = proto.Unmarshal(workflowModel.TypedInterface, &workflowInterface)
		if err != nil {
			logger.Errorf(ctx,
				"Failed to unmarshal TypedInterface for workflow [%+v] with err: %v",
				*request.Spec.WorkflowId, err)
			return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to unmarshal workflow inputs")
		}
	}
	if err := validation.ValidateLaunchPlan(ctx, request, m.db, m.config.ApplicationConfiguration(), &workflowInterface); err != nil {
		logger.Debugf(ctx, "could not create launch plan: %+v, request failed validation with err: %v", request.Id, err)
		return nil, err
	}
	ctx = getLaunchPlanContext(ctx, request.Id)
	launchPlan := transformers.CreateLaunchPlan(request, workflowInterface.Outputs)
	launchPlanDigest, err := util.GetLaunchPlanDigest(ctx, &launchPlan)
	if err != nil {
		logger.Errorf(ctx, "failed to compute launch plan digest for [%+v] with err: %v", launchPlan.Id, err)
		return nil, err
	}

	// The presence of this field indicates that this is a trigger launch plan
	// Return true and send this request over to the artifact registry instead
	if launchPlan.Spec.GetEntityMetadata() != nil && launchPlan.Spec.GetEntityMetadata().GetLaunchConditions() != nil {
		// TODO: Artifact feature gate, remove when ready
		if m.artifactRegistry.GetClient() == nil {
			logger.Debugf(ctx, "artifact feature not enabled, skipping launch plan %v", launchPlan.Id)
			return &admin.LaunchPlanCreateResponse{}, nil
		}
		err := m.artifactRegistry.RegisterTrigger(ctx, &launchPlan)
		if err != nil {
			return nil, err
		}
		logger.Debugf(ctx, "successfully registered trigger launch plan, continuing %v", launchPlan.Id)
	}

	existingLaunchPlanModel, err := util.GetLaunchPlanModel(ctx, m.db, *request.Id)
	if err == nil {
		if bytes.Equal(existingLaunchPlanModel.Digest, launchPlanDigest) {
			return nil, errors.NewLaunchPlanExistsIdenticalStructureError(ctx, &request)
		}
		existingLaunchPlan, transformerErr := transformers.FromLaunchPlanModel(existingLaunchPlanModel)
		if transformerErr != nil {
			logger.Errorf(ctx, "failed to transform launch plan from launch plan model")
			return nil, transformerErr
		}
		// A launch plan exists with different structure
		return nil, errors.NewLaunchPlanExistsDifferentStructureError(ctx, &request, existingLaunchPlan.Spec, launchPlan.Spec)
	}

	launchPlanModel, err :=
		transformers.CreateLaunchPlanModel(launchPlan, workflowModel.ID, launchPlanDigest, admin.LaunchPlanState_INACTIVE)
	if err != nil {
		logger.Errorf(ctx,
			"Failed to transform launch plan model [%+v], and workflow outputs [%+v] with err: %v",
			request, workflowInterface.Outputs, err)
		return nil, err
	}
	err = m.db.LaunchPlanRepo().Create(ctx, launchPlanModel)
	if err != nil {
		logger.Errorf(ctx, "Failed to save launch plan model %+v with err: %v", request.Id, err)
		return nil, err
	}
	m.metrics.SpecSizeBytes.Observe(float64(len(launchPlanModel.Spec)))
	m.metrics.ClosureSizeBytes.Observe(float64(len(launchPlanModel.Closure)))

	// TODO: Artifact feature gate, remove when ready
	if m.config.ApplicationConfiguration().GetTopLevelConfig().FeatureGates.EnableArtifacts {
		// As an optimization, run through the interface, and only send to artifacts service it there are artifacts
		for _, param := range launchPlan.GetSpec().GetDefaultInputs().GetParameters() {
			if param.GetArtifactId() != nil || param.GetArtifactQuery() != nil {
				err = m.artifactRegistry.RegisterArtifactConsumer(ctx, launchPlan.Id, launchPlan.Spec.DefaultInputs)
				if err != nil {
					logger.Errorf(ctx, "failed RegisterArtifactConsumer for launch plan [%+v] with err: %v", launchPlan.Id, err)
					return nil, err
				}
				break
			}
		}
	}

	return &admin.LaunchPlanCreateResponse{}, nil
}

func (m *LaunchPlanManager) updateLaunchPlanModelState(launchPlan *models.LaunchPlan, state admin.LaunchPlanState) error {
	var launchPlanClosure admin.LaunchPlanClosure
	err := proto.Unmarshal(launchPlan.Closure, &launchPlanClosure)
	if err != nil {
		logger.Errorf(context.Background(), "failed to unmarshal launch plan closure: %v", err)
		return errors.NewFlyteAdminErrorf(codes.Internal, "Failed to unmarshal launch plan closure: %v", err)
	}
	// Don't write the state in the closure - we store it only in the model column "State" and fill in the closure
	// value when transforming from a model to an admin.LaunchPlan object
	marshalledClosure, err := proto.Marshal(&launchPlanClosure)
	if err != nil {
		logger.Errorf(context.Background(), "Failed to marshal launch plan closure: %v", err)
		return errors.NewFlyteAdminErrorf(codes.Internal, "Failed to marshal launch plan closure: %v", err)
	}
	launchPlan.Closure = marshalledClosure
	stateInt := int32(state)
	launchPlan.State = &stateInt
	return nil
}

func isScheduleEmpty(launchPlanSpec admin.LaunchPlanSpec) bool {
	schedule := launchPlanSpec.GetEntityMetadata().GetSchedule()
	if schedule == nil {
		return true
	}
	if schedule.GetCronSchedule() != nil && len(schedule.GetCronSchedule().Schedule) != 0 {
		return false
	}
	if len(schedule.GetCronExpression()) != 0 {
		return false
	}
	if schedule.GetRate().GetValue() != 0 {
		return false
	}
	return true
}

func (m *LaunchPlanManager) enableSchedule(ctx context.Context, launchPlanIdentifier core.Identifier,
	launchPlanSpec admin.LaunchPlanSpec) error {
	addScheduleInput, err := m.scheduler.CreateScheduleInput(ctx,
		m.config.ApplicationConfiguration().GetSchedulerConfig(), launchPlanIdentifier,
		launchPlanSpec.EntityMetadata.Schedule)
	if err != nil {
		return err
	}

	return m.scheduler.AddSchedule(ctx, addScheduleInput)
}

func (m *LaunchPlanManager) disableSchedule(
	ctx context.Context, launchPlanIdentifier core.Identifier) error {
	return m.scheduler.RemoveSchedule(ctx, scheduleInterfaces.RemoveScheduleInput{
		Identifier:         launchPlanIdentifier,
		ScheduleNamePrefix: m.config.ApplicationConfiguration().GetSchedulerConfig().EventSchedulerConfig.ScheduleNamePrefix,
	})
}

func (m *LaunchPlanManager) updateSchedules(
	ctx context.Context, newlyActiveLaunchPlan models.LaunchPlan, formerlyActiveLaunchPlan *models.LaunchPlan) error {
	var newlyActiveLaunchPlanSpec admin.LaunchPlanSpec
	err := proto.Unmarshal(newlyActiveLaunchPlan.Spec, &newlyActiveLaunchPlanSpec)
	if err != nil {
		logger.Errorf(ctx, "failed to unmarshal newly enabled launch plan spec")
		return errors.NewFlyteAdminErrorf(codes.Internal, "failed to unmarshal newly enabled launch plan spec")
	}
	launchPlanIdentifier := core.Identifier{
		Org:     newlyActiveLaunchPlan.Org,
		Project: newlyActiveLaunchPlan.Project,
		Domain:  newlyActiveLaunchPlan.Domain,
		Name:    newlyActiveLaunchPlan.Name,
		Version: newlyActiveLaunchPlan.Version,
	}
	var formerlyActiveLaunchPlanSpec admin.LaunchPlanSpec
	if formerlyActiveLaunchPlan != nil {
		err = proto.Unmarshal(formerlyActiveLaunchPlan.Spec, &formerlyActiveLaunchPlanSpec)
		if err != nil {
			return errors.NewFlyteAdminErrorf(codes.Internal, "failed to unmarshal formerly enabled launch plan spec")
		}
	}

	if !isScheduleEmpty(formerlyActiveLaunchPlanSpec) {
		// Disable previous schedule
		formerlyActiveLaunchPlanIdentifier := core.Identifier{
			Org:     formerlyActiveLaunchPlan.Org,
			Project: formerlyActiveLaunchPlan.Project,
			Domain:  formerlyActiveLaunchPlan.Domain,
			Name:    formerlyActiveLaunchPlan.Name,
			Version: formerlyActiveLaunchPlan.Version,
		}
		if err = m.disableSchedule(ctx, formerlyActiveLaunchPlanIdentifier); err != nil {
			return err
		}
		logger.Infof(ctx, "Disabled schedules for deactivated launch plan [%+v]", launchPlanIdentifier)
	}
	if !isScheduleEmpty(newlyActiveLaunchPlanSpec) {
		// Enable new schedule
		if err = m.enableSchedule(ctx, launchPlanIdentifier, newlyActiveLaunchPlanSpec); err != nil {
			return err
		}
		logger.Infof(ctx, "Enabled schedules for activated launch plan [%+v]", launchPlanIdentifier)
	}
	return nil
}

func (m *LaunchPlanManager) disableLaunchPlan(ctx context.Context, request admin.LaunchPlanUpdateRequest) (
	*admin.LaunchPlanUpdateResponse, error) {
	if err := validation.ValidateIdentifier(request.Id, common.LaunchPlan); err != nil {
		logger.Debugf(ctx, "can't disable launch plan [%+v] with invalid identifier: %v", request.Id, err)
		return nil, err
	}
	launchPlanModel, err := util.GetLaunchPlanModel(ctx, m.db, *request.Id)
	if err != nil {
		logger.Debugf(ctx, "couldn't find launch plan [%+v] to disable with err: %v", request.Id, err)
		return nil, err
	}

	err = m.updateLaunchPlanModelState(&launchPlanModel, admin.LaunchPlanState_INACTIVE)
	if err != nil {
		logger.Debugf(ctx, "failed to disable launch plan [%+v] with err: %v", request.Id, err)
		return nil, err
	}

	var launchPlanSpec admin.LaunchPlanSpec
	err = proto.Unmarshal(launchPlanModel.Spec, &launchPlanSpec)
	if err != nil {
		logger.Errorf(ctx, "failed to unmarshal launch plan spec when disabling schedule for %+v", request.Id)
		return nil, errors.NewFlyteAdminErrorf(codes.Internal,
			"failed to unmarshal launch plan spec when disabling schedule for %+v", request.Id)
	}
	if launchPlanSpec.EntityMetadata != nil && launchPlanSpec.EntityMetadata.Schedule != nil {
		err = m.disableSchedule(ctx, core.Identifier{
			Org:     launchPlanModel.Org,
			Project: launchPlanModel.Project,
			Domain:  launchPlanModel.Domain,
			Name:    launchPlanModel.Name,
			Version: launchPlanModel.Version,
		})
		if err != nil {
			return nil, err
		}
	}

	// Send off activeness to trigger engine here, like we do for schedules above
	if m.config.ApplicationConfiguration().GetTopLevelConfig().FeatureGates.EnableArtifacts {
		if launchPlanModel.LaunchConditionType != nil && *launchPlanModel.LaunchConditionType == models.LaunchConditionTypeARTIFACT {
			err = m.artifactRegistry.DeactivateTrigger(ctx, request.Id)
			if err != nil {
				logger.Debugf(ctx, "failed to deactivate trigger for launch plan [%+v] with err: %v", request.Id, err)
				m.metrics.FailedTriggerUpdates.Inc()
				return nil, err
			}
		}
	}

	err = m.db.LaunchPlanRepo().Update(ctx, launchPlanModel)
	if err != nil {
		logger.Debugf(ctx, "Failed to update launchPlanModel with ID [%+v] with err %v", request.Id, err)
		return nil, err
	}
	logger.Debugf(ctx, "disabled launch plan: [%+v]", request.Id)
	return &admin.LaunchPlanUpdateResponse{}, nil
}

func (m *LaunchPlanManager) updateTriggers(ctx context.Context, newlyActiveLaunchPlan models.LaunchPlan, formerlyActiveLaunchPlan *models.LaunchPlan) error {
	var err error

	if newlyActiveLaunchPlan.LaunchConditionType != nil && *newlyActiveLaunchPlan.LaunchConditionType == models.LaunchConditionTypeARTIFACT {
		newID := &core.Identifier{
			Org:     newlyActiveLaunchPlan.Org,
			Project: newlyActiveLaunchPlan.Project,
			Domain:  newlyActiveLaunchPlan.Domain,
			Name:    newlyActiveLaunchPlan.Name,
			Version: newlyActiveLaunchPlan.Version,
		}
		err = m.artifactRegistry.ActivateTrigger(ctx, newID)
		if err != nil {
			logger.Infof(ctx, "failed to activate launch plan trigger [%+v] err: %v", newID, err)
			return err
		}
	}

	// Also update the old one, similar to schedule
	if formerlyActiveLaunchPlan != nil && formerlyActiveLaunchPlan.LaunchConditionType != nil && *formerlyActiveLaunchPlan.LaunchConditionType == models.LaunchConditionTypeARTIFACT {
		formerID := core.Identifier{
			Org:     formerlyActiveLaunchPlan.Org,
			Project: formerlyActiveLaunchPlan.Project,
			Domain:  formerlyActiveLaunchPlan.Domain,
			Name:    formerlyActiveLaunchPlan.Name,
			Version: formerlyActiveLaunchPlan.Version,
		}
		err = m.artifactRegistry.DeactivateTrigger(ctx, &formerID)
		if err != nil {
			logger.Infof(ctx, "failed to deactivate launch plan trigger [%+v] err: %v", formerlyActiveLaunchPlan.LaunchPlanKey, err)
			return err
		}
	}

	return nil
}

func (m *LaunchPlanManager) enableLaunchPlan(ctx context.Context, request admin.LaunchPlanUpdateRequest) (
	*admin.LaunchPlanUpdateResponse, error) {

	getUserPropertiesFunc := plugins.Get[shared.GetUserProperties](m.pluginRegistry, plugins.PluginIDUserProperties)
	userProperties := getUserPropertiesFunc(ctx)
	if userProperties.ActiveLaunchPlans > 0 {
		logger.Debugf(ctx, "user org '%s' is capped at '%d' active launch plans, verifying if [%+v] can be activated",
			userProperties.Org, userProperties.ActiveLaunchPlans, request.Id)
		// enable check to verify if the user will exceed their allowed active launch plans
		canActivateLaunchPlans, err := util.CanUserOrgActivateLaunchPlans(ctx, userProperties.Org, userProperties.ActiveLaunchPlans, m.db)
		if err != nil {
			logger.Warningf(ctx, "failed to get active launch plan count for org: %s with err: %+v", userProperties.Org, err)
			return nil, err
		}
		if !canActivateLaunchPlans {
			logger.Debugf(ctx, "org '%s' plan is only allowed '%d' active launch plans, not allowing further activation for [%+v]",
				userProperties.Org, userProperties.ActiveLaunchPlans, request.Id)
			return nil, errors.NewFlyteAdminErrorf(
				codes.ResourceExhausted,
				"org '%s' plan is only allowed '%d' active launch plans. Please disable a currently active launch plan and try again",
				userProperties.Org, userProperties.ActiveLaunchPlans)
		}
	}

	newlyActiveLaunchPlanModel, err := m.db.LaunchPlanRepo().Get(ctx, repoInterfaces.Identifier{
		Org:     request.Id.Org,
		Project: request.Id.Project,
		Domain:  request.Id.Domain,
		Name:    request.Id.Name,
		Version: request.Id.Version,
	})
	if err != nil {
		logger.Debugf(ctx, "Failed to find launch plan to enable with id [%+v] and err %v", request.Id, err)
		return nil, err
	}
	// Set desired launch plan version to active:
	err = m.updateLaunchPlanModelState(&newlyActiveLaunchPlanModel, admin.LaunchPlanState_ACTIVE)
	if err != nil {
		return nil, err
	}

	// Find currently active version, if it exists.
	filters, err := util.GetActiveLaunchPlanVersionFilters(newlyActiveLaunchPlanModel.Org, newlyActiveLaunchPlanModel.Project, newlyActiveLaunchPlanModel.Domain, newlyActiveLaunchPlanModel.Name)
	if err != nil {
		return nil, err
	}
	formerlyActiveLaunchPlanModelOutput, err := m.db.LaunchPlanRepo().List(ctx, repoInterfaces.ListResourceInput{
		InlineFilters: filters,
		Limit:         1,
	})
	var formerlyActiveLaunchPlanModel *models.LaunchPlan
	if err != nil {
		// Not found is fine, there isn't always a guaranteed active launch plan model.
		if err.(errors.FlyteAdminError).Code() != codes.NotFound {
			logger.Infof(ctx, "Failed to search for an active launch plan model with org: %s, project: %s, domain: %s, name: %s and err %v",
				request.Id.Org, request.Id.Project, request.Id.Domain, request.Id.Name, err)
			return nil, err
		}
		logger.Debugf(ctx, "No active launch plan model found to disable with org: %s, project: %s, domain: %s, name: %s",
			request.Id.Org, request.Id.Project, request.Id.Domain, request.Id.Name)
	} else if formerlyActiveLaunchPlanModelOutput.LaunchPlans != nil &&
		len(formerlyActiveLaunchPlanModelOutput.LaunchPlans) > 0 {
		formerlyActiveLaunchPlanModel = &formerlyActiveLaunchPlanModelOutput.LaunchPlans[0]
		err = m.updateLaunchPlanModelState(formerlyActiveLaunchPlanModel, admin.LaunchPlanState_INACTIVE)
		if err != nil {
			return nil, err
		}
	}
	err = m.updateSchedules(ctx, newlyActiveLaunchPlanModel, formerlyActiveLaunchPlanModel)
	if err != nil {
		m.metrics.FailedScheduleUpdates.Inc()
		return nil, err
	}

	// Send off activeness to trigger engine here, like we do for schedules above, but only if the type is of trigger.
	if m.config.ApplicationConfiguration().GetTopLevelConfig().FeatureGates.EnableArtifacts {
		err = m.updateTriggers(ctx, newlyActiveLaunchPlanModel, formerlyActiveLaunchPlanModel)
		if err != nil {
			m.metrics.FailedTriggerUpdates.Inc()
			logger.Debugf(ctx, "failed to update trigger for launch plan [%+v] with err: %v", request.Id, err)
			return nil, err
		}
	}

	// This operation is takes in the (formerly) active launch plan version as only one version can be active at a time.
	// Setting the desired launch plan to active also requires disabling the existing active launch plan version.
	err = m.db.LaunchPlanRepo().SetActive(ctx, newlyActiveLaunchPlanModel, formerlyActiveLaunchPlanModel)
	if err != nil {
		logger.Debugf(ctx,
			"Failed to set launchPlanModel with ID [%+v] to active with err %v", request.Id, err)
		return nil, err
	}

	return &admin.LaunchPlanUpdateResponse{}, nil
}

func (m *LaunchPlanManager) UpdateLaunchPlan(ctx context.Context, request admin.LaunchPlanUpdateRequest) (
	*admin.LaunchPlanUpdateResponse, error) {
	if err := validation.ValidateIdentifier(request.Id, common.LaunchPlan); err != nil {
		logger.Debugf(ctx, "can't update launch plan [%+v] state, invalid identifier: %v", request.Id, err)
	}
	ctx = getLaunchPlanContext(ctx, request.Id)
	switch request.State {
	case admin.LaunchPlanState_INACTIVE:
		return m.disableLaunchPlan(ctx, request)
	case admin.LaunchPlanState_ACTIVE:
		return m.enableLaunchPlan(ctx, request)
	default:
		return nil, errors.NewFlyteAdminErrorf(
			codes.InvalidArgument, "Unrecognized launch plan state %v for update for launch plan [%+v]",
			request.State, request.Id)
	}
}

func (m *LaunchPlanManager) GetLaunchPlan(ctx context.Context, request admin.ObjectGetRequest) (
	*admin.LaunchPlan, error) {
	if err := validation.ValidateIdentifier(request.Id, common.LaunchPlan); err != nil {
		logger.Debugf(ctx, "can't get launch plan [%+v] with invalid identifier: %v", request.Id, err)
		return nil, err
	}
	ctx = getLaunchPlanContext(ctx, request.Id)
	return util.GetLaunchPlan(ctx, m.db, *request.Id)
}

func (m *LaunchPlanManager) GetActiveLaunchPlan(ctx context.Context, request admin.ActiveLaunchPlanRequest) (
	*admin.LaunchPlan, error) {
	if err := validation.ValidateActiveLaunchPlanRequest(request); err != nil {
		logger.Debugf(ctx, "can't get active launch plan [%+v] with invalid request: %v", request.Id, err)
		return nil, err
	}
	ctx = m.getNamedEntityContext(ctx, request.Id)

	filters, err := util.GetActiveLaunchPlanVersionFilters(request.Id.Org, request.Id.Project, request.Id.Domain, request.Id.Name)
	if err != nil {
		return nil, err
	}

	listLaunchPlansInput := repoInterfaces.ListResourceInput{
		Limit:         1,
		InlineFilters: filters,
	}

	output, err := m.db.LaunchPlanRepo().List(ctx, listLaunchPlansInput)

	if err != nil {
		logger.Debugf(ctx, "Failed to list active launch plan id for request [%+v] with err %v", request, err)
		return nil, err
	}

	if len(output.LaunchPlans) != 1 {
		return nil, errors.NewFlyteAdminErrorf(codes.NotFound, "No active launch plan could be found: %s:%s:%s:%s", request.Id.Org, request.Id.Project, request.Id.Domain, request.Id.Name)
	}

	return transformers.FromLaunchPlanModel(output.LaunchPlans[0])
}

func (m *LaunchPlanManager) ListLaunchPlans(ctx context.Context, request admin.ResourceListRequest) (
	*admin.LaunchPlanList, error) {

	// Check required fields
	if err := validation.ValidateResourceListRequest(request); err != nil {
		logger.Debugf(ctx, "")
		return nil, err
	}
	ctx = m.getNamedEntityContext(ctx, request.Id)

	filters, err := util.GetDbFilters(util.FilterSpec{
		Org:            request.Id.Org,
		Project:        request.Id.Project,
		Domain:         request.Id.Domain,
		Name:           request.Id.Name,
		RequestFilters: request.Filters,
	}, common.LaunchPlan)
	if err != nil {
		return nil, err
	}

	sortParameter, err := common.NewSortParameter(request.SortBy, models.LaunchPlanColumns)
	if err != nil {
		return nil, err
	}

	offset, err := validation.ValidateToken(request.Token)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"invalid pagination token %s for ListLaunchPlans", request.Token)
	}
	listLaunchPlansInput := repoInterfaces.ListResourceInput{
		Limit:         int(request.Limit),
		Offset:        offset,
		InlineFilters: filters,
		SortParameter: sortParameter,
	}

	output, err := m.db.LaunchPlanRepo().List(ctx, listLaunchPlansInput)
	if err != nil {
		logger.Debugf(ctx, "Failed to list launch plans for request [%+v] with err %v", request, err)
		return nil, err
	}
	launchPlanList, err := transformers.FromLaunchPlanModels(output.LaunchPlans)
	if err != nil {
		logger.Errorf(ctx,
			"Failed to transform launch plan models [%+v] with err: %v", output.LaunchPlans, err)
		return nil, err
	}
	var token string
	if len(output.LaunchPlans) == int(request.Limit) {
		token = strconv.Itoa(offset + len(output.LaunchPlans))
	}
	return &admin.LaunchPlanList{
		LaunchPlans: launchPlanList,
		Token:       token,
	}, nil
}

func (m *LaunchPlanManager) ListActiveLaunchPlans(ctx context.Context, request admin.ActiveLaunchPlanListRequest) (
	*admin.LaunchPlanList, error) {

	// Check required fields
	if err := validation.ValidateActiveLaunchPlanListRequest(request); err != nil {
		logger.Debugf(ctx, "")
		return nil, err
	}
	ctx = contextutils.WithProjectDomain(ctx, request.Project, request.Domain)

	filters, err := util.ListActiveLaunchPlanVersionsFilters(request.Org, request.Project, request.Domain)
	if err != nil {
		return nil, err
	}

	sortParameter, err := common.NewSortParameter(request.SortBy, models.LaunchPlanColumns)
	if err != nil {
		return nil, err
	}

	offset, err := validation.ValidateToken(request.Token)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"invalid pagination token %s for ListActiveLaunchPlans", request.Token)
	}
	listLaunchPlansInput := repoInterfaces.ListResourceInput{
		Limit:         int(request.Limit),
		Offset:        offset,
		InlineFilters: filters,
		SortParameter: sortParameter,
	}

	output, err := m.db.LaunchPlanRepo().List(ctx, listLaunchPlansInput)
	if err != nil {
		logger.Debugf(ctx, "Failed to list active launch plans for request [%+v] with err %v", request, err)
		return nil, err
	}
	launchPlanList, err := transformers.FromLaunchPlanModels(output.LaunchPlans)
	if err != nil {
		logger.Errorf(ctx,
			"Failed to transform active launch plan models [%+v] with err: %v", output.LaunchPlans, err)
		return nil, err
	}
	var token string
	if len(output.LaunchPlans) == int(request.Limit) {
		token = strconv.Itoa(offset + len(output.LaunchPlans))
	}
	return &admin.LaunchPlanList{
		LaunchPlans: launchPlanList,
		Token:       token,
	}, nil
}

// At least project name and domain must be specified along with limit.
func (m *LaunchPlanManager) ListLaunchPlanIds(ctx context.Context, request admin.NamedEntityIdentifierListRequest) (
	*admin.NamedEntityIdentifierList, error) {
	ctx = contextutils.WithProjectDomain(ctx, request.Project, request.Domain)
	filters, err := util.GetDbFilters(util.FilterSpec{
		Org:     request.Org,
		Project: request.Project,
		Domain:  request.Domain,
	}, common.LaunchPlan)
	if err != nil {
		return nil, err
	}

	sortParameter, err := common.NewSortParameter(request.SortBy, models.LaunchPlanColumns)
	if err != nil {
		return nil, err
	}

	offset, err := validation.ValidateToken(request.Token)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "invalid pagination token %s", request.Token)
	}
	listLaunchPlansInput := repoInterfaces.ListResourceInput{
		Limit:         int(request.Limit),
		Offset:        offset,
		InlineFilters: filters,
		SortParameter: sortParameter,
	}

	output, err := m.db.LaunchPlanRepo().ListLaunchPlanIdentifiers(ctx, listLaunchPlansInput)

	if err != nil {
		logger.Debugf(ctx, "Failed to list launch plan ids for request [%+v] with err %v", request, err)
		return nil, err
	}
	var token string
	if len(output.LaunchPlans) == int(request.Limit) {
		token = strconv.Itoa(offset + len(output.LaunchPlans))
	}
	return &admin.NamedEntityIdentifierList{
		Entities: transformers.FromLaunchPlanModelsToIdentifiers(output.LaunchPlans),
		Token:    token,
	}, nil
}

func (m *LaunchPlanManager) CreateLaunchPlanFromNode(
	ctx context.Context, request admin.CreateLaunchPlanFromNodeRequest) (*admin.CreateLaunchPlanFromNodeResponse, error) {

	err := validation.ValidateCreateLaunchPlanFromNodeRequest(request)
	if err != nil {
		return nil, err
	}

	securityContext := request.GetSecurityContext()

	var subNode *core.Node
	if request.GetSubNodeIds() != nil {
		subNodeID := request.GetSubNodeIds().GetSubNodeIds()[0].GetSubNodeId()
		originalLaunchPlan, err := util.GetLaunchPlan(context.Background(), m.db, *request.GetLaunchPlanId())
		if err != nil {
			return nil, err
		}

		originalWorkflow, err := util.GetWorkflow(ctx, m.db, m.storageClient, *originalLaunchPlan.GetSpec().GetWorkflowId())
		if err != nil {
			return nil, err
		}
		subNode, err = util.GetSubNodeFromWorkflow(originalWorkflow, subNodeID)
		if err != nil {
			return nil, err
		}

		if securityContext == nil {
			securityContext = originalLaunchPlan.GetSpec().GetSecurityContext()
		}
	} else {
		subNode = request.GetSubNodeSpec()
	}

	workflowModel, err := util.CreateOrGetWorkflowFromNode(ctx, subNode, m.db, m.workflowManager, m.namedEntityManager, request.GetLaunchPlanId())
	if err != nil {
		return nil, err
	}
	workflow, err := transformers.FromWorkflowModel(*workflowModel)
	if err != nil {
		return nil, err
	}

	launchPlan, err := util.CreateOrGetLaunchPlan(ctx, m.db, m.config, workflow.Id,
		workflow.GetClosure().GetCompiledWorkflow().GetPrimary().GetTemplate().GetInterface(), workflowModel.ID,
		nil, securityContext, subNode)
	if err != nil {
		return nil, err
	}

	return &admin.CreateLaunchPlanFromNodeResponse{
		LaunchPlan: launchPlan,
	}, nil
}

func NewLaunchPlanManager(
	db repoInterfaces.Repository,
	config runtimeInterfaces.Configuration,
	scheduler scheduleInterfaces.EventScheduler,
	scope promutils.Scope,
	artifactRegistry *artifacts.ArtifactRegistry,
	pluginRegistry *plugins.Registry,
	storageClient *storage.DataStore,
	workflowManager interfaces.WorkflowInterface,
	namedEntityManager interfaces.NamedEntityInterface) interfaces.LaunchPlanInterface {

	metrics := launchPlanMetrics{
		Scope: scope,
		FailedScheduleUpdates: scope.MustNewCounter("failed_schedule_updates",
			"count of unsuccessful attempts to update the schedules when updating launch plan version"),
		FailedTriggerUpdates: scope.MustNewCounter("failed_trigger_updates", "count of failed attempts to activate/deactivate triggers"),
		SpecSizeBytes:        scope.MustNewSummary("spec_size_bytes", "size in bytes of serialized launch plan spec"),
		ClosureSizeBytes:     scope.MustNewSummary("closure_size_bytes", "size in bytes of serialized launch plan closure"),
	}
	return &LaunchPlanManager{
		db:                 db,
		config:             config,
		scheduler:          scheduler,
		metrics:            metrics,
		artifactRegistry:   artifactRegistry,
		pluginRegistry:     pluginRegistry,
		storageClient:      storageClient,
		workflowManager:    workflowManager,
		namedEntityManager: namedEntityManager,
	}
}
