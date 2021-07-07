package nodes

import (
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

func TestIsTerminalNodePhase(t *testing.T) {
	tests := []struct {
		name string
		p    core.NodeExecution_Phase
		want bool
	}{
		{"aborted", core.NodeExecution_ABORTED, true},
		{"succeeded", core.NodeExecution_SUCCEEDED, true},
		{"failed", core.NodeExecution_FAILED, true},
		{"timed_out", core.NodeExecution_TIMED_OUT, true},
		{"skipped", core.NodeExecution_SKIPPED, true},
		{"running", core.NodeExecution_RUNNING, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTerminalNodePhase(tt.p); got != tt.want {
				t.Errorf("IsTerminalNodePhase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsTerminalTaskPhase(t *testing.T) {
	tests := []struct {
		name string
		p    core.TaskExecution_Phase
		want bool
	}{
		{"aborted", core.TaskExecution_ABORTED, true},
		{"failed", core.TaskExecution_FAILED, true},
		{"succeeded", core.TaskExecution_SUCCEEDED, true},
		{"running", core.TaskExecution_RUNNING, false},
		{"waiting", core.TaskExecution_WAITING_FOR_RESOURCES, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTerminalTaskPhase(tt.p); got != tt.want {
				t.Errorf("IsTerminalTaskPhase() = %v, want %v", got, tt.want)
			}
		})
	}
}
