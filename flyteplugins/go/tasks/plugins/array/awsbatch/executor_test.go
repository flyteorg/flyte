package awsbatch

import (
	"context"
	"reflect"
	"testing"

	mocks2 "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io/mocks"

	"github.com/flyteorg/flytestdlib/storage"

	v1 "k8s.io/api/core/v1"

	"github.com/flyteorg/flytestdlib/promutils"
	"k8s.io/apimachinery/pkg/types"

	"github.com/flyteorg/flyteplugins/go/tasks/aws"

	arrayCore "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/core"

	pluginCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	pluginMocks "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	queueMocks "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/workqueue/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array"
	batchConfig "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/awsbatch/config"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/awsbatch/definition"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/awsbatch/mocks"
	cacheMocks "github.com/flyteorg/flytestdlib/cache/mocks"
	"github.com/flyteorg/flytestdlib/utils"
)

func TestExecutor_Handle(t *testing.T) {
	batchClient := NewCustomBatchClient(mocks.NewMockAwsBatchClient(), "", "",
		utils.NewRateLimiter("", 10, 20),
		utils.NewRateLimiter("", 10, 20))

	cache := &cacheMocks.AutoRefresh{}

	js := &JobStore{
		Client:      batchClient,
		AutoRefresh: cache,
		started:     true,
	}

	jc := definition.NewCache(10)

	q := &queueMocks.IndexedWorkQueue{}
	q.OnQueueMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	oa := array.OutputAssembler{
		IndexedWorkQueue: q,
	}

	e := Executor{
		jobStore:           js,
		jobDefinitionCache: jc,
		outputAssembler:    oa,
		errorAssembler:     oa,
	}

	ctx := context.Background()

	tr := &pluginMocks.TaskReader{}
	tr.OnRead(ctx).Return(&core.TaskTemplate{
		Target: &core.TaskTemplate_Container{
			Container: createSampleContainerTask(),
		},
	}, nil)

	inputState := &State{}

	pluginStateReader := &pluginMocks.PluginStateReader{}
	pluginStateReader.On("Get", mock.AnythingOfType(reflect.TypeOf(&State{}).String())).Return(
		func(v interface{}) uint8 {
			*(v.(*State)) = *inputState
			return 0
		},
		func(v interface{}) error {
			return nil
		})

	pluginStateWriter := &pluginMocks.PluginStateWriter{}
	pluginStateWriter.On("Put", mock.Anything, mock.Anything).Return(
		func(stateVersion uint8, v interface{}) error {
			assert.Equal(t, uint8(0), stateVersion)
			actualState, casted := v.(*State)
			assert.True(t, casted)

			actualPhase, _ := actualState.GetPhase()
			assert.Equal(t, arrayCore.PhasePreLaunch, actualPhase)
			return nil
		})

	tID := &pluginMocks.TaskExecutionID{}
	tID.OnGetGeneratedName().Return("notfound")
	tID.OnGetID().Return(core.TaskExecutionIdentifier{
		TaskId: &core.Identifier{
			ResourceType: core.ResourceType_TASK,
			Project:      "a",
			Domain:       "d",
			Name:         "n",
			Version:      "abc",
		},
		NodeExecutionId: &core.NodeExecutionIdentifier{
			NodeId: "node1",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "a",
				Domain:  "d",
				Name:    "exec",
			},
		},
		RetryAttempt: 0,
	})

	overrides := &pluginMocks.TaskOverrides{}
	overrides.OnGetConfig().Return(&v1.ConfigMap{Data: map[string]string{
		DynamicTaskQueueKey: "queue1",
	}})

	tMeta := &pluginMocks.TaskExecutionMetadata{}
	tMeta.OnGetTaskExecutionID().Return(tID)
	tMeta.OnGetOverrides().Return(overrides)

	dataStore, err := storage.NewDataStore(&storage.Config{
		Type: storage.TypeMemory,
	}, promutils.NewTestScope())
	assert.NoError(t, err)

	inputReader := &mocks2.InputReader{}
	inputReader.OnGetInputPrefixPath().Return("/inputs.pb")

	tCtx := &pluginMocks.TaskExecutionContext{}
	tCtx.OnPluginStateReader().Return(pluginStateReader)
	tCtx.OnPluginStateWriter().Return(pluginStateWriter)
	tCtx.OnTaskReader().Return(tr)
	tCtx.OnTaskExecutionMetadata().Return(tMeta)
	tCtx.OnDataStore().Return(dataStore)
	tCtx.OnInputReader().Return(inputReader)

	transition, err := e.Handle(ctx, tCtx)
	assert.NoError(t, err)
	assert.NotNil(t, transition)
	assert.Equal(t, pluginCore.PhaseRunning.String(), transition.Info().Phase().String())
}

func TestExecutor_Start(t *testing.T) {
	awsClient, err := aws.GetClient()
	assert.NoError(t, err)

	exec, err := NewExecutor(context.Background(), awsClient, batchConfig.GetConfig(), func(id types.NamespacedName) error {
		return nil
	}, promutils.NewTestScope())

	assert.NoError(t, err)
	assert.NotNil(t, exec)

	assert.NoError(t, exec.Start(context.Background()))
}
