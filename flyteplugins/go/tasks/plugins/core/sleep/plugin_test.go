package sleep

import (
	"context"
	"testing"
	"time"

	core "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	coreMocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	ioMocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	idlcore "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestHandleSucceedsImmediatelyForZeroDuration(t *testing.T) {
	ctx := context.Background()
	plugin := &Plugin{taskStartTimes: make(map[string]time.Time)}
	tCtx := newTaskExecutionContext(time.Duration(0), true, "task-1")

	trns, err := plugin.Handle(ctx, tCtx)
	require.NoError(t, err)
	assert.Equal(t, core.PhaseSuccess, trns.Info().Phase())
	assert.Empty(t, plugin.taskStartTimes)
}

func TestHandleWaitsUntilDurationElapses(t *testing.T) {
	ctx := context.Background()
	plugin := &Plugin{taskStartTimes: make(map[string]time.Time)}
	tCtx := newTaskExecutionContext(30*time.Second, true, "task-1")

	first, err := plugin.Handle(ctx, tCtx)
	require.NoError(t, err)
	assert.Equal(t, core.PhaseRunning, first.Info().Phase())

	plugin.Lock()
	plugin.taskStartTimes["task-1"] = time.Now().Add(-31 * time.Second)
	plugin.Unlock()

	second, err := plugin.Handle(ctx, tCtx)
	require.NoError(t, err)
	assert.Equal(t, core.PhaseSuccess, second.Info().Phase())

	require.NoError(t, plugin.Finalize(ctx, tCtx))
	assert.Empty(t, plugin.taskStartTimes)
}

func TestHandleReturnsUserFailureForInvalidInput(t *testing.T) {
	ctx := context.Background()
	plugin := &Plugin{taskStartTimes: make(map[string]time.Time)}
	tCtx := newTaskExecutionContext(time.Second, false, "task-1")

	trns, err := plugin.Handle(ctx, tCtx)
	require.NoError(t, err)
	assert.Equal(t, core.PhasePermanentFailure, trns.Info().Phase())
	require.NotNil(t, trns.Info().Err())
	assert.Equal(t, idlcore.ExecutionError_USER, trns.Info().Err().GetKind())
	assert.Contains(t, trns.Info().Err().GetMessage(), "duration")
}

func newTaskExecutionContext(sleepDuration time.Duration, validInput bool, generatedName string) *coreMocks.TaskExecutionContext {
	taskTemplate := &idlcore.TaskTemplate{
		Type: sleepTaskType,
		Interface: &idlcore.TypedInterface{
			Inputs: &idlcore.VariableMap{
				Variables: []*idlcore.VariableEntry{
					{
						Key: "duration",
						Value: &idlcore.Variable{
							Type: literalType(validInput),
						},
					},
				},
			},
		},
	}

	inputs := &idlcore.LiteralMap{
		Literals: map[string]*idlcore.Literal{
			"duration": {
				Value: &idlcore.Literal_Scalar{
					Scalar: &idlcore.Scalar{
						Value: &idlcore.Scalar_Primitive{
							Primitive: &idlcore.Primitive{
								Value: &idlcore.Primitive_Duration{
									Duration: durationpb.New(sleepDuration),
								},
							},
						},
					},
				},
			},
		},
	}

	taskReader := &coreMocks.TaskReader{}
	taskReader.EXPECT().Read(mock.Anything).Return(taskTemplate, nil)

	inputReader := &ioMocks.InputReader{}
	inputReader.EXPECT().Get(mock.Anything).Return(inputs, nil)

	taskExecutionID := &coreMocks.TaskExecutionID{}
	taskExecutionID.EXPECT().GetGeneratedName().Return(generatedName)

	metadata := &coreMocks.TaskExecutionMetadata{}
	metadata.EXPECT().GetTaskExecutionID().Return(taskExecutionID)

	tCtx := &coreMocks.TaskExecutionContext{}
	tCtx.EXPECT().TaskReader().Return(taskReader)
	tCtx.EXPECT().InputReader().Return(inputReader)
	tCtx.EXPECT().TaskExecutionMetadata().Return(metadata)

	return tCtx
}

func literalType(validInput bool) *idlcore.LiteralType {
	simpleType := idlcore.SimpleType_DURATION
	if !validInput {
		simpleType = idlcore.SimpleType_STRING
	}

	return &idlcore.LiteralType{
		Type: &idlcore.LiteralType_Simple{
			Simple: simpleType,
		},
	}
}
