package tests

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/catalog"
	catalogMocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/catalog/mocks"
	pluginCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	coreMocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/io"
	ioMocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/workqueue"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	idlCore "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

func createSampleContainerTask() *idlCore.Container {
	return &idlCore.Container{
		Command: []string{"cmd"},
		Args:    []string{"{{$inputPrefix}}"},
		Image:   "img1",
		Config: []*idlCore.KeyValuePair{
			{
				Key:   "dynamic_queue",
				Value: "queue1",
			},
		},
	}
}

func BuildTaskTemplate() *idlCore.TaskTemplate {
	return &idlCore.TaskTemplate{
		Target: &idlCore.TaskTemplate_Container{
			Container: createSampleContainerTask(),
		},
	}
}

func RunPluginEndToEndTest(t *testing.T, executor pluginCore.Plugin, template *idlCore.TaskTemplate,
	inputs *idlCore.LiteralMap, expectedOutputs *idlCore.LiteralMap, expectedFailure *idlCore.ExecutionError,
	iterationUpdate func(ctx context.Context, tCtx pluginCore.TaskExecutionContext) error) pluginCore.PhaseInfo {

	ctx := context.Background()

	ds, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)

	execID := rand.String(3)

	basePrefix := storage.DataReference("fake://bucket/prefix/" + execID)
	assert.NoError(t, ds.WriteProtobuf(ctx, basePrefix+"/inputs.pb", storage.Options{}, inputs))

	tr := &coreMocks.TaskReader{}
	tr.EXPECT().Read(ctx).Return(template, nil)

	inputReader := &ioMocks.InputReader{}
	inputReader.EXPECT().GetInputPrefixPath().Return(basePrefix)
	inputReader.EXPECT().GetInputPath().Return(basePrefix + "/inputs.pb")
	inputReader.EXPECT().Get(mock.Anything).Return(inputs, nil)

	outputWriter := &ioMocks.OutputWriter{}
	outputWriter.EXPECT().GetRawOutputPrefix().Return("/sandbox/")
	outputWriter.EXPECT().GetOutputPrefixPath().Return(basePrefix)
	outputWriter.EXPECT().GetErrorPath().Return(basePrefix + "/error.pb")
	outputWriter.EXPECT().GetOutputPath().Return(basePrefix + "/outputs.pb")
	outputWriter.EXPECT().GetCheckpointPrefix().Return("/checkpoint")
	outputWriter.EXPECT().GetPreviousCheckpointsPrefix().Return("/prev")

	outputWriter.EXPECT().Put(mock.Anything, mock.Anything).Return(nil).Run(func(ctx context.Context, reader io.OutputReader) {
		literals, ee, err := reader.Read(ctx)
		assert.NoError(t, err)

		if ee != nil {
			assert.NoError(t, ds.WriteProtobuf(ctx, outputWriter.GetErrorPath(), storage.Options{}, ee))
		}

		if literals != nil {
			assert.NoError(t, ds.WriteProtobuf(ctx, outputWriter.GetOutputPath(), storage.Options{}, literals))
		}
	})

	pluginStateWriter := &coreMocks.PluginStateWriter{}
	latestKnownState := atomic.Value{}
	pluginStateWriter.EXPECT().Put(mock.Anything, mock.Anything).Return(nil).Run(func(_ uint8, v interface{}) {
		latestKnownState.Store(v)
	})

	pluginStateWriter.EXPECT().Reset().Return(nil).Run(func() {
		latestKnownState.Store(nil)
	})

	pluginStateReader := &coreMocks.PluginStateReader{}
	pluginStateReader.EXPECT().Get(mock.Anything).Return(0, nil).Run(func(state interface{}) {
		x, err := json.Marshal(latestKnownState.Load())
		assert.NoError(t, err)
		assert.NoError(t, json.Unmarshal(x, &state))
	})
	pluginStateReader.EXPECT().GetStateVersion().Return(0)

	tID := &coreMocks.TaskExecutionID{}
	tID.EXPECT().GetGeneratedName().Return(execID + "-my-task-1")
	tID.EXPECT().GetID().Return(idlCore.TaskExecutionIdentifier{
		TaskId: &idlCore.Identifier{
			ResourceType: idlCore.ResourceType_TASK,
			Project:      "a",
			Domain:       "d",
			Name:         "n",
			Version:      "abc",
		},
		NodeExecutionId: &idlCore.NodeExecutionIdentifier{
			NodeId: "node1",
			ExecutionId: &idlCore.WorkflowExecutionIdentifier{
				Project: "a",
				Domain:  "d",
				Name:    "exec",
			},
		},
		RetryAttempt: 0,
	})
	tID.EXPECT().GetUniqueNodeID().Return("unique-node-id")

	overrides := &coreMocks.TaskOverrides{}
	overrides.EXPECT().GetConfigMap().Return(&v1.ConfigMap{Data: map[string]string{
		"dynamic-queue": "queue1",
	}})
	overrides.EXPECT().GetResources().Return(&v1.ResourceRequirements{
		Requests: map[v1.ResourceName]resource.Quantity{},
		Limits:   map[v1.ResourceName]resource.Quantity{},
	})
	overrides.EXPECT().GetExtendedResources().Return(&idlCore.ExtendedResources{})
	overrides.EXPECT().GetContainerImage().Return("")
	overrides.EXPECT().GetPodTemplate().Return(nil)

	connections := map[string]pluginCore.ConnectionWrapper{
		"my-openai": {
			Connection: idlCore.Connection{
				TaskType: "openai",
				Secrets:  map[string]string{"key": "value"},
				Configs:  map[string]string{"key": "value"},
			},
			Source: common.AttributesSource_GLOBAL,
		},
	}

	tMeta := &coreMocks.TaskExecutionMetadata{}
	tMeta.EXPECT().GetTaskExecutionID().Return(tID)
	tMeta.EXPECT().GetOverrides().Return(overrides)
	tMeta.EXPECT().GetK8sServiceAccount().Return("s")
	tMeta.EXPECT().GetNamespace().Return("fake-development")
	tMeta.EXPECT().GetMaxAttempts().Return(2)
	tMeta.EXPECT().GetSecurityContext().Return(idlCore.SecurityContext{
		RunAs: &idlCore.Identity{
			K8SServiceAccount: "s",
		},
	})
	tMeta.EXPECT().GetLabels().Return(map[string]string{"organization": "flyte", "project": "flytesnacks", "domain": "development"})
	tMeta.EXPECT().GetAnnotations().Return(map[string]string{})
	tMeta.EXPECT().IsInterruptible().Return(true)
	tMeta.EXPECT().GetOwnerReference().Return(v12.OwnerReference{})
	tMeta.EXPECT().GetOwnerID().Return(types.NamespacedName{
		Namespace: "fake-development",
		Name:      execID,
	})
	tMeta.EXPECT().GetPlatformResources().Return(&v1.ResourceRequirements{})
	tMeta.EXPECT().GetInterruptibleFailureThreshold().Return(2)
	tMeta.EXPECT().GetEnvironmentVariables().Return(nil)
	tMeta.EXPECT().GetExternalResourceAttributes().Return(pluginCore.ExternalResourceAttributes{Connections: connections})
	tMeta.EXPECT().GetConsoleURL().Return("")

	catClient := &catalogMocks.Client{}
	catData := sync.Map{}
	catClient.On("Get", mock.Anything, mock.Anything).Return(
		func(ctx context.Context, key catalog.Key) io.OutputReader {
			data, found := catData.Load(key)
			if !found {
				return nil
			}

			or := &ioMocks.OutputReader{}
			or.EXPECT().Exists(mock.Anything).Return(true, nil)
			or.EXPECT().IsError(mock.Anything).Return(false, nil)
			or.EXPECT().Read(mock.Anything).Return(data.(*idlCore.LiteralMap), nil, nil)
			return or
		},
		func(ctx context.Context, key catalog.Key) error {
			_, found := catData.Load(key)
			if !found {
				return status.Error(codes.NotFound, "No output found for key")
			}

			return nil
		})
	catClient.On(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		Run(func(args mock.Arguments) {
			key := args.Get(1).(catalog.Key)
			or := args.Get(2).(io.OutputReader)
			o, ee, err := or.Read(ctx)
			assert.NoError(t, err)
			// TODO: Outputting error is not yet supported.
			assert.Nil(t, ee)
			catData.Store(key, o)
		})
	cat, err := catalog.NewAsyncClient(catClient, catalog.Config{
		ReaderWorkqueueConfig: workqueue.Config{
			MaxRetries:         0,
			Workers:            2,
			IndexCacheMaxItems: 100,
		},
		WriterWorkqueueConfig: workqueue.Config{
			MaxRetries:         0,
			Workers:            2,
			IndexCacheMaxItems: 100,
		},
	}, promutils.NewTestScope())
	assert.NoError(t, err)
	assert.NoError(t, cat.Start(ctx))

	eRecorder := &coreMocks.EventsRecorder{}
	eRecorder.EXPECT().RecordRaw(mock.Anything, mock.Anything).Return(nil)

	resourceManager := &coreMocks.ResourceManager{}
	resourceManager.EXPECT().AllocateResource(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(pluginCore.AllocationStatusGranted, nil)
	resourceManager.EXPECT().ReleaseResource(mock.Anything, mock.Anything, mock.Anything).Return(nil)

	secretManager := &coreMocks.SecretManager{}
	secretManager.EXPECT().Get(ctx, mock.Anything).Return("fake-token", nil)

	tCtx := &coreMocks.TaskExecutionContext{}
	tCtx.EXPECT().InputReader().Return(inputReader)
	tCtx.EXPECT().TaskRefreshIndicator().Return(func(ctx context.Context) {})
	tCtx.EXPECT().OutputWriter().Return(outputWriter)
	tCtx.EXPECT().DataStore().Return(ds)
	tCtx.EXPECT().TaskReader().Return(tr)
	tCtx.EXPECT().PluginStateWriter().Return(pluginStateWriter)
	tCtx.EXPECT().PluginStateReader().Return(pluginStateReader)
	tCtx.EXPECT().TaskExecutionMetadata().Return(tMeta)
	tCtx.EXPECT().Catalog().Return(cat)
	tCtx.EXPECT().EventsRecorder().Return(eRecorder)
	tCtx.EXPECT().ResourceManager().Return(resourceManager)
	tCtx.EXPECT().SecretManager().Return(secretManager)

	trns := pluginCore.DoTransition(pluginCore.PhaseInfoQueued(time.Now(), 0, ""))
	for !trns.Info().Phase().IsTerminal() {
		trns, err = executor.Handle(ctx, tCtx)
		assert.NoError(t, err)
		if iterationUpdate != nil {
			assert.NoError(t, iterationUpdate(ctx, tCtx))
		}
	}

	assert.NoError(t, err)
	if expectedOutputs != nil {
		assert.True(t, trns.Info().Phase().IsSuccess())
		actualOutputs := &idlCore.LiteralMap{}
		assert.NoError(t, ds.ReadProtobuf(context.TODO(), outputWriter.GetOutputPath(), actualOutputs))

		if diff := deep.Equal(expectedOutputs, actualOutputs); diff != nil {
			t.Errorf("Expected != Actual. Diff: %v", diff)
		}
	} else if expectedFailure != nil {
		assert.True(t, trns.Info().Phase().IsFailure())
		actualError := &idlCore.ExecutionError{}
		assert.NoError(t, ds.ReadProtobuf(context.TODO(), outputWriter.GetErrorPath(), actualError))

		if diff := deep.Equal(expectedFailure, actualError); diff != nil {
			t.Errorf("Expected != Actual. Diff: %v", diff)
		}
	}

	return trns.Info()
}
