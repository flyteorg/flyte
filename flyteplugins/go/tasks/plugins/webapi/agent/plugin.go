package agent

import (
	"context"
	"crypto/x509"
	"encoding/gob"
	"fmt"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/config"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"google.golang.org/grpc/grpclog"

	flyteIdl "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	pluginErrors "github.com/flyteorg/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/template"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/webapi"
	"github.com/flyteorg/flytestdlib/promutils"
	"google.golang.org/grpc"
)

type GetClientFunc func(ctx context.Context, agent *Agent, connectionCache map[*Agent]*grpc.ClientConn) (service.AsyncAgentServiceClient, error)

type Plugin struct {
	metricScope     promutils.Scope
	cfg             *Config
	getClient       GetClientFunc
	connectionCache map[*Agent]*grpc.ClientConn
}

type ResourceWrapper struct {
	State   admin.State
	Outputs *flyteIdl.LiteralMap
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

	if taskTemplate.GetContainer() != nil {
		templateParameters := template.Parameters{
			TaskExecMetadata: taskCtx.TaskExecutionMetadata(),
			Inputs:           taskCtx.InputReader(),
			OutputPath:       taskCtx.OutputWriter(),
			Task:             taskCtx.TaskReader(),
		}
		modifiedArgs, err := template.Render(ctx, taskTemplate.GetContainer().Args, templateParameters)
		if err != nil {
			return nil, nil, err
		}
		taskTemplate.GetContainer().Args = modifiedArgs
	}
	outputPrefix := taskCtx.OutputWriter().GetOutputPrefixPath().String()

	agent, err := getFinalAgent(taskTemplate.Type, p.cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find agent agent with error: %v", err)
	}
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

	return &ResourceMetaWrapper{
		OutputPrefix:      outputPrefix,
		AgentResourceMeta: res.GetResourceMeta(),
		Token:             "",
		TaskType:          taskTemplate.Type,
	}, &ResourceWrapper{State: admin.State_RUNNING}, nil
}

func (p Plugin) Get(ctx context.Context, taskCtx webapi.GetContext) (latest webapi.Resource, err error) {
	metadata := taskCtx.ResourceMeta().(*ResourceMetaWrapper)

	agent, err := getFinalAgent(metadata.TaskType, p.cfg)
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

	return &ResourceWrapper{
		State:   res.Resource.State,
		Outputs: res.Resource.Outputs,
	}, nil
}

func (p Plugin) Delete(ctx context.Context, taskCtx webapi.DeleteContext) error {
	if taskCtx.ResourceMeta() == nil {
		return nil
	}
	metadata := taskCtx.ResourceMeta().(ResourceMetaWrapper)

	agent, err := getFinalAgent(metadata.TaskType, p.cfg)
	if err != nil {
		return fmt.Errorf("failed to find agent agent with error: %v", err)
	}
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
	resource := taskCtx.Resource().(*ResourceWrapper)
	taskInfo := &core.TaskInfo{}

	switch resource.State {
	case admin.State_RUNNING:
		return core.PhaseInfoRunning(core.DefaultPhaseVersion, taskInfo), nil
	case admin.State_PERMANENT_FAILURE:
		return core.PhaseInfoFailure(pluginErrors.TaskFailedWithError, "failed to run the job", taskInfo), nil
	case admin.State_RETRYABLE_FAILURE:
		return core.PhaseInfoRetryableFailure(pluginErrors.TaskFailedWithError, "failed to run the job", taskInfo), nil
	case admin.State_SUCCEEDED:
		if resource.Outputs != nil {
			err := taskCtx.OutputWriter().Put(ctx, ioutils.NewInMemoryOutputReader(resource.Outputs, nil, nil))
			if err != nil {
				return core.PhaseInfoUndefined, err
			}
		}
		return core.PhaseInfoSuccess(taskInfo), nil
	}
	return core.PhaseInfoUndefined, pluginErrors.Errorf(core.SystemErrorCode, "unknown execution phase [%v].", resource.State)
}

func getFinalAgent(taskType string, cfg *Config) (*Agent, error) {
	if id, exists := cfg.AgentForTaskTypes[taskType]; exists {
		if agent, exists := cfg.Agents[id]; exists {
			return agent, nil
		}
		return nil, fmt.Errorf("no agent definition found for ID %s that matches task type %s", id, taskType)
	}

	return &cfg.DefaultAgent, nil
}

func getClientFunc(ctx context.Context, agent *Agent, connectionCache map[*Agent]*grpc.ClientConn) (service.AsyncAgentServiceClient, error) {
	conn, ok := connectionCache[agent]
	if ok {
		return service.NewAsyncAgentServiceClient(conn), nil
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
	return service.NewAsyncAgentServiceClient(conn), nil
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

func newAgentPlugin() webapi.PluginEntry {
	supportedTaskTypes := GetConfig().SupportedTaskTypes

	return webapi.PluginEntry{
		ID:                 "agent-service",
		SupportedTaskTypes: supportedTaskTypes,
		PluginLoader: func(ctx context.Context, iCtx webapi.PluginSetupContext) (webapi.AsyncPlugin, error) {
			return &Plugin{
				metricScope:     iCtx.MetricsScope(),
				cfg:             GetConfig(),
				getClient:       getClientFunc,
				connectionCache: make(map[*Agent]*grpc.ClientConn),
			}, nil
		},
	}
}

func RegisterAgentPlugin() {
	gob.Register(ResourceMetaWrapper{})
	gob.Register(ResourceWrapper{})

	pluginmachinery.PluginRegistry().RegisterRemotePlugin(newAgentPlugin())
}
