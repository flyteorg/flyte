package connector

import (
	"context"
	"encoding/gob"
	"fmt"
	"sync"

	"google.golang.org/protobuf/types/known/structpb"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	flyteIdl "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	pluginErrors "github.com/flyteorg/flyte/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/template"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/webapi"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

type Plugin struct {
	metricScope promutils.Scope
	cfg         *Config
	cs          *ClientSet
	deployment  Connector
	mu          sync.RWMutex
}

type ResourceWrapper struct {
	Phase flyteIdl.TaskExecution_Phase
	// Deprecated: Please Use Phase instead.
	State          admin.State
	Outputs        *flyteIdl.LiteralMap
	Message        string
	LogLinks       []*flyteIdl.TaskLog
	CustomInfo     *structpb.Struct
	ConnectorError *admin.AgentError
}

// IsTerminal is used to avoid making network calls to the connector service if the resource is already in a terminal state.
func (r ResourceWrapper) IsTerminal() bool {
	return r.Phase == flyteIdl.TaskExecution_SUCCEEDED || r.Phase == flyteIdl.TaskExecution_FAILED || r.Phase == flyteIdl.TaskExecution_ABORTED
}

type ResourceMetaWrapper struct {
	OutputPrefix          string
	ConnectorResourceMeta []byte
	TaskCategory          admin.TaskCategory
}

func (p *Plugin) GetConfig() webapi.PluginConfig {
	// Return default config if deployment is nil
	if p.deployment.ConnectorDeployment == nil {
		return p.cfg.WebAPI
	}

	// Create a new config object by copying deployment's config
	config := p.deployment.ConnectorDeployment.WebAPI

	// Check if ResourceQuotas is nil
	if config.ResourceQuotas == nil {
		config.ResourceQuotas = p.cfg.WebAPI.ResourceQuotas
	}

	// Check ReadRateLimiter values individually
	if config.ReadRateLimiter.QPS == 0 {
		config.ReadRateLimiter.QPS = p.cfg.WebAPI.ReadRateLimiter.QPS
	}
	if config.ReadRateLimiter.Burst == 0 {
		config.ReadRateLimiter.Burst = p.cfg.WebAPI.ReadRateLimiter.Burst
	}

	// Check WriteRateLimiter values individually
	if config.WriteRateLimiter.QPS == 0 {
		config.WriteRateLimiter.QPS = p.cfg.WebAPI.WriteRateLimiter.QPS
	}
	if config.WriteRateLimiter.Burst == 0 {
		config.WriteRateLimiter.Burst = p.cfg.WebAPI.WriteRateLimiter.Burst
	}

	// Check Caching configuration values individually
	if config.Caching.ResyncInterval.Duration == time.Duration(0) {
		config.Caching.ResyncInterval = p.cfg.WebAPI.Caching.ResyncInterval
	}
	if config.Caching.Size == 0 {
		config.Caching.Size = p.cfg.WebAPI.Caching.Size
	}
	if config.Caching.Workers == 0 {
		config.Caching.Workers = p.cfg.WebAPI.Caching.Workers
	}
	if config.Caching.MaxSystemFailures == 0 {
		config.Caching.MaxSystemFailures = p.cfg.WebAPI.Caching.MaxSystemFailures
	}

	// Check if ResourceMeta is nil
	if config.ResourceMeta == nil {
		config.ResourceMeta = p.cfg.WebAPI.ResourceMeta
	}

	return config
}

func (p *Plugin) ResourceRequirements(_ context.Context, _ webapi.TaskExecutionContextReader) (
	namespace core.ResourceNamespace, constraints core.ResourceConstraintsSpec, err error) {

	// Resource requirements are assumed to be the same.
	return "default", p.cfg.ResourceConstraints, nil
}

func (p *Plugin) Create(ctx context.Context, taskCtx webapi.TaskExecutionContextReader) (webapi.ResourceMeta,
	webapi.Resource, error) {
	logger.Debug(ctx, "create task for deployment %s", p.deployment.ConnectorDeployment.Endpoint)
	taskTemplate, err := taskCtx.TaskReader().Read(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read task template with error: %v", err)
	}
	inputs, err := taskCtx.InputReader().Get(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read inputs with error: %v", err)
	}

	var argTemplate []string
	if taskTemplate.GetContainer() != nil {
		templateParameters := template.Parameters{
			TaskExecMetadata: taskCtx.TaskExecutionMetadata(),
			Inputs:           taskCtx.InputReader(),
			OutputPath:       taskCtx.OutputWriter(),
			Task:             taskCtx.TaskReader(),
		}
		argTemplate = taskTemplate.GetContainer().GetArgs()
		modifiedArgs, err := template.Render(ctx, taskTemplate.GetContainer().GetArgs(), templateParameters)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to render args with error: %v", err)
		}
		taskTemplate.GetContainer().Args = modifiedArgs
		defer func() {
			// Restore unrendered template for subsequent renders.
			taskTemplate.GetContainer().Args = argTemplate
		}()
	}
	outputPrefix := taskCtx.OutputWriter().GetOutputPrefixPath().String()

	taskCategory := admin.TaskCategory{Name: taskTemplate.GetType(), Version: taskTemplate.GetTaskTypeVersion()}
	connector := p.deployment.ConnectorDeployment

	taskExecutionMetadata := buildTaskExecutionMetadata(taskCtx.TaskExecutionMetadata())

	if p.deployment.IsSync {
		finalCtx, cancel := getFinalContext(ctx, "ExecuteTaskSync", connector)
		defer cancel()
		client, err := p.getSyncConnectorClient(ctx, connector)
		if err != nil {
			return nil, nil, err
		}
		header := &admin.CreateRequestHeader{Template: taskTemplate, OutputPrefix: outputPrefix, TaskExecutionMetadata: &taskExecutionMetadata}
		return p.ExecuteTaskSync(finalCtx, client, header, inputs)
	}

	finalCtx, cancel := getFinalContext(ctx, "CreateTask", connector)
	defer cancel()

	// Use async connector client
	client, err := p.getAsyncConnectorClient(ctx, connector)
	if err != nil {
		return nil, nil, err
	}
	request := &admin.CreateTaskRequest{Inputs: inputs, Template: taskTemplate, OutputPrefix: outputPrefix, TaskExecutionMetadata: &taskExecutionMetadata}
	res, err := client.CreateTask(finalCtx, request)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create task from connector with %v", err)
	}

	return ResourceMetaWrapper{
		OutputPrefix:          outputPrefix,
		ConnectorResourceMeta: res.GetResourceMeta(),
		TaskCategory:          taskCategory,
	}, nil, nil
}

func (p *Plugin) ExecuteTaskSync(
	ctx context.Context,
	client service.SyncAgentServiceClient,
	header *admin.CreateRequestHeader,
	inputs *flyteIdl.LiteralMap,
) (webapi.ResourceMeta, webapi.Resource, error) {
	stream, err := client.ExecuteTaskSync(ctx)
	if err != nil {
		logger.Errorf(ctx, "failed to execute task from connector with %v", err)
		return nil, nil, fmt.Errorf("failed to execute task from connector with %v", err)
	}

	headerProto := &admin.ExecuteTaskSyncRequest{
		Part: &admin.ExecuteTaskSyncRequest_Header{
			Header: header,
		},
	}

	err = stream.Send(headerProto)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to send headerProto with error: %w", err)
	}
	inputsProto := &admin.ExecuteTaskSyncRequest{
		Part: &admin.ExecuteTaskSyncRequest_Inputs{
			Inputs: inputs,
		},
	}
	err = stream.Send(inputsProto)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to send inputsProto with error: %w", err)
	}

	// Client is done with sending
	if err := stream.CloseSend(); err != nil {
		logger.Errorf(ctx, "failed to close stream with err %s", err.Error())
		return nil, nil, err
	}

	in, err := stream.Recv()
	if err != nil {
		logger.Errorf(ctx, "failed to receive stream from server %s", err.Error())
		return nil, nil, fmt.Errorf("failed to receive stream from server %w", err)
	}
	if in.GetHeader() == nil {
		return nil, nil, fmt.Errorf("expected header in the response, but got none")
	}
	// TODO: Read the streaming output from the connector, and merge it into the final output.
	// For now, Propeller assumes that the output is always in the header.
	resource := in.GetHeader().GetResource()

	return nil, ResourceWrapper{
		Phase:          resource.GetPhase(),
		Outputs:        resource.GetOutputs(),
		Message:        resource.GetMessage(),
		LogLinks:       resource.GetLogLinks(),
		CustomInfo:     resource.GetCustomInfo(),
		ConnectorError: resource.GetAgentError(),
	}, nil
}

func (p *Plugin) Get(ctx context.Context, taskCtx webapi.GetContext) (latest webapi.Resource, err error) {
	metadata := taskCtx.ResourceMeta().(ResourceMetaWrapper)
	connector := p.deployment.ConnectorDeployment

	client, err := p.getAsyncConnectorClient(ctx, connector)
	if err != nil {
		return nil, err
	}
	finalCtx, cancel := getFinalContext(ctx, "GetTask", connector)
	defer cancel()

	request := &admin.GetTaskRequest{
		TaskType:     metadata.TaskCategory.GetName(),
		TaskCategory: &metadata.TaskCategory,
		ResourceMeta: metadata.ConnectorResourceMeta,
		OutputPrefix: metadata.OutputPrefix,
	}
	res, err := client.GetTask(finalCtx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get task from connector with %v", err)
	}

	return ResourceWrapper{
		Phase:      res.GetResource().GetPhase(),
		State:      res.GetResource().GetState(),
		Outputs:    res.GetResource().GetOutputs(),
		Message:    res.GetResource().GetMessage(),
		LogLinks:   res.GetResource().GetLogLinks(),
		CustomInfo: res.GetResource().GetCustomInfo(),
	}, nil
}

func (p *Plugin) Delete(ctx context.Context, taskCtx webapi.DeleteContext) error {
	if taskCtx.ResourceMeta() == nil {
		return nil
	}
	metadata := taskCtx.ResourceMeta().(ResourceMetaWrapper)
	connector := p.deployment.ConnectorDeployment

	client, err := p.getAsyncConnectorClient(ctx, connector)
	if err != nil {
		return err
	}
	finalCtx, cancel := getFinalContext(ctx, "DeleteTask", connector)
	defer cancel()

	request := &admin.DeleteTaskRequest{
		TaskType:     metadata.TaskCategory.GetName(),
		TaskCategory: &metadata.TaskCategory,
		ResourceMeta: metadata.ConnectorResourceMeta,
	}
	_, err = client.DeleteTask(finalCtx, request)
	if err != nil {
		return fmt.Errorf("failed to delete task from connector with %v", err)
	}
	return nil
}

func (p *Plugin) Status(ctx context.Context, taskCtx webapi.StatusContext) (phase core.PhaseInfo, err error) {
	resource := taskCtx.Resource().(ResourceWrapper)
	taskInfo := &core.TaskInfo{Logs: resource.LogLinks, CustomInfo: resource.CustomInfo}
	errorCode := pluginErrors.TaskFailedWithError
	if resource.ConnectorError != nil && resource.ConnectorError.GetCode() != "" {
		errorCode = resource.ConnectorError.GetCode()
	}

	switch resource.Phase {
	case flyteIdl.TaskExecution_QUEUED:
		return core.PhaseInfoQueuedWithTaskInfo(core.DefaultPhaseVersion, resource.Message, taskInfo), nil
	case flyteIdl.TaskExecution_WAITING_FOR_RESOURCES:
		return core.PhaseInfoWaitingForResourcesInfo(core.DefaultPhaseVersion, resource.Message, taskInfo), nil
	case flyteIdl.TaskExecution_INITIALIZING:
		return core.PhaseInfoInitializing(core.DefaultPhaseVersion, resource.Message, taskInfo), nil
	case flyteIdl.TaskExecution_RUNNING:
		return core.PhaseInfoRunning(core.DefaultPhaseVersion, taskInfo), nil
	case flyteIdl.TaskExecution_SUCCEEDED:
		err = writeOutput(ctx, taskCtx, resource.Outputs)
		if err != nil {
			logger.Errorf(ctx, "failed to write output with err %s", err.Error())
			return core.PhaseInfoUndefined, err
		}
		return core.PhaseInfoSuccess(taskInfo), nil
	case flyteIdl.TaskExecution_ABORTED:
		return core.PhaseInfoFailure(errorCode, "failed to run the job with aborted phase.", taskInfo), nil
	case flyteIdl.TaskExecution_FAILED:
		return core.PhaseInfoFailure(errorCode, fmt.Sprintf("failed to run the job: %s", resource.Message), taskInfo), nil
	case flyteIdl.TaskExecution_RETRYABLE_FAILED:
		return core.PhaseInfoRetryableFailure(errorCode, fmt.Sprintf("failed to run the job: %s", resource.Message), taskInfo), nil
	}
	// The default phase is undefined.
	if resource.Phase != flyteIdl.TaskExecution_UNDEFINED {
		return core.PhaseInfoUndefined, pluginErrors.Errorf(core.SystemErrorCode, "unknown execution phase [%v].", resource.Phase)
	}

	// If the phase is undefined, we will use state to determine the phase.
	switch resource.State {
	case admin.State_PENDING:
		return core.PhaseInfoInitializing(core.DefaultPhaseVersion, resource.Message, taskInfo), nil
	case admin.State_RUNNING:
		return core.PhaseInfoRunning(core.DefaultPhaseVersion, taskInfo), nil
	case admin.State_PERMANENT_FAILURE:
		return core.PhaseInfoFailure(pluginErrors.TaskFailedWithError, "failed to run the job.\n"+resource.Message, taskInfo), nil
	case admin.State_RETRYABLE_FAILURE:
		return core.PhaseInfoRetryableFailure(pluginErrors.TaskFailedWithError, "failed to run the job.\n"+resource.Message, taskInfo), nil
	case admin.State_SUCCEEDED:
		err = writeOutput(ctx, taskCtx, resource.Outputs)
		if err != nil {
			logger.Errorf(ctx, "failed to write output with err %s", err.Error())
			return core.PhaseInfoUndefined, fmt.Errorf("failed to write output with err %s", err.Error())
		}
		return core.PhaseInfoSuccess(taskInfo), nil
	}
	return core.PhaseInfoUndefined, pluginErrors.Errorf(core.SystemErrorCode, "unknown execution state [%v].", resource.State)
}

func (p *Plugin) getSyncConnectorClient(ctx context.Context, connector *Deployment) (service.SyncAgentServiceClient, error) {
	client, ok := p.cs.syncConnectorClients[connector.Endpoint]
	if !ok {
		conn, err := getGrpcConnection(ctx, connector)
		if err != nil {
			return nil, err
		}
		client = service.NewSyncAgentServiceClient(conn)
		p.cs.syncConnectorClients[connector.Endpoint] = client
	}
	return client, nil
}

func (p *Plugin) getAsyncConnectorClient(ctx context.Context, connector *Deployment) (service.AsyncAgentServiceClient, error) {
	client, ok := p.cs.asyncConnectorClients[connector.Endpoint]
	if !ok {
		conn, err := getGrpcConnection(ctx, connector)
		if err != nil {
			return nil, err
		}
		client = service.NewAsyncAgentServiceClient(conn)
		p.cs.asyncConnectorClients[connector.Endpoint] = client
	}
	return client, nil
}

// UpdateDeployment updates the deployment configuration for the plugin.
// This method is thread-safe and can be called concurrently.
func (p *Plugin) UpdateDeployment(deployment Connector) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.deployment = deployment
}

func WatchConnectors(ctx context.Context) {
	cfg := GetConfig()
	go wait.Until(func() {
		childCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		clientSet := getConnectorClientSets(childCtx)
		watchConnectors(ctx, clientSet)
	}, cfg.PollInterval.Duration, ctx.Done())
}

func writeOutput(ctx context.Context, taskCtx webapi.StatusContext, outputs *flyteIdl.LiteralMap) error {
	taskTemplate, err := taskCtx.TaskReader().Read(ctx)
	if err != nil {
		return err
	}

	if taskTemplate.GetInterface() == nil || taskTemplate.GetInterface().GetOutputs() == nil || taskTemplate.Interface.Outputs.Variables == nil {
		logger.Debugf(ctx, "The task declares no outputs. Skipping writing the outputs.")
		return nil
	}

	var opReader io.OutputReader
	if outputs != nil {
		logger.Debugf(ctx, "ConnectorDeployment returned an output.")
		opReader = ioutils.NewInMemoryOutputReader(outputs, nil, nil)
	} else {
		logger.Debugf(ctx, "ConnectorDeployment didn't return any output, assuming file based outputs.")
		opReader = ioutils.NewRemoteFileOutputReader(ctx, taskCtx.DataStore(), taskCtx.OutputWriter(), 0)
	}
	return taskCtx.OutputWriter().Put(ctx, opReader)
}

func buildTaskExecutionMetadata(taskExecutionMetadata core.TaskExecutionMetadata) admin.TaskExecutionMetadata {
	taskExecutionID := taskExecutionMetadata.GetTaskExecutionID().GetID()

	return admin.TaskExecutionMetadata{
		TaskExecutionId:      &taskExecutionID,
		Namespace:            taskExecutionMetadata.GetNamespace(),
		Labels:               taskExecutionMetadata.GetLabels(),
		Annotations:          taskExecutionMetadata.GetAnnotations(),
		K8SServiceAccount:    taskExecutionMetadata.GetK8sServiceAccount(),
		EnvironmentVariables: taskExecutionMetadata.GetEnvironmentVariables(),
		Identity:             taskExecutionMetadata.GetSecurityContext().RunAs, // nolint:protogetter
	}
}

func createPluginEntry(taskType core.TaskType, taskVersion int32, deployment Deployment, clientSet *ClientSet) webapi.PluginEntry {
	versionedTaskType := fmt.Sprintf("%s_%d", taskType, taskVersion)
	plugin := &Plugin{
		metricScope: promutils.NewScope(versionedTaskType),
		cfg:         GetConfig(),
		cs:          clientSet,
		deployment:  Connector{IsSync: false, ConnectorDeployment: &deployment},
	}
	return webapi.PluginEntry{
		ID:                 versionedTaskType,
		SupportedTaskTypes: []core.TaskType{taskType},
		PluginLoader: func(ctx context.Context, iCtx webapi.PluginSetupContext) (webapi.AsyncPlugin, error) {
			return plugin, nil
		},
	}
}

// createOrUpdatePlugin handles the registration or update of a task type plugin
func createOrUpdatePlugin(ctx context.Context, taskName string, taskVersion int32, deploymentID string, connectorDeployment *Deployment, cs *ClientSet) string {
	versionedTaskType := fmt.Sprintf("%s_%d", taskName, taskVersion)

	// Register core plugin if not registered
	if !pluginmachinery.PluginRegistry().IsConnectorCorePluginRegistered(versionedTaskType, deploymentID) {
		plugin := createPluginEntry(taskName, taskVersion, *connectorDeployment, cs)
		pluginmachinery.PluginRegistry().RegisterConnectorCorePlugin(plugin, deploymentID)
	}

	// send message to Flyte Propeller TaskHandler to register or update plugin
	select {
		case pluginmachinery.PluginRegistry().GetPluginChan() <- pluginmachinery.PluginInfo{
			VersionedTaskType: versionedTaskType,
			DeploymentID: deploymentID,
		}:
		default:
			logger.Errorf(context.Background(), "Failed to create/update plugin for task type %s: channel is full", versionedTaskType)
		}
	return versionedTaskType
}

func RegisterConnectorPlugin() {
	ctx := context.Background()
	gob.Register(ResourceMetaWrapper{})
	gob.Register(ResourceWrapper{})
	WatchConnectors(ctx)
}
