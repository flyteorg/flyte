package impl

import (
	"bytes"
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strconv"

	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/resources"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/util"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/validation"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/interfaces"
	repoInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/transformers"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	workflowengine "github.com/flyteorg/flyte/flyteadmin/pkg/workflowengine/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
)

type taskMetrics struct {
	Scope            promutils.Scope
	ClosureSizeBytes prometheus.Summary
	Registered       labeled.Counter
}

type TaskManager struct {
	db              repoInterfaces.Repository
	config          runtimeInterfaces.Configuration
	compiler        workflowengine.Compiler
	metrics         taskMetrics
	resourceManager interfaces.ResourceInterface
}

func getTaskContext(ctx context.Context, identifier *core.Identifier) context.Context {
	ctx = contextutils.WithProjectDomain(ctx, identifier.GetProject(), identifier.GetDomain())
	return contextutils.WithTaskID(ctx, identifier.GetName())
}

func setDefaults(request *admin.TaskCreateRequest) (*admin.TaskCreateRequest, error) {
	if request.GetId() == nil {
		return request, errors.NewFlyteAdminError(codes.InvalidArgument,
			"missing identifier for TaskCreateRequest")
	}

	request.Spec.Template.Id = request.GetId()
	return request, nil
}

func (t *TaskManager) CreateTask(
	ctx context.Context,
	request *admin.TaskCreateRequest) (*admin.TaskCreateResponse, error) {
	platformTaskResources := util.GetTaskResources(ctx, request.GetId(), t.resourceManager, t.config.TaskResourceConfiguration())
	if err := validation.ValidateTask(ctx, request, t.db, platformTaskResources,
		t.config.WhitelistConfiguration(), t.config.ApplicationConfiguration()); err != nil {
		logger.Debugf(ctx, "Task [%+v] failed validation with err: %v", request.GetId(), err)
		return nil, err
	}
	ctx = getTaskContext(ctx, request.GetId())
	finalizedRequest, err := setDefaults(request)
	if err != nil {
		return nil, err
	}
	// Compile task and store the compiled version in the database.
	compiledTask, err := t.compiler.CompileTask(finalizedRequest.GetSpec().GetTemplate())
	if err != nil {
		logger.Debugf(ctx, "Failed to compile task with id [%+v] with err %v", request.GetId(), err)
		return nil, err
	}
	createdAt := timestamppb.Now()
	taskDigest, err := util.GetTaskDigest(ctx, compiledTask)
	if err != nil {
		logger.Errorf(ctx, "failed to compute task digest with err %v", err)
		return nil, err
	}
	// Create Task in DB
	taskModel, err := transformers.CreateTaskModel(finalizedRequest, &admin.TaskClosure{
		CompiledTask: compiledTask,
		CreatedAt:    createdAt,
	}, taskDigest)
	if err != nil {
		logger.Errorf(ctx,
			"Failed to transform task model [%+v] with err: %v", finalizedRequest, err)
		return nil, err
	}
	descriptionModel, err := transformers.CreateDescriptionEntityModel(request.GetSpec().GetDescription(), request.GetId())
	if err != nil {
		logger.Errorf(ctx,
			"Failed to transform description model [%+v] with err: %v", request.GetSpec().GetDescription(), err)
		return nil, err
	}
	if descriptionModel != nil {
		taskModel.ShortDescription = descriptionModel.ShortDescription
	}
	err = t.db.TaskRepo().Create(ctx, taskModel, descriptionModel)
	if err != nil {
		// See if an identical task already exists by checking the error code
		flyteErr, ok := err.(errors.FlyteAdminError)
		if !ok || flyteErr.Code() != codes.AlreadyExists {
			logger.Errorf(ctx, "Failed to create task model with id [%+v] with err %v", request.GetId(), err)
			return nil, err
		}
		// An identical task already exists. Fetch the existing task to verify if it has a different digest
		existingTaskModel, fetchErr := util.GetTaskModel(ctx, t.db, request.GetSpec().GetTemplate().GetId())
		if fetchErr != nil {
			logger.Errorf(ctx, "Failed to fetch existing task model for id [%+v] with err %v", request.GetId(), fetchErr)
			return nil, fetchErr
		}
		if bytes.Equal(taskDigest, existingTaskModel.Digest) {
			return nil, errors.NewTaskExistsIdenticalStructureError(ctx, request)
		}
		existingTask, transformerErr := transformers.FromTaskModel(*existingTaskModel)
		if transformerErr != nil {
			logger.Errorf(ctx, "Failed to transform task from task model for id [%+v]", request.GetId())
			return nil, transformerErr
		}
		return nil, errors.NewTaskExistsDifferentStructureError(ctx, request, existingTask.GetClosure().GetCompiledTask(), compiledTask)
	}
	t.metrics.ClosureSizeBytes.Observe(float64(len(taskModel.Closure)))
	if finalizedRequest.GetSpec().GetTemplate().GetMetadata() != nil {
		contextWithRuntimeMeta := context.WithValue(
			ctx, common.RuntimeTypeKey, finalizedRequest.GetSpec().GetTemplate().GetMetadata().GetRuntime().GetType().String())
		contextWithRuntimeMeta = context.WithValue(
			contextWithRuntimeMeta, common.RuntimeVersionKey, finalizedRequest.GetSpec().GetTemplate().GetMetadata().GetRuntime().GetVersion())
		t.metrics.Registered.Inc(contextWithRuntimeMeta)
	}

	return &admin.TaskCreateResponse{}, nil
}

func (t *TaskManager) GetTask(ctx context.Context, request *admin.ObjectGetRequest) (*admin.Task, error) {
	if err := validation.ValidateIdentifier(request.GetId(), common.Task); err != nil {
		logger.Debugf(ctx, "invalid identifier [%+v]: %v", request.GetId(), err)
	}
	ctx = getTaskContext(ctx, request.GetId())
	task, err := util.GetTask(ctx, t.db, request.GetId())
	if err != nil {
		logger.Debugf(ctx, "Failed to get task with id [%+v] with err %v", err, request.GetId())
		return nil, err
	}
	return task, nil
}

func (t *TaskManager) ListTasks(ctx context.Context, request *admin.ResourceListRequest) (*admin.TaskList, error) {
	// Check required fields
	if err := validation.ValidateResourceListRequest(request); err != nil {
		logger.Debugf(ctx, "Invalid request [%+v]: %v", request, err)
		return nil, err
	}
	ctx = contextutils.WithProjectDomain(ctx, request.GetId().GetProject(), request.GetId().GetDomain())
	ctx = contextutils.WithTaskID(ctx, request.GetId().GetName())
	spec := util.FilterSpec{
		Project:        request.GetId().GetProject(),
		Domain:         request.GetId().GetDomain(),
		Name:           request.GetId().GetName(),
		RequestFilters: request.GetFilters(),
	}

	filters, err := util.GetDbFilters(spec, common.Task)
	if err != nil {
		return nil, err
	}

	sortParameter, err := common.NewSortParameter(request.GetSortBy(), models.TaskColumns)
	if err != nil {
		return nil, err
	}

	offset, err := validation.ValidateToken(request.GetToken())
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"invalid pagination token %s for ListTasks", request.GetToken())
	}
	// And finally, query the database
	listTasksInput := repoInterfaces.ListResourceInput{
		Limit:         int(request.GetLimit()),
		Offset:        offset,
		InlineFilters: filters,
		SortParameter: sortParameter,
	}
	output, err := t.db.TaskRepo().List(ctx, listTasksInput)
	if err != nil {
		logger.Debugf(ctx, "Failed to list tasks with id [%+v] with err %v", request.GetId(), err)
		return nil, err
	}
	taskList, err := transformers.FromTaskModels(output.Tasks)
	if err != nil {
		logger.Errorf(ctx,
			"Failed to transform task models [%+v] with err: %v", output.Tasks, err)
		return nil, err
	}

	var token string
	if len(taskList) == int(request.GetLimit()) {
		token = strconv.Itoa(offset + len(taskList))
	}
	return &admin.TaskList{
		Tasks: taskList,
		Token: token,
	}, nil
}

// This queries the unique tasks for the given query parameters.  At least the project and domain must be specified.
// It will return all tasks, but only the one of each even if there are multiple versions.
func (t *TaskManager) ListUniqueTaskIdentifiers(ctx context.Context, request *admin.NamedEntityIdentifierListRequest) (
	*admin.NamedEntityIdentifierList, error) {
	if err := validation.ValidateNamedEntityIdentifierListRequest(request); err != nil {
		logger.Debugf(ctx, "invalid request [%+v]: %v", request, err)
		return nil, err
	}
	ctx = contextutils.WithProjectDomain(ctx, request.GetProject(), request.GetDomain())
	filters, err := util.GetDbFilters(util.FilterSpec{
		Project: request.GetProject(),
		Domain:  request.GetDomain(),
	}, common.Task)
	if err != nil {
		return nil, err
	}

	sortParameter, err := common.NewSortParameter(request.GetSortBy(), models.TaskColumns)
	if err != nil {
		return nil, err
	}

	offset, err := validation.ValidateToken(request.GetToken())
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"invalid pagination token %s for ListUniqueTaskIdentifiers", request.GetToken())
	}
	listTasksInput := repoInterfaces.ListResourceInput{
		Limit:         int(request.GetLimit()),
		Offset:        offset,
		InlineFilters: filters,
		SortParameter: sortParameter,
	}

	output, err := t.db.TaskRepo().ListTaskIdentifiers(ctx, listTasksInput)
	if err != nil {
		logger.Debugf(ctx, "Failed to list tasks ids with project: %s and domain: %s with err %v",
			request.GetProject(), request.GetDomain(), err)
		return nil, err
	}

	idList := transformers.FromTaskModelsToIdentifiers(output.Tasks)
	var token string
	if len(idList) == int(request.GetLimit()) {
		token = strconv.Itoa(offset + len(idList))
	}
	return &admin.NamedEntityIdentifierList{
		Entities: idList,
		Token:    token,
	}, nil
}

func NewTaskManager(
	db repoInterfaces.Repository,
	config runtimeInterfaces.Configuration, compiler workflowengine.Compiler,
	scope promutils.Scope) interfaces.TaskInterface {

	metrics := taskMetrics{
		Scope:            scope,
		ClosureSizeBytes: scope.MustNewSummary("closure_size_bytes", "size in bytes of serialized task closure"),
		Registered:       labeled.NewCounter("num_registered", "count of registered tasks", scope),
	}
	resourceManager := resources.NewResourceManager(db, config.ApplicationConfiguration())
	return &TaskManager{
		db:              db,
		config:          config,
		compiler:        compiler,
		metrics:         metrics,
		resourceManager: resourceManager,
	}
}
