package transformers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	flyteWorkflow "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

func TestToTaskSpec(t *testing.T) {
	ctx := context.Background()
	spec := &flyteWorkflow.TaskSpec{
		ShortName: "test-task",
		Environment: &flyteWorkflow.Environment{
			Name:        "prod",
			Description: "Production environment",
		},
	}

	specModel, err := models.NewTaskSpecModel(ctx, spec)
	require.NoError(t, err)

	result, err := ToTaskSpec(specModel)
	require.NoError(t, err)
	assert.Equal(t, "test-task", result.ShortName)
	assert.Equal(t, "prod", result.Environment.Name)
}

func TestToTaskSpec_InvalidData(t *testing.T) {
	specModel := &models.TaskSpec{
		Digest: "invalid",
		Spec:   []byte("invalid proto data"),
	}

	_, err := ToTaskSpec(specModel)
	assert.Error(t, err)
}

func TestToTraceSpec(t *testing.T) {
	ctx := context.Background()
	traceSpec := &flyteWorkflow.TraceSpec{
		Interface: &core.TypedInterface{},
	}

	specModel, err := models.NewTaskSpecModelFromTraceSpec(ctx, traceSpec)
	require.NoError(t, err)

	result, err := ToTraceSpec(specModel)
	require.NoError(t, err)
	assert.NotNil(t, result.Interface)
}

func TestToTraceSpec_InvalidData(t *testing.T) {
	specModel := &models.TaskSpec{
		Digest: "invalid",
		Spec:   []byte("invalid proto data"),
	}

	_, err := ToTraceSpec(specModel)
	assert.Error(t, err)
}

func TestToTaskSpec_RoundTrip(t *testing.T) {
	ctx := context.Background()
	original := &flyteWorkflow.TaskSpec{
		ShortName: "my-task",
		Environment: &flyteWorkflow.Environment{
			Name:        "prod",
			Description: "Production environment",
		},
		Documentation: &flyteWorkflow.DocumentationEntity{
			ShortDescription: "Test task",
		},
	}

	specModel, err := models.NewTaskSpecModel(ctx, original)
	require.NoError(t, err)

	result, err := ToTaskSpec(specModel)
	require.NoError(t, err)
	assert.True(t, proto.Equal(original, result))
}

func TestToTraceSpec_RoundTrip(t *testing.T) {
	ctx := context.Background()
	original := &flyteWorkflow.TraceSpec{
		Interface: &core.TypedInterface{},
	}

	specModel, err := models.NewTaskSpecModelFromTraceSpec(ctx, original)
	require.NoError(t, err)

	result, err := ToTraceSpec(specModel)
	require.NoError(t, err)
	assert.True(t, proto.Equal(original, result))
}
