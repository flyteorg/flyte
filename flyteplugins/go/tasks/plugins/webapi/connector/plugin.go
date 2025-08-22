package connector

import (
	"context"
	"encoding/gob"
	"fmt"
	"slices"
	"sync"

	"golang.org/x/exp/maps"
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

const ID = "connector-service"

type ConnectorService struct {
	mu                 sync.RWMutex
	supportedTaskTypes []string
	CorePlugin         core.Plugin
}

// ContainTaskType check if connector supports this task type.
func (p *ConnectorService) ContainTaskType(taskType string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return slices.Contains(p.supportedTaskTypes, taskType)
}

// SetSupportedTaskType set supportTaskType in the connector service.
func (p *ConnectorService) SetSupportedTaskType(taskTypes []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.supportedTaskTypes = taskTypes
}

type Registry map[string]map[int32]*Connector // map[taskTypeName][taskTypeVersion] => Connector

type Plugin struct {
	metricScope promutils.Scope
	cfg         *Config
	cs          *ClientSet
	registry    Registry
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

func (p *Plugin) setRegistry(r Registry) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.registry = r
}

func (p *Plugin) GetConfig() webapi.PluginConfig {
	return GetConfig().WebAPI
}

func (p *Plugin) ResourceRequirements(_ context.Context, _ webapi.TaskExecutionContextReader) (
	namespace core.ResourceNamespace, constraints core.ResourceConstraintsSpec, err error) {

	// Resource requirements are assumed to be the same.
	return "default", p.cfg.ResourceConstraints, nil
}

func (p *Plugin) Create(ctx context.Context, taskCtx webapi.TaskExecutionContextReader) (webapi.ResourceMeta,
	webapi.Resource, error) {
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
	connector, isSync := p.getFinalConnector(&taskCategory, p.cfg)

	taskExecutionMetadata := buildTaskExecutionMetadata(taskCtx.TaskExecutionMetadata())

	if isSync {
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
	connector, _ := p.getFinalConnector(&metadata.TaskCategory, p.cfg)

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
	connector, _ := p.getFinalConnector(&metadata.TaskCategory, p.cfg)

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

func (p *Plugin) watchConnectors(ctx context.Context, connectorService *ConnectorService) {
	go wait.Until(func() {
		childCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		clientSet := getConnectorClientSets(childCtx)
		connectorRegistry := getConnectorRegistry(childCtx, clientSet)
		p.setRegistry(connectorRegistry)
		connectorService.SetSupportedTaskType(maps.Keys(connectorRegistry))
	}, p.cfg.PollInterval.Duration, ctx.Done())
}

func (p *Plugin) getFinalConnector(taskCategory *admin.TaskCategory, cfg *Config) (*Deployment, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if connector, exists := p.registry[taskCategory.GetName()][taskCategory.GetVersion()]; exists {
		return connector.ConnectorDeployment, connector.IsSync
	}
	return &cfg.DefaultConnector, false
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

func newConnectorPlugin(connectorService *ConnectorService) webapi.PluginEntry {
	ctx := context.Background()
	gob.Register(ResourceMetaWrapper{})
	gob.Register(ResourceWrapper{})

	clientSet := getConnectorClientSets(ctx)
	connectorRegistry := getConnectorRegistry(ctx, clientSet)
	supportedTaskTypes := maps.Keys(connectorRegistry)
	connectorService.SetSupportedTaskType(supportedTaskTypes)

	plugin := &Plugin{
		metricScope: promutils.NewScope("connector_plugin"),
		cfg:         GetConfig(),
		cs:          clientSet,
		registry:    connectorRegistry,
	}
	plugin.watchConnectors(ctx, connectorService)

	return webapi.PluginEntry{
		ID:                 ID,
		SupportedTaskTypes: supportedTaskTypes,
		PluginLoader: func(ctx context.Context, iCtx webapi.PluginSetupContext) (webapi.AsyncPlugin, error) {
			return plugin, nil
		},
	}
}

func RegisterConnectorPlugin(connectorService *ConnectorService) {
	gob.Register(ResourceMetaWrapper{})
	gob.Register(ResourceWrapper{})
	pluginmachinery.PluginRegistry().RegisterRemotePlugin(newConnectorPlugin(connectorService))
}
