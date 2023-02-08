package core

import (
	"context"
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins"
	"github.com/golang/protobuf/proto"

	"github.com/flyteorg/flytestdlib/bitarray"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/arraystatus"
	"github.com/stretchr/testify/assert"
)

func TestInvertBitSet(t *testing.T) {
	input := bitarray.NewBitSet(4)
	input.Set(0)
	input.Set(2)
	input.Set(3)

	expected := bitarray.NewBitSet(4)
	expected.Set(1)
	expected.Set(4)

	actual := InvertBitSet(input, 4)
	assertBitSetsEqual(t, expected, actual, 4)
}

func assertBitSetsEqual(t testing.TB, b1, b2 *bitarray.BitSet, len int) {
	if b1 == nil {
		assert.Nil(t, b2)
	} else if b2 == nil {
		assert.FailNow(t, "b2 must not be nil")
	}

	assert.Equal(t, b1.Cap(), b2.Cap())
	for i := uint(0); i < uint(len); i++ {
		assert.Equal(t, b1.IsSet(i), b2.IsSet(i), "At index %v", i)
	}
}

func TestMapArrayStateToPluginPhase(t *testing.T) {
	ctx := context.Background()

	subTaskCount := 3

	detailedArray := NewPhasesCompactArray(uint(subTaskCount))
	indexesToCache := InvertBitSet(bitarray.NewBitSet(uint(subTaskCount)), uint(subTaskCount))
	retryAttemptsArray, err := bitarray.NewCompactArray(uint(subTaskCount), bitarray.Item(1))
	assert.NoError(t, err)

	t.Run("start", func(t *testing.T) {
		s := State{
			CurrentPhase: PhaseStart,
			ArrayStatus: arraystatus.ArrayStatus{
				Detailed: detailedArray,
			},
			IndexesToCache: indexesToCache,
			RetryAttempts:  retryAttemptsArray,
		}
		phaseInfo, err := MapArrayStateToPluginPhase(ctx, &s, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, core.PhaseInitializing, phaseInfo.Phase())
	})

	t.Run("launch", func(t *testing.T) {
		s := State{
			CurrentPhase: PhaseLaunch,
			PhaseVersion: 0,
			ArrayStatus: arraystatus.ArrayStatus{
				Detailed: detailedArray,
			},
			IndexesToCache: indexesToCache,
			RetryAttempts:  retryAttemptsArray,
		}

		phaseInfo, err := MapArrayStateToPluginPhase(ctx, &s, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, core.PhaseRunning, phaseInfo.Phase())
	})

	t.Run("monitoring subtasks", func(t *testing.T) {
		s := State{
			CurrentPhase:       PhaseCheckingSubTaskExecutions,
			PhaseVersion:       8,
			OriginalArraySize:  10,
			ExecutionArraySize: 5,
			ArrayStatus: arraystatus.ArrayStatus{
				Detailed: detailedArray,
			},
			IndexesToCache: indexesToCache,
			RetryAttempts:  retryAttemptsArray,
		}

		phaseInfo, err := MapArrayStateToPluginPhase(ctx, &s, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, core.PhaseRunning, phaseInfo.Phase())
		assert.Equal(t, uint32(8), phaseInfo.Version())
	})

	t.Run("write to discovery", func(t *testing.T) {
		s := State{
			CurrentPhase:       PhaseWriteToDiscovery,
			PhaseVersion:       8,
			OriginalArraySize:  10,
			ExecutionArraySize: 5,
			ArrayStatus: arraystatus.ArrayStatus{
				Detailed: detailedArray,
			},
			IndexesToCache: indexesToCache,
			RetryAttempts:  retryAttemptsArray,
		}

		phaseInfo, err := MapArrayStateToPluginPhase(ctx, &s, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, core.PhaseRunning, phaseInfo.Phase())
		assert.Equal(t, uint32(8), phaseInfo.Version())
	})

	t.Run("success", func(t *testing.T) {
		s := State{
			CurrentPhase: PhaseSuccess,
			PhaseVersion: 0,
			ArrayStatus: arraystatus.ArrayStatus{
				Detailed: detailedArray,
			},
			IndexesToCache: indexesToCache,
			RetryAttempts:  retryAttemptsArray,
		}

		phaseInfo, err := MapArrayStateToPluginPhase(ctx, &s, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, core.PhaseSuccess, phaseInfo.Phase())
	})

	t.Run("retryable failure", func(t *testing.T) {
		s := State{
			CurrentPhase: PhaseRetryableFailure,
			PhaseVersion: 0,
			ArrayStatus: arraystatus.ArrayStatus{
				Detailed: detailedArray,
			},
			IndexesToCache: indexesToCache,
			RetryAttempts:  retryAttemptsArray,
		}

		phaseInfo, err := MapArrayStateToPluginPhase(ctx, &s, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, core.PhaseRetryableFailure, phaseInfo.Phase())
	})

	t.Run("permanent failure", func(t *testing.T) {
		s := State{
			CurrentPhase: PhasePermanentFailure,
			PhaseVersion: 0,
			ArrayStatus: arraystatus.ArrayStatus{
				Detailed: detailedArray,
			},
			IndexesToCache: indexesToCache,
			RetryAttempts:  retryAttemptsArray,
		}

		phaseInfo, err := MapArrayStateToPluginPhase(ctx, &s, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, core.PhasePermanentFailure, phaseInfo.Phase())
	})

	t.Run("All phases", func(t *testing.T) {
		for _, p := range PhaseValues() {
			s := State{
				CurrentPhase: p,
				ArrayStatus: arraystatus.ArrayStatus{
					Detailed: detailedArray,
				},
				IndexesToCache: indexesToCache,
				RetryAttempts:  retryAttemptsArray,
			}

			phaseInfo, err := MapArrayStateToPluginPhase(ctx, &s, nil, nil)
			assert.NoError(t, err)
			assert.NotEqual(t, core.PhaseUndefined, phaseInfo.Phase())
		}
	})
}

func Test_calculateOriginalIndex(t *testing.T) {
	t.Run("BitSet is set", func(t *testing.T) {
		inputArr := bitarray.NewBitSet(7)
		inputArr.Set(3)
		inputArr.Set(4)
		inputArr.Set(5)

		tests := []struct {
			name     string
			childIdx int
			want     int
		}{
			{"0", 0, 3},
			{"1", 1, 4},
			{"2", 2, 5},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := CalculateOriginalIndex(tt.childIdx, inputArr); got != tt.want {
					t.Errorf("calculateOriginalIndex() = %v, want %v", got, tt.want)
				}
			})
		}
	})

	t.Run("Bitset is half set", func(t *testing.T) {
		inputArr := bitarray.NewBitSet(3)
		inputArr.Set(1)
		inputArr.Set(2)

		tests := []struct {
			name     string
			childIdx int
			want     int
		}{
			{"0", 0, 1},
			{"1", 1, 2},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := CalculateOriginalIndex(tt.childIdx, inputArr); got != tt.want {
					t.Errorf("calculateOriginalIndex() = %v, want %v", got, tt.want)
				}
			})
		}
	})

	t.Run("Bitset is all set", func(t *testing.T) {
		inputArr := bitarray.NewBitSet(3)
		inputArr.Set(0)
		inputArr.Set(1)
		inputArr.Set(2)

		tests := []struct {
			name     string
			childIdx int
			want     int
		}{
			{"0", 0, 0},
			{"1", 1, 1},
			{"2", 2, 2},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := CalculateOriginalIndex(tt.childIdx, inputArr); got != tt.want {
					t.Errorf("calculateOriginalIndex() = %v, want %v", got, tt.want)
				}
			})
		}
	})
}

func TestToArrayJob(t *testing.T) {
	t.Run("task_type_version == 0", func(t *testing.T) {
		arrayJob, err := ToArrayJob(nil, 0)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(arrayJob, &plugins.ArrayJob{
			Parallelism: 0,
			Size:        1,
			SuccessCriteria: &plugins.ArrayJob_MinSuccesses{
				MinSuccesses: 1,
			},
		}))
	})

	t.Run("task_type_version == 1", func(t *testing.T) {
		arrayJob, err := ToArrayJob(nil, 1)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(arrayJob, &plugins.ArrayJob{
			Parallelism: 0,
			Size:        1,
			SuccessCriteria: &plugins.ArrayJob_MinSuccessRatio{
				MinSuccessRatio: 1.0,
			},
		}))
	})
}

func TestSummaryToPhase(t *testing.T) {
	minSuccesses := int64(10)
	tests := []struct {
		name    string
		phase   Phase
		summary map[core.Phase]int64
	}{
		{
			"FailOnTooFewTasks",
			PhaseWriteToDiscoveryThenFail,
			map[core.Phase]int64{},
		},
		{
			"ContinueOnRetryableFailures",
			PhaseCheckingSubTaskExecutions,
			map[core.Phase]int64{
				core.PhaseRetryableFailure: 1,
				core.PhaseUndefined:        9,
			},
		},
		{
			"FailOnTooManyPermanentFailures",
			PhaseWriteToDiscoveryThenFail,
			map[core.Phase]int64{
				core.PhasePermanentFailure: 1,
				core.PhaseSuccess:          9,
			},
		},
		{
			"CheckWaitingForResources",
			PhaseWaitingForResources,
			map[core.Phase]int64{
				core.PhaseWaitingForResources: 1,
				core.PhaseUndefined:           9,
			},
		},
		{
			"WaitForAllSubtasksToComplete",
			PhaseCheckingSubTaskExecutions,
			map[core.Phase]int64{
				core.PhaseUndefined: 1,
				core.PhaseSuccess:   9,
			},
		},
		{
			"SuccessfullyCompleted",
			PhaseWriteToDiscovery,
			map[core.Phase]int64{
				core.PhaseSuccess: 10,
			},
		},
		{
			"FailedToRetry",
			PhaseWriteToDiscoveryThenFail,
			map[core.Phase]int64{
				core.PhaseSuccess:          5,
				core.PhasePermanentFailure: 5,
			},
		},
		{
			"Retrying",
			PhaseCheckingSubTaskExecutions,
			map[core.Phase]int64{
				core.PhaseSuccess:          5,
				core.PhaseRetryableFailure: 5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.phase, SummaryToPhase(context.TODO(), minSuccesses, tt.summary))
		})
	}
}
