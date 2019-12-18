// +build manualintegration
// Be sure to add this to your goland build settings Tags in order to get linting/testing
// Set the QUBOLE_API_KEY environment variable in your test to be a real token

package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/struct"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/plugins"
	config2 "github.com/lyft/flytestdlib/config"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/lyft/flyteplugins/go/tasks/v1/qubole_collection"
	"github.com/lyft/flyteplugins/go/tasks/v1/qubole_collection/config"
	"github.com/lyft/flyteplugins/go/tasks/v1/types"
	tasksMocks "github.com/lyft/flyteplugins/go/tasks/v1/types/mocks"
	"github.com/lyft/flyteplugins/go/tasks/v1/utils"
)

func getMockTaskContext() *tasksMocks.TaskContext {
	taskCtx := &tasksMocks.TaskContext{}
	taskCtx.On("GetNamespace").Return("ns")
	taskCtx.On("GetAnnotations").Return(map[string]string{"aKey": "aVal"})
	taskCtx.On("GetLabels").Return(map[string]string{"lKey": "lVal"})
	taskCtx.On("GetOwnerReference").Return(metav1.OwnerReference{Name: "x"})
	taskCtx.On("GetInputsFile").Return(storage.DataReference("/fake/inputs.pb"))
	taskCtx.On("GetDataDir").Return(storage.DataReference("/fake/"))
	taskCtx.On("GetErrorFile").Return(storage.DataReference("/fake/error.pb"))
	taskCtx.On("GetOutputsFile").Return(storage.DataReference("/fake/inputs.pb"))
	taskCtx.On("GetPhaseVersion").Return(uint32(1))

	id := &tasksMocks.TaskExecutionID{}
	id.On("GetGeneratedName").Return("flyteplugins_integration")
	id.On("GetID").Return(core.TaskExecutionIdentifier{})
	taskCtx.On("GetTaskExecutionID").Return(id)

	return taskCtx
}

func getTaskTemplate() core.TaskTemplate {
	hiveJob := plugins.QuboleHiveJob{
		ClusterLabel: "default",
		Tags:         []string{"flyte_plugin_test"},
		QueryCollection: &plugins.HiveQueryCollection{
			Queries: []*plugins.HiveQuery{
				{TimeoutSec: 500,
					Query:      "select 'one'",
					RetryCount: 0},
			},
		},
	}
	stObj := &structpb.Struct{}
	utils.MarshalStruct(&hiveJob, stObj)
	tt := core.TaskTemplate{
		Type:   "hive",
		Custom: stObj,
		Id: &core.Identifier{
			Name:         "integrationtest1",
			Project:      "flyteplugins",
			Version:      "1",
			ResourceType: core.ResourceType_TASK,
		},
	}

	return tt
}

func TestHappy(t *testing.T) {
	_ = logger.SetConfig(&logger.Config{
		IncludeSourceCode: true,
		Level:             6,
	})
	ctx := context.Background()
	testScope := promutils.NewTestScope()
	mockEventRecorder := tasksMocks.EventRecorder{}
	mockEventRecorder.On("RecordTaskEvent", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		fmt.Printf("Event: %v\n", args)
	})

	assert.NoError(t, config.SetQuboleConfig(&config.Config{
		RedisHostPath:          "localhost:6379",
		RedisHostKey:           "mypassword",
		LookasideBufferPrefix:  "test",
		LookasideExpirySeconds: config2.Duration{Duration: time.Second * 1000},
		LruCacheSize:           100,
	}))

	executor, err := qubole_collection.NewHiveTaskExecutorWithCache(ctx)
	assert.NoError(t, err)
	err = executor.Initialize(ctx, types.ExecutorInitializationParameters{
		MetricsScope:  testScope,
		EventRecorder: &mockEventRecorder})
	assert.NoError(t, err)

	var taskCtx *tasksMocks.TaskContext
	taskCtx = getMockTaskContext()
	taskTemplate := getTaskTemplate()

	statuses, err := executor.StartTask(ctx, taskCtx, &taskTemplate, nil)
	assert.NoError(t, err)
	assert.Equal(t, types.TaskPhaseQueued, statuses.Phase)

	// The initial custom state returned by the StartTask function
	customState0 := statuses.State
	work := customState0["flyteplugins_integration_0"].(qubole_collection.QuboleWorkItem)
	assert.Equal(t, qubole_collection.QuboleWorkNotStarted, work.Status)
	assert.Equal(t, 0, work.Retries)
	assert.Equal(t, "flyteplugins_integration_0", work.ID())

	taskCtx.On("GetCustomState").Return(customState0)
	taskCtx.On("GetPhase").Return(types.TaskPhaseQueued)

	for true {
		taskStatus, err := executor.CheckTaskStatus(ctx, taskCtx, &taskTemplate)
		assert.NoError(t, err)
		fmt.Printf("New status phase %s custom state %v\n", taskStatus.Phase, taskStatus.State)
		if taskStatus.Phase == types.TaskPhaseSucceeded {
			fmt.Println("success")
			break
		}
		taskCtx = getMockTaskContext()
		taskCtx.On("GetCustomState").Return(taskStatus.State)
		taskCtx.On("GetPhase").Return(types.TaskPhaseRunning)
		time.Sleep(15 * time.Second)
	}
}

func TestMultipleCallbacks(t *testing.T) {
	_ = logger.SetConfig(&logger.Config{
		IncludeSourceCode: true,
		Level:             6,
	})
	ctx := context.Background()
	testScope := promutils.NewTestScope()
	mockEventRecorder := tasksMocks.EventRecorder{}
	mockEventRecorder.On("RecordTaskEvent", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		fmt.Printf("Event: %v\n", args)
	})

	assert.NoError(t, config.SetQuboleConfig(&config.Config{
		RedisHostPath:          "localhost:6379",
		RedisHostKey:           "mypassword",
		LookasideBufferPrefix:  "test",
		LookasideExpirySeconds: config2.Duration{Duration: time.Second * 1000},
		LruCacheSize:           100,
	}))

	executor, err := qubole_collection.NewHiveTaskExecutorWithCache(ctx)
	assert.NoError(t, err)
	err = executor.Initialize(ctx, types.ExecutorInitializationParameters{
		MetricsScope:  testScope,
		EventRecorder: &mockEventRecorder})
	assert.NoError(t, err)

	var taskCtx *tasksMocks.TaskContext
	taskCtx = getMockTaskContext()
	taskTemplate := getTaskTemplate()

	statuses, err := executor.StartTask(ctx, taskCtx, &taskTemplate, nil)
	assert.NoError(t, err)
	assert.Equal(t, types.TaskPhaseQueued, statuses.Phase)

	// The initial custom state returned by the StartTask function
	customState0 := statuses.State
	work := customState0["flyteplugins_integration_0"].(qubole_collection.QuboleWorkItem)
	assert.Equal(t, qubole_collection.QuboleWorkNotStarted, work.Status)
	assert.Equal(t, 0, work.Retries)
	assert.Equal(t, "flyteplugins_integration_0", work.ID())

	taskCtx.On("GetCustomState").Return(customState0)
	taskCtx.On("GetPhase").Return(types.TaskPhaseQueued)

	for true {
		taskStatus, err := executor.CheckTaskStatus(ctx, taskCtx, &taskTemplate)
		assert.NoError(t, err)
		for i := 0; i < 100; i++ {
			taskStatus, err = executor.CheckTaskStatus(ctx, taskCtx, &taskTemplate)
			assert.NoError(t, err)
		}
		fmt.Printf("New status phase %s custom state %v\n", taskStatus.Phase, taskStatus.State)
		if taskStatus.Phase == types.TaskPhaseSucceeded {
			fmt.Println("success")
			break
		}
		taskCtx = getMockTaskContext()
		taskCtx.On("GetCustomState").Return(taskStatus.State)
		taskCtx.On("GetPhase").Return(types.TaskPhaseRunning)
		time.Sleep(15 * time.Second)
	}
}