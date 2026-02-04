package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/utils/strings/slices"

	"github.com/flyteorg/flyte/v2/flyteidl2/clients/go/coreutils"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	pluginCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	pluginCoreMocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	ioMocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/webapi"
	"github.com/flyteorg/flyte/v2/flyteplugins/tests"
	"github.com/flyteorg/flyte/v2/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/flytestdlib/utils"
	connectorPb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/connector"
	connectorMocks "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/connector/mocks"
	flyteIdlCore "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/plugins"
)

func TestEndToEnd(t *testing.T) {
	iter := func(ctx context.Context, tCtx pluginCore.TaskExecutionContext) error {
		return nil
	}

	cfg := defaultConfig
	cfg.WebAPI.ResourceQuotas = map[core.ResourceNamespace]int{}
	cfg.WebAPI.Caching.Workers = 1
	cfg.WebAPI.Caching.ResyncInterval.Duration = 5 * time.Second
	cfg.DefaultConnector.Endpoint = "localhost:8000"
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
		pluginEntry := pluginmachinery.CreateRemotePlugin(newMockAsyncConnectorPlugin())
		plugin, err := pluginEntry.LoadPlugin(context.TODO(), newFakeSetupContext("async task"))
		assert.NoError(t, err)

		phase := tests.RunPluginEndToEndTest(t, plugin, &template, inputs, nil, nil, iter)
		assert.Equal(t, true, phase.Phase().IsSuccess())

		template.Type = "spark"
		phase = tests.RunPluginEndToEndTest(t, plugin, &template, inputs, nil, nil, iter)
		assert.Equal(t, true, phase.Phase().IsSuccess())
	})

	t.Run("failed to create a job", func(t *testing.T) {
		connectorPlugin := newMockAsyncConnectorPlugin()
		connectorPlugin.PluginLoader = func(ctx context.Context, iCtx webapi.PluginSetupContext) (webapi.AsyncPlugin, error) {
			return &Plugin{
				metricScope: iCtx.MetricsScope(),
				cfg:         GetConfig(),
				cs: &ClientSet{
					asyncConnectorClients:    map[string]connectorPb.AsyncConnectorServiceClient{},
					connectorMetadataClients: map[string]connectorPb.ConnectorMetadataServiceClient{},
				},
			}, nil
		}
		pluginEntry := pluginmachinery.CreateRemotePlugin(connectorPlugin)
		plugin, err := pluginEntry.LoadPlugin(context.TODO(), newFakeSetupContext("test2"))
		assert.NoError(t, err)

		tCtx := getTaskContext(t)
		tr := &pluginCoreMocks.TaskReader{}
		tr.On("Read", mock.Anything).Return(&template, nil)
		tCtx.On("TaskReader").Return(tr)
		inputReader := &ioMocks.InputReader{}
		inputReader.On("GetInputPrefixPath").Return(basePrefix)
		inputReader.On("GetInputPath").Return(basePrefix + "/inputs.pb")
		inputReader.On("Get", mock.Anything).Return(inputs, nil)
		tCtx.On("InputReader").Return(inputReader)

		trns, err := plugin.Handle(context.Background(), tCtx)
		assert.Nil(t, err)
		assert.Equal(t, trns.Info().Phase(), core.PhaseRetryableFailure)
		err = plugin.Abort(context.Background(), tCtx)
		assert.Nil(t, err)
	})

	t.Run("failed to read task template", func(t *testing.T) {
		tCtx := getTaskContext(t)

		tr := &pluginCoreMocks.TaskReader{}
		tr.On("Read", mock.Anything).Return(nil, fmt.Errorf("read fail"))
		tCtx.On("TaskReader").Return(tr)

		connectorPlugin := newMockAsyncConnectorPlugin()
		pluginEntry := pluginmachinery.CreateRemotePlugin(connectorPlugin)
		plugin, err := pluginEntry.LoadPlugin(context.TODO(), newFakeSetupContext("test3"))
		assert.NoError(t, err)

		trns, err := plugin.Handle(context.Background(), tCtx)
		assert.Nil(t, err)
		assert.Equal(t, trns.Info().Phase(), core.PhaseRetryableFailure)
	})

	t.Run("failed to read inputs", func(t *testing.T) {
		tCtx := getTaskContext(t)
		tr := &pluginCoreMocks.TaskReader{}
		tr.On("Read", mock.Anything).Return(&template, nil)
		tCtx.On("TaskReader").Return(tr)
		inputReader := &ioMocks.InputReader{}
		inputReader.On("GetInputPrefixPath").Return(basePrefix)
		inputReader.On("GetInputPath").Return(basePrefix + "/inputs.pb")
		inputReader.On("Get", mock.Anything).Return(nil, fmt.Errorf("read fail"))
		tCtx.On("InputReader").Return(inputReader)

		connectorPlugin := newMockAsyncConnectorPlugin()
		pluginEntry := pluginmachinery.CreateRemotePlugin(connectorPlugin)
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
	pluginStateReader.On("Get", mock.Anything).Return(uint8(0), nil).Run(func(args mock.Arguments) {
		o := args.Get(0)
		x, err := json.Marshal(latestKnownState.Load())
		assert.NoError(t, err)
		assert.NoError(t, json.Unmarshal(x, &o))
	})

	pluginStateWriter := &pluginCoreMocks.PluginStateWriter{}
	pluginStateWriter.On("Put", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		latestKnownState.Store(args.Get(1))
	})
	pluginStateWriter.On("Reset").Return(nil).Run(func(args mock.Arguments) {
		latestKnownState.Store(nil)
	})

	execID := rand.String(3)
	tID := &pluginCoreMocks.TaskExecutionID{}
	tID.On("GetGeneratedName").Return(execID + "-my-task-1")
	tID.On("GetID").Return(flyteIdlCore.TaskExecutionIdentifier{
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
	tMeta.On("GetTaskExecutionID").Return(tID)
	tMeta.On("GetNamespace").Return("test-namespace")
	tMeta.On("GetLabels").Return(map[string]string{"foo": "bar"})
	tMeta.On("GetAnnotations").Return(map[string]string{"foo": "bar"})
	tMeta.On("GetK8sServiceAccount").Return("k8s-account")
	tMeta.On("GetEnvironmentVariables").Return(map[string]string{"foo": "bar"})
	tMeta.On("GetSecurityContext").Return(flyteIdlCore.SecurityContext{
		RunAs: &flyteIdlCore.Identity{ExecutionIdentity: "execution-identity"},
	})

	resourceManager := &pluginCoreMocks.ResourceManager{}
	resourceManager.On("AllocateResource", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(pluginCore.AllocationStatusGranted, nil)
	resourceManager.On("ReleaseResource", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	basePrefix := storage.DataReference("fake://bucket/prefix/" + execID)
	outputWriter := &ioMocks.OutputWriter{}
	outputWriter.On("GetRawOutputPrefix").Return(storage.DataReference("/sandbox/"))
	outputWriter.On("GetOutputPrefixPath").Return(basePrefix)
	outputWriter.On("GetErrorPath").Return(basePrefix + "/error.pb")
	outputWriter.On("GetOutputPath").Return(basePrefix + "/outputs.pb")
	outputWriter.On("GetCheckpointPrefix").Return(storage.DataReference("/checkpoint"))
	outputWriter.On("GetPreviousCheckpointsPrefix").Return(storage.DataReference("/prev"))

	tCtx := &pluginCoreMocks.TaskExecutionContext{}
	tCtx.On("OutputWriter").Return(outputWriter)
	tCtx.On("ResourceManager").Return(resourceManager)
	tCtx.On("PluginStateReader").Return(pluginStateReader)
	tCtx.On("PluginStateWriter").Return(pluginStateWriter)
	tCtx.On("TaskExecutionMetadata").Return(tMeta)

	return tCtx
}

func newMockAsyncConnectorPlugin() webapi.PluginEntry {
	asyncConnectorClient := new(connectorMocks.AsyncConnectorServiceClient)
	registryKey := RegistryKey{domain: "", taskTypeName: "spark", taskTypeVersion: defaultTaskTypeVersion}
	connectorRegistry := Registry{registryKey: {ConnectorDeployment: &Deployment{Endpoint: defaultConnectorEndpoint}}}

	mockCreateRequestMatcher := mock.MatchedBy(func(request *connectorPb.CreateTaskRequest) bool {
		expectedArgs := []string{"pyflyte-fast-execute", "--output-prefix", "/tmp/123"}
		return slices.Equal(request.Template.GetContainer().Args, expectedArgs)
	})
	asyncConnectorClient.On("CreateTask", mock.Anything, mockCreateRequestMatcher).Return(&connectorPb.CreateTaskResponse{
		ResourceMeta: []byte{1, 2, 3, 4}}, nil)

	mockGetRequestMatcher := mock.MatchedBy(func(request *connectorPb.GetTaskRequest) bool {
		return request.GetTaskCategory().GetName() == "spark"
	})
	asyncConnectorClient.On("GetTask", mock.Anything, mockGetRequestMatcher).Return(
		&connectorPb.GetTaskResponse{Resource: &connectorPb.Resource{Phase: flyteIdlCore.TaskExecution_SUCCEEDED}}, nil)

	asyncConnectorClient.On("DeleteTask", mock.Anything, mock.Anything).Return(
		&connectorPb.DeleteTaskResponse{}, nil)

	cfg := defaultConfig
	cfg.DefaultConnector.Endpoint = "localhost:8000"

	return webapi.PluginEntry{
		ID:                 "connector-service",
		SupportedTaskTypes: []core.TaskType{"bigquery_query_job_task", "spark"},
		PluginLoader: func(ctx context.Context, iCtx webapi.PluginSetupContext) (webapi.AsyncPlugin, error) {
			return &Plugin{
				metricScope: iCtx.MetricsScope(),
				cfg:         &cfg,
				cs: &ClientSet{
					asyncConnectorClients: map[string]connectorPb.AsyncConnectorServiceClient{
						defaultConnectorEndpoint: asyncConnectorClient,
					},
				},
				registry: connectorRegistry,
			}, nil
		},
	}
}

func newFakeSetupContext(name string) *pluginCoreMocks.SetupContext {
	fakeResourceRegistrar := pluginCoreMocks.ResourceRegistrar{}
	fakeResourceRegistrar.On("RegisterResourceQuota", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	labeled.SetMetricKeys(contextutils.NamespaceKey)

	fakeSetupContext := pluginCoreMocks.SetupContext{}
	fakeSetupContext.On("MetricsScope").Return(promutils.NewScope(name))
	fakeSetupContext.On("ResourceRegistrar").Return(&fakeResourceRegistrar)

	return &fakeSetupContext
}
