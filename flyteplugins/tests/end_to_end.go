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

	idlCore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog"
	catalogMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog/mocks"
	pluginCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	coreMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io"
	ioMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/workqueue"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
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
	tr.OnRead(ctx).Return(template, nil)

	inputReader := &ioMocks.InputReader{}
	inputReader.OnGetInputPrefixPath().Return(basePrefix)
	inputReader.OnGetInputPath().Return(basePrefix + "/inputs.pb")
	inputReader.OnGetMatch(mock.Anything).Return(inputs, nil)

	outputWriter := &ioMocks.OutputWriter{}
	outputWriter.OnGetRawOutputPrefix().Return("/sandbox/")
	outputWriter.OnGetOutputPrefixPath().Return(basePrefix)
	outputWriter.OnGetErrorPath().Return(basePrefix + "/error.pb")
	outputWriter.OnGetOutputPath().Return(basePrefix + "/outputs.pb")
	outputWriter.OnGetCheckpointPrefix().Return("/checkpoint")
	outputWriter.OnGetPreviousCheckpointsPrefix().Return("/prev")

	outputWriter.OnPutMatch(mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		or := args.Get(1).(io.OutputReader)
		literals, ee, err := or.Read(ctx)
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
	pluginStateWriter.OnPutMatch(mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		latestKnownState.Store(args.Get(1))
	})

	pluginStateWriter.OnReset().Return(nil).Run(func(args mock.Arguments) {
		latestKnownState.Store(nil)
	})

	pluginStateReader := &coreMocks.PluginStateReader{}
	pluginStateReader.OnGetMatch(mock.Anything).Return(0, nil).Run(func(args mock.Arguments) {
		o := args.Get(0)
		x, err := json.Marshal(latestKnownState.Load())
		assert.NoError(t, err)
		assert.NoError(t, json.Unmarshal(x, &o))
	})
	pluginStateReader.OnGetStateVersion().Return(0)

	tID := &coreMocks.TaskExecutionID{}
	tID.OnGetGeneratedName().Return(execID + "-my-task-1")
	tID.OnGetID().Return(idlCore.TaskExecutionIdentifier{
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
	tID.OnGetUniqueNodeID().Return("unique-node-id")

	overrides := &coreMocks.TaskOverrides{}
	overrides.OnGetConfig().Return(&v1.ConfigMap{Data: map[string]string{
		"dynamic-queue": "queue1",
	}})
	overrides.OnGetResources().Return(&v1.ResourceRequirements{
		Requests: map[v1.ResourceName]resource.Quantity{},
		Limits:   map[v1.ResourceName]resource.Quantity{},
	})
	overrides.OnGetExtendedResources().Return(&idlCore.ExtendedResources{})
	overrides.OnGetContainerImage().Return("")

	tMeta := &coreMocks.TaskExecutionMetadata{}
	tMeta.OnGetTaskExecutionID().Return(tID)
	tMeta.OnGetOverrides().Return(overrides)
	tMeta.OnGetK8sServiceAccount().Return("s")
	tMeta.OnGetNamespace().Return("fake-development")
	tMeta.OnGetMaxAttempts().Return(2)
	tMeta.OnGetSecurityContext().Return(idlCore.SecurityContext{
		RunAs: &idlCore.Identity{
			K8SServiceAccount: "s",
		},
	})
	tMeta.OnGetLabels().Return(map[string]string{})
	tMeta.OnGetAnnotations().Return(map[string]string{})
	tMeta.OnIsInterruptible().Return(true)
	tMeta.OnGetOwnerReference().Return(v12.OwnerReference{})
	tMeta.OnGetOwnerID().Return(types.NamespacedName{
		Namespace: "fake-development",
		Name:      execID,
	})
	tMeta.OnGetPlatformResources().Return(&v1.ResourceRequirements{})
	tMeta.OnGetInterruptibleFailureThreshold().Return(2)
	tMeta.OnGetEnvironmentVariables().Return(nil)
	tMeta.OnGetConsoleURL().Return("")

	catClient := &catalogMocks.Client{}
	catData := sync.Map{}
	catClient.On("Get", mock.Anything, mock.Anything).Return(
		func(ctx context.Context, key catalog.Key) io.OutputReader {
			data, found := catData.Load(key)
			if !found {
				return nil
			}

			or := &ioMocks.OutputReader{}
			or.OnExistsMatch(mock.Anything).Return(true, nil)
			or.OnIsErrorMatch(mock.Anything).Return(false, nil)
			or.OnReadMatch(mock.Anything).Return(data.(*idlCore.LiteralMap), nil, nil)
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
	eRecorder.OnRecordRawMatch(mock.Anything, mock.Anything).Return(nil)

	resourceManager := &coreMocks.ResourceManager{}
	resourceManager.OnAllocateResourceMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(pluginCore.AllocationStatusGranted, nil)
	resourceManager.OnReleaseResourceMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil)

	secretManager := &coreMocks.SecretManager{}
	secretManager.OnGet(ctx, mock.Anything).Return("fake-token", nil)

	tCtx := &coreMocks.TaskExecutionContext{}
	tCtx.OnInputReader().Return(inputReader)
	tCtx.OnTaskRefreshIndicator().Return(func(ctx context.Context) {})
	tCtx.OnOutputWriter().Return(outputWriter)
	tCtx.OnDataStore().Return(ds)
	tCtx.OnTaskReader().Return(tr)
	tCtx.OnPluginStateWriter().Return(pluginStateWriter)
	tCtx.OnPluginStateReader().Return(pluginStateReader)
	tCtx.OnTaskExecutionMetadata().Return(tMeta)
	tCtx.OnCatalog().Return(cat)
	tCtx.OnEventsRecorder().Return(eRecorder)
	tCtx.OnResourceManager().Return(resourceManager)
	tCtx.OnSecretManager().Return(secretManager)

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
