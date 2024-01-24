package agent

import (
	"context"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	flyteIdl "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	pluginErrors "github.com/flyteorg/flyte/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/template"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/webapi"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"golang.org/x/exp/maps"
)

type Plugin struct {
	metricScope   promutils.Scope
	cfg           *Config
	cs            *ClientSet
	agentRegistry map[string]*Agent // map[taskType] => Agent
}

type ResourceWrapper struct {
	State    admin.State
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

	client := p.cs.agentClients[agent.Endpoint]
	if client == nil {
		return nil, nil, fmt.Errorf("default agent:[%v] is not connected, please check if the default agent is up and running", agent)
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

	client := p.cs.agentClients[agent.Endpoint]
	finalCtx, cancel := getFinalContext(ctx, "GetTask", agent)
	defer cancel()

	res, err := client.GetTask(finalCtx, &admin.GetTaskRequest{TaskType: metadata.TaskType, ResourceMeta: metadata.AgentResourceMeta})
	if err != nil {
		return nil, err
	}

	return ResourceWrapper{
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

	client := p.cs.agentClients[agent.Endpoint]
	finalCtx, cancel := getFinalContext(ctx, "DeleteTask", agent)
	defer cancel()

	_, err := client.DeleteTask(finalCtx, &admin.DeleteTaskRequest{TaskType: metadata.TaskType, ResourceMeta: metadata.AgentResourceMeta})
	return err
}

func (p Plugin) Status(ctx context.Context, taskCtx webapi.StatusContext) (phase core.PhaseInfo, err error) {
	resource := taskCtx.Resource().(ResourceWrapper)
	taskInfo := &core.TaskInfo{Logs: resource.LogLinks}

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
	return core.PhaseInfoUndefined, pluginErrors.Errorf(core.SystemErrorCode, "unknown execution phase [%v].", resource.State)
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

func newAgentPlugin() webapi.PluginEntry {
	cs, err := initializeClients(context.Background())
	if err != nil {
		// We should wait for all agents to be up and running before starting the server
		panic(fmt.Sprintf("failed to initialize clients with error: %v", err))
	}

	agentRegistry, err := initializeAgentRegistry(cs)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize agent registry with error: %v", err))
	}

	cfg := GetConfig()
	supportedTaskTypes := append(maps.Keys(agentRegistry), cfg.SupportedTaskTypes...)
	logger.Infof(context.Background(), "Agent supports task types: %v", supportedTaskTypes)

	return webapi.PluginEntry{
		ID:                 "agent-service",
		SupportedTaskTypes: supportedTaskTypes,
		PluginLoader: func(ctx context.Context, iCtx webapi.PluginSetupContext) (webapi.AsyncPlugin, error) {
			return &Plugin{
				metricScope:   iCtx.MetricsScope(),
				cfg:           cfg,
				cs:            cs,
				agentRegistry: agentRegistry,
			}, nil
		},
	}
}

func RegisterAgentPlugin() {
	gob.Register(ResourceMetaWrapper{})
	gob.Register(ResourceWrapper{})

	pluginmachinery.PluginRegistry().RegisterRemotePlugin(newAgentPlugin())
}
