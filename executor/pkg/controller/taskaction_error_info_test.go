package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/gpufault"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

func TestToActionErrorInfo(t *testing.T) {
	fault := gpufault.ToProto(
		gpufault.Fault{
			Kind: gpufault.KindXid, Code: 79, Name: gpufault.NameFor(gpufault.KindXid, 79),
			Severity: gpufault.SeverityCritical, PCI: "0000:3b:00.0",
		},
		gpufault.Attribution{NodeName: "ip-10-0-0-1", GPUIndex: 0},
	)

	tests := []struct {
		name string
		err  *core.ExecutionError
		want *workflow.ErrorInfo
	}{
		{
			name: "no error",
			err:  nil,
			want: nil,
		},
		{
			name: "user error keeps its code",
			err:  &core.ExecutionError{Code: "OOMKilled", Message: "container was oom killed", Kind: core.ExecutionError_USER},
			want: &workflow.ErrorInfo{Code: "OOMKilled", Message: "container was oom killed", Kind: workflow.ErrorInfo_KIND_USER},
		},
		{
			name: "system error keeps its code",
			err:  &core.ExecutionError{Code: "NodeShutdown", Message: "node is shutting down", Kind: core.ExecutionError_SYSTEM},
			want: &workflow.ErrorInfo{Code: "NodeShutdown", Message: "node is shutting down", Kind: workflow.ErrorInfo_KIND_SYSTEM},
		},
		{
			name: "unknown kind and no code",
			err:  &core.ExecutionError{Message: "something went wrong"},
			want: &workflow.ErrorInfo{Message: "something went wrong", Kind: workflow.ErrorInfo_KIND_UNSPECIFIED},
		},
		{
			name: "gpu fault travels with the error",
			err: &core.ExecutionError{
				Code: gpufault.CodeGpuFallenOffBus, Message: "the gpu fell off the bus",
				Kind: core.ExecutionError_SYSTEM, GpuFault: fault,
			},
			want: &workflow.ErrorInfo{
				Code: gpufault.CodeGpuFallenOffBus, Message: "the gpu fell off the bus",
				Kind: workflow.ErrorInfo_KIND_SYSTEM, GpuFault: fault,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toActionErrorInfo(tt.err)
			if tt.want == nil {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.Equal(t, tt.want.GetCode(), got.GetCode())
			assert.Equal(t, tt.want.GetMessage(), got.GetMessage())
			assert.Equal(t, tt.want.GetKind(), got.GetKind())
			assert.Equal(t, tt.want.GetGpuFault(), got.GetGpuFault())
		})
	}
}

// The CR-persisted ErrorState is the other route a failure takes to the user, when the
// executor reports the action from the stored status rather than from the live phase.
func TestErrorStateFromExecError(t *testing.T) {
	tests := []struct {
		name        string
		err         *core.ExecutionError
		wantCode    string
		wantKind    string
		wantMessage string
		wantNil     bool
	}{
		{name: "no error", err: nil, wantNil: true},
		{
			name:        "user error",
			err:         &core.ExecutionError{Code: "OOMKilled", Message: "oom", Kind: core.ExecutionError_USER},
			wantCode:    "OOMKilled",
			wantKind:    "USER",
			wantMessage: "oom",
		},
		{
			name:        "gpu fault code survives the round trip through the CR",
			err:         &core.ExecutionError{Code: gpufault.CodeGpuFallenOffBus, Message: "gpu gone", Kind: core.ExecutionError_SYSTEM},
			wantCode:    gpufault.CodeGpuFallenOffBus,
			wantKind:    "SYSTEM",
			wantMessage: "gpu gone",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := errorStateFromExecError(tt.err)
			if tt.wantNil {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.Equal(t, tt.wantCode, got.Code)
			assert.Equal(t, tt.wantKind, got.Kind)
			assert.Equal(t, tt.wantMessage, got.Message)
		})
	}
}
