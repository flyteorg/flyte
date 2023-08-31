package impl

import (
	"bytes"
	"context"
	"strconv"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	"github.com/golang/protobuf/ptypes"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/resources"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/util"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/validation"
	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	repoInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteadmin/pkg/repositories/transformers"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	workflowengine "github.com/flyteorg/flyteadmin/pkg/workflowengine/interfaces"
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
	ctx = contextutils.WithProjectDomain(ctx, identifier.Project, identifier.Domain)
	return contextutils.WithTaskID(ctx, identifier.Name)
}

func setDefaults(request admin.TaskCreateRequest) (admin.TaskCreateRequest, error) {
	if request.Id == nil {
		return request, errors.NewFlyteAdminError(codes.InvalidArgument,
			"missing identifier for TaskCreateRequest")
	}

	request.Spec.Template.Id = request.Id
	return request, nil
}

func (t *TaskManager) CreateTask(
	ctx context.Context,
	request admin.TaskCreateRequest) (*admin.TaskCreateResponse, error) {
	platformTaskResources := util.GetTaskResources(ctx, request.Id, t.resourceManager, t.config.TaskResourceConfiguration())
	if err := validation.ValidateTask(ctx, request, t.db, platformTaskResources,
		t.config.WhitelistConfiguration(), t.config.ApplicationConfiguration()); err != nil {
		logger.Debugf(ctx, "Task [%+v] failed validation with err: %v", request.Id, err)
		return nil, err
	}
	ctx = getTaskContext(ctx, request.Id)
	finalizedRequest, err := setDefaults(request)
	if err != nil {
		return nil, err
	}
	// Compile task and store the compiled version in the database.
	compiledTask, err := t.compiler.CompileTask(finalizedRequest.Spec.Template)
	if err != nil {
		logger.Debugf(ctx, "Failed to compile task with id [%+v] with err %v", request.Id, err)
		return nil, err
	}
	createdAt, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal,
			"Failed to serialize CreatedAt: %v when creating task: %+v", err, request.Id)
	}
	taskDigest, err := util.GetTaskDigest(ctx, compiledTask)
	if err != nil {
		logger.Errorf(ctx, "failed to compute task digest with err %v", err)
		return nil, err
	}
	// See if a task exists and confirm whether it's an identical task or one that with a separate definition.
	existingTask, err := util.GetTaskModel(ctx, t.db, request.Spec.Template.Id)
	if err == nil {
		if bytes.Equal(taskDigest, existingTask.Digest) {
			return nil, errors.NewFlyteAdminErrorf(codes.AlreadyExists,
				"identical task already exists with id %s", request.Id)
		}
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"task with different structure already exists with id %v", request.Id)
	}
	taskModel, err := transformers.CreateTaskModel(finalizedRequest, admin.TaskClosure{
		CompiledTask: compiledTask,
		CreatedAt:    createdAt,
	}, taskDigest)
	if err != nil {
		logger.Errorf(ctx,
			"Failed to transform task model [%+v] with err: %v", finalizedRequest, err)
		return nil, err
	}

	descriptionModel, err := transformers.CreateDescriptionEntityModel(request.Spec.Description, *request.Id)
	if err != nil {
		logger.Errorf(ctx,
			"Failed to transform description model [%+v] with err: %v", request.Spec.Description, err)
		return nil, err
	}
	if descriptionModel != nil {
		taskModel.ShortDescription = descriptionModel.ShortDescription
	}
	err = t.db.TaskRepo().Create(ctx, taskModel, descriptionModel)
	if err != nil {
		logger.Debugf(ctx, "Failed to create task model with id [%+v] with err %v", request.Id, err)
		return nil, err
	}
	t.metrics.ClosureSizeBytes.Observe(float64(len(taskModel.Closure)))
	if finalizedRequest.Spec.Template.Metadata != nil {
		contextWithRuntimeMeta := context.WithValue(
			ctx, common.RuntimeTypeKey, finalizedRequest.Spec.Template.Metadata.Runtime.Type.String())
		contextWithRuntimeMeta = context.WithValue(
			contextWithRuntimeMeta, common.RuntimeVersionKey, finalizedRequest.Spec.Template.Metadata.Runtime.Version)
		t.metrics.Registered.Inc(contextWithRuntimeMeta)
	}

	return &admin.TaskCreateResponse{}, nil
}

func (t *TaskManager) GetTask(ctx context.Context, request admin.ObjectGetRequest) (*admin.Task, error) {
	if err := validation.ValidateIdentifier(request.Id, common.Task); err != nil {
		logger.Debugf(ctx, "invalid identifier [%+v]: %v", request.Id, err)
	}
	ctx = getTaskContext(ctx, request.Id)
	task, err := util.GetTask(ctx, t.db, *request.Id)
	if err != nil {
		logger.Debugf(ctx, "Failed to get task with id [%+v] with err %v", err, request.Id)
		return nil, err
	}
	return task, nil
}

func (t *TaskManager) ListTasks(ctx context.Context, request admin.ResourceListRequest) (*admin.TaskList, error) {
	// Check required fields
	if err := validation.ValidateResourceListRequest(request); err != nil {
		logger.Debugf(ctx, "Invalid request [%+v]: %v", request, err)
		return nil, err
	}
	ctx = contextutils.WithProjectDomain(ctx, request.Id.Project, request.Id.Domain)
	ctx = contextutils.WithTaskID(ctx, request.Id.Name)
	spec := util.FilterSpec{
		Project:        request.Id.Project,
		Domain:         request.Id.Domain,
		Name:           request.Id.Name,
		RequestFilters: request.Filters,
	}

	filters, err := util.GetDbFilters(spec, common.Task)
	if err != nil {
		return nil, err
	}

	sortParameter, err := common.NewSortParameter(request.SortBy, models.TaskColumns)
	if err != nil {
		return nil, err
	}

	offset, err := validation.ValidateToken(request.Token)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"invalid pagination token %s for ListTasks", request.Token)
	}
	// And finally, query the database
	listTasksInput := repoInterfaces.ListResourceInput{
		Limit:         int(request.Limit),
		Offset:        offset,
		InlineFilters: filters,
		SortParameter: sortParameter,
	}
	output, err := t.db.TaskRepo().List(ctx, listTasksInput)
	if err != nil {
		logger.Debugf(ctx, "Failed to list tasks with id [%+v] with err %v", request.Id, err)
		return nil, err
	}
	taskList, err := transformers.FromTaskModels(output.Tasks)
	if err != nil {
		logger.Errorf(ctx,
			"Failed to transform task models [%+v] with err: %v", output.Tasks, err)
		return nil, err
	}

	var token string
	if len(taskList) == int(request.Limit) {
		token = strconv.Itoa(offset + len(taskList))
	}
	return &admin.TaskList{
		Tasks: taskList,
		Token: token,
	}, nil
}

// This queries the unique tasks for the given query parameters.  At least the project and domain must be specified.
// It will return all tasks, but only the one of each even if there are multiple versions.
func (t *TaskManager) ListUniqueTaskIdentifiers(ctx context.Context, request admin.NamedEntityIdentifierListRequest) (
	*admin.NamedEntityIdentifierList, error) {
	if err := validation.ValidateNamedEntityIdentifierListRequest(request); err != nil {
		logger.Debugf(ctx, "invalid request [%+v]: %v", request, err)
		return nil, err
	}
	ctx = contextutils.WithProjectDomain(ctx, request.Project, request.Domain)
	filters, err := util.GetDbFilters(util.FilterSpec{
		Project: request.Project,
		Domain:  request.Domain,
	}, common.Task)
	if err != nil {
		return nil, err
	}

	sortParameter, err := common.NewSortParameter(request.SortBy, models.TaskColumns)
	if err != nil {
		return nil, err
	}

	offset, err := validation.ValidateToken(request.Token)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"invalid pagination token %s for ListUniqueTaskIdentifiers", request.Token)
	}
	listTasksInput := repoInterfaces.ListResourceInput{
		Limit:         int(request.Limit),
		Offset:        offset,
		InlineFilters: filters,
		SortParameter: sortParameter,
	}

	output, err := t.db.TaskRepo().ListTaskIdentifiers(ctx, listTasksInput)
	if err != nil {
		logger.Debugf(ctx, "Failed to list tasks ids with project: %s and domain: %s with err %v",
			request.Project, request.Domain, err)
		return nil, err
	}

	idList := transformers.FromTaskModelsToIdentifiers(output.Tasks)
	var token string
	if len(idList) == int(request.Limit) {
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
