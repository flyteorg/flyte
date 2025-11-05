package impl

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/auth"
	cloudeventInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/async/cloudevent/interfaces"
	eventWriter "github.com/flyteorg/flyte/flyteadmin/pkg/async/events/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/async/notifications"
	notificationInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/async/notifications/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	dataInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/data/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/executions"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/resources"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/util"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/validation"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/interfaces"
	repositoryInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/transformers"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	workflowengineInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/workflowengine/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/plugins"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

const childContainerQueueKey = "child_queue"

var activeExecutionPhases = []string{
	core.WorkflowExecution_QUEUED.String(),
	core.WorkflowExecution_RUNNING.String(),
}

type executionSystemMetrics struct {
	Scope                      promutils.Scope
	ActiveExecutions           prometheus.Gauge
	ExecutionsCreated          prometheus.Counter
	ExecutionsTerminated       labeled.Counter
	ExecutionEventsCreated     prometheus.Counter
	PropellerFailures          prometheus.Counter
	PublishNotificationError   prometheus.Counter
	TransformerError           prometheus.Counter
	UnexpectedDataError        prometheus.Counter
	SpecSizeBytes              prometheus.Summary
	ClosureSizeBytes           prometheus.Summary
	AcceptanceDelay            prometheus.Summary
	PublishEventError          prometheus.Counter
	TerminateExecutionFailures prometheus.Counter
	ConcurrencyCheckDuration   labeled.StopWatch
	ConcurrencyLimitHits       *prometheus.CounterVec
}

type executionUserMetrics struct {
	Scope                        promutils.Scope
	ScheduledExecutionDelays     sync.Map // Map of [project] -> map of [domain] -> stop watch
	WorkflowExecutionDurations   sync.Map // Map of [project] -> map of [domain] -> stop watch
	WorkflowExecutionInputBytes  prometheus.Summary
	WorkflowExecutionOutputBytes prometheus.Summary
}

type ExecutionManager struct {
	db                        repositoryInterfaces.Repository
	config                    runtimeInterfaces.Configuration
	storageClient             *storage.DataStore
	queueAllocator            executions.QueueAllocator
	_clock                    clock.Clock
	systemMetrics             executionSystemMetrics
	userMetrics               *executionUserMetrics
	notificationClient        notificationInterfaces.Publisher
	urlData                   dataInterfaces.RemoteURLInterface
	workflowManager           interfaces.WorkflowInterface
	namedEntityManager        interfaces.NamedEntityInterface
	resourceManager           interfaces.ResourceInterface
	qualityOfServiceAllocator executions.QualityOfServiceAllocator
	eventPublisher            notificationInterfaces.Publisher
	cloudEventPublisher       notificationInterfaces.Publisher
	dbEventWriter             eventWriter.WorkflowExecutionEventWriter
	pluginRegistry            *plugins.Registry
}

func getExecutionContext(ctx context.Context, id *core.WorkflowExecutionIdentifier) context.Context {
	ctx = contextutils.WithExecutionID(ctx, id.GetName())
	return contextutils.WithProjectDomain(ctx, id.GetProject(), id.GetDomain())
}

// Returns the unique string which identifies the authenticated end user (if any).
func getUser(ctx context.Context) string {
	identityContext := auth.IdentityContextFromContext(ctx)
	return identityContext.UserID()
}

func (m *ExecutionManager) populateExecutionQueue(
	ctx context.Context, identifier *core.Identifier, compiledWorkflow *core.CompiledWorkflowClosure) {
	queueConfig := m.queueAllocator.GetQueue(ctx, identifier)
	for _, task := range compiledWorkflow.GetTasks() {
		container := task.GetTemplate().GetContainer()
		if container == nil {
			// Unrecognized target type, nothing to do
			continue
		}

		if queueConfig.DynamicQueue != "" {
			logger.Debugf(ctx, "Assigning %s as child queue for task %+v", queueConfig.DynamicQueue, task.GetTemplate().GetId())
			container.Config = append(container.GetConfig(), &core.KeyValuePair{
				Key:   childContainerQueueKey,
				Value: queueConfig.DynamicQueue,
			})
		}
	}
}

func validateMapSize(maxEntries int, candidate map[string]string, candidateName string) error {
	if maxEntries == 0 {
		// Treat the max as unset
		return nil
	}
	if len(candidate) > maxEntries {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument, "%s has too many entries [%v > %v]",
			candidateName, len(candidate), maxEntries)
	}
	return nil
}

type mapWithValues interface {
	GetValues() map[string]string
}

func resolveStringMap(preferredValues, defaultValues mapWithValues, valueName string, maxEntries int) (map[string]string, error) {
	var response = make(map[string]string)
	if preferredValues != nil && preferredValues.GetValues() != nil {
		response = preferredValues.GetValues()
	} else if defaultValues != nil && defaultValues.GetValues() != nil {
		response = defaultValues.GetValues()
	}

	err := validateMapSize(maxEntries, response, valueName)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (m *ExecutionManager) addPluginOverrides(ctx context.Context, executionID *core.WorkflowExecutionIdentifier,
	workflowName, launchPlanName string) ([]*admin.PluginOverride, error) {
	override, err := m.resourceManager.GetResource(ctx, interfaces.ResourceRequest{
		Project:      executionID.GetProject(),
		Domain:       executionID.GetDomain(),
		Workflow:     workflowName,
		LaunchPlan:   launchPlanName,
		ResourceType: admin.MatchableResource_PLUGIN_OVERRIDE,
	})
	if err != nil && !errors.IsDoesNotExistError(err) {
		return nil, err
	}
	if override != nil && override.Attributes != nil && override.Attributes.GetPluginOverrides() != nil {
		return override.Attributes.GetPluginOverrides().GetOverrides(), nil
	}
	return nil, nil
}

// TODO: Delete this code usage after the flyte v0.17.0 release
// Assumes input contains a compiled task with a valid container resource execConfig.
//
// Note: The system will assign a system-default value for request but for limit it will deduce it from the request
// itself => Limit := Min([Some-Multiplier X Request], System-Max). For now we are using a multiplier of 1. In
// general we recommend the users to set limits close to requests for more predictability in the system.
func (m *ExecutionManager) setCompiledTaskDefaults(ctx context.Context, task *core.CompiledTask,
	platformTaskResources workflowengineInterfaces.TaskResources) {

	if task == nil {
		logger.Warningf(ctx, "Can't set default resources for nil task.")
		return
	}

	if task.GetTemplate() == nil || task.GetTemplate().GetContainer() == nil {
		// Nothing to do
		logger.Debugf(ctx, "Not setting default resources for task [%+v], no container resources found to check", task)
		return
	}

	if task.GetTemplate().GetContainer().GetResources() == nil {
		// In case of no resources on the container, create empty requests and limits
		// so the container will still have resources configure properly
		task.Template.GetContainer().Resources = &core.Resources{
			Requests: []*core.Resources_ResourceEntry{},
			Limits:   []*core.Resources_ResourceEntry{},
		}
	}

	var finalizedResourceRequests = make([]*core.Resources_ResourceEntry, 0)
	var finalizedResourceLimits = make([]*core.Resources_ResourceEntry, 0)

	// The IDL representation for container-type tasks represents resources as a list with string quantities.
	// In order to easily reason about them we convert them to a set where we can O(1) fetch specific resources (e.g. CPU)
	// and represent them as comparable quantities rather than strings.
	taskResourceRequirements := util.GetCompleteTaskResourceRequirements(ctx, task.GetTemplate().GetId(), task)

	cpu := flytek8s.AdjustOrDefaultResource(taskResourceRequirements.Defaults.CPU, taskResourceRequirements.Limits.CPU,
		platformTaskResources.Defaults.CPU, platformTaskResources.Limits.CPU)
	finalizedResourceRequests = append(finalizedResourceRequests, &core.Resources_ResourceEntry{
		Name:  core.Resources_CPU,
		Value: cpu.Request.String(),
	})
	finalizedResourceLimits = append(finalizedResourceLimits, &core.Resources_ResourceEntry{
		Name:  core.Resources_CPU,
		Value: cpu.Limit.String(),
	})

	memory := flytek8s.AdjustOrDefaultResource(taskResourceRequirements.Defaults.Memory, taskResourceRequirements.Limits.Memory,
		platformTaskResources.Defaults.Memory, platformTaskResources.Limits.Memory)
	finalizedResourceRequests = append(finalizedResourceRequests, &core.Resources_ResourceEntry{
		Name:  core.Resources_MEMORY,
		Value: memory.Request.String(),
	})
	finalizedResourceLimits = append(finalizedResourceLimits, &core.Resources_ResourceEntry{
		Name:  core.Resources_MEMORY,
		Value: memory.Limit.String(),
	})

	// Only assign ephemeral storage when it is either requested or limited in the task definition, or a platform
	// default exists.
	if !taskResourceRequirements.Defaults.EphemeralStorage.IsZero() ||
		!taskResourceRequirements.Limits.EphemeralStorage.IsZero() ||
		!platformTaskResources.Defaults.EphemeralStorage.IsZero() {
		ephemeralStorage := flytek8s.AdjustOrDefaultResource(taskResourceRequirements.Defaults.EphemeralStorage, taskResourceRequirements.Limits.EphemeralStorage,
			platformTaskResources.Defaults.EphemeralStorage, platformTaskResources.Limits.EphemeralStorage)
		finalizedResourceRequests = append(finalizedResourceRequests, &core.Resources_ResourceEntry{
			Name:  core.Resources_EPHEMERAL_STORAGE,
			Value: ephemeralStorage.Request.String(),
		})
		finalizedResourceLimits = append(finalizedResourceLimits, &core.Resources_ResourceEntry{
			Name:  core.Resources_EPHEMERAL_STORAGE,
			Value: ephemeralStorage.Limit.String(),
		})
	}

	// Only assign gpu when it is either requested or limited in the task definition, or a platform default exists.
	if !taskResourceRequirements.Defaults.GPU.IsZero() ||
		!taskResourceRequirements.Limits.GPU.IsZero() ||
		!platformTaskResources.Defaults.GPU.IsZero() {
		gpu := flytek8s.AdjustOrDefaultResource(taskResourceRequirements.Defaults.GPU, taskResourceRequirements.Limits.GPU,
			platformTaskResources.Defaults.GPU, platformTaskResources.Limits.GPU)
		finalizedResourceRequests = append(finalizedResourceRequests, &core.Resources_ResourceEntry{
			Name:  core.Resources_GPU,
			Value: gpu.Request.String(),
		})
		finalizedResourceLimits = append(finalizedResourceLimits, &core.Resources_ResourceEntry{
			Name:  core.Resources_GPU,
			Value: gpu.Limit.String(),
		})
	}

	task.Template.GetContainer().Resources = &core.Resources{
		Requests: finalizedResourceRequests,
		Limits:   finalizedResourceLimits,
	}
}

// Fetches inherited execution metadata including the parent node execution db model id and the source execution model id
// as well as sets request spec metadata with the inherited principal and adjusted nesting data.
func (m *ExecutionManager) getInheritedExecMetadata(ctx context.Context, requestSpec *admin.ExecutionSpec,
	workflowExecutionID *core.WorkflowExecutionIdentifier) (parentNodeExecutionID uint, sourceExecutionID uint, err error) {
	if requestSpec.GetMetadata() == nil || requestSpec.GetMetadata().GetParentNodeExecution() == nil {
		return parentNodeExecutionID, sourceExecutionID, nil
	}
	parentNodeExecutionModel, err := util.GetNodeExecutionModel(ctx, m.db, requestSpec.GetMetadata().GetParentNodeExecution())
	if err != nil {
		logger.Errorf(ctx, "Failed to get node execution [%+v] that launched this execution [%+v] with error %v",
			requestSpec.GetMetadata().GetParentNodeExecution(), workflowExecutionID, err)
		return parentNodeExecutionID, sourceExecutionID, err
	}

	parentNodeExecutionID = parentNodeExecutionModel.ID

	sourceExecutionModel, err := util.GetExecutionModel(ctx, m.db, requestSpec.GetMetadata().GetParentNodeExecution().GetExecutionId())
	if err != nil {
		logger.Errorf(ctx, "Failed to get workflow execution [%+v] that launched this execution [%+v] with error %v",
			requestSpec.GetMetadata().GetParentNodeExecution(), workflowExecutionID, err)
		return parentNodeExecutionID, sourceExecutionID, err
	}
	sourceExecutionID = sourceExecutionModel.ID
	requestSpec.Metadata.Principal = sourceExecutionModel.User
	sourceExecution, err := transformers.FromExecutionModel(ctx, *sourceExecutionModel, transformers.DefaultExecutionTransformerOptions)
	if err != nil {
		logger.Errorf(ctx, "Failed transform parent execution model for child execution [%+v] with err: %v", workflowExecutionID, err)
		return parentNodeExecutionID, sourceExecutionID, err
	}
	if sourceExecution.GetSpec().GetMetadata() != nil {
		requestSpec.Metadata.Nesting = sourceExecution.GetSpec().GetMetadata().GetNesting() + 1
	} else {
		requestSpec.Metadata.Nesting = 1
	}

	// If the source execution has a cluster label, inherit it.
	if sourceExecution.GetSpec().GetExecutionClusterLabel() != nil {
		logger.Infof(ctx, "Inherited execution label from source execution [%+v]", sourceExecution.GetSpec().GetExecutionClusterLabel().GetValue())
		requestSpec.ExecutionClusterLabel = sourceExecution.GetSpec().GetExecutionClusterLabel()
	}
	return parentNodeExecutionID, sourceExecutionID, nil
}

// Produces execution-time attributes for workflow execution.
// Defaults to overridable execution values set in the execution create request, then looks at the launch plan values
// (if any) before defaulting to values set in the matchable resource db and further if matchable resources don't
// exist then defaults to one set in application configuration
func (m *ExecutionManager) getExecutionConfig(ctx context.Context, request *admin.ExecutionCreateRequest,
	launchPlan *admin.LaunchPlan) (*admin.WorkflowExecutionConfig, error) {

	workflowExecConfig := &admin.WorkflowExecutionConfig{}
	// Merge the request spec into workflowExecConfig
	workflowExecConfig = util.MergeIntoExecConfig(workflowExecConfig, request.GetSpec())

	var workflowName string
	if launchPlan != nil && launchPlan.GetSpec() != nil {
		// Merge the launch plan spec into workflowExecConfig
		workflowExecConfig = util.MergeIntoExecConfig(workflowExecConfig, launchPlan.GetSpec())
		if launchPlan.GetSpec().GetWorkflowId() != nil {
			workflowName = launchPlan.GetSpec().GetWorkflowId().GetName()
		}
	}

	// This will get the most specific Workflow Execution Config.
	matchableResource, err := util.GetMatchableResource(ctx, m.resourceManager,
		admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG, request.GetProject(), request.GetDomain(), workflowName)
	if err != nil {
		return nil, err
	}
	if matchableResource != nil && matchableResource.Attributes.GetWorkflowExecutionConfig() != nil {
		// merge the matchable resource workflow execution config into workflowExecConfig
		workflowExecConfig = util.MergeIntoExecConfig(workflowExecConfig,
			matchableResource.Attributes.GetWorkflowExecutionConfig())
	}

	// To match what the front-end will display to the user, we need to do the project level query too.
	// This searches only for a direct match, and will not merge in system config level defaults like the
	// GetProjectAttributes call does, since that's done below.
	// The reason we need to do the project level query is for the case where some configs (say max parallelism)
	// is set on the project level, but other items (say service account) is set on the project-domain level.
	// In this case you want to use the project-domain service account, the project-level max parallelism, and
	// system level defaults for the rest.
	// See FLYTE-2322 for more background information.
	projectMatchableResource, err := util.GetMatchableResource(ctx, m.resourceManager,
		admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG, request.GetProject(), "", "")
	if err != nil {
		return nil, err
	}
	if projectMatchableResource != nil && projectMatchableResource.Attributes.GetWorkflowExecutionConfig() != nil {
		// merge the matchable resource workflow execution config into workflowExecConfig
		workflowExecConfig = util.MergeIntoExecConfig(workflowExecConfig,
			projectMatchableResource.Attributes.GetWorkflowExecutionConfig())
	}

	// Backward compatibility changes to get security context from auth role.
	// Older authRole or auth fields in the launchplan spec or execution request need to be used over application defaults.
	// This portion of the code makes sure if newer way of setting security context is empty i.e
	// K8sServiceAccount and  IamRole is empty then get the values from the deprecated fields.
	resolvedAuthRole := resolveAuthRole(request, launchPlan)
	resolvedSecurityCtx := resolveSecurityCtx(ctx, workflowExecConfig.GetSecurityContext(), resolvedAuthRole)
	if workflowExecConfig.GetSecurityContext() == nil &&
		(len(resolvedSecurityCtx.GetRunAs().GetK8SServiceAccount()) > 0 ||
			len(resolvedSecurityCtx.GetRunAs().GetIamRole()) > 0) {
		workflowExecConfig.SecurityContext = resolvedSecurityCtx
	}

	// Merge the application config into workflowExecConfig. If even the deprecated fields are not set
	workflowExecConfig = util.MergeIntoExecConfig(workflowExecConfig, m.config.ApplicationConfiguration().GetTopLevelConfig())
	// Explicitly set the security context if its nil since downstream we expect this settings to be available
	if workflowExecConfig.GetSecurityContext() == nil {
		workflowExecConfig.SecurityContext = &core.SecurityContext{
			RunAs: &core.Identity{},
		}
	}

	if workflowExecConfig.GetSecurityContext().GetRunAs() == nil {
		workflowExecConfig.SecurityContext.RunAs = &core.Identity{}
	}

	// In the case of reference_launch_plan subworkflow, the context comes from flytepropeller instead of the user side, so user auth is missing.
	// We skip getUserIdentityFromContext but can still get ExecUserId because flytepropeller passes it in the execution request.
	// https://github.com/flyteorg/flytepropeller/blob/03a6672960ed04e7687ba4f790fee9a02a4057fb/pkg/controller/nodes/subworkflow/launchplan/admin.go#L114
	if workflowExecConfig.GetSecurityContext().GetRunAs().GetExecutionIdentity() == "" {
		workflowExecConfig.SecurityContext.RunAs.ExecutionIdentity = auth.IdentityContextFromContext(ctx).ExecutionIdentity()
	}

	logger.Infof(ctx, "getting the workflow execution config from application configuration")
	// Defaults to one from the application config
	return workflowExecConfig, nil
}

func (m *ExecutionManager) getClusterAssignment(ctx context.Context, req *admin.ExecutionCreateRequest) (*admin.ClusterAssignment, error) {
	storedAssignment, err := m.fetchClusterAssignment(ctx, req.GetProject(), req.GetDomain())
	if err != nil {
		return nil, err
	}

	reqAssignment := req.GetSpec().GetClusterAssignment()
	reqPool := reqAssignment.GetClusterPoolName()
	storedPool := storedAssignment.GetClusterPoolName()
	if reqPool == "" {
		return storedAssignment, nil
	}

	if storedPool == "" {
		return reqAssignment, nil
	}

	if reqPool != storedPool {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "execution with project %q and domain %q cannot run on cluster pool %q, because its configured to run on pool %q", req.GetProject(), req.GetDomain(), reqPool, storedPool)
	}

	return storedAssignment, nil
}

func (m *ExecutionManager) fetchClusterAssignment(ctx context.Context, project, domain string) (*admin.ClusterAssignment, error) {
	resource, err := m.resourceManager.GetResource(ctx, interfaces.ResourceRequest{
		Project:      project,
		Domain:       domain,
		ResourceType: admin.MatchableResource_CLUSTER_ASSIGNMENT,
	})
	if err != nil && !errors.IsDoesNotExistError(err) {
		logger.Errorf(ctx, "Failed to get cluster assignment overrides with error: %v", err)
		return nil, err
	}
	if resource != nil && resource.Attributes.GetClusterAssignment() != nil {
		return resource.Attributes.GetClusterAssignment(), nil
	}

	var clusterAssignment *admin.ClusterAssignment
	domainAssignment := m.config.ClusterPoolAssignmentConfiguration().GetClusterPoolAssignments()[domain]
	if domainAssignment.Pool != "" {
		clusterAssignment = &admin.ClusterAssignment{ClusterPoolName: domainAssignment.Pool}
	}
	return clusterAssignment, nil
}

func (m *ExecutionManager) launchSingleTaskExecution(
	ctx context.Context, request *admin.ExecutionCreateRequest, requestedAt time.Time) (
	context.Context, *models.Execution, error) {

	taskModel, err := m.db.TaskRepo().Get(ctx, repositoryInterfaces.Identifier{
		Project: request.GetSpec().GetLaunchPlan().GetProject(),
		Domain:  request.GetSpec().GetLaunchPlan().GetDomain(),
		Name:    request.GetSpec().GetLaunchPlan().GetName(),
		Version: request.GetSpec().GetLaunchPlan().GetVersion(),
	})
	if err != nil {
		return nil, nil, err
	}
	task, err := transformers.FromTaskModel(taskModel)
	if err != nil {
		return nil, nil, err
	}

	// Prepare a skeleton workflow and launch plan
	taskIdentifier := request.GetSpec().GetLaunchPlan()
	workflowModel, err :=
		util.CreateOrGetWorkflowModel(ctx, request, m.db, m.workflowManager, m.namedEntityManager, taskIdentifier, &task)
	if err != nil {
		logger.Debugf(ctx, "Failed to created skeleton workflow for [%+v] with err: %v", taskIdentifier, err)
		return nil, nil, err
	}
	workflow, err := transformers.FromWorkflowModel(*workflowModel)
	if err != nil {
		return nil, nil, err
	}

	launchPlan, err := util.CreateOrGetLaunchPlan(ctx, m.db, m.config, m.namedEntityManager, taskIdentifier,
		workflow.GetClosure().GetCompiledWorkflow().GetPrimary().GetTemplate().GetInterface(), workflowModel.ID, request.GetSpec())
	if err != nil {
		return nil, nil, err
	}

	executionInputs, err := validation.CheckAndFetchInputsForExecution(
		request.GetInputs(),
		launchPlan.GetSpec().GetFixedInputs(),
		launchPlan.GetClosure().GetExpectedInputs(),
	)
	if err != nil {
		logger.Debugf(ctx, "Failed to CheckAndFetchInputsForExecution with request.Inputs: %+v"+
			"fixed inputs: %+v and expected inputs: %+v with err %v",
			request.GetInputs(), launchPlan.GetSpec().GetFixedInputs(), launchPlan.GetClosure().GetExpectedInputs(), err)
		return nil, nil, err
	}

	name := util.GetExecutionName(request)
	workflowExecutionID := &core.WorkflowExecutionIdentifier{
		Project: request.GetProject(),
		Domain:  request.GetDomain(),
		Name:    name,
	}

	// Overlap the blob store reads and writes
	getClosureGroup, getClosureGroupCtx := errgroup.WithContext(ctx)
	var closure *admin.WorkflowClosure
	getClosureGroup.Go(func() error {
		var err error
		closure, err = util.FetchAndGetWorkflowClosure(getClosureGroupCtx, m.storageClient, workflowModel.RemoteClosureIdentifier)
		return err
	})

	offloadInputsGroup, offloadInputsGroupCtx := errgroup.WithContext(ctx)
	var inputsURI storage.DataReference
	offloadInputsGroup.Go(func() error {
		var err error
		inputsURI, err = common.OffloadLiteralMap(offloadInputsGroupCtx, m.storageClient, executionInputs, // or request.Inputs?
			workflowExecutionID.GetProject(), workflowExecutionID.GetDomain(), workflowExecutionID.GetName(), shared.Inputs)
		return err
	})

	var userInputsURI storage.DataReference
	offloadInputsGroup.Go(func() error {
		var err error
		userInputsURI, err = common.OffloadLiteralMap(offloadInputsGroupCtx, m.storageClient, request.GetInputs(),
			workflowExecutionID.GetProject(), workflowExecutionID.GetDomain(), workflowExecutionID.GetName(), shared.UserInputs)
		return err
	})

	err = getClosureGroup.Wait()
	if err != nil {
		return nil, nil, err
	}
	closure.CreatedAt = workflow.GetClosure().GetCreatedAt()
	workflow.Closure = closure

	ctx = getExecutionContext(ctx, workflowExecutionID)
	namespace := common.GetNamespaceName(
		m.config.NamespaceMappingConfiguration().GetNamespaceTemplate(), workflowExecutionID.GetProject(), workflowExecutionID.GetDomain())

	requestSpec := request.GetSpec()
	if requestSpec.GetMetadata() == nil {
		requestSpec.Metadata = &admin.ExecutionMetadata{}
	}
	requestSpec.Metadata.Principal = getUser(ctx)

	// Get the node execution (if any) that launched this execution
	var parentNodeExecutionID uint
	var sourceExecutionID uint
	parentNodeExecutionID, sourceExecutionID, err = m.getInheritedExecMetadata(ctx, requestSpec, workflowExecutionID)
	if err != nil {
		return nil, nil, err
	}

	// Dynamically assign task resource defaults.
	platformTaskResources := util.GetTaskResources(ctx, workflow.GetId(), m.resourceManager, m.config.TaskResourceConfiguration())
	for _, t := range workflow.GetClosure().GetCompiledWorkflow().GetTasks() {
		m.setCompiledTaskDefaults(ctx, t, platformTaskResources)
	}

	// Dynamically assign execution queues.
	m.populateExecutionQueue(ctx, workflow.GetId(), workflow.GetClosure().GetCompiledWorkflow())

	executionConfig, err := m.getExecutionConfig(ctx, request, nil)
	if err != nil {
		return nil, nil, err
	}

	var labels map[string]string
	if executionConfig.GetLabels() != nil {
		labels = executionConfig.GetLabels().GetValues()
	}

	labels, err = m.addProjectLabels(ctx, request.GetProject(), labels)
	if err != nil {
		return nil, nil, err
	}

	var annotations map[string]string
	if executionConfig.GetAnnotations() != nil {
		annotations = executionConfig.GetAnnotations().GetValues()
	}

	annotations = m.addIdentityAnnotations(ctx, annotations)

	var rawOutputDataConfig *admin.RawOutputDataConfig
	if executionConfig.GetRawOutputDataConfig() != nil {
		rawOutputDataConfig = executionConfig.GetRawOutputDataConfig()
	}

	clusterAssignment, err := m.getClusterAssignment(ctx, request)
	if err != nil {
		return nil, nil, err
	}

	var executionClusterLabel *admin.ExecutionClusterLabel
	if requestSpec.GetExecutionClusterLabel() != nil {
		executionClusterLabel = requestSpec.GetExecutionClusterLabel()
	}
	executionParameters := workflowengineInterfaces.ExecutionParameters{
		Inputs:                executionInputs,
		AcceptedAt:            requestedAt,
		Labels:                labels,
		Annotations:           annotations,
		ExecutionConfig:       executionConfig,
		TaskResources:         &platformTaskResources,
		EventVersion:          m.config.ApplicationConfiguration().GetTopLevelConfig().EventVersion,
		RoleNameKey:           m.config.ApplicationConfiguration().GetTopLevelConfig().RoleNameKey,
		RawOutputDataConfig:   rawOutputDataConfig,
		ClusterAssignment:     clusterAssignment,
		ExecutionClusterLabel: executionClusterLabel,
	}

	overrides, err := m.addPluginOverrides(ctx, workflowExecutionID, workflowExecutionID.GetName(), "")
	if err != nil {
		return nil, nil, err
	}
	if overrides != nil {
		executionParameters.TaskPluginOverrides = overrides
	}
	if request.GetSpec().GetMetadata() != nil && request.GetSpec().GetMetadata().GetReferenceExecution() != nil &&
		request.GetSpec().GetMetadata().GetMode() == admin.ExecutionMetadata_RECOVERED {
		executionParameters.RecoveryExecution = request.GetSpec().GetMetadata().GetReferenceExecution()
	}

	err = offloadInputsGroup.Wait()
	if err != nil {
		return nil, nil, err
	}

	workflowExecutor := plugins.Get[workflowengineInterfaces.WorkflowExecutor](m.pluginRegistry, plugins.PluginIDWorkflowExecutor)
	execInfo, err := workflowExecutor.Execute(ctx, workflowengineInterfaces.ExecutionData{
		Namespace:                namespace,
		ExecutionID:              workflowExecutionID,
		ReferenceWorkflowName:    workflow.GetId().GetName(),
		ReferenceLaunchPlanName:  launchPlan.GetId().GetName(),
		WorkflowClosure:          workflow.GetClosure().GetCompiledWorkflow(),
		WorkflowClosureReference: storage.DataReference(workflowModel.RemoteClosureIdentifier),
		ExecutionParameters:      executionParameters,
		OffloadedInputsReference: inputsURI,
	})

	if err != nil {
		m.systemMetrics.PropellerFailures.Inc()
		logger.Infof(ctx, "Failed to execute workflow %+v with execution id %+v and inputs %+v with err %v",
			request, &workflowExecutionID, request.GetInputs(), err)
		return nil, nil, err
	}
	executionCreatedAt := time.Now()
	acceptanceDelay := executionCreatedAt.Sub(requestedAt)
	m.systemMetrics.AcceptanceDelay.Observe(acceptanceDelay.Seconds())

	// Request notification settings takes precedence over the launch plan settings.
	// If there is no notification in the request and DisableAll is not true, use the settings from the launch plan.
	var notificationsSettings []*admin.Notification
	if launchPlan.GetSpec().GetEntityMetadata() != nil {
		notificationsSettings = launchPlan.GetSpec().GetEntityMetadata().GetNotifications()
	}
	if request.GetSpec().GetNotifications() != nil && request.GetSpec().GetNotifications().GetNotifications() != nil &&
		len(request.GetSpec().GetNotifications().GetNotifications()) > 0 {
		notificationsSettings = request.GetSpec().GetNotifications().GetNotifications()
	} else if request.GetSpec().GetDisableAll() {
		notificationsSettings = make([]*admin.Notification, 0)
	}

	executionModel, err := transformers.CreateExecutionModel(transformers.CreateExecutionModelInput{
		WorkflowExecutionID: workflowExecutionID,
		RequestSpec:         requestSpec,
		TaskID:              taskModel.ID,
		WorkflowID:          workflowModel.ID,
		// The execution is not considered running until the propeller sends a specific event saying so.
		CreatedAt:             m._clock.Now(),
		Notifications:         notificationsSettings,
		WorkflowIdentifier:    workflow.GetId(),
		ParentNodeExecutionID: parentNodeExecutionID,
		SourceExecutionID:     sourceExecutionID,
		Cluster:               execInfo.Cluster,
		InputsURI:             inputsURI,
		UserInputsURI:         userInputsURI,
		SecurityContext:       executionConfig.GetSecurityContext(),
		LaunchEntity:          taskIdentifier.GetResourceType(),
		Namespace:             namespace,
	})
	if err != nil {
		logger.Infof(ctx, "Failed to create execution model in transformer for id: [%+v] with err: %v",
			workflowExecutionID, err)
		return nil, nil, err
	}
	m.userMetrics.WorkflowExecutionInputBytes.Observe(float64(proto.Size(request.GetInputs())))
	return ctx, executionModel, nil
}

func resolveAuthRole(request *admin.ExecutionCreateRequest, launchPlan *admin.LaunchPlan) *admin.AuthRole {
	if request.GetSpec().GetAuthRole() != nil {
		return request.GetSpec().GetAuthRole()
	}

	if launchPlan == nil || launchPlan.GetSpec() == nil {
		return &admin.AuthRole{}
	}

	// Set role permissions based on launch plan Auth values.
	// The branched-ness of this check is due to the presence numerous deprecated fields
	if launchPlan.GetSpec().GetAuthRole() != nil {
		return launchPlan.GetSpec().GetAuthRole()
	} else if launchPlan.GetSpec().GetAuth() != nil {
		return &admin.AuthRole{
			AssumableIamRole:         launchPlan.GetSpec().GetAuth().GetAssumableIamRole(),
			KubernetesServiceAccount: launchPlan.GetSpec().GetAuth().GetKubernetesServiceAccount(),
		}
	} else if len(launchPlan.GetSpec().GetRole()) > 0 {
		return &admin.AuthRole{
			AssumableIamRole: launchPlan.GetSpec().GetRole(),
		}
	}

	return &admin.AuthRole{}
}

func resolveSecurityCtx(ctx context.Context, executionConfigSecurityCtx *core.SecurityContext,
	resolvedAuthRole *admin.AuthRole) *core.SecurityContext {
	// Use security context from the executionConfigSecurityCtx if its set and non empty or else resolve from authRole
	if executionConfigSecurityCtx != nil && executionConfigSecurityCtx.GetRunAs() != nil &&
		(len(executionConfigSecurityCtx.GetRunAs().GetK8SServiceAccount()) > 0 ||
			len(executionConfigSecurityCtx.GetRunAs().GetIamRole()) > 0 ||
			len(executionConfigSecurityCtx.GetRunAs().GetExecutionIdentity()) > 0) {
		return executionConfigSecurityCtx
	}
	logger.Warn(ctx, "Setting security context from auth Role")
	return &core.SecurityContext{
		RunAs: &core.Identity{
			IamRole:           resolvedAuthRole.GetAssumableIamRole(),
			K8SServiceAccount: resolvedAuthRole.GetKubernetesServiceAccount(),
		},
	}
}

// getStringFromInput should be called when a tag or partition value is a binding to an input. the input is looked up
// from the input map and the binding, and an error is returned if the input key is not in the map.
func (m *ExecutionManager) getStringFromInput(ctx context.Context, inputBinding *core.InputBindingData, inputs map[string]*core.Literal) (string, error) {

	inputName := inputBinding.GetVar()
	if inputName == "" {
		return "", errors.NewFlyteAdminErrorf(codes.InvalidArgument, "input binding var is empty")
	}
	if inputVal, ok := inputs[inputName]; ok {
		if inputVal.GetScalar() == nil || inputVal.GetScalar().GetPrimitive() == nil {
			return "", errors.NewFlyteAdminErrorf(codes.InvalidArgument, "invalid input value [%+v]", inputVal)
		}
		var strVal string
		p := inputVal.GetScalar().GetPrimitive()
		switch p.GetValue().(type) {
		case *core.Primitive_Integer:
			strVal = p.GetStringValue()
		case *core.Primitive_Datetime:
			t := time.Unix(p.GetDatetime().GetSeconds(), int64(p.GetDatetime().GetNanos()))
			t = t.In(time.UTC)
			strVal = t.Format("2006-01-02")
		case *core.Primitive_StringValue:
			strVal = p.GetStringValue()
		case *core.Primitive_FloatValue:
			strVal = fmt.Sprintf("%.2f", p.GetFloatValue())
		case *core.Primitive_Boolean:
			strVal = fmt.Sprintf("%t", p.GetBoolean())
		default:
			strVal = fmt.Sprintf("%s", p.GetValue())
		}

		logger.Debugf(ctx, "String templating returning [%s] for [%+v]", strVal, inputVal)
		return strVal, nil
	}
	return "", errors.NewFlyteAdminErrorf(codes.InvalidArgument, "input binding var [%s] not found", inputName)
}

func (m *ExecutionManager) getLabelValue(ctx context.Context, l *core.LabelValue, inputs map[string]*core.Literal) (string, error) {
	if l == nil {
		return "", errors.NewFlyteAdminErrorf(codes.InvalidArgument, "label value is nil")
	}
	if l.GetInputBinding() != nil {
		return m.getStringFromInput(ctx, l.GetInputBinding(), inputs)
	}
	if l.GetStaticValue() != "" {
		return l.GetStaticValue(), nil
	}
	return "", errors.NewFlyteAdminErrorf(codes.InvalidArgument, "label value is empty")
}

func (m *ExecutionManager) fillInTemplateArgs(ctx context.Context, query *core.ArtifactQuery, inputs map[string]*core.Literal) (*core.ArtifactQuery, error) {
	if query.GetUri() != "" {
		// If a query string, then just pass it through, nothing to fill in.
		return query, nil
	} else if query.GetArtifactId() != nil {
		artifactID := query.GetArtifactId()
		ak := artifactID.GetArtifactKey()
		if ak == nil {
			return query, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "id doesn't have key")
		}
		var project, domain string
		if ak.GetProject() == "" {
			project = contextutils.Value(ctx, contextutils.ProjectKey)
		} else {
			project = ak.GetProject()
		}
		if ak.GetDomain() == "" {
			domain = contextutils.Value(ctx, contextutils.DomainKey)
		} else {
			domain = ak.GetDomain()
		}

		var partitions map[string]*core.LabelValue

		if artifactID.GetPartitions().GetValue() != nil {
			partitions = make(map[string]*core.LabelValue, len(artifactID.GetPartitions().GetValue()))
			for k, v := range artifactID.GetPartitions().GetValue() {
				newValue, err := m.getLabelValue(ctx, v, inputs)
				if err != nil {
					logger.Errorf(ctx, "Failed to resolve partition input string [%s] [%+v] [%v]", k, v, err)
					return query, err
				}
				partitions[k] = &core.LabelValue{Value: &core.LabelValue_StaticValue{StaticValue: newValue}}
			}
		}

		var timePartition *core.TimePartition
		if artifactID.GetTimePartition().GetValue() != nil {
			if artifactID.GetTimePartition().GetValue().GetTimeValue() != nil {
				// If the time value is set, then just pass it through, nothing to fill in.
				timePartition = artifactID.GetTimePartition()
			} else if artifactID.GetTimePartition().GetValue().GetInputBinding() != nil {
				// Evaluate the time partition input binding
				lit, ok := inputs[artifactID.GetTimePartition().GetValue().GetInputBinding().GetVar()]
				if !ok {
					return query, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "time partition input binding var [%s] not found in inputs %v", artifactID.GetTimePartition().GetValue().GetInputBinding().GetVar(), inputs)
				}

				if lit.GetScalar().GetPrimitive().GetDatetime() == nil {
					return query, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
						"time partition binding to input var [%s] failing because %v is not a datetime",
						artifactID.GetTimePartition().GetValue().GetInputBinding().GetVar(), lit)
				}
				timePartition = &core.TimePartition{
					Value: &core.LabelValue{
						Value: &core.LabelValue_TimeValue{
							TimeValue: lit.GetScalar().GetPrimitive().GetDatetime(),
						},
					},
				}
			} else {
				return query, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "time partition value cannot be empty when evaluating query: %v", query)
			}
		}

		return &core.ArtifactQuery{
			Identifier: &core.ArtifactQuery_ArtifactId{
				ArtifactId: &core.ArtifactID{
					ArtifactKey: &core.ArtifactKey{
						Project: project,
						Domain:  domain,
						Name:    ak.GetName(),
					},
					Partitions: &core.Partitions{
						Value: partitions,
					},
					TimePartition: timePartition,
				},
			},
		}, nil
	}
	return query, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "query doesn't have uri, tag, or id")
}

func (m *ExecutionManager) launchExecutionAndPrepareModel(
	ctx context.Context, request *admin.ExecutionCreateRequest, requestedAt time.Time) (
	context.Context, *models.Execution, []*models.ExecutionTag, error) {

	err := validation.ValidateExecutionRequest(ctx, request, m.db, m.config.ApplicationConfiguration())
	if err != nil {
		logger.Debugf(ctx, "Failed to validate ExecutionCreateRequest %+v with err %v", request, err)
		return nil, nil, nil, err
	}

	if request.GetSpec().GetLaunchPlan().GetResourceType() == core.ResourceType_TASK {
		logger.Debugf(ctx, "Launching single task execution with [%+v]", request.GetSpec().GetLaunchPlan())
		// When tasks can have defaults this will need to handle Artifacts as well.
		ctx, model, err := m.launchSingleTaskExecution(ctx, request, requestedAt)
		return ctx, model, nil, err
	}
	return m.launchExecution(ctx, request, requestedAt)
}

func (m *ExecutionManager) launchExecution(
	ctx context.Context, request *admin.ExecutionCreateRequest, requestedAt time.Time) (context.Context, *models.Execution, []*models.ExecutionTag, error) {
	launchPlanModel, err := util.GetLaunchPlanModel(ctx, m.db, request.GetSpec().GetLaunchPlan())
	if err != nil {
		logger.Debugf(ctx, "Failed to get launch plan model for ExecutionCreateRequest %+v with err %v", request, err)
		return nil, nil, nil, err
	}
	launchPlan, err := transformers.FromLaunchPlanModel(launchPlanModel)
	if err != nil {
		logger.Debugf(ctx, "Failed to transform launch plan model %+v with err %v", launchPlanModel, err)
		return nil, nil, nil, err
	}

	var lpExpectedInputs *core.ParameterMap
	var usedArtifactIDs []*core.ArtifactID
	lpExpectedInputs = launchPlan.GetClosure().GetExpectedInputs()

	// Artifacts retrieved will need to be stored somewhere to ensure that we can re-emit events if necessary
	// in the future, and also to make sure that relaunch and recover can use it if necessary.
	executionInputs, err := validation.CheckAndFetchInputsForExecution(
		request.GetInputs(),
		launchPlan.GetSpec().GetFixedInputs(),
		lpExpectedInputs,
	)

	if err != nil {
		logger.Debugf(ctx, "Failed to CheckAndFetchInputsForExecution with request.Inputs: %+v"+
			"fixed inputs: %+v and expected inputs: %+v with err %v",
			request.GetInputs(), launchPlan.GetSpec().GetFixedInputs(), lpExpectedInputs, err)
		return nil, nil, nil, err
	}

	workflowModel, err := util.GetWorkflowModel(ctx, m.db, launchPlan.GetSpec().GetWorkflowId())
	if err != nil {
		logger.Debugf(ctx, "Failed to get workflow with id %+v with err %v", launchPlan.GetSpec().GetWorkflowId(), err)
		return nil, nil, nil, err
	}
	workflow, err := transformers.FromWorkflowModel(workflowModel)
	if err != nil {
		logger.Debugf(ctx, "Failed to get workflow with id %+v with err %v", launchPlan.GetSpec().GetWorkflowId(), err)
		return nil, nil, nil, err
	}

	name := util.GetExecutionName(request)
	workflowExecutionID := &core.WorkflowExecutionIdentifier{
		Project: request.GetProject(),
		Domain:  request.GetDomain(),
		Name:    name,
	}

	// Overlap the blob store reads and writes
	group, groupCtx := errgroup.WithContext(ctx)
	var closure *admin.WorkflowClosure
	group.Go(func() error {
		var err error
		closure, err = util.FetchAndGetWorkflowClosure(groupCtx, m.storageClient, workflowModel.RemoteClosureIdentifier)
		if err != nil {
			logger.Debugf(ctx, "Failed to get workflow with id %+v with err %v", launchPlan.GetSpec().GetWorkflowId(), err)
		}
		return err
	})

	var inputsURI storage.DataReference
	group.Go(func() error {
		var err error
		inputsURI, err = common.OffloadLiteralMap(groupCtx, m.storageClient, executionInputs,
			workflowExecutionID.GetProject(), workflowExecutionID.GetDomain(), workflowExecutionID.GetName(), shared.Inputs)
		return err
	})

	var userInputsURI storage.DataReference
	group.Go(func() error {
		var err error
		userInputsURI, err = common.OffloadLiteralMap(groupCtx, m.storageClient, request.GetInputs(),
			workflowExecutionID.GetProject(), workflowExecutionID.GetDomain(), workflowExecutionID.GetName(), shared.UserInputs)
		return err
	})

	err = group.Wait()
	if err != nil {
		return nil, nil, nil, err
	}
	closure.CreatedAt = workflow.GetClosure().GetCreatedAt()
	workflow.Closure = closure

	ctx = getExecutionContext(ctx, workflowExecutionID)
	var requestSpec = request.GetSpec()
	if requestSpec.GetMetadata() == nil {
		requestSpec.Metadata = &admin.ExecutionMetadata{}
	}
	requestSpec.Metadata.Principal = getUser(ctx)
	requestSpec.Metadata.ArtifactIds = usedArtifactIDs

	// Get the node and parent execution (if any) that launched this execution
	var parentNodeExecutionID uint
	var sourceExecutionID uint
	parentNodeExecutionID, sourceExecutionID, err = m.getInheritedExecMetadata(ctx, requestSpec, workflowExecutionID)
	if err != nil {
		return nil, nil, nil, err
	}

	// Dynamically assign task resource defaults.
	platformTaskResources := util.GetTaskResources(ctx, workflow.GetId(), m.resourceManager, m.config.TaskResourceConfiguration())
	for _, task := range workflow.GetClosure().GetCompiledWorkflow().GetTasks() {
		m.setCompiledTaskDefaults(ctx, task, platformTaskResources)
	}

	// Dynamically assign execution queues.
	m.populateExecutionQueue(ctx, workflow.GetId(), workflow.GetClosure().GetCompiledWorkflow())

	executionConfig, err := m.getExecutionConfig(ctx, request, launchPlan)
	if err != nil {
		return nil, nil, nil, err
	}

	namespace := common.GetNamespaceName(
		m.config.NamespaceMappingConfiguration().GetNamespaceTemplate(), workflowExecutionID.GetProject(), workflowExecutionID.GetDomain())

	labels, err := resolveStringMap(executionConfig.GetLabels(), launchPlan.GetSpec().GetLabels(), "labels", m.config.RegistrationValidationConfiguration().GetMaxLabelEntries())
	if err != nil {
		return nil, nil, nil, err
	}
	labels, err = m.addProjectLabels(ctx, request.GetProject(), labels)
	if err != nil {
		return nil, nil, nil, err
	}
	annotations, err := resolveStringMap(executionConfig.GetAnnotations(), launchPlan.GetSpec().GetAnnotations(), "annotations", m.config.RegistrationValidationConfiguration().GetMaxAnnotationEntries())
	if err != nil {
		return nil, nil, nil, err
	}

	annotations = m.addIdentityAnnotations(ctx, annotations)

	var rawOutputDataConfig *admin.RawOutputDataConfig
	if executionConfig.GetRawOutputDataConfig() != nil {
		rawOutputDataConfig = executionConfig.GetRawOutputDataConfig()
	}

	clusterAssignment, err := m.getClusterAssignment(ctx, request)
	if err != nil {
		return nil, nil, nil, err
	}

	var executionClusterLabel *admin.ExecutionClusterLabel
	if requestSpec.GetExecutionClusterLabel() != nil {
		executionClusterLabel = requestSpec.GetExecutionClusterLabel()
	}

	executionParameters := workflowengineInterfaces.ExecutionParameters{
		Inputs:                executionInputs,
		AcceptedAt:            requestedAt,
		Labels:                labels,
		Annotations:           annotations,
		ExecutionConfig:       executionConfig,
		TaskResources:         &platformTaskResources,
		EventVersion:          m.config.ApplicationConfiguration().GetTopLevelConfig().EventVersion,
		RoleNameKey:           m.config.ApplicationConfiguration().GetTopLevelConfig().RoleNameKey,
		RawOutputDataConfig:   rawOutputDataConfig,
		ClusterAssignment:     clusterAssignment,
		ExecutionClusterLabel: executionClusterLabel,
	}

	overrides, err := m.addPluginOverrides(ctx, workflowExecutionID, launchPlan.GetSpec().GetWorkflowId().GetName(), launchPlan.GetId().GetName())
	if err != nil {
		return nil, nil, nil, err
	}
	if overrides != nil {
		executionParameters.TaskPluginOverrides = overrides
	}

	if request.GetSpec().GetMetadata() != nil && request.GetSpec().GetMetadata().GetReferenceExecution() != nil &&
		request.GetSpec().GetMetadata().GetMode() == admin.ExecutionMetadata_RECOVERED {
		executionParameters.RecoveryExecution = request.GetSpec().GetMetadata().GetReferenceExecution()
	}

	executionCreatedAt := time.Now()
	acceptanceDelay := executionCreatedAt.Sub(requestedAt)

	// Request notification settings takes precedence over the launch plan settings.
	// If there is no notification in the request and DisableAll is not true, use the settings from the launch plan.
	var notificationsSettings []*admin.Notification
	if launchPlan.GetSpec().GetEntityMetadata() != nil {
		notificationsSettings = launchPlan.GetSpec().GetEntityMetadata().GetNotifications()
	}
	if requestSpec.GetNotifications() != nil && requestSpec.GetNotifications().GetNotifications() != nil &&
		len(requestSpec.GetNotifications().GetNotifications()) > 0 {
		notificationsSettings = requestSpec.GetNotifications().GetNotifications()
	} else if requestSpec.GetDisableAll() {
		notificationsSettings = make([]*admin.Notification, 0)
	}

	// Set the max parallelism based on the execution config (calculated based on multiple levels of settings)
	requestSpec.MaxParallelism = executionConfig.GetMaxParallelism()

	createExecModelInput := transformers.CreateExecutionModelInput{
		WorkflowExecutionID: workflowExecutionID,
		RequestSpec:         requestSpec,
		LaunchPlanID:        launchPlanModel.ID,
		WorkflowID:          launchPlanModel.WorkflowID,
		// The execution is not considered running until the propeller sends a specific event saying so.
		CreatedAt:             m._clock.Now(),
		Notifications:         notificationsSettings,
		WorkflowIdentifier:    workflow.GetId(),
		ParentNodeExecutionID: parentNodeExecutionID,
		SourceExecutionID:     sourceExecutionID,
		InputsURI:             inputsURI,
		UserInputsURI:         userInputsURI,
		SecurityContext:       executionConfig.GetSecurityContext(),
		LaunchEntity:          launchPlan.GetId().GetResourceType(),
		Namespace:             namespace,
	}

	// Check ConcurrencyPolicy for LaunchPlan, this is a means to limit the concurrency of a launch plan across versions
	// NOTE: There's a potential race condition here. Multiple concurrent requests
	// might pass this check before the database reflects the newly created executions,
	// potentially leading to more than 'Max' concurrent executions.
	if launchPlan.GetSpec().GetConcurrencyPolicy() != nil {
		if err := checkLaunchPlanConcurrency(ctx, launchPlan, m.db.ExecutionRepo(), m.systemMetrics); err != nil {
			return nil, nil, nil, err
		}
	}

	workflowExecutor := plugins.Get[workflowengineInterfaces.WorkflowExecutor](m.pluginRegistry, plugins.PluginIDWorkflowExecutor)
	execInfo, execErr := workflowExecutor.Execute(ctx, workflowengineInterfaces.ExecutionData{
		Namespace:                namespace,
		ExecutionID:              workflowExecutionID,
		ReferenceWorkflowName:    workflow.GetId().GetName(),
		ReferenceLaunchPlanName:  launchPlan.GetId().GetName(),
		WorkflowClosure:          workflow.GetClosure().GetCompiledWorkflow(),
		WorkflowClosureReference: storage.DataReference(workflowModel.RemoteClosureIdentifier),
		ExecutionParameters:      executionParameters,
		OffloadedInputsReference: inputsURI,
	})
	if execErr != nil {
		createExecModelInput.Error = execErr
		m.systemMetrics.PropellerFailures.Inc()
		logger.Infof(ctx, "failed to execute workflow %+v with execution id %+v and inputs %+v with err %v",
			request, workflowExecutionID, executionInputs, execErr)
	} else {
		m.systemMetrics.AcceptanceDelay.Observe(acceptanceDelay.Seconds())
		createExecModelInput.Cluster = execInfo.Cluster
	}

	executionModel, err := transformers.CreateExecutionModel(createExecModelInput)
	if err != nil {
		logger.Infof(ctx, "Failed to create execution model in transformer for id: [%+v] with err: %v",
			workflowExecutionID, err)
		return nil, nil, nil, err
	}

	executionTagModel, err := transformers.CreateExecutionTagModel(createExecModelInput)
	if err != nil {
		logger.Infof(ctx, "Failed to create execution tag model in transformer for id: [%+v] with err: %v",
			workflowExecutionID, err)
		return nil, nil, nil, err
	}

	return ctx, executionModel, executionTagModel, nil
}

// Inserts an execution model into the database store and emits platform metrics.
func (m *ExecutionManager) createExecutionModel(
	ctx context.Context, executionModel *models.Execution, executionTagModel []*models.ExecutionTag) (*core.WorkflowExecutionIdentifier, error) {
	workflowExecutionIdentifier := &core.WorkflowExecutionIdentifier{
		Project: executionModel.ExecutionKey.Project,
		Domain:  executionModel.ExecutionKey.Domain,
		Name:    executionModel.ExecutionKey.Name,
	}
	err := m.db.ExecutionRepo().Create(ctx, *executionModel, executionTagModel)
	if err != nil {
		logger.Debugf(ctx, "failed to save newly created execution [%+v] with id %+v to db with err %v",
			workflowExecutionIdentifier, workflowExecutionIdentifier, err)
		return nil, err
	}
	m.systemMetrics.ActiveExecutions.Inc()
	m.systemMetrics.ExecutionsCreated.Inc()
	m.systemMetrics.SpecSizeBytes.Observe(float64(len(executionModel.Spec)))
	m.systemMetrics.ClosureSizeBytes.Observe(float64(len(executionModel.Closure)))
	return workflowExecutionIdentifier, nil
}

func checkLaunchPlanConcurrency(ctx context.Context, launchPlan *admin.LaunchPlan, executionRepo repositoryInterfaces.ExecutionRepoInterface, metrics executionSystemMetrics) error {
	lpID := launchPlan.GetId()
	lpProject := lpID.GetProject()
	lpDomain := lpID.GetDomain()
	lpName := lpID.GetName()
	ctxForTimer := contextutils.WithProjectDomain(ctx, lpProject, lpDomain)
	ctxForTimer = contextutils.WithLaunchPlanID(ctxForTimer, lpName)
	defer metrics.ConcurrencyCheckDuration.Start(ctxForTimer).Stop()

	logger.Debugf(ctx, "checking concurrency limits for launch plan %v with policy %+v", lpID, launchPlan.GetSpec().GetConcurrencyPolicy())

	projectFilter, err := common.NewSingleValueFilter(common.Execution, common.Equal, "project", lpProject)
	if err != nil {
		return fmt.Errorf("failed to create project filter for concurrency check: %w", err)
	}
	domainFilter, err := common.NewSingleValueFilter(common.Execution, common.Equal, "domain", lpDomain)
	if err != nil {
		return fmt.Errorf("failed to create domain filter for concurrency check: %w", err)

	}

	lpNameFilter, err := common.NewSingleValueFilter(common.LaunchPlan, common.Equal, "name", lpName)
	if err != nil {
		return fmt.Errorf("failed to create launch plan name filter for concurrency check (JOIN): %w", err)
	}

	phaseFilter, err := common.NewRepeatedValueFilter(common.Execution, common.ValueIn, "phase", activeExecutionPhases)
	if err != nil {
		return fmt.Errorf("failed to create phase filter for concurrency check (JOIN): %w", err)
	}

	/*
		Count active executions for this launch plan, including an inner join with launch_plans table.
		This query joins 'executions' with 'launch_plans' to filter by launch plan name across all its versions,
		and then filters by active execution phases. This benefits from the existing indexes on the join keys (project, domain, name on launch_plans) and the execution phase. In plain SQL this is the expression:

		SELECT
			COUNT(executions.id)
		FROM
			executions
		INNER JOIN
			launch_plans ON executions.launch_plan_id = launch_plans.id
		WHERE
			executions.project = '${lpProject}'
			AND executions.domain = '${lpDomain}'
			AND launch_plans.name = '${lpName}'
			AND executions.phase IN (
				'QUEUED',
				'RUNNING',
			); -- Values from the activePhases slice
	*/

	count, err := executionRepo.Count(ctx, repositoryInterfaces.CountResourceInput{
		InlineFilters: []common.InlineFilter{projectFilter, domainFilter, lpNameFilter, phaseFilter},
		JoinTableEntities: map[common.Entity]bool{
			common.LaunchPlan: true,
		},
	})

	if err != nil {
		logger.Errorf(ctx, "failed to count active executions using JOIN for launch plan %s.%s.%s: %v", lpProject, lpDomain, lpName, err)
		// We still proceed to log the count as 0 and potentially hit the concurrency limit if Max is 0.
	}

	logger.Debugf(ctx, "found %d active executions for launch plan %s.%s.%s (any version)", count, lpProject, lpDomain, lpName)

	// Check against the policy limit
	if count >= int64(launchPlan.GetSpec().GetConcurrencyPolicy().GetMax()) {
		behavior := launchPlan.GetSpec().GetConcurrencyPolicy().GetBehavior()

		switch behavior {
		case admin.ConcurrencyLimitBehavior_CONCURRENCY_LIMIT_BEHAVIOR_SKIP:
			metrics.ConcurrencyLimitHits.WithLabelValues(lpProject, lpDomain, lpName).Inc()
			logger.Warningf(ctx, "skipping execution creation for launch plan %v due to concurrency limit", lpID)
			return errors.NewFlyteAdminErrorf(
				codes.AlreadyExists,
				"concurrency limit (%d) reached for launch plan %s; skipping execution",
				launchPlan.GetSpec().GetConcurrencyPolicy().GetMax(), lpName)

		case admin.ConcurrencyLimitBehavior_CONCURRENCY_LIMIT_BEHAVIOR_UNSPECIFIED:
			// fall through
		default:
			return fmt.Errorf("unsupported concurrency-limit behavior: %v", behavior)
		}
	}
	return nil
}

func (m *ExecutionManager) CreateExecution(
	ctx context.Context, request *admin.ExecutionCreateRequest, requestedAt time.Time) (
	*admin.ExecutionCreateResponse, error) {

	// Prior to flyteidl v0.15.0, Inputs was held in ExecutionSpec. Ensure older clients continue to work.
	if request.GetInputs() == nil || len(request.GetInputs().GetLiterals()) == 0 {
		request.Inputs = request.GetSpec().GetInputs()
	}
	var executionModel *models.Execution
	var executionTagModel []*models.ExecutionTag
	var err error
	ctx, executionModel, executionTagModel, err = m.launchExecutionAndPrepareModel(ctx, request, requestedAt)
	if err != nil {
		return nil, err
	}
	workflowExecutionIdentifier, err := m.createExecutionModel(ctx, executionModel, executionTagModel)
	if err != nil {
		return nil, err
	}
	return &admin.ExecutionCreateResponse{
		Id: workflowExecutionIdentifier,
	}, nil
}

func (m *ExecutionManager) RelaunchExecution(
	ctx context.Context, request *admin.ExecutionRelaunchRequest, requestedAt time.Time) (
	*admin.ExecutionCreateResponse, error) {
	existingExecutionModel, err := util.GetExecutionModel(ctx, m.db, request.GetId())
	if err != nil {
		logger.Debugf(ctx, "Failed to get execution model for request [%+v] with err %v", request, err)
		return nil, err
	}
	existingExecution, err := transformers.FromExecutionModel(ctx, *existingExecutionModel, transformers.DefaultExecutionTransformerOptions)
	if err != nil {
		return nil, err
	}

	executionSpec := existingExecution.GetSpec()
	if executionSpec.GetMetadata() == nil {
		executionSpec.Metadata = &admin.ExecutionMetadata{}
	}
	var inputs *core.LiteralMap
	if len(existingExecutionModel.UserInputsURI) > 0 {
		inputs = &core.LiteralMap{}
		if err := m.storageClient.ReadProtobuf(ctx, existingExecutionModel.UserInputsURI, inputs); err != nil {
			return nil, err
		}
	} else {
		// For old data, inputs are held in the spec
		var spec admin.ExecutionSpec
		err = proto.Unmarshal(existingExecutionModel.Spec, &spec)
		if err != nil {
			return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to unmarshal spec")
		}
		inputs = spec.GetInputs()
	}
	executionSpec.Metadata.Mode = admin.ExecutionMetadata_RELAUNCH
	executionSpec.Metadata.ReferenceExecution = existingExecution.GetId()
	executionSpec.OverwriteCache = request.GetOverwriteCache()
	var executionModel *models.Execution
	var executionTagModel []*models.ExecutionTag
	ctx, executionModel, executionTagModel, err = m.launchExecutionAndPrepareModel(ctx, &admin.ExecutionCreateRequest{
		Project: request.GetId().GetProject(),
		Domain:  request.GetId().GetDomain(),
		Name:    request.GetName(),
		Spec:    executionSpec,
		Inputs:  inputs,
	}, requestedAt)
	if err != nil {
		return nil, err
	}
	executionModel.SourceExecutionID = existingExecutionModel.ID
	workflowExecutionIdentifier, err := m.createExecutionModel(ctx, executionModel, executionTagModel)
	if err != nil {
		return nil, err
	}
	logger.Debugf(ctx, "Successfully relaunched [%+v] as [%+v]", request.GetId(), workflowExecutionIdentifier)
	return &admin.ExecutionCreateResponse{
		Id: workflowExecutionIdentifier,
	}, nil
}

func (m *ExecutionManager) RecoverExecution(
	ctx context.Context, request *admin.ExecutionRecoverRequest, requestedAt time.Time) (
	*admin.ExecutionCreateResponse, error) {
	existingExecutionModel, err := util.GetExecutionModel(ctx, m.db, request.GetId())
	if err != nil {
		logger.Debugf(ctx, "Failed to get execution model for request [%+v] with err %v", request, err)
		return nil, err
	}
	existingExecution, err := transformers.FromExecutionModel(ctx, *existingExecutionModel, transformers.DefaultExecutionTransformerOptions)
	if err != nil {
		return nil, err
	}

	executionSpec := existingExecution.GetSpec()
	if executionSpec.GetMetadata() == nil {
		executionSpec.Metadata = &admin.ExecutionMetadata{}
	}
	var inputs *core.LiteralMap
	if len(existingExecutionModel.UserInputsURI) > 0 {
		inputs = &core.LiteralMap{}
		if err := m.storageClient.ReadProtobuf(ctx, existingExecutionModel.UserInputsURI, inputs); err != nil {
			return nil, err
		}
	}
	if request.GetMetadata() != nil {
		executionSpec.Metadata.ParentNodeExecution = request.GetMetadata().GetParentNodeExecution()
	}
	executionSpec.Metadata.Mode = admin.ExecutionMetadata_RECOVERED
	executionSpec.Metadata.ReferenceExecution = existingExecution.GetId()
	var executionModel *models.Execution
	var executionTagModel []*models.ExecutionTag
	ctx, executionModel, executionTagModel, err = m.launchExecutionAndPrepareModel(ctx, &admin.ExecutionCreateRequest{
		Project: request.GetId().GetProject(),
		Domain:  request.GetId().GetDomain(),
		Name:    request.GetName(),
		Spec:    executionSpec,
		Inputs:  inputs,
	}, requestedAt)
	if err != nil {
		return nil, err
	}
	executionModel.SourceExecutionID = existingExecutionModel.ID
	workflowExecutionIdentifier, err := m.createExecutionModel(ctx, executionModel, executionTagModel)
	if err != nil {
		return nil, err
	}
	logger.Infof(ctx, "Successfully recovered [%+v] as [%+v]", request.GetId(), workflowExecutionIdentifier)
	return &admin.ExecutionCreateResponse{
		Id: workflowExecutionIdentifier,
	}, nil
}

func (m *ExecutionManager) emitScheduledWorkflowMetrics(
	ctx context.Context, executionModel *models.Execution, runningEventTimeProto *timestamp.Timestamp) {
	if executionModel == nil || runningEventTimeProto == nil {
		logger.Warningf(context.Background(),
			"tried to calculate scheduled workflow execution stats with a nil execution or event time")
		return
	}
	// Find the reference launch plan to get the kickoff time argument
	execution, err := transformers.FromExecutionModel(ctx, *executionModel, transformers.DefaultExecutionTransformerOptions)
	if err != nil {
		logger.Warningf(context.Background(),
			"failed to transform execution model when emitting scheduled workflow execution stats with for "+
				"[%s/%s/%s]", executionModel.Project, executionModel.Domain, executionModel.Name)
		return
	}
	launchPlan, err := util.GetLaunchPlan(context.Background(), m.db, execution.GetSpec().GetLaunchPlan())
	if err != nil {
		logger.Warningf(context.Background(),
			"failed to find launch plan when emitting scheduled workflow execution stats with for "+
				"execution: [%+v] and launch plan [%+v]", execution.GetId(), execution.GetSpec().GetLaunchPlan())
		return
	}

	if launchPlan.GetSpec().GetEntityMetadata() == nil ||
		launchPlan.GetSpec().GetEntityMetadata().GetSchedule() == nil ||
		launchPlan.GetSpec().GetEntityMetadata().GetSchedule().GetKickoffTimeInputArg() == "" {
		// Kickoff time arguments aren't always required for scheduled workflows.
		logger.Debugf(context.Background(), "no kickoff time to report for scheduled workflow execution [%+v]",
			execution.GetId())
		return
	}

	var inputs core.LiteralMap
	err = m.storageClient.ReadProtobuf(ctx, executionModel.InputsURI, &inputs)
	if err != nil {
		logger.Errorf(ctx, "Failed to find inputs for emitting schedule delay event from uri: [%v]", executionModel.InputsURI)
		return
	}
	scheduledKickoffTimeProto := inputs.GetLiterals()[launchPlan.GetSpec().GetEntityMetadata().GetSchedule().GetKickoffTimeInputArg()]
	if scheduledKickoffTimeProto == nil || scheduledKickoffTimeProto.GetScalar() == nil ||
		scheduledKickoffTimeProto.GetScalar().GetPrimitive() == nil ||
		scheduledKickoffTimeProto.GetScalar().GetPrimitive().GetDatetime() == nil {
		logger.Warningf(context.Background(),
			"failed to find scheduled kickoff time datetime value for scheduled workflow execution [%+v] "+
				"although one was expected", execution.GetId())
		return
	}
	scheduledKickoffTime, err := ptypes.Timestamp(scheduledKickoffTimeProto.GetScalar().GetPrimitive().GetDatetime())
	if err != nil {
		// Timestamps are serialized by flyteadmin and should always be valid
		return
	}
	runningEventTime, err := ptypes.Timestamp(runningEventTimeProto)
	if err != nil {
		// Timestamps are always sent from propeller and should always be valid
		return
	}

	projectKey := execution.GetId().GetProject()
	val, ok := m.userMetrics.ScheduledExecutionDelays.Load(projectKey)
	if !ok {
		val = &sync.Map{}
		m.userMetrics.ScheduledExecutionDelays.Store(projectKey, val)
	}

	domainCounterMap := val.(*sync.Map)

	domainKey := execution.GetId().GetDomain()
	watchVal, ok := domainCounterMap.Load(domainKey)
	if !ok {
		newWatch, err := m.systemMetrics.Scope.NewSubScope(execution.GetId().GetProject()).NewSubScope(domainKey).NewStopWatch(
			"scheduled_execution_delay",
			"delay between scheduled execution time and time execution was observed running",
			time.Nanosecond)
		if err != nil {
			// Could be related to a concurrent exception.
			logger.Debugf(context.Background(),
				"failed to emit scheduled workflow execution delay stat, couldn't find or create counter")
			return
		}
		watchVal = &newWatch
		domainCounterMap.Store(domainKey, watchVal)
	}

	watch := watchVal.(*promutils.StopWatch)
	watch.Observe(scheduledKickoffTime, runningEventTime)
}

func (m *ExecutionManager) emitOverallWorkflowExecutionTime(
	executionModel *models.Execution, terminalEventTimeProto *timestamp.Timestamp) {
	if executionModel == nil || terminalEventTimeProto == nil {
		logger.Warningf(context.Background(),
			"tried to calculate scheduled workflow execution stats with a nil execution or event time")
		return
	}

	projectKey := executionModel.Project
	val, ok := m.userMetrics.WorkflowExecutionDurations.Load(projectKey)
	if !ok {
		val = &sync.Map{}
		m.userMetrics.WorkflowExecutionDurations.Store(projectKey, val)
	}

	domainCounterMap := val.(*sync.Map)

	domainKey := executionModel.Domain
	watchVal, ok := domainCounterMap.Load(domainKey)
	if !ok {
		newWatch, err := m.systemMetrics.Scope.NewSubScope(executionModel.Project).NewSubScope(domainKey).NewStopWatch(
			"workflow_execution_duration",
			"overall time from when when a workflow create request was sent to k8s to the workflow terminating",
			time.Nanosecond)
		if err != nil {
			// Could be related to a concurrent exception.
			logger.Debugf(context.Background(),
				"failed to emit workflow execution duration stat, couldn't find or create counter")
			return
		}
		watchVal = &newWatch
		domainCounterMap.Store(domainKey, watchVal)
	}

	watch := watchVal.(*promutils.StopWatch)

	terminalEventTime, err := ptypes.Timestamp(terminalEventTimeProto)
	if err != nil {
		// Timestamps are always sent from propeller and should always be valid
		return
	}

	if executionModel.ExecutionCreatedAt == nil {
		logger.Warningf(context.Background(), "found execution with nil ExecutionCreatedAt: [%s/%s/%s]",
			executionModel.Project, executionModel.Domain, executionModel.Name)
		return
	}
	watch.Observe(*executionModel.ExecutionCreatedAt, terminalEventTime)
}

func (m *ExecutionManager) CreateWorkflowEvent(ctx context.Context, request *admin.WorkflowExecutionEventRequest) (
	*admin.WorkflowExecutionEventResponse, error) {
	err := validation.ValidateCreateWorkflowEventRequest(request, m.config.ApplicationConfiguration().GetRemoteDataConfig().MaxSizeInBytes)
	if err != nil {
		logger.Debugf(ctx, "received invalid CreateWorkflowEventRequest [%s]: %v", request.GetRequestId(), err)
		return nil, err
	}
	ctx = getExecutionContext(ctx, request.GetEvent().GetExecutionId())
	logger.Debugf(ctx, "Received workflow execution event for [%+v] transitioning to phase [%v]",
		request.GetEvent().GetExecutionId(), request.GetEvent().GetPhase())

	executionModel, err := util.GetExecutionModel(ctx, m.db, request.GetEvent().GetExecutionId())
	if err != nil {
		logger.Debugf(ctx, "failed to find execution [%+v] for recorded event [%s]: %v",
			request.GetEvent().GetExecutionId(), request.GetRequestId(), err)
		return nil, err
	}

	wfExecPhase := core.WorkflowExecution_Phase(core.WorkflowExecution_Phase_value[executionModel.Phase])
	// Subsequent queued events announcing a cluster reassignment are permitted.
	if request.GetEvent().GetPhase() != core.WorkflowExecution_QUEUED {
		if wfExecPhase == request.GetEvent().GetPhase() {
			logger.Debugf(ctx, "This phase %s was already recorded for workflow execution %v",
				wfExecPhase.String(), request.GetEvent().GetExecutionId())
			return nil, errors.NewFlyteAdminErrorf(codes.AlreadyExists,
				"This phase %s was already recorded for workflow execution %v",
				wfExecPhase.String(), request.GetEvent().GetExecutionId())
		} else if err := validation.ValidateCluster(ctx, executionModel.Cluster, request.GetEvent().GetProducerId()); err != nil {
			// Only perform event cluster validation **after** an execution has moved on from QUEUED.
			return nil, err
		}
	}

	if common.IsExecutionTerminal(wfExecPhase) {
		// Cannot go backwards in time from a terminal state to anything else
		curPhase := wfExecPhase.String()
		errorMsg := fmt.Sprintf("Invalid phase change from %s to %s for workflow execution %v", curPhase, request.GetEvent().GetPhase().String(), request.GetEvent().GetExecutionId())
		return nil, errors.NewAlreadyInTerminalStateError(ctx, errorMsg, curPhase)
	} else if wfExecPhase == core.WorkflowExecution_RUNNING && request.GetEvent().GetPhase() == core.WorkflowExecution_QUEUED {
		// Cannot go back in time from RUNNING -> QUEUED
		return nil, errors.NewFlyteAdminErrorf(codes.FailedPrecondition,
			"Cannot go from %s to %s for workflow execution %v",
			wfExecPhase.String(), request.GetEvent().GetPhase().String(), request.GetEvent().GetExecutionId())
	} else if wfExecPhase == core.WorkflowExecution_ABORTING && !common.IsExecutionTerminal(request.GetEvent().GetPhase()) {
		return nil, errors.NewFlyteAdminErrorf(codes.FailedPrecondition,
			"Invalid phase change from aborting to %s for workflow execution %v", request.GetEvent().GetPhase().String(), request.GetEvent().GetExecutionId())
	}

	err = transformers.UpdateExecutionModelState(ctx, executionModel, request, m.config.ApplicationConfiguration().GetRemoteDataConfig().InlineEventDataPolicy, m.storageClient)
	if err != nil {
		logger.Debugf(ctx, "failed to transform updated workflow execution model [%+v] after receiving event with err: %v",
			request.GetEvent().GetExecutionId(), err)
		return nil, err
	}
	err = m.db.ExecutionRepo().Update(ctx, *executionModel)
	if err != nil {
		logger.Debugf(ctx, "Failed to update execution with CreateWorkflowEvent [%+v] with err %v",
			request, err)
		return nil, err
	}
	m.dbEventWriter.Write(request)

	if request.GetEvent().GetPhase() == core.WorkflowExecution_RUNNING {
		// Workflow executions are created in state "UNDEFINED". All the time up until a RUNNING event is received is
		// considered system-induced delay.
		if executionModel.Mode == int32(admin.ExecutionMetadata_SCHEDULED) {
			go m.emitScheduledWorkflowMetrics(ctx, executionModel, request.GetEvent().GetOccurredAt())
		}
	} else if common.IsExecutionTerminal(request.GetEvent().GetPhase()) {
		if request.GetEvent().GetPhase() == core.WorkflowExecution_FAILED {
			// request.Event is expected to be of type WorkflowExecutionEvent_Error when workflow fails.
			// if not, log the error and continue
			if err := request.GetEvent().GetError(); err != nil {
				ctx = context.WithValue(ctx, common.ErrorKindKey, err.GetKind().String())
			} else {
				logger.Warning(ctx, "Failed to parse error for FAILED request [%+v]", request)
			}
		}

		m.systemMetrics.ActiveExecutions.Dec()
		m.systemMetrics.ExecutionsTerminated.Inc(contextutils.WithPhase(ctx, request.GetEvent().GetPhase().String()))
		go m.emitOverallWorkflowExecutionTime(executionModel, request.GetEvent().GetOccurredAt())
		if request.GetEvent().GetOutputData() != nil {
			m.userMetrics.WorkflowExecutionOutputBytes.Observe(float64(proto.Size(request.GetEvent().GetOutputData())))
		}

		err = m.publishNotifications(ctx, request, *executionModel)
		if err != nil {
			// The only errors that publishNotifications will forward are those related
			// to unexpected data and transformation errors.
			logger.Debugf(ctx, "failed to publish notifications for CreateWorkflowEvent [%+v] due to err: %v",
				request, err)
			return nil, err
		}
	}

	if err := m.eventPublisher.Publish(ctx, proto.MessageName(request), request); err != nil {
		m.systemMetrics.PublishEventError.Inc()
		logger.Infof(ctx, "error publishing event [%+v] with err: [%v]", request.GetRequestId(), err)
	}

	go func() {
		ceCtx := context.TODO()
		if err := m.cloudEventPublisher.Publish(ceCtx, proto.MessageName(request), request); err != nil {
			m.systemMetrics.PublishEventError.Inc()
			logger.Infof(ctx, "error publishing cloud event [%+v] with err: [%v]", request.GetRequestId(), err)
		}
	}()

	return &admin.WorkflowExecutionEventResponse{}, nil
}

func (m *ExecutionManager) GetExecution(
	ctx context.Context, request *admin.WorkflowExecutionGetRequest) (*admin.Execution, error) {
	if err := validation.ValidateWorkflowExecutionIdentifier(request.GetId()); err != nil {
		logger.Debugf(ctx, "GetExecution request [%+v] failed validation with err: %v", request, err)
		return nil, err
	}
	ctx = getExecutionContext(ctx, request.GetId())
	executionModel, err := util.GetExecutionModel(ctx, m.db, request.GetId())
	if err != nil {
		logger.Debugf(ctx, "Failed to get execution model for request [%+v] with err: %v", request, err)
		return nil, err
	}
	namespace := common.GetNamespaceName(
		m.config.NamespaceMappingConfiguration().GetNamespaceTemplate(), request.GetId().GetProject(), request.GetId().GetDomain())
	execution, transformerErr := transformers.FromExecutionModel(ctx, *executionModel, &transformers.ExecutionTransformerOptions{
		DefaultNamespace: namespace,
	})
	if transformerErr != nil {
		logger.Debugf(ctx, "Failed to transform execution model [%+v] to proto object with err: %v", request.GetId(),
			transformerErr)
		return nil, transformerErr
	}

	return execution, nil
}

func (m *ExecutionManager) UpdateExecution(ctx context.Context, request *admin.ExecutionUpdateRequest,
	requestedAt time.Time) (*admin.ExecutionUpdateResponse, error) {
	if err := validation.ValidateWorkflowExecutionIdentifier(request.GetId()); err != nil {
		logger.Debugf(ctx, "UpdateExecution request [%+v] failed validation with err: %v", request, err)
		return nil, err
	}
	ctx = getExecutionContext(ctx, request.GetId())
	executionModel, err := util.GetExecutionModel(ctx, m.db, request.GetId())
	if err != nil {
		logger.Debugf(ctx, "Failed to get execution model for request [%+v] with err: %v", request, err)
		return nil, err
	}

	if err = transformers.UpdateExecutionModelStateChangeDetails(executionModel, request.GetState(), requestedAt,
		getUser(ctx)); err != nil {
		return nil, err
	}

	if err := m.db.ExecutionRepo().Update(ctx, *executionModel); err != nil {
		return nil, err
	}

	return &admin.ExecutionUpdateResponse{}, nil
}

func (m *ExecutionManager) GetExecutionData(
	ctx context.Context, request *admin.WorkflowExecutionGetDataRequest) (*admin.WorkflowExecutionGetDataResponse, error) {
	ctx = getExecutionContext(ctx, request.GetId())
	executionModel, err := util.GetExecutionModel(ctx, m.db, request.GetId())
	if err != nil {
		logger.Debugf(ctx, "Failed to get execution model for request [%+v] with err: %v", request, err)
		return nil, err
	}
	execution, err := transformers.FromExecutionModel(ctx, *executionModel, transformers.DefaultExecutionTransformerOptions)
	if err != nil {
		logger.Debugf(ctx, "Failed to transform execution model [%+v] to proto object with err: %v", request.GetId(), err)
		return nil, err
	}
	// Prior to flyteidl v0.15.0, Inputs were held in ExecutionClosure and were not offloaded. Ensure we can return the inputs as expected.
	if len(executionModel.InputsURI) == 0 {
		closure := &admin.ExecutionClosure{}
		// We must not use the FromExecutionModel method because it empties deprecated fields.
		if err := proto.Unmarshal(executionModel.Closure, closure); err != nil {
			return nil, err
		}
		newInputsURI, err := common.OffloadLiteralMap(ctx, m.storageClient, closure.GetComputedInputs(), request.GetId().GetProject(), request.GetId().GetDomain(), request.GetId().GetName(), shared.Inputs)
		if err != nil {
			return nil, err
		}
		// Update model so as not to offload again.
		executionModel.InputsURI = newInputsURI
		if err := m.db.ExecutionRepo().Update(ctx, *executionModel); err != nil {
			return nil, err
		}
	}

	var inputs *core.LiteralMap
	var inputURLBlob *admin.UrlBlob
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		var err error
		inputs, inputURLBlob, err = util.GetInputs(groupCtx, m.urlData, m.config.ApplicationConfiguration().GetRemoteDataConfig(),
			m.storageClient, executionModel.InputsURI.String())
		return err
	})

	var outputs *core.LiteralMap
	var outputURLBlob *admin.UrlBlob
	group.Go(func() error {
		var err error
		outputs, outputURLBlob, err = util.GetOutputs(groupCtx, m.urlData, m.config.ApplicationConfiguration().GetRemoteDataConfig(),
			m.storageClient, util.ToExecutionClosureInterface(execution.GetClosure()))
		return err
	})

	err = group.Wait()
	if err != nil {
		return nil, err
	}

	response := &admin.WorkflowExecutionGetDataResponse{
		Inputs:      inputURLBlob,
		Outputs:     outputURLBlob,
		FullInputs:  inputs,
		FullOutputs: outputs,
	}

	m.userMetrics.WorkflowExecutionInputBytes.Observe(float64(response.GetInputs().GetBytes()))
	if response.GetOutputs().GetBytes() > 0 {
		m.userMetrics.WorkflowExecutionOutputBytes.Observe(float64(response.GetOutputs().GetBytes()))
	} else if response.GetFullOutputs() != nil {
		m.userMetrics.WorkflowExecutionOutputBytes.Observe(float64(proto.Size(response.GetFullOutputs())))
	}
	return response, nil
}

func (m *ExecutionManager) ListExecutions(
	ctx context.Context, request *admin.ResourceListRequest) (*admin.ExecutionList, error) {
	// Check required fields
	if err := validation.ValidateResourceListRequest(request); err != nil {
		logger.Debugf(ctx, "ListExecutions request [%+v] failed validation with err: %v", request, err)
		return nil, err
	}
	ctx = contextutils.WithProjectDomain(ctx, request.GetId().GetProject(), request.GetId().GetDomain())
	filters, err := util.GetDbFilters(util.FilterSpec{
		Project:        request.GetId().GetProject(),
		Domain:         request.GetId().GetDomain(),
		Name:           request.GetId().GetName(), // Optional, may be empty.
		RequestFilters: request.GetFilters(),
	}, common.Execution)
	if err != nil {
		return nil, err
	}

	sortParameter, err := common.NewSortParameter(request.GetSortBy(), models.ExecutionColumns)
	if err != nil {
		return nil, err
	}

	offset, err := validation.ValidateToken(request.GetToken())
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "invalid pagination token %s for ListExecutions",
			request.GetToken())
	}
	joinTableEntities := make(map[common.Entity]bool)
	for _, filter := range filters {
		joinTableEntities[filter.GetEntity()] = true
	}

	// Check if state filter exists and if not then add filter to fetch only ACTIVE executions
	if filters, err = addStateFilter(filters); err != nil {
		return nil, err
	}

	listExecutionsInput := repositoryInterfaces.ListResourceInput{
		Limit:             int(request.GetLimit()),
		Offset:            offset,
		InlineFilters:     filters,
		SortParameter:     sortParameter,
		JoinTableEntities: joinTableEntities,
	}
	output, err := m.db.ExecutionRepo().List(ctx, listExecutionsInput)
	if err != nil {
		logger.Debugf(ctx, "Failed to list executions using input [%+v] with err %v", listExecutionsInput, err)
		return nil, err
	}
	executionList, err := transformers.FromExecutionModels(ctx, output.Executions, transformers.ListExecutionTransformerOptions)
	if err != nil {
		logger.Errorf(ctx,
			"Failed to transform execution models [%+v] with err: %v", output.Executions, err)
		return nil, err
	}
	// TODO: TO BE DELETED
	// Clear deprecated fields during migration phase. Once migration is complete, these will be cleared in the database.
	// Thus this will be redundant
	for _, execution := range executionList {
		execution.Spec.Inputs = nil
		execution.Closure.ComputedInputs = nil
	}

	// END TO BE DELETED
	var token string
	if len(executionList) == int(request.GetLimit()) {
		token = strconv.Itoa(offset + len(executionList))
	}
	return &admin.ExecutionList{
		Executions: executionList,
		Token:      token,
	}, nil
}

// publishNotifications will only forward major errors because the assumption made is all of the objects
// that are being manipulated have already been validated/manipulated by Flyte itself.
// Note: This method should be refactored somewhere else once the interaction with pushing to SNS.
func (m *ExecutionManager) publishNotifications(ctx context.Context, request *admin.WorkflowExecutionEventRequest,
	execution models.Execution) error {
	// Notifications are stored in the Spec object of an admin.Execution object.
	adminExecution, err := transformers.FromExecutionModel(ctx, execution, transformers.DefaultExecutionTransformerOptions)
	if err != nil {
		// This shouldn't happen because execution manager marshaled the data into models.Execution.
		m.systemMetrics.TransformerError.Inc()
		return errors.NewFlyteAdminErrorf(codes.Internal, "Failed to transform execution [%+v] with err: %v", request.GetEvent().GetExecutionId(), err)
	}
	var notificationsList = adminExecution.GetClosure().GetNotifications()
	logger.Debugf(ctx, "publishing notifications for execution [%+v] in state [%+v] for notifications [%+v]",
		request.GetEvent().GetExecutionId(), request.GetEvent().GetPhase(), notificationsList)
	for _, notification := range notificationsList {
		// Check if the notification phase matches the current one.
		var matchPhase = false
		for _, phase := range notification.GetPhases() {
			if phase == request.GetEvent().GetPhase() {
				matchPhase = true
			}
		}

		// The current phase doesn't match; no notifications will be sent for the current notification option.
		if !matchPhase {
			continue
		}

		// Currently all three supported notifications use email underneath to send the notification.
		// Convert Slack and PagerDuty into an EmailNotification type.
		emailNotification := &admin.EmailNotification{}
		if notification.GetEmail() != nil {
			emailNotification.RecipientsEmail = notification.GetEmail().GetRecipientsEmail()
		} else if notification.GetPagerDuty() != nil {
			emailNotification.RecipientsEmail = notification.GetPagerDuty().GetRecipientsEmail()
		} else if notification.GetSlack() != nil {
			emailNotification.RecipientsEmail = notification.GetSlack().GetRecipientsEmail()
		} else {
			logger.Debugf(ctx, "failed to publish notification, encountered unrecognized type: %v", notification.GetType())
			m.systemMetrics.UnexpectedDataError.Inc()
			// Unsupported notification types should have been caught when the launch plan was being created.
			return errors.NewFlyteAdminErrorf(codes.Internal, "Unsupported notification type [%v] for execution [%+v]",
				notification.GetType(), request.GetEvent().GetExecutionId())
		}

		// Convert the email Notification into an email message to be published.
		// Currently there are no possible errors while creating an email message.
		// Once customizable content is specified, errors are possible.
		email := notifications.ToEmailMessageFromWorkflowExecutionEvent(
			*m.config.ApplicationConfiguration().GetNotificationsConfig(), emailNotification, request, adminExecution)
		// Errors seen while publishing a message are considered non-fatal to the method and will not result
		// in the method returning an error.
		if err = m.notificationClient.Publish(ctx, proto.MessageName(emailNotification), email); err != nil {
			m.systemMetrics.PublishNotificationError.Inc()
			logger.Infof(ctx, "error publishing email notification [%+v] with err: [%v]", notification, err)
		}
	}
	return nil
}

func (m *ExecutionManager) TerminateExecution(
	ctx context.Context, request *admin.ExecutionTerminateRequest) (*admin.ExecutionTerminateResponse, error) {
	if err := validation.ValidateWorkflowExecutionIdentifier(request.GetId()); err != nil {
		logger.Debugf(ctx, "received terminate execution request: %v with invalid identifier: %v", request, err)
		return nil, err
	}
	ctx = getExecutionContext(ctx, request.GetId())
	// Save the abort reason (best effort)
	executionModel, err := m.db.ExecutionRepo().Get(ctx, repositoryInterfaces.Identifier{
		Project: request.GetId().GetProject(),
		Domain:  request.GetId().GetDomain(),
		Name:    request.GetId().GetName(),
	})
	if err != nil {
		logger.Infof(ctx, "couldn't find execution [%+v] to save termination cause", request.GetId())
		return nil, err
	}

	if common.IsExecutionTerminal(core.WorkflowExecution_Phase(core.WorkflowExecution_Phase_value[executionModel.Phase])) {
		return nil, errors.NewAlreadyInTerminalStateError(ctx, "Cannot abort an already terminated workflow execution", executionModel.Phase)
	}

	err = transformers.SetExecutionAborting(&executionModel, request.GetCause(), getUser(ctx))
	if err != nil {
		logger.Debugf(ctx, "failed to add abort metadata for execution [%+v] with err: %v", request.GetId(), err)
		return nil, err
	}

	err = m.db.ExecutionRepo().Update(ctx, executionModel)
	if err != nil {
		logger.Debugf(ctx, "failed to save abort cause for terminated execution: %+v with err: %v", request.GetId(), err)
		return nil, err
	}

	workflowExecutor := plugins.Get[workflowengineInterfaces.WorkflowExecutor](m.pluginRegistry, plugins.PluginIDWorkflowExecutor)
	err = workflowExecutor.Abort(ctx, workflowengineInterfaces.AbortData{
		Namespace: common.GetNamespaceName(
			m.config.NamespaceMappingConfiguration().GetNamespaceTemplate(), request.GetId().GetProject(), request.GetId().GetDomain()),

		ExecutionID: request.GetId(),
		Cluster:     executionModel.Cluster,
	})
	if err != nil {
		m.systemMetrics.TerminateExecutionFailures.Inc()
		return nil, err
	}
	return &admin.ExecutionTerminateResponse{}, nil
}

func newExecutionSystemMetrics(scope promutils.Scope) executionSystemMetrics {
	return executionSystemMetrics{
		Scope: scope,
		ActiveExecutions: scope.MustNewGauge("active_executions",
			"overall count of active workflow executions"),
		ExecutionsCreated: scope.MustNewCounter("executions_created",
			"overall count of successfully completed CreateExecutionRequests"),
		ExecutionsTerminated: labeled.NewCounter("executions_terminated",
			"overall count of terminated workflow executions", scope),
		ExecutionEventsCreated: scope.MustNewCounter("execution_events_created",
			"overall count of successfully completed WorkflowExecutionEventRequest"),
		PropellerFailures: scope.MustNewCounter("propeller_failures",
			"propeller failures in creating workflow executions"),
		TransformerError: scope.MustNewCounter("transformer_error",
			"overall count of errors when transforming models and messages"),
		UnexpectedDataError: scope.MustNewCounter("unexpected_data_error",
			"overall count of unexpected data for previously validated objects"),
		PublishNotificationError: scope.MustNewCounter("publish_error",
			"overall count of publish notification errors when invoking publish()"),
		SpecSizeBytes:    scope.MustNewSummary("spec_size_bytes", "size in bytes of serialized execution spec"),
		ClosureSizeBytes: scope.MustNewSummary("closure_size_bytes", "size in bytes of serialized execution closure"),
		AcceptanceDelay: scope.MustNewSummary("acceptance_delay",
			"delay in seconds from when an execution was requested to be created and when it actually was"),
		PublishEventError: scope.MustNewCounter("publish_event_error",
			"overall count of publish event errors when invoking publish()"),
		TerminateExecutionFailures: scope.MustNewCounter("execution_termination_failure",
			"count of failed workflow executions terminations"),
		ConcurrencyCheckDuration: labeled.NewStopWatch("concurrency_check_duration",
			"time spent checking concurrency limits for launch plans", time.Millisecond, scope),
		ConcurrencyLimitHits: scope.MustNewCounterVec("concurrency_limit_hits",
			"count of times concurrency limits were hit for launch plans", "project", "domain", "name"),
	}
}

func NewExecutionManager(db repositoryInterfaces.Repository, pluginRegistry *plugins.Registry, config runtimeInterfaces.Configuration,
	storageClient *storage.DataStore, systemScope promutils.Scope, userScope promutils.Scope,
	publisher notificationInterfaces.Publisher, urlData dataInterfaces.RemoteURLInterface,
	workflowManager interfaces.WorkflowInterface, namedEntityManager interfaces.NamedEntityInterface,
	eventPublisher notificationInterfaces.Publisher, cloudEventPublisher cloudeventInterfaces.Publisher,
	eventWriter eventWriter.WorkflowExecutionEventWriter) interfaces.ExecutionInterface {

	queueAllocator := executions.NewQueueAllocator(config, db)
	systemMetrics := newExecutionSystemMetrics(systemScope)

	userMetrics := executionUserMetrics{
		Scope:                      userScope,
		ScheduledExecutionDelays:   sync.Map{},
		WorkflowExecutionDurations: sync.Map{},
		WorkflowExecutionInputBytes: userScope.MustNewSummary("input_size_bytes",
			"size in bytes of serialized execution inputs"),
		WorkflowExecutionOutputBytes: userScope.MustNewSummary("output_size_bytes",
			"size in bytes of serialized execution outputs"),
	}

	resourceManager := resources.NewResourceManager(db, config.ApplicationConfiguration())
	return &ExecutionManager{
		db:                        db,
		config:                    config,
		storageClient:             storageClient,
		queueAllocator:            queueAllocator,
		_clock:                    clock.New(),
		systemMetrics:             systemMetrics,
		userMetrics:               &userMetrics,
		notificationClient:        publisher,
		urlData:                   urlData,
		workflowManager:           workflowManager,
		namedEntityManager:        namedEntityManager,
		resourceManager:           resourceManager,
		qualityOfServiceAllocator: executions.NewQualityOfServiceAllocator(config, resourceManager),
		eventPublisher:            eventPublisher,
		cloudEventPublisher:       cloudEventPublisher,
		dbEventWriter:             eventWriter,
		pluginRegistry:            pluginRegistry,
	}
}

// Adds project labels with higher precedence to workflow labels. Project labels are ignored if a corresponding label is set on the workflow.
func (m *ExecutionManager) addProjectLabels(ctx context.Context, projectName string, initialLabels map[string]string) (map[string]string, error) {
	project, err := m.db.ProjectRepo().Get(ctx, projectName)
	if err != nil {
		logger.Errorf(ctx, "Failed to get project for [%+v] with error: %v", project, err)
		return nil, err
	}
	// passing nil domain as not needed to retrieve labels
	projectLabels := transformers.FromProjectModel(project, nil).GetLabels().GetValues()

	if initialLabels == nil {
		initialLabels = make(map[string]string)
	}

	for k, v := range projectLabels {
		if _, ok := initialLabels[k]; !ok {
			initialLabels[k] = v
		}
	}
	return initialLabels, nil
}

// addIdentityAnnotations automatically injects identity information (user or app) as annotations when enabled in config.
// This allows tracking which identity submitted each workflow execution and enables identity-based authorization.
func (m *ExecutionManager) addIdentityAnnotations(ctx context.Context, initialAnnotations map[string]string) map[string]string {
	// Check if identity annotation injection is enabled
	if !m.config.ApplicationConfiguration().GetTopLevelConfig().GetInjectIdentityAnnotations() {
		return initialAnnotations
	}

	// Get identity from authentication context
	identityContext := auth.IdentityContextFromContext(ctx)

	if initialAnnotations == nil {
		initialAnnotations = make(map[string]string)
	}

	prefix := m.config.ApplicationConfiguration().GetTopLevelConfig().GetIdentityAnnotationPrefix()
	keys := m.config.ApplicationConfiguration().GetTopLevelConfig().GetIdentityAnnotationKeys()

	// Determine if this is an app or user identity
	isAppIdentity := identityContext.AppID() != ""
	isUserIdentity := identityContext.UserInfo() != nil

	if !isAppIdentity && !isUserIdentity {
		logger.Debugf(ctx, "No identity information found in context, skipping identity annotation injection")
		return initialAnnotations
	}

	// Add annotations based on identity type
	if isAppIdentity {
		// Handle app-based identity
		appID := identityContext.AppID()
		for _, key := range keys {
			annotationKey := prefix + "/app-" + key
			if _, exists := initialAnnotations[annotationKey]; !exists {
				var value string
				switch key {
				case "email", "sub", "id":
					// For app identities, use the app ID for these fields
					value = appID
				default:
					// Skip unknown keys for app identities
					continue
				}
				if value != "" {
					initialAnnotations[annotationKey] = value
					logger.Debugf(ctx, "Injected app identity annotation %s=%s", annotationKey, value)
				}
			}
		}
	} else if isUserIdentity {
		// Handle user-based identity
		userInfo := identityContext.UserInfo()
		for _, key := range keys {
			annotationKey := prefix + "/user-" + key
			if _, exists := initialAnnotations[annotationKey]; !exists {
				var value string
				switch key {
				case "email":
					value = userInfo.GetEmail()
				case "sub":
					value = userInfo.GetSubject()
				default:
					// Skip unknown keys
					continue
				}
				if value != "" {
					initialAnnotations[annotationKey] = value
					logger.Debugf(ctx, "Injected user identity annotation %s=%s", annotationKey, value)
				}
			}
		}
	}

	return initialAnnotations
}

func addStateFilter(filters []common.InlineFilter) ([]common.InlineFilter, error) {
	var stateFilterExists bool
	for _, inlineFilter := range filters {
		if inlineFilter.GetField() == shared.State {
			stateFilterExists = true
		}
	}

	if !stateFilterExists {
		stateFilter, err := common.NewSingleValueFilter(common.Execution, common.Equal, shared.State,
			admin.ExecutionState_EXECUTION_ACTIVE)
		if err != nil {
			return filters, err
		}
		filters = append(filters, stateFilter)
	}
	return filters, nil
}
