package impl

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/flyteorg/flyte/flyteadmin/auth"
	"github.com/flyteorg/flyte/flyteadmin/pkg/artifacts"
	cloudeventInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/async/cloudevent/interfaces"
	eventWriter "github.com/flyteorg/flyte/flyteadmin/pkg/async/events/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/async/notifications"
	notificationInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/async/notifications/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	dataInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/data/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/executions"
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
	artifactsIdl "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/artifacts"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

const childContainerQueueKey = "child_queue"
const artifactTrackerKey = "_ua"

// Map of [project] -> map of [domain] -> stop watch
type projectDomainScopedStopWatchMap = map[string]map[string]*promutils.StopWatch

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
}

type executionUserMetrics struct {
	Scope                        promutils.Scope
	ScheduledExecutionDelays     projectDomainScopedStopWatchMap
	WorkflowExecutionDurations   projectDomainScopedStopWatchMap
	WorkflowExecutionInputBytes  prometheus.Summary
	WorkflowExecutionOutputBytes prometheus.Summary
	MaxActiveExecutionsReached   prometheus.Counter
}

type ExecutionManager struct {
	db                        repositoryInterfaces.Repository
	config                    runtimeInterfaces.Configuration
	storageClient             *storage.DataStore
	queueAllocator            executions.QueueAllocator
	_clock                    clock.Clock
	systemMetrics             executionSystemMetrics
	userMetrics               executionUserMetrics
	notificationClient        notificationInterfaces.Publisher
	urlData                   dataInterfaces.RemoteURLInterface
	workflowManager           interfaces.WorkflowInterface
	launchPlanManager         interfaces.LaunchPlanInterface
	namedEntityManager        interfaces.NamedEntityInterface
	resourceManager           interfaces.ResourceInterface
	qualityOfServiceAllocator executions.QualityOfServiceAllocator
	eventPublisher            notificationInterfaces.Publisher
	cloudEventPublisher       notificationInterfaces.Publisher
	dbEventWriter             eventWriter.WorkflowExecutionEventWriter
	pluginRegistry            *plugins.Registry
	artifactRegistry          *artifacts.ArtifactRegistry
}

func getExecutionContext(ctx context.Context, id *core.WorkflowExecutionIdentifier) context.Context {
	ctx = contextutils.WithExecutionID(ctx, id.Name)
	return contextutils.WithProjectDomain(ctx, id.Project, id.Domain)
}

// Returns the unique string which identifies the authenticated end user (if any).
func getUser(ctx context.Context) string {
	identityContext := auth.IdentityContextFromContext(ctx)
	return identityContext.UserID()
}

func (m *ExecutionManager) populateExecutionQueue(
	ctx context.Context, identifier core.Identifier, compiledWorkflow *core.CompiledWorkflowClosure) {
	queueConfig := m.queueAllocator.GetQueue(ctx, identifier)
	for _, task := range compiledWorkflow.Tasks {
		container := task.Template.GetContainer()
		if container == nil {
			// Unrecognized target type, nothing to do
			continue
		}

		if queueConfig.DynamicQueue != "" {
			logger.Debugf(ctx, "Assigning %s as child queue for task %+v", queueConfig.DynamicQueue, task.Template.Id)
			container.Config = append(container.Config, &core.KeyValuePair{
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
		Project:      executionID.Project,
		Domain:       executionID.Domain,
		Workflow:     workflowName,
		LaunchPlan:   launchPlanName,
		ResourceType: admin.MatchableResource_PLUGIN_OVERRIDE,
	})
	if err != nil && !errors.IsDoesNotExistError(err) {
		return nil, err
	}
	if override != nil && override.Attributes != nil && override.Attributes.GetPluginOverrides() != nil {
		return override.Attributes.GetPluginOverrides().Overrides, nil
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

	if task.Template == nil || task.Template.GetContainer() == nil {
		// Nothing to do
		logger.Debugf(ctx, "Not setting default resources for task [%+v], no container resources found to check", task)
		return
	}

	if task.Template.GetContainer().Resources == nil {
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
	taskResourceRequirements := util.GetCompleteTaskResourceRequirements(ctx, task.Template.Id, task)

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
	if requestSpec.Metadata == nil || requestSpec.Metadata.ParentNodeExecution == nil {
		return parentNodeExecutionID, sourceExecutionID, nil
	}
	parentNodeExecutionModel, err := util.GetNodeExecutionModel(ctx, m.db, requestSpec.Metadata.ParentNodeExecution)
	if err != nil {
		logger.Errorf(ctx, "Failed to get node execution [%+v] that launched this execution [%+v] with error %v",
			requestSpec.Metadata.ParentNodeExecution, workflowExecutionID, err)
		return parentNodeExecutionID, sourceExecutionID, err
	}

	parentNodeExecutionID = parentNodeExecutionModel.ID

	sourceExecutionModel, err := util.GetExecutionModel(ctx, m.db, *requestSpec.Metadata.ParentNodeExecution.ExecutionId)
	if err != nil {
		logger.Errorf(ctx, "Failed to get workflow execution [%+v] that launched this execution [%+v] with error %v",
			requestSpec.Metadata.ParentNodeExecution, workflowExecutionID, err)
		return parentNodeExecutionID, sourceExecutionID, err
	}
	sourceExecutionID = sourceExecutionModel.ID
	requestSpec.Metadata.Principal = sourceExecutionModel.User
	sourceExecution, err := transformers.FromExecutionModel(ctx, *sourceExecutionModel, transformers.DefaultExecutionTransformerOptions)
	if err != nil {
		logger.Errorf(ctx, "Failed transform parent execution model for child execution [%+v] with err: %v", workflowExecutionID, err)
		return parentNodeExecutionID, sourceExecutionID, err
	}
	if sourceExecution.Spec.Metadata != nil {
		requestSpec.Metadata.Nesting = sourceExecution.Spec.Metadata.Nesting + 1
	} else {
		requestSpec.Metadata.Nesting = 1
	}
	return parentNodeExecutionID, sourceExecutionID, nil
}

// Produces execution-time attributes for workflow execution.
// Defaults to overridable execution values set in the execution create request, then looks at the launch plan values
// (if any) before defaulting to values set in the matchable resource db and further if matchable resources don't
// exist then defaults to one set in application configuration
func (m *ExecutionManager) getExecutionConfig(ctx context.Context, request *admin.ExecutionCreateRequest,
	launchPlan *admin.LaunchPlan) (*admin.WorkflowExecutionConfig, error) {

	workflowExecConfig := admin.WorkflowExecutionConfig{}
	// Merge the request spec into workflowExecConfig
	workflowExecConfig = util.MergeIntoExecConfig(workflowExecConfig, request.Spec)

	var workflowName string
	if launchPlan != nil && launchPlan.Spec != nil {
		// Merge the launch plan spec into workflowExecConfig
		workflowExecConfig = util.MergeIntoExecConfig(workflowExecConfig, launchPlan.Spec)
		if launchPlan.Spec.WorkflowId != nil {
			workflowName = launchPlan.Spec.WorkflowId.Name
		}
	}

	// This will get the most specific Workflow Execution Config.
	matchableResource, err := util.GetMatchableResource(ctx, m.resourceManager,
		admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG, request.Org, request.Project, request.Domain, workflowName)
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
		admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG, request.Org, request.Project, "", "")
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
	return &workflowExecConfig, nil
}

func (m *ExecutionManager) getClusterAssignment(ctx context.Context, req *admin.ExecutionCreateRequest) (*admin.ClusterAssignment, error) {
	storedAssignment, err := m.fetchClusterAssignment(ctx, req.Org, req.Project, req.Domain)
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
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "execution with project %q and domain %q cannot run on cluster pool %q, because its configured to run on pool %q", req.Project, req.Domain, reqPool, storedPool)
	}

	return storedAssignment, nil
}

func (m *ExecutionManager) fetchClusterAssignment(ctx context.Context, org, project, domain string) (*admin.ClusterAssignment, error) {
	resource, err := m.resourceManager.GetResource(ctx, interfaces.ResourceRequest{
		Org:          org,
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

func (m *ExecutionManager) getExternalResourceAttributes(ctx context.Context, request *admin.ExecutionCreateRequest) (
	*workflowengineInterfaces.ExternalResourceAttributes, error) {
	// Merge connections from all sources
	// The order of precedence is Project-Domain > Project > Global
	// conn1 (openai) <- project-domain
	// conn2 (databricks) <- project
	// conn3 (s3) <- global
	// conn4 (databricks) <- global
	// conn5 (openai) <- global
	// Merged connections: conn1, conn2, conn3
	externalResourceAttributes := &workflowengineInterfaces.ExternalResourceAttributes{}

	// TODO: Get the global level external resource attributes

	// Project level external resource attributes
	projectResource, err := m.resourceManager.GetResource(ctx, interfaces.ResourceRequest{
		Org:          request.Org,
		Project:      request.Project,
		ResourceType: admin.MatchableResource_EXTERNAL_RESOURCE,
	})
	if err != nil && !errors.IsDoesNotExistError(err) {
		errString := fmt.Sprintf("Failed to get external resource with error: %v", err)
		logger.Error(ctx, errString)
		return nil, status.Errorf(codes.Internal, errString)
	}
	if projectResource != nil && projectResource.Attributes.GetExternalResourceAttributes() != nil {
		attr := projectResource.Attributes.GetExternalResourceAttributes()
		for name, connection := range attr.Connections {
			externalResourceAttributes.AddConnection(name, workflowengineInterfaces.Connection{
				TaskType: connection.GetTaskType(),
				Secrets:  connection.GetSecrets(),
				Configs:  connection.GetConfigs(),
				Source:   workflowengineInterfaces.AttributesSource(admin.AttributesSource_PROJECT),
			})
		}
	}

	// Project-Domain level external resource attributes
	projectDomainResource, err := m.resourceManager.GetResource(ctx, interfaces.ResourceRequest{
		Org:          request.Org,
		Project:      request.Project,
		Domain:       request.Domain,
		ResourceType: admin.MatchableResource_EXTERNAL_RESOURCE,
	})
	if err != nil && !errors.IsDoesNotExistError(err) {
		errString := fmt.Sprintf("Failed to get external resource with error: %v", err)
		logger.Error(ctx, errString)
		return nil, status.Errorf(codes.Internal, errString)
	}
	if projectDomainResource != nil && projectDomainResource.Attributes.GetExternalResourceAttributes() != nil {
		attr := projectDomainResource.Attributes.GetExternalResourceAttributes()
		for name, connection := range attr.Connections {
			externalResourceAttributes.AddConnection(name, workflowengineInterfaces.Connection{
				TaskType: connection.GetTaskType(),
				Secrets:  connection.GetSecrets(),
				Configs:  connection.GetConfigs(),
				Source:   workflowengineInterfaces.AttributesSource(admin.AttributesSource_PROJECT_DOMAIN),
			})
		}
	}

	return externalResourceAttributes, nil
}

func (m *ExecutionManager) launchSingleTaskExecution(
	ctx context.Context, request admin.ExecutionCreateRequest, requestedAt time.Time) (
	context.Context, *models.Execution, []*models.ExecutionTag, error) {

	taskModel, err := m.db.TaskRepo().Get(ctx, repositoryInterfaces.Identifier{
		Org:     request.Spec.LaunchPlan.Org,
		Project: request.Spec.LaunchPlan.Project,
		Domain:  request.Spec.LaunchPlan.Domain,
		Name:    request.Spec.LaunchPlan.Name,
		Version: request.Spec.LaunchPlan.Version,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	task, err := transformers.FromTaskModel(taskModel)
	if err != nil {
		return nil, nil, nil, err
	}

	// Prepare a skeleton workflow and launch plan
	taskIdentifier := request.Spec.LaunchPlan
	workflowModel, err :=
		util.CreateOrGetWorkflowModel(ctx, request, m.db, m.workflowManager, m.namedEntityManager, taskIdentifier, &task)
	if err != nil {
		logger.Debugf(ctx, "Failed to created skeleton workflow for [%+v] with err: %v", taskIdentifier, err)
		return nil, nil, nil, err
	}
	workflow, err := transformers.FromWorkflowModel(*workflowModel)
	if err != nil {
		return nil, nil, nil, err
	}

	launchPlan, err := util.CreateOrGetLaunchPlan(ctx, m.db, m.config, m.namedEntityManager, taskIdentifier,
		workflow.Closure.CompiledWorkflow.Primary.Template.Interface, workflowModel.ID,
		request.Spec.AuthRole, request.Spec.SecurityContext)
	if err != nil {
		return nil, nil, nil, err
	}

	executionInputs, err := validation.CheckAndFetchInputsForExecution(
		request.Inputs,
		launchPlan.Spec.FixedInputs,
		launchPlan.Closure.ExpectedInputs,
	)
	if err != nil {
		logger.Debugf(ctx, "Failed to CheckAndFetchInputsForExecution with request.Inputs: %+v"+
			"fixed inputs: %+v and expected inputs: %+v with err %v",
			request.Inputs, launchPlan.Spec.FixedInputs, launchPlan.Closure.ExpectedInputs, err)
		return nil, nil, nil, err
	}

	name := util.GetExecutionName(request)
	workflowExecutionID := core.WorkflowExecutionIdentifier{
		Org:     request.Org,
		Project: request.Project,
		Domain:  request.Domain,
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
			workflowExecutionID.Org, workflowExecutionID.Project, workflowExecutionID.Domain, workflowExecutionID.Name, shared.Inputs)
		return err
	})

	var userInputsURI storage.DataReference
	offloadInputsGroup.Go(func() error {
		var err error
		userInputsURI, err = common.OffloadLiteralMap(offloadInputsGroupCtx, m.storageClient, request.Inputs,
			workflowExecutionID.Org, workflowExecutionID.Project, workflowExecutionID.Domain, workflowExecutionID.Name, shared.UserInputs)
		return err
	})

	err = getClosureGroup.Wait()
	if err != nil {
		return nil, nil, nil, err
	}
	closure.CreatedAt = workflow.Closure.CreatedAt
	workflow.Closure = closure

	ctx = getExecutionContext(ctx, &workflowExecutionID)
	namespace := common.GetNamespaceName(
		m.config.NamespaceMappingConfiguration().GetNamespaceTemplate(), workflowExecutionID.Org, workflowExecutionID.Project, workflowExecutionID.Domain)

	requestSpec := request.Spec
	if requestSpec.Metadata == nil {
		requestSpec.Metadata = &admin.ExecutionMetadata{}
	}
	requestSpec.Metadata.Principal = getUser(ctx)

	// Get the node execution (if any) that launched this execution
	var parentNodeExecutionID uint
	var sourceExecutionID uint
	parentNodeExecutionID, sourceExecutionID, err = m.getInheritedExecMetadata(ctx, requestSpec, &workflowExecutionID)
	if err != nil {
		return nil, nil, nil, err
	}

	// Dynamically assign task resource defaults.
	var taskResources workflowengineInterfaces.TaskResources
	if requestSpec.TaskResourceAttributes != nil {
		taskResources = workflowengineInterfaces.TaskResources{
			Defaults: util.FromAdminProtoTaskResourceSpec(ctx, requestSpec.TaskResourceAttributes.Defaults),
			Limits:   util.FromAdminProtoTaskResourceSpec(ctx, requestSpec.TaskResourceAttributes.Limits),
		}
	} else {
		taskResources = util.GetTaskResources(ctx, workflow.Id, m.resourceManager, m.config.TaskResourceConfiguration())
	}
	for _, t := range workflow.Closure.CompiledWorkflow.Tasks {
		m.setCompiledTaskDefaults(ctx, t, taskResources)
	}

	// Dynamically assign execution queues.
	m.populateExecutionQueue(ctx, *workflow.Id, workflow.Closure.CompiledWorkflow)

	executionConfig, err := m.getExecutionConfig(ctx, &request, nil)
	if err != nil {
		return nil, nil, nil, err
	}

	var labels map[string]string
	if executionConfig.Labels != nil {
		labels = executionConfig.Labels.Values
	}

	labels, err = m.addProjectLabels(ctx, request.Project, request.Org, labels)
	if err != nil {
		return nil, nil, nil, err
	}

	var annotations map[string]string
	if executionConfig.Annotations != nil {
		annotations = executionConfig.Annotations.Values
	}

	var rawOutputDataConfig *admin.RawOutputDataConfig
	if executionConfig.RawOutputDataConfig != nil {
		rawOutputDataConfig = executionConfig.RawOutputDataConfig
	}

	clusterAssignment, err := m.getClusterAssignment(ctx, &request)
	if err != nil {
		return nil, nil, nil, err
	}

	externalResourceAttributes, err := m.getExternalResourceAttributes(ctx, &request)
	if err != nil {
		return nil, nil, nil, err
	}

	var executionClusterLabel *admin.ExecutionClusterLabel
	if requestSpec.ExecutionClusterLabel != nil {
		executionClusterLabel = requestSpec.ExecutionClusterLabel
	}
	executionParameters := workflowengineInterfaces.ExecutionParameters{
		Inputs:                     executionInputs,
		AcceptedAt:                 requestedAt,
		Labels:                     labels,
		Annotations:                annotations,
		ExecutionConfig:            executionConfig,
		TaskResources:              &taskResources,
		EventVersion:               m.config.ApplicationConfiguration().GetTopLevelConfig().EventVersion,
		RoleNameKey:                m.config.ApplicationConfiguration().GetTopLevelConfig().RoleNameKey,
		RawOutputDataConfig:        rawOutputDataConfig,
		ClusterAssignment:          clusterAssignment,
		ExecutionClusterLabel:      executionClusterLabel,
		ExternalResourceAttributes: externalResourceAttributes,
	}

	overrides, err := m.addPluginOverrides(ctx, &workflowExecutionID, workflowExecutionID.Name, "")
	if err != nil {
		return nil, nil, nil, err
	}
	if overrides != nil {
		executionParameters.TaskPluginOverrides = overrides
	}
	if request.Spec.Metadata != nil && request.Spec.Metadata.ReferenceExecution != nil &&
		request.Spec.Metadata.Mode == admin.ExecutionMetadata_RECOVERED {
		executionParameters.RecoveryExecution = request.Spec.Metadata.ReferenceExecution
	}

	if err := m.validateActiveExecutions(ctx); err != nil {
		return nil, nil, nil, err
	}

	workflowExecutor := plugins.Get[workflowengineInterfaces.WorkflowExecutor](m.pluginRegistry, plugins.PluginIDWorkflowExecutor)
	execInfo, err := workflowExecutor.Execute(ctx, workflowengineInterfaces.ExecutionData{
		Namespace:                namespace,
		ExecutionID:              &workflowExecutionID,
		ReferenceWorkflowName:    workflow.Id.Name,
		ReferenceLaunchPlanName:  launchPlan.Id.Name,
		WorkflowClosure:          workflow.Closure.CompiledWorkflow,
		WorkflowClosureReference: storage.DataReference(workflowModel.RemoteClosureIdentifier),
		ExecutionParameters:      executionParameters,
	})

	if err != nil {
		m.systemMetrics.PropellerFailures.Inc()
		logger.Infof(ctx, "Failed to execute workflow %+v with execution id %+v and inputs %+v with err %v",
			request, workflowExecutionID, request.Inputs, err)
		return nil, nil, nil, err
	}
	executionCreatedAt := time.Now()
	acceptanceDelay := executionCreatedAt.Sub(requestedAt)
	m.systemMetrics.AcceptanceDelay.Observe(acceptanceDelay.Seconds())

	// Request notification settings takes precedence over the launch plan settings.
	// If there is no notification in the request and DisableAll is not true, use the settings from the launch plan.
	var notificationsSettings []*admin.Notification
	if launchPlan.Spec.GetEntityMetadata() != nil {
		notificationsSettings = launchPlan.Spec.EntityMetadata.GetNotifications()
	}
	if request.Spec.GetNotifications() != nil && request.Spec.GetNotifications().Notifications != nil &&
		len(request.Spec.GetNotifications().Notifications) > 0 {
		notificationsSettings = request.Spec.GetNotifications().Notifications
	} else if request.Spec.GetDisableAll() {
		notificationsSettings = make([]*admin.Notification, 0)
	}

	err = offloadInputsGroup.Wait()
	if err != nil {
		return nil, nil, nil, err
	}

	resolvedSpec := proto.Clone(request.Spec).(*admin.ExecutionSpec)
	resolvedSpec = completeResolvedSpec(resolvedSpec, &executionParameters)

	createExecutionModelInput := transformers.CreateExecutionModelInput{
		WorkflowExecutionID: workflowExecutionID,
		RequestSpec:         requestSpec,
		ResolvedSpec:        resolvedSpec,
		TaskID:              taskModel.ID,
		WorkflowID:          workflowModel.ID,
		// The execution is not considered running until the propeller sends a specific event saying so.
		CreatedAt:             m._clock.Now(),
		Notifications:         notificationsSettings,
		WorkflowIdentifier:    workflow.Id,
		ParentNodeExecutionID: parentNodeExecutionID,
		SourceExecutionID:     sourceExecutionID,
		Cluster:               execInfo.Cluster,
		InputsURI:             inputsURI,
		UserInputsURI:         userInputsURI,
		SecurityContext:       executionConfig.SecurityContext,
		LaunchEntity:          taskIdentifier.ResourceType,
		Namespace:             namespace,
	}

	executionModel, err := transformers.CreateExecutionModel(createExecutionModelInput)
	if err != nil {
		logger.Infof(ctx, "Failed to create execution model in transformer for id: [%+v] with err: %v",
			workflowExecutionID, err)
		return nil, nil, nil, err
	}
	m.userMetrics.WorkflowExecutionInputBytes.Observe(float64(proto.Size(request.Inputs)))
	executionTagModel, err := transformers.CreateExecutionTagModel(createExecutionModelInput)
	if err != nil {
		logger.Infof(ctx, "Failed to create execution tag model in transformer for id: [%+v] with err: %v",
			workflowExecutionID, err)
		return nil, nil, nil, err
	}

	return ctx, executionModel, executionTagModel, nil
}

func resolveAuthRole(request *admin.ExecutionCreateRequest, launchPlan *admin.LaunchPlan) *admin.AuthRole {
	if request.Spec.AuthRole != nil {
		return request.Spec.AuthRole
	}

	if launchPlan == nil || launchPlan.Spec == nil {
		return &admin.AuthRole{}
	}

	// Set role permissions based on launch plan Auth values.
	// The branched-ness of this check is due to the presence numerous deprecated fields
	if launchPlan.Spec.GetAuthRole() != nil {
		return launchPlan.Spec.GetAuthRole()
	} else if launchPlan.GetSpec().GetAuth() != nil {
		return &admin.AuthRole{
			AssumableIamRole:         launchPlan.GetSpec().GetAuth().AssumableIamRole,
			KubernetesServiceAccount: launchPlan.GetSpec().GetAuth().KubernetesServiceAccount,
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
	if executionConfigSecurityCtx != nil && executionConfigSecurityCtx.RunAs != nil &&
		(len(executionConfigSecurityCtx.RunAs.K8SServiceAccount) > 0 ||
			len(executionConfigSecurityCtx.RunAs.IamRole) > 0 ||
			len(executionConfigSecurityCtx.RunAs.ExecutionIdentity) > 0) {
		return executionConfigSecurityCtx
	}
	logger.Warn(ctx, "Setting security context from auth Role")
	return &core.SecurityContext{
		RunAs: &core.Identity{
			IamRole:           resolvedAuthRole.AssumableIamRole,
			K8SServiceAccount: resolvedAuthRole.KubernetesServiceAccount,
		},
	}
}

// ExtractArtifactTrackers pulls out artifact tracker strings from Literals for lineage
func (m *ExecutionManager) ExtractArtifactTrackers(artifactTrackers map[string]string, input *core.Literal) {

	if input == nil {
		return
	}
	if input.GetMetadata() != nil {
		if tracker, ok := input.GetMetadata()[artifactTrackerKey]; ok {
			artifactTrackers[tracker] = ""
		}
	}
	if input.GetCollection() != nil {
		for _, v := range input.GetCollection().Literals {
			m.ExtractArtifactTrackers(artifactTrackers, v)
		}
	} else if input.GetMap() != nil {
		for _, v := range input.GetMap().Literals {
			m.ExtractArtifactTrackers(artifactTrackers, v)
		}
	}
}

// getStringFromInput should be called when a tag or partition value is a binding to an input. the input is looked up
// from the input map and the binding, and an error is returned if the input key is not in the map.
func (m *ExecutionManager) getStringFromInput(ctx context.Context, inputBinding core.InputBindingData, inputs map[string]*core.Literal) (string, error) {

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
			strVal = fmt.Sprintf("%d", p.GetInteger())
		case *core.Primitive_Datetime:
			t := time.Unix(p.GetDatetime().Seconds, int64(p.GetDatetime().Nanos))
			t = t.In(time.UTC)
			if t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 {
				strVal = t.Format("2006-01-02")
			} else {
				strVal = t.Format("2006-01-02T15:04:05Z07:00")
			}
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
		return m.getStringFromInput(ctx, *l.GetInputBinding(), inputs)
	}
	if l.GetStaticValue() != "" {
		return l.GetStaticValue(), nil
	}
	return "", errors.NewFlyteAdminErrorf(codes.InvalidArgument, "label value is empty")
}

func (m *ExecutionManager) fillInTemplateArgs(ctx context.Context, query core.ArtifactQuery, inputs map[string]*core.Literal) (core.ArtifactQuery, error) {
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
		project = contextutils.Value(ctx, contextutils.ProjectKey)
		domain = contextutils.Value(ctx, contextutils.DomainKey)

		var partitions map[string]*core.LabelValue

		if artifactID.GetPartitions().GetValue() != nil {
			partitions = make(map[string]*core.LabelValue, len(artifactID.GetPartitions().Value))
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
			if artifactID.GetTimePartition().Value.GetTimeValue() != nil {
				// If the time value is set, then just pass it through, nothing to fill in.
				timePartition = artifactID.GetTimePartition()
			} else if artifactID.GetTimePartition().Value.GetInputBinding() != nil {
				// Evaluate the time partition input binding
				lit, ok := inputs[artifactID.GetTimePartition().Value.GetInputBinding().GetVar()]
				if !ok {
					return query, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "time partition input binding var [%s] not found in inputs %v", artifactID.GetTimePartition().Value.GetInputBinding().GetVar(), inputs)
				}

				if lit.GetScalar().GetPrimitive().GetDatetime() == nil {
					return query, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
						"time partition binding to input var [%s] failing because %v is not a datetime",
						artifactID.GetTimePartition().Value.GetInputBinding().GetVar(), lit)
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

		return core.ArtifactQuery{
			Identifier: &core.ArtifactQuery_ArtifactId{
				ArtifactId: &core.ArtifactID{
					ArtifactKey: &core.ArtifactKey{
						Project: project,
						Domain:  domain,
						Name:    ak.GetName(),
						Org:     ak.GetOrg(),
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

// ResolveParameterMapArtifacts will go through the parameter map, and resolve any artifact queries.
func (m *ExecutionManager) ResolveParameterMapArtifacts(ctx context.Context, inputs *core.ParameterMap, inputsForQueryTemplating map[string]*core.Literal, executionOrg string) (*core.ParameterMap, []*core.ArtifactID, error) {

	// only top level replace for now. Need to make this recursive
	var artifactIDs []*core.ArtifactID
	if inputs == nil {
		return nil, artifactIDs, nil
	}
	outputs := map[string]*core.Parameter{}

	// copy ghHandle into request
	for k, v := range inputs.Parameters {
		if inputsForQueryTemplating != nil {
			if _, ok := inputsForQueryTemplating[k]; ok {
				// Mark these as required as they're already provided by the other two LiteralMaps
				outputs[k] = &core.Parameter{
					Var:      v.Var,
					Behavior: &core.Parameter_Required{Required: true},
				}
				continue
			}
		}
		if v.GetArtifactQuery() != nil {
			// This case handles when an Artifact query is specified as a default value.
			if m.artifactRegistry.GetClient() == nil {
				return nil, nil, errors.NewFlyteAdminErrorf(codes.Internal, "artifact client is not initialized, can't resolve queries")
			}
			filledInQuery, err := m.fillInTemplateArgs(ctx, *v.GetArtifactQuery(), inputsForQueryTemplating)
			if err != nil {
				logger.Errorf(ctx, "Failed to fill in template args for [%s] [%v]", k, err)
				return nil, nil, err
			}

			if filledInQuery.GetArtifactId().GetArtifactKey() != nil {
				filledInQuery.GetArtifactId().GetArtifactKey().Org = executionOrg
			}
			req := &artifactsIdl.GetArtifactRequest{
				Query:   &filledInQuery,
				Details: false,
			}

			resp, err := m.artifactRegistry.GetClient().GetArtifact(ctx, req)
			if err != nil {
				return nil, nil, err
			}
			artifactIDs = append(artifactIDs, resp.Artifact.GetArtifactId())
			logger.Debugf(ctx, "Resolved query for [%s] to [%+v]", k, resp.Artifact.ArtifactId)
			outputs[k] = &core.Parameter{
				Var:      v.Var,
				Behavior: &core.Parameter_Default{Default: resp.Artifact.Spec.Value},
			}
		} else if v.GetArtifactId() != nil {
			// This case is for when someone hard-codes a known ArtifactID as a default value.
			req := &artifactsIdl.GetArtifactRequest{
				Query: &core.ArtifactQuery{
					Identifier: &core.ArtifactQuery_ArtifactId{
						ArtifactId: v.GetArtifactId(),
					},
				},
				Details: false,
			}
			resp, err := m.artifactRegistry.GetClient().GetArtifact(ctx, req)
			if err != nil {
				return nil, nil, err
			}
			artifactIDs = append(artifactIDs, v.GetArtifactId())
			logger.Debugf(ctx, "Using specified artifactID for [%+v] for [%s]", v.GetArtifactId(), k)
			outputs[k] = &core.Parameter{
				Var:      v.Var,
				Behavior: &core.Parameter_Default{Default: resp.Artifact.Spec.Value},
			}
		} else {
			outputs[k] = v
		}
	}
	pm := &core.ParameterMap{Parameters: outputs}
	return pm, artifactIDs, nil
}

func (m *ExecutionManager) validateActiveExecutions(ctx context.Context) error {
	getUserPropertiesFunc := plugins.Get[shared.GetUserProperties](m.pluginRegistry, plugins.PluginIDUserProperties)
	if getUserPropertiesFunc == nil {
		logger.Warningf(ctx, "failed to find plugin for: %s, not validating", plugins.PluginIDUserProperties)
		return nil
	}
	userProperties := getUserPropertiesFunc(ctx)
	if userProperties.ActiveExecutions == 0 {
		logger.Debugf(ctx, "not validating whether execution can be triggered, since active executions limit is set to 0")
		return nil
	}
	aWeekAgo := time.Now().AddDate(0, 0, -7)
	aWeekAgoFilter, err := util.NewTimestampFilter(common.Execution, common.GreaterThanOrEqual, shared.ExecutionCreatedAt, aWeekAgo)
	if err != nil {
		logger.Error(ctx, "failed to add a week ago timestamp filter when filtering on active executions for org %s, aWeekAgo %+v and err: %v", userProperties.Org, aWeekAgo, err)
		return err
	}

	orgFilter, err := util.GetSingleValueEqualityFilter(common.Execution, shared.Org, userProperties.Org)
	if err != nil {
		logger.Error(ctx, "failed to add org filter when filtering on active executions for org %s and err: %v", userProperties.Org, err)
		return err
	}
	activeExecutionFilter, err := common.NewRepeatedValueFilter(common.Execution, common.ValueIn, shared.Phase, common.ActiveExecutionPhases)
	if err != nil {
		logger.Errorf(ctx, "failed to construct active execution phase filter with err %v", err)
		return err
	}
	filters := []common.InlineFilter{aWeekAgoFilter, orgFilter, activeExecutionFilter}

	countExecutionByPhaseInput := repositoryInterfaces.CountResourceInput{
		InlineFilters: filters,
	}
	countExecutionByPhaseOutput, err := m.db.ExecutionRepo().CountByPhase(ctx, countExecutionByPhaseInput)
	if err != nil {
		logger.Debugf(ctx, "failed to get execution counts using input [%+v] with err %v", countExecutionByPhaseInput, err)
		return err
	}

	executionCounts, err := transformers.FromExecutionCountsByPhase(ctx, countExecutionByPhaseOutput)
	if err != nil {
		logger.Errorf(ctx, "failed to transform execution by phase output [%+v] with err %v", countExecutionByPhaseOutput, err)
		return err
	}
	var activeExecutions int64
	for _, execCount := range executionCounts {
		if !common.IsExecutionTerminal(execCount.Phase) {
			activeExecutions += execCount.Count
		}
	}
	if activeExecutions > int64(userProperties.ActiveExecutions) {
		m.userMetrics.MaxActiveExecutionsReached.Inc()
		logger.Infof(ctx, "user org '%s' has '%d' active executions which exceeds their account limit '%d': blocking creation of additional executions",
			userProperties.Org, activeExecutions, userProperties.ActiveExecutions)
		return errors.NewFlyteAdminErrorf(codes.ResourceExhausted, "Your account only allows %d active executions at a time. Please wait for some of your existing executions to complete before starting new ones.", userProperties.ActiveExecutions)
	}
	logger.Debugf(ctx, "user org '%s' has '%d' active executions which is permitted under their account limit '%d'",
		userProperties.Org, activeExecutions, userProperties.ActiveExecutions)
	return nil
}

func (m *ExecutionManager) launchExecutionAndPrepareModel(
	ctx context.Context, request admin.ExecutionCreateRequest, requestedAt time.Time) (context.Context, *models.Execution, []*models.ExecutionTag, error) {
	err := validation.ValidateExecutionRequest(ctx, request, m.db, m.config.ApplicationConfiguration())
	if err != nil {
		logger.Debugf(ctx, "Failed to validate ExecutionCreateRequest %+v with err %v", request, err)
		return nil, nil, nil, err
	}

	// This section is for serverless user namespace checking. In BYOC, ValidateUserOrg always returns nil.
	err = plugins.Get[executions.PreExecutionValidationPlugin](m.pluginRegistry, plugins.PluginIDPreExecutionValidation).ValidateCreateExecutionRequest(ctx, request)
	if err != nil {
		logger.Errorf(ctx, "Failed to validate execution request [%+v] with err %v", request, err)
		return nil, nil, nil, err
	}

	if request.Spec.LaunchPlan.ResourceType == core.ResourceType_TASK {
		logger.Debugf(ctx, "Launching single task execution with [%+v]", request.Spec.LaunchPlan)
		nodeId := request.GetSpec().GetMetadata().GetParentNodeExecution().GetNodeId()
		executionId := request.GetSpec().GetMetadata().GetParentNodeExecution().GetExecutionId()
		if nodeId != "" && executionId != nil {
			subNodeIds, err := util.GetSubNodesFromParentNodeExecution(ctx, m.db, request.GetSpec().GetMetadata().GetParentNodeExecution())
			if err != nil {
				logger.Errorf(ctx, "failed to fetch sub nodes for the parent node: %v %s with err: %v", executionId, nodeId, err)
				return nil, nil, nil, err
			}

			execution, err := m.GetExecution(ctx, admin.WorkflowExecutionGetRequest{
				Id: executionId,
			})
			if err != nil {
				logger.Errorf(ctx, "failed to fetch parent node's execution with err: %v", err)
				return nil, nil, nil, err
			}

			launchPlanID := execution.GetSpec().GetLaunchPlan()
			if launchPlanID == nil {
				logger.Errorf(ctx, "Launch plan id is nil for the parent node execution: %v %s", executionId, nodeId)
				return nil, nil, nil, errors.NewFlyteAdminErrorf(codes.Internal, "Parent node's launch plan id is nil")
			}

			parentNodeWorkflowIdentifier := execution.GetClosure().GetWorkflowId()
			if parentNodeWorkflowIdentifier == nil {
				logger.Errorf(ctx, "Workflow identifier is nil for the parent node: %v %s", executionId, nodeId)
				return nil, nil, nil, errors.NewFlyteAdminErrorf(codes.Internal, "Parent node's workflow identifier is nil")
			}

			parentNodeWorkflow, err := util.GetWorkflow(ctx, m.db, m.storageClient, *parentNodeWorkflowIdentifier)
			if err != nil {
				logger.Errorf(ctx, "failed to fetch workflow of the parent node execution: %v %s with err: %v", executionId, nodeId, err)
				return nil, nil, nil, err
			}

			subNode, err := util.GetSubNodeFromWorkflow(parentNodeWorkflow, subNodeIds)
			// launchSingleTaskExecution if node is not an ArrayNode or cannot find node in the workflow
			if err == nil && subNode.GetArrayNode() != nil {
				// check if the fetched array node matches the task passed in
				arrayNodeTaskId := subNode.GetArrayNode().GetNode().GetTaskNode().GetReferenceId()
				if arrayNodeTaskId.GetName() != request.GetSpec().GetLaunchPlan().GetName() ||
					arrayNodeTaskId.GetProject() != request.GetSpec().GetLaunchPlan().GetProject() ||
					arrayNodeTaskId.GetDomain() != request.GetSpec().GetLaunchPlan().GetDomain() ||
					arrayNodeTaskId.GetOrg() != request.GetSpec().GetLaunchPlan().GetOrg() {
					return nil, nil, nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "Task spec does not match the node under the parent node execution id")
				}

				// if versions differ - check if the subNode is the same across versions
				// Note: the launch form doesn't get updated across versions so the interfaces might differ and will error
				if arrayNodeTaskId.GetVersion() != request.GetSpec().GetLaunchPlan().GetVersion() {
					diffVersionWorkflowIdentifier := &core.Identifier{
						Name:    parentNodeWorkflowIdentifier.GetName(),
						Project: parentNodeWorkflowIdentifier.GetProject(),
						Domain:  parentNodeWorkflowIdentifier.GetDomain(),
						Org:     parentNodeWorkflowIdentifier.GetOrg(),
						Version: request.GetSpec().GetLaunchPlan().GetVersion(),
					}
					diffVersionWorkflow, err := util.GetWorkflow(ctx, m.db, m.storageClient, *diffVersionWorkflowIdentifier)
					if err != nil {
						logger.Errorf(ctx, "failed to fetch workflow for different version %s of the execution for the parent node: %v %s with err: %v", request.GetSpec().GetLaunchPlan().GetVersion(), executionId, nodeId, err)
						return nil, nil, nil, err
					}
					diffVersionSubNode, err := util.GetSubNodeFromWorkflow(diffVersionWorkflow, subNodeIds)
					if err != nil {
						logger.Errorf(ctx, "failed to fetch sub node for different version %s of the execution for the parent node: %v %s with err: %v", request.GetSpec().GetLaunchPlan().GetVersion(), executionId, nodeId, err)
						return nil, nil, nil, err
					}
					diffVersionArrayNodeTaskId := diffVersionSubNode.GetArrayNode().GetNode().GetTaskNode().GetReferenceId()
					if arrayNodeTaskId.GetName() != diffVersionArrayNodeTaskId.GetName() ||
						arrayNodeTaskId.GetProject() != diffVersionArrayNodeTaskId.GetProject() ||
						arrayNodeTaskId.GetDomain() != diffVersionArrayNodeTaskId.GetDomain() ||
						arrayNodeTaskId.GetOrg() != diffVersionArrayNodeTaskId.GetOrg() {
						return nil, nil, nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "Selected version does not match the node under the parent node execution id")
					}

					// this id is only utilized for its values and not fetching the launch plan, so it's fine if this doesn't exist
					launchPlanID.Version = request.GetSpec().GetLaunchPlan().GetVersion()
					subNode = diffVersionSubNode
				}

				createLaunchPlanResponse, err := m.launchPlanManager.CreateLaunchPlanFromNode(ctx, admin.CreateLaunchPlanFromNodeRequest{
					SubNodes: &admin.CreateLaunchPlanFromNodeRequest_SubNodeSpec{
						SubNodeSpec: subNode,
					},
					Name:            executionId.GetName() + "-" + strings.Join(subNodeIds, "-"),
					LaunchPlanId:    launchPlanID,
					SecurityContext: execution.GetSpec().GetSecurityContext(),
				})
				if err != nil {
					logger.Errorf(ctx, "failed to create launch plan from node spec for the parent node: %v %s with err: %v", executionId, nodeId, err)
					return nil, nil, nil, err
				}
				request.Spec.LaunchPlan = createLaunchPlanResponse.GetLaunchPlan().GetId()
				return m.launchExecution(ctx, request, requestedAt)
			}
		}
		// When tasks can have defaults this will need to handle Artifacts as well.
		return m.launchSingleTaskExecution(ctx, request, requestedAt)
	}
	return m.launchExecution(ctx, request, requestedAt)
}

func (m *ExecutionManager) launchExecution(
	ctx context.Context, request admin.ExecutionCreateRequest, requestedAt time.Time) (context.Context, *models.Execution, []*models.ExecutionTag, error) {
	launchPlanModel, err := util.GetLaunchPlanModel(ctx, m.db, *request.Spec.LaunchPlan)
	if err != nil {
		logger.Debugf(ctx, "Failed to get launch plan model for ExecutionCreateRequest %+v with err %v", request, err)
		return nil, nil, nil, err
	}
	launchPlan, err := transformers.FromLaunchPlanModel(launchPlanModel)
	if err != nil {
		logger.Debugf(ctx, "Failed to transform launch plan model %+v with err %v", launchPlanModel, err)
		return nil, nil, nil, err
	}

	// TODO: Artifact feature gate, remove when ready
	var lpExpectedInputs *core.ParameterMap
	var artifactTrackers = make(map[string]string)
	var usedArtifactIDs []*core.ArtifactID
	if m.artifactRegistry.GetClient() != nil {
		// Literals may have an artifact key in the metadata field. This is something the artifact service should have
		// added. Pull these back out so we can keep track of them for lineage purposes. Use a dummy wrapper object for
		// easier recursion.
		requestInputMap := &core.Literal{
			Value: &core.Literal_Map{Map: request.Inputs},
		}
		fixedInputMap := &core.Literal{
			Value: &core.Literal_Map{Map: launchPlan.Spec.FixedInputs},
		}
		m.ExtractArtifactTrackers(artifactTrackers, requestInputMap)
		m.ExtractArtifactTrackers(artifactTrackers, fixedInputMap)

		// Put together the inputs that we've already resolved so that the artifact querying bit can fill them in.
		// This is to support artifact queries that depend on other inputs using the {{ .inputs.var }} construct.
		var inputsForQueryTemplating = make(map[string]*core.Literal)
		if request.Inputs != nil {
			for k, v := range request.Inputs.Literals {
				inputsForQueryTemplating[k] = v
			}
		}
		if launchPlan.Spec.FixedInputs != nil {
			for k, v := range launchPlan.Spec.FixedInputs.Literals {
				inputsForQueryTemplating[k] = v
			}
		}
		logger.Debugf(ctx, "Inputs for query templating: [%+v]", inputsForQueryTemplating)

		// Resolve artifact queries
		//   Within the launch plan, the artifact will be in the Parameter map, and can come in form of an ArtifactID,
		//     or as an ArtifactQuery.
		// Also send in the inputsForQueryTemplating for two reasons, so we don't run queries for things we don't need to
		// and so we can fill in template args.
		// ArtifactIDs are also returned for lineage purposes.
		ctxPD := contextutils.WithProjectDomain(ctx, request.Project, request.Domain)
		lpExpectedInputs, usedArtifactIDs, err = m.ResolveParameterMapArtifacts(ctxPD, launchPlan.Closure.ExpectedInputs, inputsForQueryTemplating, request.Org)
		if err != nil {
			logger.Errorf(ctx, "Error looking up launch plan closure parameter map: %v", err)
			return nil, nil, nil, err
		}

		logger.Debugf(ctx, "Resolved launch plan closure expected inputs from [%+v] to [%+v]", launchPlan.Closure.ExpectedInputs, lpExpectedInputs)
		logger.Debugf(ctx, "Found artifact trackers: %v", artifactTrackers)
		logger.Debugf(ctx, "Found artifact IDs: %v", usedArtifactIDs)

	} else {
		lpExpectedInputs = launchPlan.Closure.ExpectedInputs
	}

	// Artifacts retrieved will need to be stored somewhere to ensure that we can re-emit events if necessary
	// in the future, and also to make sure that relaunch and recover can use it if necessary.
	executionInputs, err := validation.CheckAndFetchInputsForExecution(
		request.Inputs,
		launchPlan.Spec.FixedInputs,
		lpExpectedInputs,
	)

	if err != nil {
		logger.Debugf(ctx, "Failed to CheckAndFetchInputsForExecution with request.Inputs: %+v"+
			"fixed inputs: %+v and expected inputs: %+v with err %v",
			request.Inputs, launchPlan.Spec.FixedInputs, lpExpectedInputs, err)
		return nil, nil, nil, err
	}

	workflowModel, err := util.GetWorkflowModel(ctx, m.db, *launchPlan.Spec.WorkflowId)
	if err != nil {
		logger.Debugf(ctx, "Failed to get workflow with id %+v with err %v", launchPlan.Spec.WorkflowId, err)
		return nil, nil, nil, err
	}
	workflow, err := transformers.FromWorkflowModel(workflowModel)
	if err != nil {
		logger.Debugf(ctx, "Failed to get workflow with id %+v with err %v", launchPlan.Spec.WorkflowId, err)
		return nil, nil, nil, err
	}

	name := util.GetExecutionName(request)
	workflowExecutionID := core.WorkflowExecutionIdentifier{
		Org:     request.Org,
		Project: request.Project,
		Domain:  request.Domain,
		Name:    name,
	}

	// Overlap the blob store reads and writes
	group, groupCtx := errgroup.WithContext(ctx)
	var closure *admin.WorkflowClosure
	group.Go(func() error {
		var err error
		closure, err = util.FetchAndGetWorkflowClosure(groupCtx, m.storageClient, workflowModel.RemoteClosureIdentifier)
		if err != nil {
			logger.Debugf(ctx, "Failed to get workflow with id %+v with err %v", launchPlan.Spec.WorkflowId, err)
		}
		return err
	})

	var inputsURI storage.DataReference
	group.Go(func() error {
		var err error
		inputsURI, err = common.OffloadLiteralMap(groupCtx, m.storageClient, executionInputs,
			workflowExecutionID.Org, workflowExecutionID.Project, workflowExecutionID.Domain, workflowExecutionID.Name, shared.Inputs)
		return err
	})

	var userInputsURI storage.DataReference
	group.Go(func() error {
		var err error
		userInputsURI, err = common.OffloadLiteralMap(groupCtx, m.storageClient, request.Inputs,
			workflowExecutionID.Org, workflowExecutionID.Project, workflowExecutionID.Domain, workflowExecutionID.Name, shared.UserInputs)
		return err
	})

	err = group.Wait()
	if err != nil {
		return nil, nil, nil, err
	}
	closure.CreatedAt = workflow.Closure.CreatedAt
	workflow.Closure = closure

	ctx = getExecutionContext(ctx, &workflowExecutionID)
	var requestSpec = request.Spec
	if requestSpec.Metadata == nil {
		requestSpec.Metadata = &admin.ExecutionMetadata{}
	}
	requestSpec.Metadata.Principal = getUser(ctx)
	requestSpec.Metadata.ArtifactIds = usedArtifactIDs

	// Get the node and parent execution (if any) that launched this execution
	var parentNodeExecutionID uint
	var sourceExecutionID uint
	parentNodeExecutionID, sourceExecutionID, err = m.getInheritedExecMetadata(ctx, requestSpec, &workflowExecutionID)
	if err != nil {
		return nil, nil, nil, err
	}

	var taskResources workflowengineInterfaces.TaskResources
	if requestSpec.TaskResourceAttributes != nil {
		taskResources = workflowengineInterfaces.TaskResources{
			Defaults: util.FromAdminProtoTaskResourceSpec(ctx, requestSpec.TaskResourceAttributes.Defaults),
			Limits:   util.FromAdminProtoTaskResourceSpec(ctx, requestSpec.TaskResourceAttributes.Limits),
		}
	} else {
		taskResources = util.GetTaskResources(ctx, workflow.Id, m.resourceManager, m.config.TaskResourceConfiguration())
	}
	// Dynamically assign task resource defaults.
	for _, task := range workflow.Closure.CompiledWorkflow.Tasks {
		m.setCompiledTaskDefaults(ctx, task, taskResources)
	}

	// Dynamically assign execution queues.
	m.populateExecutionQueue(ctx, *workflow.Id, workflow.Closure.CompiledWorkflow)

	executionConfig, err := m.getExecutionConfig(ctx, &request, launchPlan)
	if err != nil {
		return nil, nil, nil, err
	}

	namespace := common.GetNamespaceName(
		m.config.NamespaceMappingConfiguration().GetNamespaceTemplate(), workflowExecutionID.Org, workflowExecutionID.Project, workflowExecutionID.Domain)

	labels, err := resolveStringMap(executionConfig.GetLabels(), launchPlan.Spec.Labels, "labels", m.config.RegistrationValidationConfiguration().GetMaxLabelEntries())
	if err != nil {
		return nil, nil, nil, err
	}
	labels, err = m.addProjectLabels(ctx, request.Project, request.Org, labels)
	if err != nil {
		return nil, nil, nil, err
	}
	annotations, err := resolveStringMap(executionConfig.GetAnnotations(), launchPlan.Spec.Annotations, "annotations", m.config.RegistrationValidationConfiguration().GetMaxAnnotationEntries())
	if err != nil {
		return nil, nil, nil, err
	}
	var rawOutputDataConfig *admin.RawOutputDataConfig
	if executionConfig.RawOutputDataConfig != nil {
		rawOutputDataConfig = executionConfig.RawOutputDataConfig
	}

	clusterAssignment, err := m.getClusterAssignment(ctx, &request)
	if err != nil {
		return nil, nil, nil, err
	}

	var executionClusterLabel *admin.ExecutionClusterLabel
	if requestSpec.ExecutionClusterLabel != nil {
		executionClusterLabel = requestSpec.ExecutionClusterLabel
	}

	externalResourceAttributes, err := m.getExternalResourceAttributes(ctx, &request)
	if err != nil {
		return nil, nil, nil, err
	}

	executionParameters := workflowengineInterfaces.ExecutionParameters{
		Inputs:                     executionInputs,
		AcceptedAt:                 requestedAt,
		Labels:                     labels,
		Annotations:                annotations,
		ExecutionConfig:            executionConfig,
		TaskResources:              &taskResources,
		EventVersion:               m.config.ApplicationConfiguration().GetTopLevelConfig().EventVersion,
		RoleNameKey:                m.config.ApplicationConfiguration().GetTopLevelConfig().RoleNameKey,
		RawOutputDataConfig:        rawOutputDataConfig,
		ClusterAssignment:          clusterAssignment,
		ExecutionClusterLabel:      executionClusterLabel,
		ExternalResourceAttributes: externalResourceAttributes,
	}

	overrides, err := m.addPluginOverrides(ctx, &workflowExecutionID, launchPlan.GetSpec().WorkflowId.Name, launchPlan.Id.Name)
	if err != nil {
		return nil, nil, nil, err
	}
	if overrides != nil {
		executionParameters.TaskPluginOverrides = overrides
	}

	if request.Spec.Metadata != nil && request.Spec.Metadata.ReferenceExecution != nil &&
		request.Spec.Metadata.Mode == admin.ExecutionMetadata_RECOVERED {
		executionParameters.RecoveryExecution = request.Spec.Metadata.ReferenceExecution
	}

	executionCreatedAt := time.Now()
	acceptanceDelay := executionCreatedAt.Sub(requestedAt)

	// Request notification settings takes precedence over the launch plan settings.
	// If there is no notification in the request and DisableAll is not true, use the settings from the launch plan.
	var notificationsSettings []*admin.Notification
	if launchPlan.Spec.GetEntityMetadata() != nil {
		notificationsSettings = launchPlan.Spec.EntityMetadata.GetNotifications()
	}
	if requestSpec.GetNotifications() != nil && requestSpec.GetNotifications().Notifications != nil &&
		len(requestSpec.GetNotifications().Notifications) > 0 {
		notificationsSettings = requestSpec.GetNotifications().Notifications
	} else if requestSpec.GetDisableAll() {
		notificationsSettings = make([]*admin.Notification, 0)
	}

	// Publish of event is also gated on the artifact client being available, even though it's not directly required.
	// TODO: Artifact feature gate, remove when ready
	if m.artifactRegistry.GetClient() != nil {
		// TODO: Add principal
		m.publishExecutionStart(ctx, workflowExecutionID, request.Spec.LaunchPlan, workflow.Id, artifactTrackers, usedArtifactIDs)
	}

	resolvedSpec := proto.Clone(request.Spec).(*admin.ExecutionSpec)
	resolvedSpec = completeResolvedSpec(resolvedSpec, &executionParameters)

	createExecModelInput := transformers.CreateExecutionModelInput{
		WorkflowExecutionID: workflowExecutionID,
		RequestSpec:         requestSpec,
		ResolvedSpec:        resolvedSpec,
		LaunchPlanID:        launchPlanModel.ID,
		WorkflowID:          launchPlanModel.WorkflowID,
		// The execution is not considered running until the propeller sends a specific event saying so.
		CreatedAt:             m._clock.Now(),
		Notifications:         notificationsSettings,
		WorkflowIdentifier:    workflow.Id,
		ParentNodeExecutionID: parentNodeExecutionID,
		SourceExecutionID:     sourceExecutionID,
		InputsURI:             inputsURI,
		UserInputsURI:         userInputsURI,
		SecurityContext:       executionConfig.SecurityContext,
		LaunchEntity:          launchPlan.Id.ResourceType,
		Namespace:             namespace,
	}

	if err := m.validateActiveExecutions(ctx); err != nil {
		return nil, nil, nil, err
	}

	workflowExecutor := plugins.Get[workflowengineInterfaces.WorkflowExecutor](m.pluginRegistry, plugins.PluginIDWorkflowExecutor)
	execInfo, execErr := workflowExecutor.Execute(ctx, workflowengineInterfaces.ExecutionData{
		Namespace:                namespace,
		ExecutionID:              &workflowExecutionID,
		ReferenceWorkflowName:    workflow.Id.Name,
		ReferenceLaunchPlanName:  launchPlan.Id.Name,
		WorkflowClosure:          workflow.Closure.CompiledWorkflow,
		WorkflowClosureReference: storage.DataReference(workflowModel.RemoteClosureIdentifier),
		ExecutionParameters:      executionParameters,
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

// This function will store the values for an execution in the execution spec.
// Be sure to update this function when adding new fields to the execution spec.
func completeResolvedSpec(spec *admin.ExecutionSpec, executionParameters *workflowengineInterfaces.ExecutionParameters) *admin.ExecutionSpec {
	// We can skip fields that are in input spec, since they are already copied when proto.clone() is called.
	// Skipped: LaunchPlan, Metadata, NotificationOverrides, tags, sub_node_ids
	spec.Labels = &admin.Labels{Values: executionParameters.Labels}
	spec.Annotations = &admin.Annotations{Values: executionParameters.Annotations}
	spec.SecurityContext = executionParameters.ExecutionConfig.SecurityContext
	spec.MaxParallelism = executionParameters.ExecutionConfig.MaxParallelism
	spec.RawOutputDataConfig = executionParameters.RawOutputDataConfig
	spec.ClusterAssignment = executionParameters.ClusterAssignment
	spec.Interruptible = executionParameters.ExecutionConfig.Interruptible
	if spec.GetInterruptible() == nil {
		spec.Interruptible = &wrapperspb.BoolValue{
			Value: false,
		}
	}
	spec.OverwriteCache = executionParameters.ExecutionConfig.OverwriteCache
	spec.Envs = executionParameters.ExecutionConfig.Envs
	spec.ExecutionClusterLabel = executionParameters.ExecutionClusterLabel
	spec.ExecutionEnvAssignments = executionParameters.ExecutionConfig.ExecutionEnvAssignments
	spec.TaskResourceAttributes = &admin.TaskResourceAttributes{
		Defaults: util.ToAdminProtoTaskResourceSpec(&executionParameters.TaskResources.Defaults),
		Limits:   util.ToAdminProtoTaskResourceSpec(&executionParameters.TaskResources.Limits),
	}
	return spec
}

// publishExecutionStart is an event that Admin publishes for artifact lineage.
func (m *ExecutionManager) publishExecutionStart(ctx context.Context, executionID core.WorkflowExecutionIdentifier,
	launchPlanID *core.Identifier, workflowID *core.Identifier, artifactTrackers map[string]string, usedArtifactIDs []*core.ArtifactID) {

	var artifactTrackerList []string
	// Use a list instead of the fake set
	for k := range artifactTrackers {
		artifactTrackerList = append(artifactTrackerList, k)
	}

	if len(artifactTrackerList) > 0 || len(usedArtifactIDs) > 0 {
		logger.Debugf(ctx, "Sending execution start event for execution [%+v] with trackers [%+v] and artifact ids [%+v]", executionID, artifactTrackerList, usedArtifactIDs)

		request := event.CloudEventExecutionStart{
			ExecutionId:      &executionID,
			LaunchPlanId:     launchPlanID,
			WorkflowId:       workflowID,
			ArtifactIds:      usedArtifactIDs,
			ArtifactTrackers: artifactTrackerList,
		}
		go func() {
			ceCtx := context.TODO()
			if err := m.cloudEventPublisher.Publish(ceCtx, proto.MessageName(&request), &request); err != nil {
				m.systemMetrics.PublishEventError.Inc()
				logger.Infof(ctx, "error publishing cloud event [%+v] with err: [%v]", request, err)
			}
		}()
	}
}

// Inserts an execution model into the database store and emits platform metrics.
func (m *ExecutionManager) createExecutionModel(
	ctx context.Context, executionModel *models.Execution, executionTagModel []*models.ExecutionTag) (*core.WorkflowExecutionIdentifier, error) {
	workflowExecutionIdentifier := core.WorkflowExecutionIdentifier{
		Org:     executionModel.ExecutionKey.Org,
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
	return &workflowExecutionIdentifier, nil
}

func (m *ExecutionManager) CreateExecution(
	ctx context.Context, request admin.ExecutionCreateRequest, requestedAt time.Time) (
	*admin.ExecutionCreateResponse, error) {

	// Prior to flyteidl v0.15.0, Inputs was held in ExecutionSpec. Ensure older clients continue to work.
	if request.Inputs == nil || len(request.Inputs.Literals) == 0 {
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
	ctx context.Context, request admin.ExecutionRelaunchRequest, requestedAt time.Time) (
	*admin.ExecutionCreateResponse, error) {
	existingExecutionModel, err := util.GetExecutionModel(ctx, m.db, *request.Id)
	if err != nil {
		logger.Debugf(ctx, "Failed to get execution model for request [%+v] with err %v", request, err)
		return nil, err
	}
	existingExecution, err := transformers.FromExecutionModel(ctx, *existingExecutionModel, transformers.DefaultExecutionTransformerOptions)
	if err != nil {
		return nil, err
	}

	executionSpec := existingExecution.Spec
	if executionSpec.Metadata == nil {
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
		inputs = spec.Inputs
	}
	executionSpec.Metadata.Mode = admin.ExecutionMetadata_RELAUNCH
	executionSpec.Metadata.ReferenceExecution = existingExecution.Id
	executionSpec.OverwriteCache = request.GetOverwriteCache()
	var executionModel *models.Execution
	var executionTagModel []*models.ExecutionTag
	ctx, executionModel, executionTagModel, err = m.launchExecutionAndPrepareModel(ctx, admin.ExecutionCreateRequest{
		Org:     request.Id.Org,
		Project: request.Id.Project,
		Domain:  request.Id.Domain,
		Name:    request.Name,
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
	logger.Debugf(ctx, "Successfully relaunched [%+v] as [%+v]", request.Id, workflowExecutionIdentifier)
	return &admin.ExecutionCreateResponse{
		Id: workflowExecutionIdentifier,
	}, nil
}

func (m *ExecutionManager) RecoverExecution(
	ctx context.Context, request admin.ExecutionRecoverRequest, requestedAt time.Time) (
	*admin.ExecutionCreateResponse, error) {
	existingExecutionModel, err := util.GetExecutionModel(ctx, m.db, *request.Id)
	if err != nil {
		logger.Debugf(ctx, "Failed to get execution model for request [%+v] with err %v", request, err)
		return nil, err
	}
	existingExecution, err := transformers.FromExecutionModel(ctx, *existingExecutionModel, transformers.DefaultExecutionTransformerOptions)
	if err != nil {
		return nil, err
	}

	executionSpec := existingExecution.Spec
	if executionSpec.Metadata == nil {
		executionSpec.Metadata = &admin.ExecutionMetadata{}
	}
	var inputs *core.LiteralMap
	if len(existingExecutionModel.UserInputsURI) > 0 {
		inputs = &core.LiteralMap{}
		if err := m.storageClient.ReadProtobuf(ctx, existingExecutionModel.UserInputsURI, inputs); err != nil {
			return nil, err
		}
	}
	if request.Metadata != nil {
		executionSpec.Metadata.ParentNodeExecution = request.Metadata.ParentNodeExecution
	}
	executionSpec.Metadata.Mode = admin.ExecutionMetadata_RECOVERED
	executionSpec.Metadata.ReferenceExecution = existingExecution.Id
	var executionModel *models.Execution
	var executionTagModel []*models.ExecutionTag
	ctx, executionModel, executionTagModel, err = m.launchExecutionAndPrepareModel(ctx, admin.ExecutionCreateRequest{
		Org:     request.Id.Org,
		Project: request.Id.Project,
		Domain:  request.Id.Domain,
		Name:    request.Name,
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
	logger.Infof(ctx, "Successfully recovered [%+v] as [%+v]", request.Id, workflowExecutionIdentifier)
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
				"[%s/%s/%s/%s]", executionModel.Org, executionModel.Project, executionModel.Domain, executionModel.Name)
		return
	}
	launchPlan, err := util.GetLaunchPlan(context.Background(), m.db, *execution.Spec.LaunchPlan)
	if err != nil {
		logger.Warningf(context.Background(),
			"failed to find launch plan when emitting scheduled workflow execution stats with for "+
				"execution: [%+v] and launch plan [%+v]", execution.Id, execution.Spec.LaunchPlan)
		return
	}

	if launchPlan.Spec.EntityMetadata == nil ||
		launchPlan.Spec.EntityMetadata.Schedule == nil ||
		launchPlan.Spec.EntityMetadata.Schedule.KickoffTimeInputArg == "" {
		// Kickoff time arguments aren't always required for scheduled workflows.
		logger.Debugf(context.Background(), "no kickoff time to report for scheduled workflow execution [%+v]",
			execution.Id)
		return
	}

	var inputs core.LiteralMap
	err = m.storageClient.ReadProtobuf(ctx, executionModel.InputsURI, &inputs)
	if err != nil {
		logger.Errorf(ctx, "Failed to find inputs for emitting schedule delay event from uri: [%v]", executionModel.InputsURI)
		return
	}
	scheduledKickoffTimeProto := inputs.Literals[launchPlan.Spec.EntityMetadata.Schedule.KickoffTimeInputArg]
	if scheduledKickoffTimeProto == nil || scheduledKickoffTimeProto.GetScalar() == nil ||
		scheduledKickoffTimeProto.GetScalar().GetPrimitive() == nil ||
		scheduledKickoffTimeProto.GetScalar().GetPrimitive().GetDatetime() == nil {
		logger.Warningf(context.Background(),
			"failed to find scheduled kickoff time datetime value for scheduled workflow execution [%+v] "+
				"although one was expected", execution.Id)
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

	domainCounterMap, ok := m.userMetrics.ScheduledExecutionDelays[execution.Id.Project]
	if !ok {
		domainCounterMap = make(map[string]*promutils.StopWatch)
		m.userMetrics.ScheduledExecutionDelays[execution.Id.Project] = domainCounterMap
	}

	var watch *promutils.StopWatch
	watch, ok = domainCounterMap[execution.Id.Domain]
	if !ok {
		newWatch, err := m.systemMetrics.Scope.NewSubScope(execution.Id.Project).NewSubScope(execution.Id.Domain).NewStopWatch(
			"scheduled_execution_delay",
			"delay between scheduled execution time and time execution was observed running",
			time.Nanosecond)
		if err != nil {
			// Could be related to a concurrent exception.
			logger.Debugf(context.Background(),
				"failed to emit scheduled workflow execution delay stat, couldn't find or create counter")
			return
		}
		watch = &newWatch
		domainCounterMap[execution.Id.Domain] = watch
	}
	watch.Observe(scheduledKickoffTime, runningEventTime)
}

func (m *ExecutionManager) emitOverallWorkflowExecutionTime(
	executionModel *models.Execution, terminalEventTimeProto *timestamp.Timestamp) {
	if executionModel == nil || terminalEventTimeProto == nil {
		logger.Warningf(context.Background(),
			"tried to calculate scheduled workflow execution stats with a nil execution or event time")
		return
	}

	domainCounterMap, ok := m.userMetrics.WorkflowExecutionDurations[executionModel.Project]
	if !ok {
		domainCounterMap = make(map[string]*promutils.StopWatch)
		m.userMetrics.WorkflowExecutionDurations[executionModel.Project] = domainCounterMap
	}

	var watch *promutils.StopWatch
	watch, ok = domainCounterMap[executionModel.Domain]
	if !ok {
		newWatch, err := m.systemMetrics.Scope.NewSubScope(executionModel.Project).NewSubScope(executionModel.Domain).NewStopWatch(
			"workflow_execution_duration",
			"overall time from when when a workflow create request was sent to k8s to the workflow terminating",
			time.Nanosecond)
		if err != nil {
			// Could be related to a concurrent exception.
			logger.Debugf(context.Background(),
				"failed to emit workflow execution duration stat, couldn't find or create counter")
			return
		}
		watch = &newWatch
		domainCounterMap[executionModel.Domain] = watch
	}

	terminalEventTime, err := ptypes.Timestamp(terminalEventTimeProto)
	if err != nil {
		// Timestamps are always sent from propeller and should always be valid
		return
	}

	if executionModel.ExecutionCreatedAt == nil {
		logger.Warningf(context.Background(), "found execution with nil ExecutionCreatedAt: [%s/%s/%s/%s]",
			executionModel.Org, executionModel.Project, executionModel.Domain, executionModel.Name)
		return
	}
	watch.Observe(*executionModel.ExecutionCreatedAt, terminalEventTime)
}

func (m *ExecutionManager) CreateWorkflowEvent(ctx context.Context, request admin.WorkflowExecutionEventRequest) (
	*admin.WorkflowExecutionEventResponse, error) {
	err := validation.ValidateCreateWorkflowEventRequest(request, m.config.ApplicationConfiguration().GetRemoteDataConfig().MaxSizeInBytes)
	if err != nil {
		logger.Debugf(ctx, "received invalid CreateWorkflowEventRequest [%s]: %v", request.RequestId, err)
		return nil, err
	}
	ctx = getExecutionContext(ctx, request.Event.ExecutionId)
	logger.Debugf(ctx, "Received workflow execution event for [%+v] transitioning to phase [%v]",
		request.Event.ExecutionId, request.Event.Phase)

	executionModel, err := util.GetExecutionModel(ctx, m.db, *request.Event.ExecutionId)
	if err != nil {
		logger.Debugf(ctx, "failed to find execution [%+v] for recorded event [%s]: %v",
			request.Event.ExecutionId, request.RequestId, err)
		return nil, err
	}

	wfExecPhase := core.WorkflowExecution_Phase(core.WorkflowExecution_Phase_value[executionModel.Phase])
	// Subsequent queued events announcing a cluster reassignment are permitted.
	if request.Event.Phase != core.WorkflowExecution_QUEUED {
		if wfExecPhase == request.Event.Phase {
			logger.Debugf(ctx, "This phase %s was already recorded for workflow execution %v",
				wfExecPhase.String(), request.Event.ExecutionId)
			return nil, errors.NewFlyteAdminErrorf(codes.AlreadyExists,
				"This phase %s was already recorded for workflow execution %v",
				wfExecPhase.String(), request.Event.ExecutionId)
		} else if err := validation.ValidateCluster(ctx, executionModel.Cluster, request.Event.ProducerId); err != nil {
			// Only perform event cluster validation **after** an execution has moved on from QUEUED.
			return nil, err
		}
	}

	if common.IsExecutionTerminal(wfExecPhase) {
		// Cannot go backwards in time from a terminal state to anything else
		curPhase := wfExecPhase.String()
		errorMsg := fmt.Sprintf("Invalid phase change from %s to %s for workflow execution %v", curPhase, request.Event.Phase.String(), request.Event.ExecutionId)
		return nil, errors.NewAlreadyInTerminalStateError(ctx, errorMsg, curPhase)
	} else if wfExecPhase == core.WorkflowExecution_RUNNING && request.Event.Phase == core.WorkflowExecution_QUEUED {
		// Cannot go back in time from RUNNING -> QUEUED
		return nil, errors.NewFlyteAdminErrorf(codes.FailedPrecondition,
			"Cannot go from %s to %s for workflow execution %v",
			wfExecPhase.String(), request.Event.Phase.String(), request.Event.ExecutionId)
	} else if wfExecPhase == core.WorkflowExecution_ABORTING && !common.IsExecutionTerminal(request.Event.Phase) {
		return nil, errors.NewFlyteAdminErrorf(codes.FailedPrecondition,
			"Invalid phase change from aborting to %s for workflow execution %v", request.Event.Phase.String(), request.Event.ExecutionId)
	}

	err = transformers.UpdateExecutionModelState(ctx, executionModel, request, m.config.ApplicationConfiguration().GetRemoteDataConfig().InlineEventDataPolicy, m.storageClient)
	if err != nil {
		logger.Debugf(ctx, "failed to transform updated workflow execution model [%+v] after receiving event with err: %v",
			request.Event.ExecutionId, err)
		return nil, err
	}
	err = m.db.ExecutionRepo().Update(ctx, *executionModel)
	if err != nil {
		logger.Debugf(ctx, "Failed to update execution with CreateWorkflowEvent [%+v] with err %v",
			request, err)
		return nil, err
	}
	m.dbEventWriter.Write(request)

	if request.Event.Phase == core.WorkflowExecution_RUNNING {
		// Workflow executions are created in state "UNDEFINED". All the time up until a RUNNING event is received is
		// considered system-induced delay.
		if executionModel.Mode == int32(admin.ExecutionMetadata_SCHEDULED) {
			go m.emitScheduledWorkflowMetrics(ctx, executionModel, request.Event.OccurredAt)
		}
	} else if common.IsExecutionTerminal(request.Event.Phase) {
		if request.Event.Phase == core.WorkflowExecution_FAILED {
			// request.Event is expected to be of type WorkflowExecutionEvent_Error when workflow fails.
			// if not, log the error and continue
			if err := request.Event.GetError(); err != nil {
				ctx = context.WithValue(ctx, common.ErrorKindKey, err.Kind.String())
			} else {
				logger.Warning(ctx, "Failed to parse error for FAILED request [%+v]", request)
			}
		}

		m.systemMetrics.ActiveExecutions.Dec()
		m.systemMetrics.ExecutionsTerminated.Inc(contextutils.WithPhase(ctx, request.Event.Phase.String()))
		go m.emitOverallWorkflowExecutionTime(executionModel, request.Event.OccurredAt)
		if request.Event.GetOutputData() != nil {
			m.userMetrics.WorkflowExecutionOutputBytes.Observe(float64(proto.Size(request.Event.GetOutputData())))
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

	if err := m.eventPublisher.Publish(ctx, proto.MessageName(&request), &request); err != nil {
		m.systemMetrics.PublishEventError.Inc()
		logger.Infof(ctx, "error publishing event [%+v] with err: [%v]", request.RequestId, err)
	}

	go func() {
		ceCtx := context.TODO()
		if err := m.cloudEventPublisher.Publish(ceCtx, proto.MessageName(&request), &request); err != nil {
			m.systemMetrics.PublishEventError.Inc()
			logger.Infof(ctx, "error publishing cloud event [%+v] with err: [%v]", request.RequestId, err)
		}
	}()

	return &admin.WorkflowExecutionEventResponse{}, nil
}

func (m *ExecutionManager) GetExecution(
	ctx context.Context, request admin.WorkflowExecutionGetRequest) (*admin.Execution, error) {
	if err := validation.ValidateWorkflowExecutionIdentifier(request.Id); err != nil {
		logger.Debugf(ctx, "GetExecution request [%+v] failed validation with err: %v", request, err)
		return nil, err
	}
	ctx = getExecutionContext(ctx, request.Id)
	executionModel, err := util.GetExecutionModel(ctx, m.db, *request.Id)
	if err != nil {
		logger.Debugf(ctx, "Failed to get execution model for request [%+v] with err: %v", request, err)
		return nil, err
	}
	namespace := common.GetNamespaceName(
		m.config.NamespaceMappingConfiguration().GetNamespaceTemplate(), request.GetId().GetOrg(), request.GetId().GetProject(), request.GetId().GetDomain())
	execution, transformerErr := transformers.FromExecutionModel(ctx, *executionModel, &transformers.ExecutionTransformerOptions{
		DefaultNamespace: namespace,
	})
	if transformerErr != nil {
		logger.Debugf(ctx, "Failed to transform execution model [%+v] to proto object with err: %v", request.Id,
			transformerErr)
		return nil, transformerErr
	}

	return execution, nil
}

func (m *ExecutionManager) UpdateExecution(ctx context.Context, request admin.ExecutionUpdateRequest,
	requestedAt time.Time) (*admin.ExecutionUpdateResponse, error) {
	if err := validation.ValidateWorkflowExecutionIdentifier(request.Id); err != nil {
		logger.Debugf(ctx, "UpdateExecution request [%+v] failed validation with err: %v", request, err)
		return nil, err
	}
	ctx = getExecutionContext(ctx, request.Id)
	executionModel, err := util.GetExecutionModel(ctx, m.db, *request.Id)
	if err != nil {
		logger.Debugf(ctx, "Failed to get execution model for request [%+v] with err: %v", request, err)
		return nil, err
	}

	if err = transformers.UpdateExecutionModelStateChangeDetails(executionModel, request.State, requestedAt,
		getUser(ctx)); err != nil {
		return nil, err
	}

	if err := m.db.ExecutionRepo().Update(ctx, *executionModel); err != nil {
		return nil, err
	}

	return &admin.ExecutionUpdateResponse{}, nil
}

func (m *ExecutionManager) GetExecutionData(
	ctx context.Context, request admin.WorkflowExecutionGetDataRequest) (*admin.WorkflowExecutionGetDataResponse, error) {
	ctx = getExecutionContext(ctx, request.Id)
	executionModel, err := util.GetExecutionModel(ctx, m.db, *request.Id)
	if err != nil {
		logger.Debugf(ctx, "Failed to get execution model for request [%+v] with err: %v", request, err)
		return nil, err
	}
	execution, err := transformers.FromExecutionModel(ctx, *executionModel, transformers.DefaultExecutionTransformerOptions)
	if err != nil {
		logger.Debugf(ctx, "Failed to transform execution model [%+v] to proto object with err: %v", request.Id, err)
		return nil, err
	}
	// Prior to flyteidl v0.15.0, Inputs were held in ExecutionClosure and were not offloaded. Ensure we can return the inputs as expected.
	if len(executionModel.InputsURI) == 0 {
		closure := &admin.ExecutionClosure{}
		// We must not use the FromExecutionModel method because it empties deprecated fields.
		if err := proto.Unmarshal(executionModel.Closure, closure); err != nil {
			return nil, err
		}
		newInputsURI, err := common.OffloadLiteralMap(ctx, m.storageClient, closure.ComputedInputs, request.Id.Org, request.Id.Project, request.Id.Domain, request.Id.Name, shared.Inputs)
		if err != nil {
			return nil, err
		}
		// Update model so as not to offload again.
		executionModel.InputsURI = newInputsURI
		if err := m.db.ExecutionRepo().Update(ctx, *executionModel); err != nil {
			return nil, err
		}
	}

	cluster := execution.GetSpec().GetMetadata().GetSystemMetadata().GetExecutionCluster()

	id := request.GetId()
	objectStore := plugins.Get[util.ObjectStore](m.pluginRegistry, plugins.PluginIDObjectStore)
	var inputs *core.LiteralMap
	var inputURLBlob *admin.UrlBlob
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		var err error
		inputs, inputURLBlob, err = util.GetInputs(groupCtx,
			m.urlData,
			m.config.ApplicationConfiguration().GetRemoteDataConfig(),
			m.storageClient,
			cluster,
			id.Org,
			id.Project,
			id.Domain,
			executionModel.InputsURI.String(),
			objectStore,
		)
		if err != nil {
			logger.Errorf(ctx, "failed to read inputs during execution [%v]: %v", id, err)
		}
		return err
	})

	var outputs *core.LiteralMap
	var outputURLBlob *admin.UrlBlob
	group.Go(func() error {
		var err error
		outputs, outputURLBlob, err = util.GetOutputs(groupCtx,
			m.urlData,
			m.config.ApplicationConfiguration().GetRemoteDataConfig(),
			m.storageClient,
			util.ToExecutionClosureInterface(execution.Closure),
			cluster,
			id.Org,
			id.Project,
			id.Domain,
			objectStore,
		)
		if err != nil {
			logger.Errorf(ctx, "failed to read outputs during execution [%v]: %v", id, err)
		}
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

	m.userMetrics.WorkflowExecutionInputBytes.Observe(float64(response.Inputs.Bytes))
	if response.Outputs.Bytes > 0 {
		m.userMetrics.WorkflowExecutionOutputBytes.Observe(float64(response.Outputs.Bytes))
	} else if response.FullOutputs != nil {
		m.userMetrics.WorkflowExecutionOutputBytes.Observe(float64(proto.Size(response.FullOutputs)))
	}
	return response, nil
}

func (m *ExecutionManager) addDefaultRequestFilters(ctx context.Context, filters []common.InlineFilter) []common.InlineFilter {
	sawModeFilter := false
	for _, filter := range filters {
		if filter.GetField() == shared.Mode {
			sawModeFilter = true
			break
		}
	}
	if sawModeFilter {
		return filters
	}
	stateFilter, err := common.NewInlineFilter(common.Execution, "ne", shared.Mode, admin.ExecutionMetadata_WORKSPACE)
	if err != nil {
		logger.Errorf(ctx, "failed to generate mode filter to exclude workspace executions: %v", err)
	}
	filters = append(filters, stateFilter)
	return filters
}

func (m *ExecutionManager) ListExecutions(
	ctx context.Context, request admin.ResourceListRequest) (*admin.ExecutionList, error) {
	// Check required fields
	if err := validation.ValidateResourceListRequest(request); err != nil {
		logger.Debugf(ctx, "ListExecutions request [%+v] failed validation with err: %v", request, err)
		return nil, err
	}
	ctx = contextutils.WithProjectDomain(ctx, request.Id.Project, request.Id.Domain)
	filters, err := util.GetDbFilters(util.FilterSpec{
		Org:            request.Id.Org,
		Project:        request.Id.Project,
		Domain:         request.Id.Domain,
		Name:           request.Id.Name, // Optional, may be empty.
		RequestFilters: request.Filters,
	}, common.Execution)
	if err != nil {
		return nil, err
	}
	filters = m.addDefaultRequestFilters(ctx, filters)

	sortParameter, err := common.NewSortParameter(request.SortBy, models.ExecutionColumns)
	if err != nil {
		return nil, err
	}

	offset, err := validation.ValidateToken(request.Token)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "invalid pagination token %s for ListExecutions",
			request.Token)
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
		Limit:             int(request.Limit),
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
	if len(executionList) == int(request.Limit) {
		token = strconv.Itoa(offset + len(executionList))
	}
	return &admin.ExecutionList{
		Executions: executionList,
		Token:      token,
	}, nil
}

func (m *ExecutionManager) GetExecutionCounts(
	ctx context.Context, request admin.ExecutionCountsGetRequest) (*admin.ExecutionCountsGetResponse, error) {
	// Check required fields
	if err := validation.ValidateExecutionCountsGetRequest(request); err != nil {
		logger.Debugf(ctx, "ExecutionCounts request [%+v] failed validation with err: %v", request, err)
		return nil, err
	}
	ctx = contextutils.WithProjectDomain(ctx, request.Project, request.Domain)
	filters, err := util.GetDbFilters(util.FilterSpec{
		Org:            request.Org,
		Project:        request.Project,
		Domain:         request.Domain,
		RequestFilters: request.Filters,
	}, common.Execution)
	if err != nil {
		return nil, err
	}

	countExecutionByPhaseInput := repositoryInterfaces.CountResourceInput{
		InlineFilters: filters,
	}
	countExecutionByPhaseOutput, err := m.db.ExecutionRepo().CountByPhase(ctx, countExecutionByPhaseInput)
	if err != nil {
		logger.Debugf(ctx, "Failed to get execution counts using input [%+v] with err %v", countExecutionByPhaseInput, err)
		return nil, err
	}

	executionCounts, err := transformers.FromExecutionCountsByPhase(ctx, countExecutionByPhaseOutput)
	if err != nil {
		logger.Errorf(ctx, "Failed to transform execution by phase output [%+v] with err %v", countExecutionByPhaseOutput, err)
		return nil, err
	}

	return &admin.ExecutionCountsGetResponse{
		ExecutionCounts: executionCounts,
	}, nil
}

func (m *ExecutionManager) GetRunningExecutionsCount(
	ctx context.Context, request admin.RunningExecutionsCountGetRequest) (*admin.RunningExecutionsCountGetResponse, error) {
	// Check required fields
	if err := validation.ValidateRunningExecutionsGetRequest(request); err != nil {
		logger.Debugf(ctx, "RunningExecutionsCount request [%+v] failed validation with err: %v", request, err)
		return nil, err
	}
	ctx = contextutils.WithProjectDomain(ctx, request.Project, request.Domain)
	filters, err := util.GetDbFilters(util.FilterSpec{
		Org:     request.Org,
		Project: request.Project,
		Domain:  request.Domain,
	}, common.Execution)
	if err != nil {
		return nil, err
	}

	// Add filter to fetch only RUNNING executions
	if filters, err = addPhaseFilter(filters, core.WorkflowExecution_RUNNING); err != nil {
		return nil, err
	}

	countRunningExecutionsInput := repositoryInterfaces.CountResourceInput{
		InlineFilters: filters,
	}
	countRunningExecutionsOutput, err := m.db.ExecutionRepo().Count(ctx, countRunningExecutionsInput)
	if err != nil {
		logger.Debugf(ctx, "Failed to get running executions count using input [%+v] with err %v", countRunningExecutionsOutput, err)
		return nil, err
	}

	return &admin.RunningExecutionsCountGetResponse{
		Count: countRunningExecutionsOutput,
	}, nil
}

// publishNotifications will only forward major errors because the assumption made is all of the objects
// that are being manipulated have already been validated/manipulated by Flyte itself.
// Note: This method should be refactored somewhere else once the interaction with pushing to SNS.
func (m *ExecutionManager) publishNotifications(ctx context.Context, request admin.WorkflowExecutionEventRequest,
	execution models.Execution) error {
	// Notifications are stored in the Spec object of an admin.Execution object.
	adminExecution, err := transformers.FromExecutionModel(ctx, execution, transformers.DefaultExecutionTransformerOptions)
	if err != nil {
		// This shouldn't happen because execution manager marshaled the data into models.Execution.
		m.systemMetrics.TransformerError.Inc()
		return errors.NewFlyteAdminErrorf(codes.Internal, "Failed to transform execution [%+v] with err: %v", request.Event.ExecutionId, err)
	}
	var notificationsList = adminExecution.Closure.Notifications
	logger.Debugf(ctx, "publishing notifications for execution [%+v] in state [%+v] for notifications [%+v]",
		request.Event.ExecutionId, request.Event.Phase, notificationsList)
	for _, notification := range notificationsList {
		// Check if the notification phase matches the current one.
		var matchPhase = false
		for _, phase := range notification.Phases {
			if phase == request.Event.Phase {
				matchPhase = true
			}
		}

		// The current phase doesn't match; no notifications will be sent for the current notification option.
		if !matchPhase {
			continue
		}

		// Currently all three supported notifications use email underneath to send the notification.
		// Convert Slack and PagerDuty into an EmailNotification type.
		var emailNotification admin.EmailNotification
		if notification.GetEmail() != nil {
			emailNotification.RecipientsEmail = notification.GetEmail().GetRecipientsEmail()
		} else if notification.GetPagerDuty() != nil {
			emailNotification.RecipientsEmail = notification.GetPagerDuty().GetRecipientsEmail()
		} else if notification.GetSlack() != nil {
			emailNotification.RecipientsEmail = notification.GetSlack().GetRecipientsEmail()
		} else {
			logger.Debugf(ctx, "failed to publish notification, encountered unrecognized type: %v", notification.Type)
			m.systemMetrics.UnexpectedDataError.Inc()
			// Unsupported notification types should have been caught when the launch plan was being created.
			return errors.NewFlyteAdminErrorf(codes.Internal, "Unsupported notification type [%v] for execution [%+v]",
				notification.Type, request.Event.ExecutionId)
		}

		// Convert the email Notification into an email message to be published.
		// Currently there are no possible errors while creating an email message.
		// Once customizable content is specified, errors are possible.
		email := notifications.ToEmailMessageFromWorkflowExecutionEvent(
			*m.config.ApplicationConfiguration().GetNotificationsConfig(), emailNotification, request, adminExecution)
		// Errors seen while publishing a message are considered non-fatal to the method and will not result
		// in the method returning an error.
		if err = m.notificationClient.Publish(ctx, proto.MessageName(&emailNotification), email); err != nil {
			m.systemMetrics.PublishNotificationError.Inc()
			logger.Infof(ctx, "error publishing email notification [%+v] with err: [%v]", notification, err)
		}
	}
	return nil
}

func (m *ExecutionManager) TerminateExecution(
	ctx context.Context, request admin.ExecutionTerminateRequest) (*admin.ExecutionTerminateResponse, error) {
	if err := validation.ValidateWorkflowExecutionIdentifier(request.Id); err != nil {
		logger.Debugf(ctx, "received terminate execution request: %v with invalid identifier: %v", request, err)
		return nil, err
	}
	ctx = getExecutionContext(ctx, request.Id)
	// Save the abort reason (best effort)
	executionModel, err := m.db.ExecutionRepo().Get(ctx, repositoryInterfaces.Identifier{
		Org:     request.Id.Org,
		Project: request.Id.Project,
		Domain:  request.Id.Domain,
		Name:    request.Id.Name,
	})
	if err != nil {
		logger.Infof(ctx, "couldn't find execution [%+v] to save termination cause", request.Id)
		return nil, err
	}

	if common.IsExecutionTerminal(core.WorkflowExecution_Phase(core.WorkflowExecution_Phase_value[executionModel.Phase])) {
		return nil, errors.NewAlreadyInTerminalStateError(ctx, "Cannot abort an already terminate workflow execution", executionModel.Phase)
	}

	err = transformers.SetExecutionAborting(&executionModel, request.Cause, getUser(ctx))
	if err != nil {
		logger.Debugf(ctx, "failed to add abort metadata for execution [%+v] with err: %v", request.Id, err)
		return nil, err
	}

	err = m.db.ExecutionRepo().Update(ctx, executionModel)
	if err != nil {
		logger.Debugf(ctx, "failed to save abort cause for terminated execution: %+v with err: %v", request.Id, err)
		return nil, err
	}

	workflowExecutor := plugins.Get[workflowengineInterfaces.WorkflowExecutor](m.pluginRegistry, plugins.PluginIDWorkflowExecutor)
	err = workflowExecutor.Abort(ctx, workflowengineInterfaces.AbortData{
		Namespace: common.GetNamespaceName(
			m.config.NamespaceMappingConfiguration().GetNamespaceTemplate(), request.Id.Org, request.Id.Project, request.Id.Domain),

		ExecutionID: request.Id,
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
	}
}

func NewExecutionManager(db repositoryInterfaces.Repository, pluginRegistry *plugins.Registry, config runtimeInterfaces.Configuration,
	storageClient *storage.DataStore, systemScope promutils.Scope, userScope promutils.Scope,
	publisher notificationInterfaces.Publisher, urlData dataInterfaces.RemoteURLInterface,
	workflowManager interfaces.WorkflowInterface, launchPlanManager interfaces.LaunchPlanInterface, namedEntityManager interfaces.NamedEntityInterface,
	eventPublisher notificationInterfaces.Publisher, cloudEventPublisher cloudeventInterfaces.Publisher,
	eventWriter eventWriter.WorkflowExecutionEventWriter, artifactRegistry *artifacts.ArtifactRegistry, resourceManager interfaces.ResourceInterface) interfaces.ExecutionInterface {

	queueAllocator := executions.NewQueueAllocator(config, db, resourceManager)
	systemMetrics := newExecutionSystemMetrics(systemScope)

	userMetrics := executionUserMetrics{
		Scope:                      userScope,
		ScheduledExecutionDelays:   make(map[string]map[string]*promutils.StopWatch),
		WorkflowExecutionDurations: make(map[string]map[string]*promutils.StopWatch),
		WorkflowExecutionInputBytes: userScope.MustNewSummary("input_size_bytes",
			"size in bytes of serialized execution inputs"),
		WorkflowExecutionOutputBytes: userScope.MustNewSummary("output_size_bytes",
			"size in bytes of serialized execution outputs"),
		MaxActiveExecutionsReached: userScope.MustNewCounter("max_active_executions_reached",
			"count of times the maximum number of active executions was reached for a user org"),
	}

	return &ExecutionManager{
		db:                        db,
		config:                    config,
		storageClient:             storageClient,
		queueAllocator:            queueAllocator,
		_clock:                    clock.New(),
		systemMetrics:             systemMetrics,
		userMetrics:               userMetrics,
		notificationClient:        publisher,
		urlData:                   urlData,
		workflowManager:           workflowManager,
		launchPlanManager:         launchPlanManager,
		namedEntityManager:        namedEntityManager,
		resourceManager:           resourceManager,
		qualityOfServiceAllocator: executions.NewQualityOfServiceAllocator(config, resourceManager),
		eventPublisher:            eventPublisher,
		cloudEventPublisher:       cloudEventPublisher,
		dbEventWriter:             eventWriter,
		pluginRegistry:            pluginRegistry,
		artifactRegistry:          artifactRegistry,
	}
}

// Adds project labels with higher precedence to workflow labels. Project labels are ignored if a corresponding label is set on the workflow.
func (m *ExecutionManager) addProjectLabels(ctx context.Context, projectName, org string, initialLabels map[string]string) (map[string]string, error) {
	project, err := m.db.ProjectRepo().Get(ctx, projectName, org)
	if err != nil {
		logger.Errorf(ctx, "Failed to get project for [%+v] with error: %v", project, err)
		return nil, err
	}
	// passing nil domain as not needed to retrieve labels
	projectLabels := transformers.FromProjectModel(project, nil).Labels.GetValues()

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

func addPhaseFilter(filters []common.InlineFilter, phase core.WorkflowExecution_Phase) ([]common.InlineFilter, error) {
	phaseFilter, err := common.NewSingleValueFilter(common.Execution, common.Equal, shared.Phase,
		phase.String())
	if err != nil {
		return filters, err
	}
	filters = append(filters, phaseFilter)
	return filters, nil
}
