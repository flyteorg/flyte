package array

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	idlcore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

func TestAppendLiteral(t *testing.T) {
	outputLiterals := make(map[string]*idlcore.Literal)
	literalMaps := []map[string]*idlcore.Literal{
		map[string]*idlcore.Literal{
			"foo": nilLiteral,
			"bar": nilLiteral,
		},
		map[string]*idlcore.Literal{
			"foo": nilLiteral,
			"bar": nilLiteral,
		},
	}

	for _, m := range literalMaps {
		for k, v := range m {
			appendLiteral(k, v, outputLiterals, len(literalMaps))
		}
	}

	for _, v := range outputLiterals {
		collection, ok := v.Value.(*idlcore.Literal_Collection)
		assert.True(t, ok)

		assert.Equal(t, 2, len(collection.Collection.Literals))
	}
}

func TestInferParallelism(t *testing.T) {
	ctx := context.TODO()
	zero := uint32(0)
	one := uint32(1)

	tests := []struct {
		name                   string
		parallelism            *uint32
		parallelismBehavior    string
		remainingParallelism   int
		arrayNodeSize          int
		expectedIncrement      bool
		expectedMaxParallelism int
	}{
		{
			name:                   "NilParallelismWorkflowBehavior",
			parallelism:            nil,
			parallelismBehavior:    "workflow",
			remainingParallelism:   2,
			arrayNodeSize:          3,
			expectedIncrement:      true,
			expectedMaxParallelism: 2,
		},
		{
			name:                   "NilParallelismHybridBehavior",
			parallelism:            nil,
			parallelismBehavior:    "hybrid",
			remainingParallelism:   2,
			arrayNodeSize:          3,
			expectedIncrement:      true,
			expectedMaxParallelism: 2,
		},
		{
			name:                   "NilParallelismUnlimitedBehavior",
			parallelism:            nil,
			parallelismBehavior:    "unlimited",
			remainingParallelism:   2,
			arrayNodeSize:          3,
			expectedIncrement:      false,
			expectedMaxParallelism: 3,
		},
		{
			name:                   "ZeroParallelismWorkflowBehavior",
			parallelism:            &zero,
			parallelismBehavior:    "workflow",
			remainingParallelism:   2,
			arrayNodeSize:          3,
			expectedIncrement:      true,
			expectedMaxParallelism: 2,
		},
		{
			name:                   "ZeroParallelismHybridBehavior",
			parallelism:            &zero,
			parallelismBehavior:    "hybrid",
			remainingParallelism:   2,
			arrayNodeSize:          3,
			expectedIncrement:      false,
			expectedMaxParallelism: 3,
		},
		{
			name:                   "ZeroParallelismUnlimitedBehavior",
			parallelism:            &zero,
			parallelismBehavior:    "unlimited",
			remainingParallelism:   2,
			arrayNodeSize:          3,
			expectedIncrement:      false,
			expectedMaxParallelism: 3,
		},
		{
			name:                   "OneParallelismWorkflowBehavior",
			parallelism:            &one,
			parallelismBehavior:    "workflow",
			remainingParallelism:   2,
			arrayNodeSize:          3,
			expectedIncrement:      false,
			expectedMaxParallelism: 1,
		},
		{
			name:                   "OneParallelismHybridBehavior",
			parallelism:            &one,
			parallelismBehavior:    "hybrid",
			remainingParallelism:   2,
			arrayNodeSize:          3,
			expectedIncrement:      false,
			expectedMaxParallelism: 1,
		},
		{
			name:                   "OneParallelismUnlimitedBehavior",
			parallelism:            &one,
			parallelismBehavior:    "unlimited",
			remainingParallelism:   2,
			arrayNodeSize:          3,
			expectedIncrement:      false,
			expectedMaxParallelism: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			increment, maxParallelism := inferParallelism(ctx, tt.parallelism, tt.parallelismBehavior, tt.remainingParallelism, tt.arrayNodeSize)
			assert.Equal(t, tt.expectedIncrement, increment)
			assert.Equal(t, tt.expectedMaxParallelism, maxParallelism)
		})
	}
}

func TestShouldIncrementTaskPhaseVersion(t *testing.T) {

	tests := []struct {
		name                      string
		subNodeKind               v1alpha1.NodeKind
		subNodeStatus             *v1alpha1.NodeStatus
		previousNodePhase         v1alpha1.NodePhase
		previousTaskPhase         int
		incrementTaskPhaseVersion bool
	}{
		{
			name: "DifferentNodePhase",
			subNodeStatus: &v1alpha1.NodeStatus{
				Phase: v1alpha1.NodePhaseSucceeded,
			},
			previousNodePhase:         v1alpha1.NodePhaseRunning,
			previousTaskPhase:         0,
			incrementTaskPhaseVersion: true,
		},
		{
			name: "DifferentTaskNodePhase",
			subNodeStatus: &v1alpha1.NodeStatus{
				Phase: v1alpha1.NodePhaseRunning,
				TaskNodeStatus: &v1alpha1.TaskNodeStatus{
					Phase: 1,
				},
			},
			previousNodePhase:         v1alpha1.NodePhaseRunning,
			previousTaskPhase:         0,
			incrementTaskPhaseVersion: true,
		},
		{
			name: "SameTaskNodePhase",
			subNodeStatus: &v1alpha1.NodeStatus{
				Phase: v1alpha1.NodePhaseRunning,
				TaskNodeStatus: &v1alpha1.TaskNodeStatus{
					Phase: 0,
				},
			},
			previousNodePhase:         v1alpha1.NodePhaseRunning,
			previousTaskPhase:         0,
			incrementTaskPhaseVersion: false,
		},
		{
			name: "DifferentWorkflowNodePhase",
			subNodeStatus: &v1alpha1.NodeStatus{
				Phase: v1alpha1.NodePhaseRunning,
				WorkflowNodeStatus: &v1alpha1.WorkflowNodeStatus{
					Phase: 1,
				},
			},
			previousNodePhase:         v1alpha1.NodePhaseRunning,
			previousTaskPhase:         0,
			incrementTaskPhaseVersion: true,
		},
		{
			name: "SameWorkflowNodePhase",
			subNodeStatus: &v1alpha1.NodeStatus{
				Phase: v1alpha1.NodePhaseRunning,
				WorkflowNodeStatus: &v1alpha1.WorkflowNodeStatus{
					Phase: 1,
				},
			},
			previousNodePhase:         v1alpha1.NodePhaseRunning,
			previousTaskPhase:         1,
			incrementTaskPhaseVersion: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			increment := shouldIncrementTaskPhaseVersion(tt.subNodeStatus, tt.previousNodePhase, tt.previousTaskPhase)
			assert.Equal(t, tt.incrementTaskPhaseVersion, increment)
		})
	}
}
