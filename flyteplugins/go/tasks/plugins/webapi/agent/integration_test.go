package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/utils/strings/slices"

	agentMocks "github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyte/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	flyteIdlCore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/plugins"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	pluginCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	pluginCoreMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	ioMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/webapi"
	"github.com/flyteorg/flyte/flyteplugins/tests"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/flytestdlib/storage"
	"github.com/flyteorg/flyte/flytestdlib/utils"
)

func TestEndToEnd(t *testing.T) {
	iter := func(ctx context.Context, tCtx pluginCore.TaskExecutionContext) error {
		return nil
	}

	cfg := defaultConfig
	cfg.WebAPI.ResourceQuotas = map[core.ResourceNamespace]int{}
	cfg.WebAPI.Caching.Workers = 1
	cfg.WebAPI.Caching.ResyncInterval.Duration = 5 * time.Second
	cfg.DefaultAgent.Endpoint = "localhost:8000"
	err := SetConfig(&cfg)
	assert.NoError(t, err)

	databricksConfDict := map[string]interface{}{
		"name": "flytekit databricks plugin example",
		"new_cluster": map[string]string{
			"spark_version": "11.0.x-scala2.12",
			"node_type_id":  "r3.xlarge",
			"num_workers":   "4",
		},
		"timeout_seconds": 3600,
		"max_retries":     1,
	}
	databricksConfig, err := utils.MarshalObjToStruct(databricksConfDict)
	assert.NoError(t, err)
	sparkJob := plugins.SparkJob{DatabricksConf: databricksConfig, DatabricksToken: "token", SparkConf: map[string]string{"spark.driver.bindAddress": "127.0.0.1"}}
	st, err := utils.MarshalPbToStruct(&sparkJob)
	assert.NoError(t, err)

	inputs, _ := coreutils.MakeLiteralMap(map[string]interface{}{"x": 1})
	template := flyteIdlCore.TaskTemplate{
		Type:   "spark",
		Custom: st,
		Target: &flyteIdlCore.TaskTemplate_Container{
			Container: &flyteIdlCore.Container{Args: []string{"pyflyte-fast-execute", "--output-prefix", "/tmp/123"}},
		},
	}
	basePrefix := storage.DataReference("fake://bucket/prefix/")

	t.Run("run an async task", func(t *testing.T) {
		pluginEntry := pluginmachinery.CreateRemotePlugin(newMockAsyncAgentPlugin())
		plugin, err := pluginEntry.LoadPlugin(context.TODO(), newFakeSetupContext("async task"))
		assert.NoError(t, err)

		phase := tests.RunPluginEndToEndTest(t, plugin, &template, inputs, nil, nil, iter)
		assert.Equal(t, true, phase.Phase().IsSuccess())

		template.Type = "spark"
		phase = tests.RunPluginEndToEndTest(t, plugin, &template, inputs, nil, nil, iter)
		assert.Equal(t, true, phase.Phase().IsSuccess())
	})

	t.Run("run a sync task", func(t *testing.T) {
		pluginEntry := pluginmachinery.CreateRemotePlugin(newMockSyncAgentPlugin())
		plugin, err := pluginEntry.LoadPlugin(context.TODO(), newFakeSetupContext("sync task"))
		assert.NoError(t, err)

		template.Type = "openai"
		template.Interface = &flyteIdlCore.TypedInterface{
			Outputs: &flyteIdlCore.VariableMap{
				Variables: map[string]*flyteIdlCore.Variable{
					"x": {Type: &flyteIdlCore.LiteralType{
						Type: &flyteIdlCore.LiteralType_Simple{
							Simple: flyteIdlCore.SimpleType_INTEGER,
						},
					},
					},
				},
			},
		}
		expectedOutputs, err := coreutils.MakeLiteralMap(map[string]interface{}{"x": 1})
		assert.NoError(t, err)
		phase := tests.RunPluginEndToEndTest(t, plugin, &template, inputs, expectedOutputs, nil, iter)
		assert.Equal(t, true, phase.Phase().IsSuccess())
	})

	t.Run("failed to create a job", func(t *testing.T) {
		agentPlugin := newMockAsyncAgentPlugin()
		agentPlugin.PluginLoader = func(ctx context.Context, iCtx webapi.PluginSetupContext) (webapi.AsyncPlugin, error) {
			return Plugin{
				metricScope: iCtx.MetricsScope(),
				cfg:         GetConfig(),
				cs: &ClientSet{
					asyncAgentClients:    map[string]service.AsyncAgentServiceClient{},
					agentMetadataClients: map[string]service.AgentMetadataServiceClient{},
				},
			}, nil
		}
		pluginEntry := pluginmachinery.CreateRemotePlugin(agentPlugin)
		plugin, err := pluginEntry.LoadPlugin(context.TODO(), newFakeSetupContext("test2"))
		assert.NoError(t, err)

		tCtx := getTaskContext(t)
		tr := &pluginCoreMocks.TaskReader{}
		tr.OnRead(context.Background()).Return(&template, nil)
		tCtx.OnTaskReader().Return(tr)
		inputReader := &ioMocks.InputReader{}
		inputReader.OnGetInputPrefixPath().Return(basePrefix)
		inputReader.OnGetInputPath().Return(basePrefix + "/inputs.pb")
		inputReader.OnGetMatch(mock.Anything).Return(inputs, nil)
		tCtx.OnInputReader().Return(inputReader)

		trns, err := plugin.Handle(context.Background(), tCtx)
		assert.Nil(t, err)
		assert.Equal(t, trns.Info().Phase(), core.PhaseRetryableFailure)
		err = plugin.Abort(context.Background(), tCtx)
		assert.Nil(t, err)
	})

	t.Run("failed to read task template", func(t *testing.T) {
		tCtx := getTaskContext(t)
		tr := &pluginCoreMocks.TaskReader{}
		tr.OnRead(context.Background()).Return(nil, fmt.Errorf("read fail"))
		tCtx.OnTaskReader().Return(tr)

		agentPlugin := newMockAsyncAgentPlugin()
		pluginEntry := pluginmachinery.CreateRemotePlugin(agentPlugin)
		plugin, err := pluginEntry.LoadPlugin(context.TODO(), newFakeSetupContext("test3"))
		assert.NoError(t, err)

		trns, err := plugin.Handle(context.Background(), tCtx)
		assert.Nil(t, err)
		assert.Equal(t, trns.Info().Phase(), core.PhaseRetryableFailure)
	})

	t.Run("failed to read inputs", func(t *testing.T) {
		tCtx := getTaskContext(t)
		tr := &pluginCoreMocks.TaskReader{}
		tr.OnRead(context.Background()).Return(&template, nil)
		tCtx.OnTaskReader().Return(tr)
		inputReader := &ioMocks.InputReader{}
		inputReader.OnGetInputPrefixPath().Return(basePrefix)
		inputReader.OnGetInputPath().Return(basePrefix + "/inputs.pb")
		inputReader.OnGetMatch(mock.Anything).Return(nil, fmt.Errorf("read fail"))
		tCtx.OnInputReader().Return(inputReader)

		agentPlugin := newMockAsyncAgentPlugin()
		pluginEntry := pluginmachinery.CreateRemotePlugin(agentPlugin)
		plugin, err := pluginEntry.LoadPlugin(context.TODO(), newFakeSetupContext("test4"))
		assert.NoError(t, err)

		trns, err := plugin.Handle(context.Background(), tCtx)
		assert.Nil(t, err)
		assert.Equal(t, trns.Info().Phase(), core.PhaseRetryableFailure)
	})
}

func getTaskContext(t *testing.T) *pluginCoreMocks.TaskExecutionContext {
	latestKnownState := atomic.Value{}
	pluginStateReader := &pluginCoreMocks.PluginStateReader{}
	pluginStateReader.OnGetMatch(mock.Anything).Return(0, nil).Run(func(args mock.Arguments) {
		o := args.Get(0)
		x, err := json.Marshal(latestKnownState.Load())
		assert.NoError(t, err)
		assert.NoError(t, json.Unmarshal(x, &o))
	})
	pluginStateWriter := &pluginCoreMocks.PluginStateWriter{}
	pluginStateWriter.OnPutMatch(mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		latestKnownState.Store(args.Get(1))
	})

	pluginStateWriter.OnReset().Return(nil).Run(func(args mock.Arguments) {
		latestKnownState.Store(nil)
	})

	execID := rand.String(3)
	tID := &pluginCoreMocks.TaskExecutionID{}
	tID.OnGetGeneratedName().Return(execID + "-my-task-1")
	tID.OnGetID().Return(flyteIdlCore.TaskExecutionIdentifier{
		TaskId: &flyteIdlCore.Identifier{
			ResourceType: flyteIdlCore.ResourceType_TASK,
			Project:      "a",
			Domain:       "d",
			Name:         "n",
			Version:      "abc",
		},
		NodeExecutionId: &flyteIdlCore.NodeExecutionIdentifier{
			NodeId: "node1",
			ExecutionId: &flyteIdlCore.WorkflowExecutionIdentifier{
				Project: "a",
				Domain:  "d",
				Name:    "exec",
			},
		},
		RetryAttempt: 0,
	})
	tMeta := &pluginCoreMocks.TaskExecutionMetadata{}
	tMeta.OnGetTaskExecutionID().Return(tID)
	tMeta.OnGetNamespace().Return("test-namespace")
	tMeta.OnGetLabels().Return(map[string]string{"foo": "bar"})
	tMeta.OnGetAnnotations().Return(map[string]string{"foo": "bar"})
	tMeta.OnGetK8sServiceAccount().Return("k8s-account")
	tMeta.OnGetEnvironmentVariables().Return(map[string]string{"foo": "bar"})
	tMeta.OnGetSecurityContext().Return(flyteIdlCore.SecurityContext{
		RunAs: &flyteIdlCore.Identity{ExecutionIdentity: "execution-identity"},
	})
	resourceManager := &pluginCoreMocks.ResourceManager{}
	resourceManager.OnAllocateResourceMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(pluginCore.AllocationStatusGranted, nil)
	resourceManager.OnReleaseResourceMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil)

	basePrefix := storage.DataReference("fake://bucket/prefix/" + execID)
	outputWriter := &ioMocks.OutputWriter{}
	outputWriter.OnGetRawOutputPrefix().Return("/sandbox/")
	outputWriter.OnGetOutputPrefixPath().Return(basePrefix)
	outputWriter.OnGetErrorPath().Return(basePrefix + "/error.pb")
	outputWriter.OnGetOutputPath().Return(basePrefix + "/outputs.pb")
	outputWriter.OnGetCheckpointPrefix().Return("/checkpoint")
	outputWriter.OnGetPreviousCheckpointsPrefix().Return("/prev")

	tCtx := &pluginCoreMocks.TaskExecutionContext{}
	tCtx.OnOutputWriter().Return(outputWriter)
	tCtx.OnResourceManager().Return(resourceManager)
	tCtx.OnPluginStateReader().Return(pluginStateReader)
	tCtx.OnPluginStateWriter().Return(pluginStateWriter)
	tCtx.OnTaskExecutionMetadata().Return(tMeta)
	return tCtx
}

func newMockAsyncAgentPlugin() webapi.PluginEntry {
	asyncAgentClient := new(agentMocks.AsyncAgentServiceClient)

	mockCreateRequestMatcher := mock.MatchedBy(func(request *admin.CreateTaskRequest) bool {
		expectedArgs := []string{"pyflyte-fast-execute", "--output-prefix", "/tmp/123"}
		return slices.Equal(request.Template.GetContainer().Args, expectedArgs)
	})
	asyncAgentClient.On("CreateTask", mock.Anything, mockCreateRequestMatcher).Return(&admin.CreateTaskResponse{
		ResourceMeta: []byte{1, 2, 3, 4}}, nil)

	mockGetRequestMatcher := mock.MatchedBy(func(request *admin.GetTaskRequest) bool {
		return request.GetTaskCategory().GetName() == "spark"
	})
	asyncAgentClient.On("GetTask", mock.Anything, mockGetRequestMatcher).Return(
		&admin.GetTaskResponse{Resource: &admin.Resource{Phase: flyteIdlCore.TaskExecution_SUCCEEDED}}, nil)

	asyncAgentClient.On("DeleteTask", mock.Anything, mock.Anything).Return(
		&admin.DeleteTaskResponse{}, nil)

	cfg := defaultConfig
	cfg.DefaultAgent.Endpoint = "localhost:8000"

	return webapi.PluginEntry{
		ID:                 "agent-service",
		SupportedTaskTypes: []core.TaskType{"bigquery_query_job_task", "spark"},
		PluginLoader: func(ctx context.Context, iCtx webapi.PluginSetupContext) (webapi.AsyncPlugin, error) {
			return Plugin{
				metricScope: iCtx.MetricsScope(),
				cfg:         &cfg,
				cs: &ClientSet{
					asyncAgentClients: map[string]service.AsyncAgentServiceClient{
						defaultAgentEndpoint: asyncAgentClient,
					},
				},
			}, nil
		},
	}
}

func newMockSyncAgentPlugin() webapi.PluginEntry {
	syncAgentClient := new(agentMocks.SyncAgentServiceClient)
	output, _ := coreutils.MakeLiteralMap(map[string]interface{}{"x": 1})
	resource := &admin.Resource{Phase: flyteIdlCore.TaskExecution_SUCCEEDED, Outputs: output}

	stream := new(agentMocks.SyncAgentService_ExecuteTaskSyncClient)
	stream.OnRecv().Return(&admin.ExecuteTaskSyncResponse{
		Res: &admin.ExecuteTaskSyncResponse_Header{
			Header: &admin.ExecuteTaskSyncResponseHeader{
				Resource: resource,
			},
		},
	}, nil).Once()

	stream.OnRecv().Return(nil, io.EOF).Once()
	stream.OnSendMatch(mock.Anything).Return(nil)
	stream.OnCloseSendMatch(mock.Anything).Return(nil)

	syncAgentClient.OnExecuteTaskSyncMatch(mock.Anything).Return(stream, nil)

	cfg := defaultConfig
	cfg.DefaultAgent.Endpoint = defaultAgentEndpoint

	return webapi.PluginEntry{
		ID:                 "agent-service",
		SupportedTaskTypes: []core.TaskType{"openai"},
		PluginLoader: func(ctx context.Context, iCtx webapi.PluginSetupContext) (webapi.AsyncPlugin, error) {
			return Plugin{
				metricScope: iCtx.MetricsScope(),
				cfg:         &cfg,
				cs: &ClientSet{
					syncAgentClients: map[string]service.SyncAgentServiceClient{
						defaultAgentEndpoint: syncAgentClient,
					},
				},
				agentRegistry: Registry{"openai": {defaultTaskTypeVersion: {AgentDeployment: &Deployment{Endpoint: defaultAgentEndpoint}, IsSync: true}}},
			}, nil
		},
	}
}

func newFakeSetupContext(name string) *pluginCoreMocks.SetupContext {
	fakeResourceRegistrar := pluginCoreMocks.ResourceRegistrar{}
	fakeResourceRegistrar.On("RegisterResourceQuota", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	labeled.SetMetricKeys(contextutils.NamespaceKey)

	fakeSetupContext := pluginCoreMocks.SetupContext{}
	fakeSetupContext.OnMetricsScope().Return(promutils.NewScope(name))
	fakeSetupContext.OnResourceRegistrar().Return(&fakeResourceRegistrar)

	return &fakeSetupContext
}
