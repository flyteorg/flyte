package connector

import (
	"context"
	"encoding/gob"
	"fmt"
	"slices"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/structpb"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/utils"

	pluginErrors "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/logs"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/template"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/tasklog"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/webapi"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	connectorPb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/connector"
	flyteIdl "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
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

type RegistryKey struct {
	domain          string
	taskTypeName    string
	taskTypeVersion int32
}

type Registry map[RegistryKey]*Connector

// getSupportedTaskTypes Get all the supported task types in the registry.
func (r Registry) getSupportedTaskTypes() []string {
	var taskTypes []string
	for k := range r {
		taskTypes = append(taskTypes, k.taskTypeName)
	}
	return taskTypes
}

type Plugin struct {
	metricScope promutils.Scope
	cfg         *Config
	cs          *ClientSet
	registry    Registry
	mu          sync.RWMutex
}

type ResourceWrapper struct {
	Phase          flyteIdl.TaskExecution_Phase
	Outputs        *task.Outputs
	Message        string
	LogLinks       []*flyteIdl.TaskLog
	CustomInfo     *structpb.Struct
	ConnectorID    string
	IsConnectorApp bool
}

// IsTerminal is used to avoid making network calls to the connector service if the resource is already in a terminal state.
func (r ResourceWrapper) IsTerminal() bool {
	return r.Phase == flyteIdl.TaskExecution_SUCCEEDED || r.Phase == flyteIdl.TaskExecution_FAILED || r.Phase == flyteIdl.TaskExecution_ABORTED
}

type ResourceMetaWrapper struct {
	OutputPrefix          string
	ConnectorResourceMeta []byte
	TaskCategory          connectorPb.TaskCategory
	Domain                string
	Connection            flyteIdl.Connection
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
	literalMap, err := taskCtx.InputReader().Get(ctx)
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

	taskCategory := connectorPb.TaskCategory{Name: taskTemplate.GetType(), Version: taskTemplate.GetTaskTypeVersion()}
	connector, err := p.getFinalConnector(&taskCategory, p.cfg, taskTemplate.GetId().GetDomain())
	if err != nil {
		return nil, nil, err
	}

	connection := flyteIdl.Connection{}
	err = utils.UnmarshalStruct(taskTemplate.GetCustom(), &connection)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal connection from task template custom: %v", err)
	}
	//for k, v := range connection.GetSecrets() {
	//	secretID, err := secret.GetSecretID(v, source, labels)
	//	if err != nil {
	//		errString := fmt.Sprintf("Failed to get secret id with error: %v", err)
	//		logger.Errorf(ctx, errString)
	//		return nil, nil, status.Errorf(codes.Internal, errString)
	//	}
	//	secretVal, err := taskCtx.SecretManager().Get(ctx, secretID)
	//	if err != nil {
	//		errString := fmt.Sprintf("Failed to get secret value with error: %v", err)
	//		logger.Errorf(ctx, errString)
	//		return nil, nil, status.Errorf(codes.Internal, errString)
	//	}
	//	conn.Secrets[k] = secretVal
	//}

	taskExecutionMetadata := buildTaskExecutionMetadata(taskCtx.TaskExecutionMetadata())

	finalCtx, cancel := getFinalContext(ctx, "CreateTask", connector.ConnectorDeployment)
	defer cancel()

	// Use async connector client
	client, err := p.getAsyncConnectorClient(ctx, connector.ConnectorDeployment)
	if err != nil {
		return nil, nil, err
	}
	var inputs *task.Inputs
	if literalMap != nil {
		var literals []*task.NamedLiteral
		for name, val := range literalMap.Literals {
			literals = append(literals, &task.NamedLiteral{Name: name, Value: val})
		}
		inputs = &task.Inputs{Literals: literals}
	}

	request := &connectorPb.CreateTaskRequest{
		Inputs:                inputs,
		Template:              taskTemplate,
		OutputPrefix:          outputPrefix,
		TaskExecutionMetadata: &taskExecutionMetadata,
		Connection:            &connection,
	}
	res, err := client.CreateTask(finalCtx, request)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create task from connector with %v", err)
	}

	return ResourceMetaWrapper{
		OutputPrefix:          outputPrefix,
		ConnectorResourceMeta: res.GetResourceMeta(),
		TaskCategory:          taskCategory,
		Connection:            connection,
		Domain:                taskTemplate.GetId().GetDomain(),
	}, nil, nil
}

func (p *Plugin) Get(ctx context.Context, taskCtx webapi.GetContext) (latest webapi.Resource, err error) {
	metadata := taskCtx.ResourceMeta().(ResourceMetaWrapper)
	connector, err := p.getFinalConnector(&metadata.TaskCategory, p.cfg, metadata.Domain)
	if err != nil {
		return nil, err
	}

	client, err := p.getAsyncConnectorClient(ctx, connector.ConnectorDeployment)
	if err != nil {
		return nil, err
	}
	finalCtx, cancel := getFinalContext(ctx, "GetTask", connector.ConnectorDeployment)
	defer cancel()

	request := &connectorPb.GetTaskRequest{
		TaskCategory: &metadata.TaskCategory,
		ResourceMeta: metadata.ConnectorResourceMeta,
		OutputPrefix: metadata.OutputPrefix,
		Connection:   &metadata.Connection,
	}
	res, err := client.GetTask(finalCtx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get task from connector with %v", err)
	}

	return ResourceWrapper{
		Phase:          res.GetResource().GetPhase(),
		Outputs:        res.GetResource().GetOutputs(),
		Message:        res.GetResource().GetMessage(),
		LogLinks:       res.GetResource().GetLogLinks(),
		CustomInfo:     res.GetResource().GetCustomInfo(),
		ConnectorID:    connector.ConnectorID,
		IsConnectorApp: connector.IsConnectorApp,
	}, nil
}

func (p *Plugin) Delete(ctx context.Context, taskCtx webapi.DeleteContext) error {
	if taskCtx.ResourceMeta() == nil {
		return nil
	}
	metadata := taskCtx.ResourceMeta().(ResourceMetaWrapper)
	connector, err := p.getFinalConnector(&metadata.TaskCategory, p.cfg, metadata.Domain)
	if err != nil {
		return err
	}

	client, err := p.getAsyncConnectorClient(ctx, connector.ConnectorDeployment)
	if err != nil {
		return err
	}
	finalCtx, cancel := getFinalContext(ctx, "DeleteTask", connector.ConnectorDeployment)
	defer cancel()

	request := &connectorPb.DeleteTaskRequest{
		TaskCategory: &metadata.TaskCategory,
		ResourceMeta: metadata.ConnectorResourceMeta,
		Connection:   &metadata.Connection,
	}
	_, err = client.DeleteTask(finalCtx, request)
	if err != nil {
		return fmt.Errorf("failed to delete task from connector with %v", err)
	}
	return nil
}

func (p *Plugin) getEventInfoForConnectorApp(taskCtx webapi.StatusContext, resource ResourceWrapper) ([]*flyteIdl.TaskLog, error) {
	logPlugin, err := logs.InitializeLogPlugins(&p.cfg.Logs)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize log plugins with error: %v", err)
	}
	taskLogs := resource.LogLinks

	if taskCtx.TaskExecutionMetadata().GetTaskExecutionID() == nil || !resource.IsConnectorApp {
		return taskLogs, nil
	}

	in := tasklog.Input{
		TaskExecutionID: taskCtx.TaskExecutionMetadata().GetTaskExecutionID(),
		ConnectorID:     resource.ConnectorID,
	}

	logs, err := logPlugin.GetTaskLogs(in)
	if err != nil {
		return nil, fmt.Errorf("failed to get task logs with error: %v", err)
	}

	taskLogs = append(taskLogs, logs.TaskLogs...)

	return taskLogs, nil
}

func (p *Plugin) Status(ctx context.Context, taskCtx webapi.StatusContext) (phase core.PhaseInfo, err error) {
	resource := taskCtx.Resource().(ResourceWrapper)

	logLinks, err := p.getEventInfoForConnectorApp(taskCtx, resource)
	if err != nil {
		return core.PhaseInfoUndefined, err
	}

	taskInfo := &core.TaskInfo{Logs: logLinks, CustomInfo: resource.CustomInfo}
	errorCode := pluginErrors.TaskFailedWithError

	switch resource.Phase {
	case flyteIdl.TaskExecution_QUEUED:
		return core.PhaseInfoQueuedWithTaskInfo(time.Now(), core.DefaultPhaseVersion, resource.Message, taskInfo), nil
	case flyteIdl.TaskExecution_WAITING_FOR_RESOURCES:
		return core.PhaseInfoWaitingForResourcesInfo(time.Now(), core.DefaultPhaseVersion, resource.Message, taskInfo), nil
	case flyteIdl.TaskExecution_INITIALIZING:
		return core.PhaseInfoInitializing(time.Now(), core.DefaultPhaseVersion, resource.Message, taskInfo), nil
	case flyteIdl.TaskExecution_RUNNING:
		return core.PhaseInfoRunning(core.DefaultPhaseVersion, taskInfo), nil
	case flyteIdl.TaskExecution_SUCCEEDED:
		if resource.Outputs != nil {
			var literalMap *flyteIdl.LiteralMap
			literalMap = &flyteIdl.LiteralMap{Literals: make(map[string]*flyteIdl.Literal)}
			for _, val := range resource.Outputs.Literals {
				literalMap.Literals[val.Name] = val.Value
			}
			// TODO: Add support writing outputs proto.
			err = writeOutput(ctx, taskCtx, literalMap)
			if err != nil {
				logger.Errorf(ctx, "failed to write output with err %s", err.Error())
				return core.PhaseInfoUndefined, err
			}
		}
		return core.PhaseInfoSuccess(taskInfo), nil
	case flyteIdl.TaskExecution_ABORTED:
		return core.PhaseInfoFailure(errorCode, "failed to run the job with aborted phase.", taskInfo), nil
	case flyteIdl.TaskExecution_FAILED:
		return core.PhaseInfoFailure(errorCode, fmt.Sprintf("failed to run the job: %s", resource.Message), taskInfo), nil
	}
	// The default phase is undefined.
	return core.PhaseInfoUndefined, pluginErrors.Errorf(core.SystemErrorCode, "unknown execution phase [%v].", resource.Phase)
}

func (p *Plugin) getAsyncConnectorClient(ctx context.Context, connector *Deployment) (connectorPb.AsyncConnectorServiceClient, error) {
	client, ok := p.cs.asyncConnectorClients[connector.Endpoint]
	if !ok {
		conn, err := getGrpcConnection(ctx, connector)
		if err != nil {
			return nil, err
		}
		client = connectorPb.NewAsyncConnectorServiceClient(conn)
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
		connectorService.SetSupportedTaskType(connectorRegistry.getSupportedTaskTypes())
	}, p.cfg.PollInterval.Duration, ctx.Done())
}

func (p *Plugin) getFinalConnector(taskCategory *connectorPb.TaskCategory, cfg *Config, domain string) (*Connector, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	registryKey := RegistryKey{domain: domain, taskTypeName: taskCategory.GetName(), taskTypeVersion: taskCategory.GetVersion()}
	if connector, exists := p.registry[registryKey]; exists {
		return connector, nil
	}
	logger.Debugf(context.Background(), "No connector found for task type [%s] and version [%d] in domain [%s].", taskCategory.GetName(), taskCategory.GetVersion(), domain)

	// Use the connector that supports across all domains.
	registryKey = RegistryKey{domain: "", taskTypeName: taskCategory.GetName(), taskTypeVersion: taskCategory.GetVersion()}
	if connector, exists := p.registry[registryKey]; exists {
		return connector, nil
	}
	logger.Debugf(context.Background(), "No connector found for task type [%s] and version [%d] in any domain.", taskCategory.GetName(), taskCategory.GetVersion())

	if len(cfg.DefaultConnector.Endpoint) != 0 {
		return &Connector{
			ConnectorDeployment: &cfg.DefaultConnector,
			ConnectorID:         "defaultConnector",
			IsConnectorApp:      false,
		}, nil
	}

	logger.Errorf(context.Background(), "No connector found for task type [%s] and version [%d]. Using default connector.", taskCategory.GetName(), taskCategory.GetVersion())
	return nil, fmt.Errorf("no connector found for task type [%s] and version [%d]", taskCategory.GetName(), taskCategory.GetVersion())
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

func buildTaskExecutionMetadata(taskExecutionMetadata core.TaskExecutionMetadata) connectorPb.TaskExecutionMetadata {
	taskExecutionID := taskExecutionMetadata.GetTaskExecutionID().GetID()

	return connectorPb.TaskExecutionMetadata{
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
	supportedTaskTypes := connectorRegistry.getSupportedTaskTypes()
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
	fmt.Printf("Registering connector plugin...\n")
	gob.Register(ResourceMetaWrapper{})
	gob.Register(ResourceWrapper{})
	pluginmachinery.PluginRegistry().RegisterRemotePlugin(newConnectorPlugin(connectorService))
}
