package awsbatch

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
	v1 "k8s.io/api/core/v1"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/array/awsbatch/config"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/array/awsbatch/definition"
	batchMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/array/awsbatch/mocks"
	arrayCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/array/core"
	"github.com/flyteorg/flyte/flytestdlib/utils"
)

func TestContainerImageRepository(t *testing.T) {
	testCases := [][]string{
		{"myrepo/test", "test"},
		{"ubuntu", "ubuntu"},
		{"registry/test:1.1", "test"},
		{"ubuntu:1.1", "ubuntu"},
		{"registry/ubuntu:1.1", "ubuntu"},
		{"test.custom.domain.net:1234/prefix/key/repo:0.0.2alpha_dev", "repo"},
	}

	for _, testCase := range testCases {
		t.Run(testCase[1], func(t *testing.T) {
			actual := containerImageRepository(testCase[0])
			assert.Equal(t, testCase[1], actual)
		})
	}
}

func TestEnsureJobDefinition(t *testing.T) {
	ctx := context.Background()

	tReader := &mocks.TaskReader{}
	tReader.EXPECT().Read(mock.Anything).Return(&core.TaskTemplate{
		Interface: &core.TypedInterface{
			Outputs: &core.VariableMap{
				Variables: map[string]*core.Variable{"var1": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}},
			},
		},
		Target: &core.TaskTemplate_Container{
			Container: createSampleContainerTask(),
		},
	}, nil)

	overrides := &mocks.TaskOverrides{}
	overrides.EXPECT().GetConfig().Return(&v1.ConfigMap{Data: map[string]string{
		DynamicTaskQueueKey: "queue1",
	}})

	tID := &mocks.TaskExecutionID{}
	tID.EXPECT().GetGeneratedName().Return("found")

	tMeta := &mocks.TaskExecutionMetadata{}
	tMeta.EXPECT().GetTaskExecutionID().Return(tID)
	tMeta.EXPECT().GetOverrides().Return(overrides)
	tMeta.EXPECT().GetAnnotations().Return(map[string]string{})
	tMeta.EXPECT().GetSecurityContext().Return(core.SecurityContext{})
	tCtx := &mocks.TaskExecutionContext{}
	tCtx.EXPECT().TaskReader().Return(tReader)
	tCtx.EXPECT().TaskExecutionMetadata().Return(tMeta)

	cfg := &config.Config{}
	batchClient := NewCustomBatchClient(batchMocks.NewMockAwsBatchClient(), "", "",
		utils.NewRateLimiter("", 10, 20),
		utils.NewRateLimiter("", 10, 20))

	t.Run("Not Found", func(t *testing.T) {
		dCache := definition.NewCache(10)

		nextState, err := EnsureJobDefinition(ctx, tCtx, cfg, batchClient, dCache, &State{
			State: &arrayCore.State{},
		}, 0)

		assert.NoError(t, err)
		assert.NotNil(t, nextState)
		assert.Equal(t, "my-arn", nextState.JobDefinitionArn)
		p, v := nextState.GetPhase()
		assert.Equal(t, arrayCore.PhaseLaunch, p)
		assert.Zero(t, v)
	})

	t.Run("Found", func(t *testing.T) {
		dCache := definition.NewCache(10)
		assert.NoError(t, dCache.Put(definition.NewCacheKey("", "img1", defaultComputeEngine), "their-arn"))

		nextState, err := EnsureJobDefinition(ctx, tCtx, cfg, batchClient, dCache, &State{
			State: &arrayCore.State{},
		}, 0)
		assert.NoError(t, err)
		assert.NotNil(t, nextState)
		assert.Equal(t, "their-arn", nextState.JobDefinitionArn)
	})

	t.Run("Test New Cache Key", func(t *testing.T) {
		cacheKey := definition.NewCacheKey("default", "img1", defaultComputeEngine)
		assert.Equal(t, cacheKey.String(), "img1-default-EC2")
	})
}

func TestEnsureJobDefinitionWithSecurityContext(t *testing.T) {
	ctx := context.Background()

	tReader := &mocks.TaskReader{}
	tReader.EXPECT().Read(mock.Anything).Return(&core.TaskTemplate{
		Interface: &core.TypedInterface{
			Outputs: &core.VariableMap{
				Variables: map[string]*core.Variable{"var1": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}},
			},
		},
		Target: &core.TaskTemplate_Container{
			Container: createSampleContainerTask(),
		},
		Config: map[string]string{platformCapabilitiesConfigKey: defaultComputeEngine},
	}, nil)

	overrides := &mocks.TaskOverrides{}
	overrides.EXPECT().GetConfig().Return(&v1.ConfigMap{Data: map[string]string{
		DynamicTaskQueueKey: "queue1",
	}})

	tID := &mocks.TaskExecutionID{}
	tID.EXPECT().GetGeneratedName().Return("found")

	tMeta := &mocks.TaskExecutionMetadata{}
	tMeta.EXPECT().GetTaskExecutionID().Return(tID)
	tMeta.EXPECT().GetOverrides().Return(overrides)
	tMeta.EXPECT().GetAnnotations().Return(map[string]string{})
	tMeta.EXPECT().GetSecurityContext().Return(core.SecurityContext{
		RunAs: &core.Identity{IamRole: "new-role"},
	})
	tCtx := &mocks.TaskExecutionContext{}
	tCtx.EXPECT().TaskReader().Return(tReader)
	tCtx.EXPECT().TaskExecutionMetadata().Return(tMeta)

	cfg := &config.Config{}
	batchClient := NewCustomBatchClient(batchMocks.NewMockAwsBatchClient(), "", "",
		utils.NewRateLimiter("", 10, 20),
		utils.NewRateLimiter("", 10, 20))

	t.Run("Not Found", func(t *testing.T) {
		dCache := definition.NewCache(10)

		nextState, err := EnsureJobDefinition(ctx, tCtx, cfg, batchClient, dCache, &State{
			State: &arrayCore.State{},
		}, 0)

		assert.NoError(t, err)
		assert.NotNil(t, nextState)
		assert.Equal(t, "my-arn", nextState.JobDefinitionArn)
		p, v := nextState.GetPhase()
		assert.Equal(t, arrayCore.PhaseLaunch, p)
		assert.Zero(t, v)
	})

	t.Run("Found", func(t *testing.T) {
		dCache := definition.NewCache(10)
		assert.NoError(t, dCache.Put(definition.NewCacheKey("new-role", "img1", defaultComputeEngine), "their-arn"))

		nextState, err := EnsureJobDefinition(ctx, tCtx, cfg, batchClient, dCache, &State{
			State: &arrayCore.State{},
		}, 0)
		assert.NoError(t, err)
		assert.NotNil(t, nextState)
		assert.Equal(t, "their-arn", nextState.JobDefinitionArn)
	})
}
