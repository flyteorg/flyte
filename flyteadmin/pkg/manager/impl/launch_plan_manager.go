package impl

import (
	"bytes"
	"context"
	"strconv"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/golang/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"

	scheduleInterfaces "github.com/flyteorg/flyteadmin/pkg/async/schedule/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/util"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/validation"
	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	repoInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteadmin/pkg/repositories/transformers"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
)

type launchPlanMetrics struct {
	Scope                 promutils.Scope
	FailedScheduleUpdates prometheus.Counter
	SpecSizeBytes         prometheus.Summary
	ClosureSizeBytes      prometheus.Summary
}

type LaunchPlanManager struct {
	db        repoInterfaces.Repository
	config    runtimeInterfaces.Configuration
	scheduler scheduleInterfaces.EventScheduler
	metrics   launchPlanMetrics
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

	existingLaunchPlanModel, err := util.GetLaunchPlanModel(ctx, m.db, *request.Id)
	if err == nil {
		if bytes.Equal(existingLaunchPlanModel.Digest, launchPlanDigest) {
			return nil, errors.NewFlyteAdminErrorf(codes.AlreadyExists,
				"identical launch plan already exists with id %s", request.Id)
		}

		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"launch plan with different structure already exists with id %v", request.Id)
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
			Project: launchPlanModel.Project,
			Domain:  launchPlanModel.Domain,
			Name:    launchPlanModel.Name,
			Version: launchPlanModel.Version,
		})
		if err != nil {
			return nil, err
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

func (m *LaunchPlanManager) enableLaunchPlan(ctx context.Context, request admin.LaunchPlanUpdateRequest) (
	*admin.LaunchPlanUpdateResponse, error) {
	newlyActiveLaunchPlanModel, err := m.db.LaunchPlanRepo().Get(ctx, repoInterfaces.Identifier{
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
	filters, err := util.GetActiveLaunchPlanVersionFilters(newlyActiveLaunchPlanModel.Project, newlyActiveLaunchPlanModel.Domain, newlyActiveLaunchPlanModel.Name)
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
			logger.Infof(ctx, "Failed to search for an active launch plan model with project: %s, domain: %s, name: %s and err %v",
				request.Id.Project, request.Id.Domain, request.Id.Name, err)
			return nil, err
		}
		logger.Debugf(ctx, "No active launch plan model found to disable with project: %s, domain: %s, name: %s",
			request.Id.Project, request.Id.Domain, request.Id.Name)
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

	filters, err := util.GetActiveLaunchPlanVersionFilters(request.Id.Project, request.Id.Domain, request.Id.Name)
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
		return nil, errors.NewFlyteAdminErrorf(codes.NotFound, "No active launch plan could be found: %s:%s:%s", request.Id.Project, request.Id.Domain, request.Id.Name)
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

	filters, err := util.ListActiveLaunchPlanVersionsFilters(request.Project, request.Domain)
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

func NewLaunchPlanManager(
	db repoInterfaces.Repository,
	config runtimeInterfaces.Configuration,
	scheduler scheduleInterfaces.EventScheduler,
	scope promutils.Scope) interfaces.LaunchPlanInterface {

	metrics := launchPlanMetrics{
		Scope: scope,
		FailedScheduleUpdates: scope.MustNewCounter("failed_schedule_updates",
			"count of unsuccessful attempts to update the schedules when updating launch plan version"),
		SpecSizeBytes:    scope.MustNewSummary("spec_size_bytes", "size in bytes of serialized launch plan spec"),
		ClosureSizeBytes: scope.MustNewSummary("closure_size_bytes", "size in bytes of serialized launch plan closure"),
	}
	return &LaunchPlanManager{
		db:        db,
		config:    config,
		scheduler: scheduler,
		metrics:   metrics,
	}
}
