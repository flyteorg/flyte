package agent

import (
	"context"
	"crypto/x509"
	"encoding/gob"
	"fmt"
	"time"

	"golang.org/x/exp/maps"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"

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
	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

type GetClientFunc func(ctx context.Context, agent *Agent, connectionCache map[*Agent]*grpc.ClientConn) (service.AsyncAgentServiceClient, error)
type GetAgentMetadataClientFunc func(ctx context.Context, agent *Agent, connCache map[*Agent]*grpc.ClientConn) (service.AgentMetadataServiceClient, error)

type Plugin struct {
	metricScope     promutils.Scope
	cfg             *Config
	getClient       GetClientFunc
	connectionCache map[*Agent]*grpc.ClientConn
	agentRegistry   map[string]*Agent // map[taskType] => Agent
}

type ResourceWrapper struct {
	Phase    flyteIdl.TaskExecution_Phase
	State    admin.State // This is deprecated.
	Outputs  *flyteIdl.LiteralMap
	Message  string
	LogLinks []*flyteIdl.TaskLog
}

type ResourceMetaWrapper struct {
	OutputPrefix      string
	Token             string
	AgentResourceMeta []byte
	TaskType          string
}

func (p Plugin) GetConfig() webapi.PluginConfig {
	return GetConfig().WebAPI
}

func (p Plugin) ResourceRequirements(_ context.Context, _ webapi.TaskExecutionContextReader) (
	namespace core.ResourceNamespace, constraints core.ResourceConstraintsSpec, err error) {

	// Resource requirements are assumed to be the same.
	return "default", p.cfg.ResourceConstraints, nil
}

func (p Plugin) Create(ctx context.Context, taskCtx webapi.TaskExecutionContextReader) (webapi.ResourceMeta,
	webapi.Resource, error) {
	taskTemplate, err := taskCtx.TaskReader().Read(ctx)
	if err != nil {
		return nil, nil, err
	}
	inputs, err := taskCtx.InputReader().Get(ctx)
	if err != nil {
		return nil, nil, err
	}

	var argTemplate []string
	if taskTemplate.GetContainer() != nil {
		templateParameters := template.Parameters{
			TaskExecMetadata: taskCtx.TaskExecutionMetadata(),
			Inputs:           taskCtx.InputReader(),
			OutputPath:       taskCtx.OutputWriter(),
			Task:             taskCtx.TaskReader(),
		}
		argTemplate = taskTemplate.GetContainer().Args
		modifiedArgs, err := template.Render(ctx, taskTemplate.GetContainer().Args, templateParameters)
		if err != nil {
			return nil, nil, err
		}
		taskTemplate.GetContainer().Args = modifiedArgs
	}
	outputPrefix := taskCtx.OutputWriter().GetOutputPrefixPath().String()

	agent := getFinalAgent(taskTemplate.Type, p.cfg, p.agentRegistry)

	client, err := p.getClient(ctx, agent, p.connectionCache)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to agent with error: %v", err)
	}

	finalCtx, cancel := getFinalContext(ctx, "CreateTask", agent)
	defer cancel()

	taskExecutionMetadata := buildTaskExecutionMetadata(taskCtx.TaskExecutionMetadata())
	res, err := client.CreateTask(finalCtx, &admin.CreateTaskRequest{Inputs: inputs, Template: taskTemplate, OutputPrefix: outputPrefix, TaskExecutionMetadata: &taskExecutionMetadata})
	if err != nil {
		return nil, nil, err
	}

	// Restore unrendered template for subsequent renders.
	if taskTemplate.GetContainer() != nil {
		taskTemplate.GetContainer().Args = argTemplate
	}

	// If the agent returned a resource, we assume this is a synchronous task.
	// The state should be a terminal state, for example, SUCCEEDED, PERMANENT_FAILURE, or RETRYABLE_FAILURE.
	if res.GetResource() != nil {
		logger.Infof(ctx, "Agent is executing a synchronous task.")
		return nil,
			ResourceWrapper{
				Phase:    res.GetResource().Phase,
				State:    res.GetResource().State,
				Outputs:  res.GetResource().Outputs,
				Message:  res.GetResource().Message,
				LogLinks: res.GetResource().LogLinks,
			}, nil
	}

	logger.Infof(ctx, "Agent is executing an asynchronous task.")
	return ResourceMetaWrapper{
		OutputPrefix:      outputPrefix,
		AgentResourceMeta: res.GetResourceMeta(),
		Token:             "",
		TaskType:          taskTemplate.Type,
	}, nil, nil
}

func (p Plugin) Get(ctx context.Context, taskCtx webapi.GetContext) (latest webapi.Resource, err error) {
	metadata := taskCtx.ResourceMeta().(ResourceMetaWrapper)

	agent := getFinalAgent(metadata.TaskType, p.cfg, p.agentRegistry)
	if err != nil {
		return nil, fmt.Errorf("failed to find agent with error: %v", err)
	}

	client, err := p.getClient(ctx, agent, p.connectionCache)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to agent with error: %v", err)
	}

	finalCtx, cancel := getFinalContext(ctx, "GetTask", agent)
	defer cancel()

	res, err := client.GetTask(finalCtx, &admin.GetTaskRequest{TaskType: metadata.TaskType, ResourceMeta: metadata.AgentResourceMeta})
	if err != nil {
		return nil, err
	}

	return ResourceWrapper{
		Phase:    res.Resource.Phase,
		State:    res.Resource.State,
		Outputs:  res.Resource.Outputs,
		Message:  res.Resource.Message,
		LogLinks: res.LogLinks,
	}, nil
}

func (p Plugin) Delete(ctx context.Context, taskCtx webapi.DeleteContext) error {
	if taskCtx.ResourceMeta() == nil {
		return nil
	}
	metadata := taskCtx.ResourceMeta().(ResourceMetaWrapper)

	agent := getFinalAgent(metadata.TaskType, p.cfg, p.agentRegistry)

	client, err := p.getClient(ctx, agent, p.connectionCache)
	if err != nil {
		return fmt.Errorf("failed to connect to agent with error: %v", err)
	}

	finalCtx, cancel := getFinalContext(ctx, "DeleteTask", agent)
	defer cancel()

	_, err = client.DeleteTask(finalCtx, &admin.DeleteTaskRequest{TaskType: metadata.TaskType, ResourceMeta: metadata.AgentResourceMeta})
	return err
}

func (p Plugin) Status(ctx context.Context, taskCtx webapi.StatusContext) (phase core.PhaseInfo, err error) {
	resource := taskCtx.Resource().(ResourceWrapper)
	taskInfo := &core.TaskInfo{Logs: resource.LogLinks}

	switch resource.Phase {
	case flyteIdl.TaskExecution_QUEUED:
		return core.PhaseInfoQueuedWithTaskInfo(core.DefaultPhaseVersion, resource.Message, taskInfo), nil
	case flyteIdl.TaskExecution_WAITING_FOR_RESOURCES:
		return core.PhaseInfoWaitingForResourcesInfo(time.Now(), core.DefaultPhaseVersion, resource.Message, taskInfo), nil
	case flyteIdl.TaskExecution_INITIALIZING:
		return core.PhaseInfoInitializing(time.Now(), core.DefaultPhaseVersion, resource.Message, taskInfo), nil
	case flyteIdl.TaskExecution_RUNNING:
		return core.PhaseInfoRunning(core.DefaultPhaseVersion, taskInfo), nil
	case flyteIdl.TaskExecution_SUCCEEDED:
		err = writeOutput(ctx, taskCtx, resource)
		if err != nil {
			logger.Errorf(ctx, "Failed to write output with err %s", err.Error())
			return core.PhaseInfoUndefined, err
		}
		return core.PhaseInfoSuccess(taskInfo), nil
	case flyteIdl.TaskExecution_ABORTED:
		return core.PhaseInfoFailure(pluginErrors.TaskFailedWithError, "failed to run the job with aborted phase", taskInfo), nil
	case flyteIdl.TaskExecution_FAILED:
		return core.PhaseInfoFailure(pluginErrors.TaskFailedWithError, "failed to run the job", taskInfo), nil
	}

	// The default phase is undefined.
	if resource.Phase != flyteIdl.TaskExecution_UNDEFINED {
		return core.PhaseInfoUndefined, pluginErrors.Errorf(core.SystemErrorCode, "unknown execution phase [%v].", resource.Phase)
	}

	// If the phase is undefined, we will use state to determine the phase.
	switch resource.State {
	case admin.State_PENDING:
		return core.PhaseInfoInitializing(time.Now(), core.DefaultPhaseVersion, resource.Message, taskInfo), nil
	case admin.State_RUNNING:
		return core.PhaseInfoRunning(core.DefaultPhaseVersion, taskInfo), nil
	case admin.State_PERMANENT_FAILURE:
		return core.PhaseInfoFailure(pluginErrors.TaskFailedWithError, "failed to run the job", taskInfo), nil
	case admin.State_RETRYABLE_FAILURE:
		return core.PhaseInfoRetryableFailure(pluginErrors.TaskFailedWithError, "failed to run the job", taskInfo), nil
	case admin.State_SUCCEEDED:
		err = writeOutput(ctx, taskCtx, resource)
		if err != nil {
			logger.Errorf(ctx, "Failed to write output with err %s", err.Error())
			return core.PhaseInfoUndefined, err
		}
		return core.PhaseInfoSuccess(taskInfo), nil
	}
	return core.PhaseInfoUndefined, pluginErrors.Errorf(core.SystemErrorCode, "unknown execution state [%v].", resource.State)
}

func writeOutput(ctx context.Context, taskCtx webapi.StatusContext, resource ResourceWrapper) error {
	taskTemplate, err := taskCtx.TaskReader().Read(ctx)
	if err != nil {
		return err
	}

	if taskTemplate.Interface == nil || taskTemplate.Interface.Outputs == nil || taskTemplate.Interface.Outputs.Variables == nil {
		logger.Debugf(ctx, "The task declares no outputs. Skipping writing the outputs.")
		return nil
	}

	var opReader io.OutputReader
	if resource.Outputs != nil {
		logger.Debugf(ctx, "Agent returned an output.")
		opReader = ioutils.NewInMemoryOutputReader(resource.Outputs, nil, nil)
	} else {
		logger.Debugf(ctx, "Agent didn't return any output, assuming file based outputs.")
		opReader = ioutils.NewRemoteFileOutputReader(ctx, taskCtx.DataStore(), taskCtx.OutputWriter(), taskCtx.MaxDatasetSizeBytes())
	}
	return taskCtx.OutputWriter().Put(ctx, opReader)
}

func getFinalAgent(taskType string, cfg *Config, agentRegistry map[string]*Agent) *Agent {
	if agent, exists := agentRegistry[taskType]; exists {
		return agent
	}

	return &cfg.DefaultAgent
}

func getGrpcConnection(ctx context.Context, agent *Agent, connectionCache map[*Agent]*grpc.ClientConn) (*grpc.ClientConn, error) {
	conn, ok := connectionCache[agent]
	if ok {
		return conn, nil
	}
	var opts []grpc.DialOption

	if agent.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		pool, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}

		creds := credentials.NewClientTLSFromCert(pool, "")
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}

	if len(agent.DefaultServiceConfig) != 0 {
		opts = append(opts, grpc.WithDefaultServiceConfig(agent.DefaultServiceConfig))
	}

	var err error
	conn, err = grpc.Dial(agent.Endpoint, opts...)
	if err != nil {
		return nil, err
	}
	connectionCache[agent] = conn
	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", agent, cerr)
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", agent, cerr)
			}
		}()
	}()

	return conn, nil
}

func getClientFunc(ctx context.Context, agent *Agent, connectionCache map[*Agent]*grpc.ClientConn) (service.AsyncAgentServiceClient, error) {
	conn, err := getGrpcConnection(ctx, agent, connectionCache)
	if err != nil {
		return nil, err
	}

	return service.NewAsyncAgentServiceClient(conn), nil
}

func getAgentMetadataClientFunc(ctx context.Context, agent *Agent, connectionCache map[*Agent]*grpc.ClientConn) (service.AgentMetadataServiceClient, error) {
	conn, err := getGrpcConnection(ctx, agent, connectionCache)
	if err != nil {
		return nil, err
	}

	return service.NewAgentMetadataServiceClient(conn), nil
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
	}
}

func getFinalTimeout(operation string, agent *Agent) config.Duration {
	if t, exists := agent.Timeouts[operation]; exists {
		return t
	}

	return agent.DefaultTimeout
}

func getFinalContext(ctx context.Context, operation string, agent *Agent) (context.Context, context.CancelFunc) {
	timeout := getFinalTimeout(operation, agent).Duration
	if timeout == 0 {
		return ctx, func() {}
	}

	return context.WithTimeout(ctx, timeout)
}

func initializeAgentRegistry(cfg *Config, connectionCache map[*Agent]*grpc.ClientConn, getAgentMetadataClientFunc GetAgentMetadataClientFunc) (map[string]*Agent, error) {
	agentRegistry := make(map[string]*Agent)
	var agentDeployments []*Agent

	// Ensure that the old configuration is backward compatible
	for taskType, agentID := range cfg.AgentForTaskTypes {
		agentRegistry[taskType] = cfg.Agents[agentID]
	}

	if len(cfg.DefaultAgent.Endpoint) != 0 {
		agentDeployments = append(agentDeployments, &cfg.DefaultAgent)
	}
	agentDeployments = append(agentDeployments, maps.Values(cfg.Agents)...)
	for _, agentDeployment := range agentDeployments {
		client, err := getAgentMetadataClientFunc(context.Background(), agentDeployment, connectionCache)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to agent [%v] with error: [%v]", agentDeployment, err)
		}

		finalCtx, cancel := getFinalContext(context.Background(), "ListAgents", agentDeployment)
		defer cancel()

		res, err := client.ListAgents(finalCtx, &admin.ListAgentsRequest{})
		if err != nil {
			grpcStatus, ok := status.FromError(err)
			if grpcStatus.Code() == codes.Unimplemented {
				// we should not panic here, as we want to continue to support old agent settings
				logger.Infof(context.Background(), "list agent method not implemented for agent: [%v]", agentDeployment)
				continue
			}

			if !ok {
				return nil, fmt.Errorf("failed to list agent: [%v] with a non-gRPC error: [%v]", agentDeployment, err)
			}

			return nil, fmt.Errorf("failed to list agent: [%v] with error: [%v]", agentDeployment, err)
		}

		agents := res.GetAgents()
		for _, agent := range agents {
			supportedTaskTypes := agent.SupportedTaskTypes
			for _, supportedTaskType := range supportedTaskTypes {
				agentRegistry[supportedTaskType] = agentDeployment
			}
		}
	}

	return agentRegistry, nil
}

func newAgentPlugin() webapi.PluginEntry {
	cfg := GetConfig()
	connectionCache := make(map[*Agent]*grpc.ClientConn)
	agentRegistry, err := initializeAgentRegistry(cfg, connectionCache, getAgentMetadataClientFunc)
	if err != nil {
		// We should wait for all agents to be up and running before starting the server
		panic(err)
	}

	supportedTaskTypes := append(maps.Keys(agentRegistry), cfg.SupportedTaskTypes...)
	logger.Infof(context.Background(), "Agent supports task types: %v", supportedTaskTypes)

	return webapi.PluginEntry{
		ID:                 "agent-service",
		SupportedTaskTypes: supportedTaskTypes,
		PluginLoader: func(ctx context.Context, iCtx webapi.PluginSetupContext) (webapi.AsyncPlugin, error) {
			return &Plugin{
				metricScope:     iCtx.MetricsScope(),
				cfg:             cfg,
				getClient:       getClientFunc,
				connectionCache: connectionCache,
				agentRegistry:   agentRegistry,
			}, nil
		},
	}
}

func RegisterAgentPlugin() {
	gob.Register(ResourceMetaWrapper{})
	gob.Register(ResourceWrapper{})

	pluginmachinery.PluginRegistry().RegisterRemotePlugin(newAgentPlugin())
}
