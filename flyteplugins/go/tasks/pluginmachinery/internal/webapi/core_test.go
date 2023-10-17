package webapi

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/util/rand"

	flyteIdlCore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	ioMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/webapi"
	webapiMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/webapi/mocks"
	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

func TestHandle(t *testing.T) {
	ctx := context.TODO()
	tCtx := getTaskContext(t)
	taskReader := new(mocks.TaskReader)

	template := flyteIdlCore.TaskTemplate{
		Type: "api_task",
		Metadata: &flyteIdlCore.TaskMetadata{
			Runtime: &flyteIdlCore.RuntimeMetadata{
				Flavor: "sync_plugin",
			},
		},
	}

	taskReader.On("Read", mock.Anything).Return(&template, nil)

	tCtx.On("TaskReader").Return(taskReader)

	p := new(webapiMocks.SyncPlugin)
	p.On("Do", ctx, tCtx).Return(core.PhaseInfo{}, nil)

	c := CorePlugin{
		id: "test",
		p:  p,
	}

	transition, err := c.Handle(ctx, tCtx)
	assert.NoError(t, err)
	assert.NotNil(t, transition)
}

func Test_validateConfig(t *testing.T) {
	t.Run("In range", func(t *testing.T) {
		cfg := webapi.PluginConfig{
			ReadRateLimiter: webapi.RateLimiterConfig{
				QPS:   10,
				Burst: 100,
			},
			WriteRateLimiter: webapi.RateLimiterConfig{
				QPS:   10,
				Burst: 100,
			},
			Caching: webapi.CachingConfig{
				Size:           10,
				ResyncInterval: config.Duration{Duration: 10 * time.Second},
				Workers:        10,
			},
		}

		assert.NoError(t, validateConfig(cfg))
	})

	t.Run("Below min", func(t *testing.T) {
		cfg := webapi.PluginConfig{
			ReadRateLimiter: webapi.RateLimiterConfig{
				QPS:   0,
				Burst: 0,
			},
			WriteRateLimiter: webapi.RateLimiterConfig{
				QPS:   0,
				Burst: 0,
			},
			Caching: webapi.CachingConfig{
				Size:           0,
				ResyncInterval: config.Duration{Duration: 0 * time.Second},
				Workers:        0,
			},
		}

		err := validateConfig(cfg)
		assert.Error(t, err)
		assert.Equal(t, "\ncache size is expected to be between 10 and 500000. Provided value is 0\nworkers count is expected to be between 1 and 100. Provided value is 0\nresync interval is expected to be between 5 and 3600. Provided value is 0\nread burst is expected to be between 5 and 10000. Provided value is 0\nread qps is expected to be between 1 and 100000. Provided value is 0\nwrite burst is expected to be between 5 and 10000. Provided value is 0\nwrite qps is expected to be between 1 and 100000. Provided value is 0", err.Error())
	})

	t.Run("Above max", func(t *testing.T) {
		cfg := webapi.PluginConfig{
			ReadRateLimiter: webapi.RateLimiterConfig{
				QPS:   1000,
				Burst: 1000000,
			},
			WriteRateLimiter: webapi.RateLimiterConfig{
				QPS:   1000,
				Burst: 1000000,
			},
			Caching: webapi.CachingConfig{
				Size:           1000000000,
				ResyncInterval: config.Duration{Duration: 10000 * time.Hour},
				Workers:        1000000000,
			},
		}

		err := validateConfig(cfg)
		assert.Error(t, err)
		assert.Equal(t, "\ncache size is expected to be between 10 and 500000. Provided value is 1000000000\nworkers count is expected to be between 1 and 100. Provided value is 1000000000\nresync interval is expected to be between 5 and 3600. Provided value is 3.6e+07\nread burst is expected to be between 5 and 10000. Provided value is 1000000\nwrite burst is expected to be between 5 and 10000. Provided value is 1000000", err.Error())
	})
}

func TestCreateRemotePlugin(t *testing.T) {
	CreateRemotePlugin(webapi.PluginEntry{
		ID:                 "MyTestPlugin",
		SupportedTaskTypes: []core.TaskType{"test-task"},
		PluginLoader: func(ctx context.Context, iCtx webapi.PluginSetupContext) (webapi.Plugin, error) {
			return newPluginWithProperties(webapi.PluginConfig{
				Caching: webapi.CachingConfig{
					Size: 10,
				},
			}), nil
		},
		IsDefault:           false,
		DefaultForTaskTypes: []core.TaskType{"test-task"},
	})
}

func getTaskContext(t *testing.T) *mocks.TaskExecutionContext {
	latestKnownState := atomic.Value{}
	pluginStateReader := &mocks.PluginStateReader{}
	pluginStateReader.OnGetMatch(mock.Anything).Return(0, nil).Run(func(args mock.Arguments) {
		o := args.Get(0)
		x, err := json.Marshal(latestKnownState.Load())
		assert.NoError(t, err)
		assert.NoError(t, json.Unmarshal(x, &o))
	})
	pluginStateWriter := &mocks.PluginStateWriter{}
	pluginStateWriter.OnPutMatch(mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		latestKnownState.Store(args.Get(1))
	})

	pluginStateWriter.OnReset().Return(nil).Run(func(args mock.Arguments) {
		latestKnownState.Store(nil)
	})

	execID := rand.String(3)
	tID := &mocks.TaskExecutionID{}
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
	tMeta := &mocks.TaskExecutionMetadata{}
	tMeta.OnGetTaskExecutionID().Return(tID)
	resourceManager := &mocks.ResourceManager{}
	resourceManager.OnAllocateResourceMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(core.AllocationStatusGranted, nil)
	resourceManager.OnReleaseResourceMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil)

	basePrefix := storage.DataReference("fake://bucket/prefix/" + execID)
	outputWriter := &ioMocks.OutputWriter{}
	outputWriter.OnGetRawOutputPrefix().Return("/sandbox/")
	outputWriter.OnGetOutputPrefixPath().Return(basePrefix)
	outputWriter.OnGetErrorPath().Return(basePrefix + "/error.pb")
	outputWriter.OnGetOutputPath().Return(basePrefix + "/outputs.pb")
	outputWriter.OnGetCheckpointPrefix().Return("/checkpoint")
	outputWriter.OnGetPreviousCheckpointsPrefix().Return("/prev")

	tCtx := &mocks.TaskExecutionContext{}
	tCtx.OnOutputWriter().Return(outputWriter)
	tCtx.OnResourceManager().Return(resourceManager)
	tCtx.OnPluginStateReader().Return(pluginStateReader)
	tCtx.OnPluginStateWriter().Return(pluginStateWriter)
	tCtx.OnTaskExecutionMetadata().Return(tMeta)
	return tCtx
}
