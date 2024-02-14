package v1alpha1

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func TestNodeKindString(t *testing.T) {
	tests := []struct {
		kind     NodeKind
		expected string
	}{
		{NodeKindTask, "task"},
		{NodeKindBranch, "branch"},
		{NodeKindWorkflow, "workflow"},
		{NodeKindGate, "gate"},
		{NodeKindArray, "array"},
		{NodeKindStart, "start"},
		{NodeKindEnd, "end"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			assert.Equal(t, test.expected, test.kind.String())
		})
	}
}

func TestNodePhaseString(t *testing.T) {
	tests := []struct {
		phase    NodePhase
		expected string
	}{
		{NodePhaseNotYetStarted, "NotYetStarted"},
		{NodePhaseQueued, "Queued"},
		{NodePhaseRunning, "Running"},
		{NodePhaseTimingOut, "NodePhaseTimingOut"},
		{NodePhaseTimedOut, "NodePhaseTimedOut"},
		{NodePhaseSucceeding, "Succeeding"},
		{NodePhaseSucceeded, "Succeeded"},
		{NodePhaseFailing, "Failing"},
		{NodePhaseFailed, "Failed"},
		{NodePhaseSkipped, "Skipped"},
		{NodePhaseRetryableFailure, "RetryableFailure"},
		{NodePhaseDynamicRunning, "DynamicRunning"},
		{NodePhaseRecovered, "NodePhaseRecovered"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			assert.Equal(t, test.expected, test.phase.String())
		})
	}
}

func TestWorkflowPhaseString(t *testing.T) {
	tests := []struct {
		phase    WorkflowPhase
		expected string
	}{
		{WorkflowPhaseReady, "Ready"},
		{WorkflowPhaseRunning, "Running"},
		{WorkflowPhaseSuccess, "Succeeded"},
		{WorkflowPhaseFailed, "Failed"},
		{WorkflowPhaseFailing, "Failing"},
		{WorkflowPhaseSucceeding, "Succeeding"},
		{WorkflowPhaseAborted, "Aborted"},
		{WorkflowPhaseHandlingFailureNode, "HandlingFailureNode"},
		{-1, "Unknown"},
		// Add more cases as needed
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			assert.Equal(t, test.expected, test.phase.String())
		})
	}
}

func TestBranchNodePhaseString(t *testing.T) {
	tests := []struct {
		phase    BranchNodePhase
		expected string
	}{
		{BranchNodeNotYetEvaluated, "NotYetEvaluated"},
		{BranchNodeSuccess, "BranchEvalSuccess"},
		{BranchNodeError, "BranchEvalFailed"},
		{-1, "Undefined"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			assert.Equal(t, test.expected, test.phase.String())
		})
	}
}

func TestWorkflowOnFailurePolicyStringError(t *testing.T) {
	_, err := WorkflowOnFailurePolicyString("NON_EXISTENT_POLICY")
	assert.Error(t, err)
}

func TestWorkflowOnFailurePolicyJSONMarshalling(t *testing.T) {
	tests := []struct {
		policy  WorkflowOnFailurePolicy
		jsonStr string
	}{
		{WorkflowOnFailurePolicy(core.WorkflowMetadata_FAIL_IMMEDIATELY), `"FAIL_IMMEDIATELY"`},
		{WorkflowOnFailurePolicy(core.WorkflowMetadata_FAIL_AFTER_EXECUTABLE_NODES_COMPLETE), `"FAIL_AFTER_EXECUTABLE_NODES_COMPLETE"`},
	}

	for _, test := range tests {
		t.Run(test.jsonStr, func(t *testing.T) {
			// Testing marshalling
			data, err := json.Marshal(test.policy)
			require.NoError(t, err)
			assert.Equal(t, test.jsonStr, string(data))

			// Testing unmarshalling
			var unmarshalledPolicy WorkflowOnFailurePolicy
			err = json.Unmarshal(data, &unmarshalledPolicy)
			require.NoError(t, err)
			assert.Equal(t, test.policy, unmarshalledPolicy)
		})
	}

	invalidTest := `123`
	var policy WorkflowOnFailurePolicy
	err := json.Unmarshal([]byte(invalidTest), &policy)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "WorkflowOnFailurePolicy should be a string, got 123")

}

func TestGetOutputsFile(t *testing.T) {
	tests := []struct {
		outputDir DataReference
		expected  DataReference
	}{
		{"dir1", "dir1/outputs.pb"},
		{"dir2", "dir2/outputs.pb"},
	}

	for _, tt := range tests {
		t.Run(string(tt.outputDir), func(t *testing.T) { // Convert DataReference to string here
			assert.Equal(t, tt.expected, GetOutputsFile(tt.outputDir))
		})
	}
}

func TestGetInputsFile(t *testing.T) {
	tests := []struct {
		inputDir DataReference
		expected DataReference
	}{
		{"dir1", "dir1/inputs_old.pb"},
		{"dir2", "dir2/inputs_old.pb"},
	}

	for _, tt := range tests {
		t.Run(string(tt.inputDir), func(t *testing.T) {
			assert.Equal(t, tt.expected, GetInputsFile(tt.inputDir))
		})
	}
}

func TestGetDeckFile(t *testing.T) {
	tests := []struct {
		inputDir DataReference
		expected DataReference
	}{
		{"dir1", "dir1/deck.html"},
		{"dir2", "dir2/deck.html"},
	}

	for _, tt := range tests {
		t.Run(string(tt.inputDir), func(t *testing.T) {
			assert.Equal(t, tt.expected, GetDeckFile(tt.inputDir))
		})
	}
}
