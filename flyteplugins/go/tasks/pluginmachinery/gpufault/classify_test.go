package gpufault

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

func gpuFault(code int, severity Severity) *core.GpuFault {
	return ToProto(
		Fault{Kind: KindXid, Code: code, Name: NameFor(KindXid, code), Severity: severity, PCI: "0000:3b:00.0"},
		Attribution{NodeName: "ip-10-0-0-1", GPUUUID: testUUID, GPUIndex: 0},
	)
}

func sxidFault(code int) *core.GpuFault {
	return ToProto(
		Fault{Kind: KindSXid, Code: code, Name: NameFor(KindSXid, code), Severity: SeverityCritical, PCI: "0000:05:00.0"},
		Attribution{NodeName: "ip-10-0-0-1", GPUIndex: UnknownGPUIndex},
	)
}

func TestCodeFor(t *testing.T) {
	tests := []struct {
		name  string
		fault Fault
		want  string
	}{
		{name: "fallen off the bus", fault: Fault{Kind: KindXid, Code: 79}, want: CodeGpuFallenOffBus},
		{name: "double bit ecc", fault: Fault{Kind: KindXid, Code: 48}, want: CodeGpuEccUncorrectable},
		{name: "contained ecc", fault: Fault{Kind: KindXid, Code: 94}, want: CodeGpuEccUncorrectable},
		{name: "uncontained ecc", fault: Fault{Kind: KindXid, Code: 95}, want: CodeGpuEccUncorrectable},
		{name: "unrecovered ecc", fault: Fault{Kind: KindXid, Code: 140}, want: CodeGpuEccUncorrectable},
		{name: "row remap recorded", fault: Fault{Kind: KindXid, Code: 63}, want: CodeGpuRowRemapPending},
		{name: "row remap failed", fault: Fault{Kind: KindXid, Code: 64}, want: CodeGpuRowRemapPending},
		{name: "nvlink", fault: Fault{Kind: KindXid, Code: 74}, want: CodeGpuNvlinkError},
		{name: "gsp rpc timeout", fault: Fault{Kind: KindXid, Code: 119}, want: CodeGpuGspError},
		{name: "gsp error", fault: Fault{Kind: KindXid, Code: 120}, want: CodeGpuGspError},
		{name: "workload fault", fault: Fault{Kind: KindXid, Code: 31}, want: CodeGpuXidError},
		{name: "unknown code", fault: Fault{Kind: KindXid, Code: 4242}, want: CodeGpuXidError},
		{name: "every sxid is an nvlink error", fault: Fault{Kind: KindSXid, Code: 12028}, want: CodeGpuNvlinkError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, CodeFor(tt.fault))
		})
	}
}

func TestIsGenericCode(t *testing.T) {
	tests := []struct {
		code string
		want bool
	}{
		{code: "", want: true},
		{code: "Unknown", want: true},
		{code: "UnknownError", want: true},
		{code: "Error", want: true},
		{code: "1", want: true},
		{code: "137", want: true},
		{code: "ExitCode1", want: true},
		{code: "exit-code-137", want: true},
		{code: "OOMKilled", want: false},
		{code: "Interrupted", want: false},
		{code: "PrimaryContainerNotFound", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			assert.Equal(t, tt.want, isGenericCode(tt.code))
		})
	}
}

func TestClassifyFailureLeavesNonFailuresAlone(t *testing.T) {
	tests := []struct {
		name  string
		phase pluginsCore.PhaseInfo
	}{
		{name: "success", phase: pluginsCore.PhaseInfoSuccess(nil)},
		{name: "running", phase: pluginsCore.PhaseInfoRunning(1, nil)},
		{name: "aborted", phase: pluginsCore.PhaseInfoAborted(time.Now(), 1, "user aborted")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyFailure(tt.phase, []*core.GpuFault{gpuFault(79, SeverityCritical)})
			assert.Equal(t, tt.phase.Phase(), got.Phase())
			assert.Nil(t, got.Err())
		})
	}
}

func TestClassifyFailureWithoutFaults(t *testing.T) {
	phase := pluginsCore.PhaseInfoRetryableFailure("OOMKilled", "Pod failed", nil)

	got := ClassifyFailure(phase, nil)

	assert.Equal(t, "OOMKilled", got.Err().GetCode())
	assert.Equal(t, "Pod failed", got.Err().GetMessage())
	assert.Nil(t, got.Err().GetGpuFault())
}

func TestClassifyFailureCritical(t *testing.T) {
	tests := []struct {
		name          string
		phase         pluginsCore.PhaseInfo
		faults        []*core.GpuFault
		wantPhase     pluginsCore.Phase
		wantCode      string
		wantFaultCode uint32
	}{
		{
			name:          "user retryable failure becomes a system one",
			phase:         pluginsCore.PhaseInfoRetryableFailure("UnknownError", "Pod failed", nil),
			faults:        []*core.GpuFault{gpuFault(79, SeverityCritical)},
			wantCode:      CodeGpuFallenOffBus,
			wantFaultCode: 79,
		},
		{
			name:          "permanent failure stays permanent but becomes the system's",
			phase:         pluginsCore.PhaseInfoFailure("Error", "Pod failed", nil),
			faults:        []*core.GpuFault{gpuFault(48, SeverityCritical)},
			wantPhase:     pluginsCore.PhasePermanentFailure,
			wantCode:      CodeGpuEccUncorrectable,
			wantFaultCode: 48,
		},
		{
			name:          "an nvswitch fault is critical too",
			phase:         pluginsCore.PhaseInfoRetryableFailure("Error", "Pod failed", nil),
			faults:        []*core.GpuFault{sxidFault(12028)},
			wantCode:      CodeGpuNvlinkError,
			wantFaultCode: 12028,
		},
		{
			name:          "a critical fault outranks an earlier user one",
			phase:         pluginsCore.PhaseInfoRetryableFailure("Error", "Pod failed", nil),
			faults:        []*core.GpuFault{gpuFault(31, SeverityUser), gpuFault(63, SeverityCritical)},
			wantCode:      CodeGpuRowRemapPending,
			wantFaultCode: 63,
		},
		{
			name:          "the first critical fault wins",
			phase:         pluginsCore.PhaseInfoRetryableFailure("Error", "Pod failed", nil),
			faults:        []*core.GpuFault{gpuFault(74, SeverityCritical), gpuFault(79, SeverityCritical)},
			wantCode:      CodeGpuNvlinkError,
			wantFaultCode: 74,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyFailure(tt.phase, tt.faults)

			wantPhase := tt.wantPhase
			if wantPhase == pluginsCore.PhaseUndefined {
				wantPhase = pluginsCore.PhaseRetryableFailure
			}
			assert.Equal(t, wantPhase, got.Phase())
			require.NotNil(t, got.Err())
			assert.Equal(t, tt.wantCode, got.Err().GetCode())
			assert.Equal(t, core.ExecutionError_SYSTEM, got.Err().GetKind())
			assert.Equal(t, tt.wantFaultCode, got.Err().GetGpuFault().GetCode())
			assert.Contains(t, got.Err().GetMessage(), MessagePrefix)
			assert.True(t, len(got.Err().GetMessage()) > len("Pod failed"))
		})
	}
}

func TestClassifyFailureCriticalMessageAndCarriedFields(t *testing.T) {
	timestamp := timestamppb.New(time.Unix(1700000000, 0))
	phase := pluginsCore.PhaseInfoRetryableFailure("UnknownError", "Pod failed. No message received from kubernetes.", nil)
	phase.Err().ErrorUri = "s3://bucket/error.pb"
	phase.Err().Timestamp = timestamp
	phase.Err().Worker = "worker-1"
	phase.Err().Recoverability = core.ContainerError_RECOVERABLE
	phase = phase.WithVersion(4)
	phase.WithReason("pod event")

	fault := gpuFault(79, SeverityCritical)
	got := ClassifyFailure(phase, []*core.GpuFault{fault})

	assert.Equal(t,
		"[gpu-health] [CRITICAL] Xid 79 (GPU has fallen off the bus) on GPU 0 "+testUUID+
			". Pod failed. No message received from kubernetes.",
		got.Err().GetMessage())
	assert.Equal(t, "s3://bucket/error.pb", got.Err().GetErrorUri())
	assert.Equal(t, timestamp, got.Err().GetTimestamp())
	assert.Equal(t, "worker-1", got.Err().GetWorker())
	assert.Equal(t, core.ContainerError_RECOVERABLE, got.Err().GetRecoverability())
	assert.Equal(t, uint32(4), got.Version())
	assert.Equal(t, "pod event", got.Reason())
	assert.Equal(t, fault, got.Err().GetGpuFault())
	assert.NotNil(t, got.Info())
}

func TestClassifyFailureUser(t *testing.T) {
	tests := []struct {
		name        string
		phase       pluginsCore.PhaseInfo
		wantPhase   pluginsCore.Phase
		wantCode    string
		wantErrKind core.ExecutionError_ErrorKind
	}{
		{
			name:        "a generic code is replaced by the fault",
			phase:       pluginsCore.PhaseInfoRetryableFailure("UnknownError", "exit code 1", nil),
			wantPhase:   pluginsCore.PhaseRetryableFailure,
			wantCode:    CodeGpuXidError,
			wantErrKind: core.ExecutionError_USER,
		},
		{
			name:        "an exit status reported as a code is replaced too",
			phase:       pluginsCore.PhaseInfoRetryableFailure("137", "exit code 137", nil),
			wantPhase:   pluginsCore.PhaseRetryableFailure,
			wantCode:    CodeGpuXidError,
			wantErrKind: core.ExecutionError_USER,
		},
		{
			name:        "a specific code is kept",
			phase:       pluginsCore.PhaseInfoRetryableFailure("OOMKilled", "container was oom killed", nil),
			wantPhase:   pluginsCore.PhaseRetryableFailure,
			wantCode:    "OOMKilled",
			wantErrKind: core.ExecutionError_USER,
		},
		{
			name:        "a permanent failure is never downgraded to retryable",
			phase:       pluginsCore.PhaseInfoFailure("UnknownError", "exit code 1", nil),
			wantPhase:   pluginsCore.PhasePermanentFailure,
			wantCode:    CodeGpuXidError,
			wantErrKind: core.ExecutionError_USER,
		},
		{
			name:        "a system failure keeps its kind",
			phase:       pluginsCore.PhaseInfoSystemRetryableFailure("Interrupted", "node shut down", nil),
			wantPhase:   pluginsCore.PhaseRetryableFailure,
			wantCode:    "Interrupted",
			wantErrKind: core.ExecutionError_SYSTEM,
		},
	}

	fault := gpuFault(31, SeverityUser)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalMessage := tt.phase.Err().GetMessage()

			got := ClassifyFailure(tt.phase, []*core.GpuFault{gpuFault(92, SeverityWarn), fault})

			assert.Equal(t, tt.wantPhase, got.Phase())
			assert.Equal(t, tt.wantCode, got.Err().GetCode())
			assert.Equal(t, tt.wantErrKind, got.Err().GetKind())
			assert.Equal(t, fault, got.Err().GetGpuFault())
			assert.Equal(t,
				"[gpu-health] [USER] Xid 31 (GPU memory page fault) on GPU 0 "+testUUID+". "+originalMessage,
				got.Err().GetMessage())
		})
	}
}

func TestClassifyFailureWarnOnly(t *testing.T) {
	phase := pluginsCore.PhaseInfoFailure("OOMKilled", "container was oom killed", nil)
	phase = phase.WithVersion(2)

	first := gpuFault(92, SeverityWarn)
	got := ClassifyFailure(phase, []*core.GpuFault{first, gpuFault(4242, SeverityWarn)})

	assert.Equal(t, pluginsCore.PhasePermanentFailure, got.Phase())
	assert.Equal(t, "OOMKilled", got.Err().GetCode())
	assert.Equal(t, core.ExecutionError_USER, got.Err().GetKind())
	assert.Equal(t, "container was oom killed", got.Err().GetMessage())
	assert.Equal(t, first, got.Err().GetGpuFault())
	assert.Equal(t, uint32(2), got.Version())
}

// Classification must not write through to the PhaseInfo it was handed: the caller
// keeps using the original when nothing about it should change.
func TestClassifyFailureDoesNotMutateInput(t *testing.T) {
	phase := pluginsCore.PhaseInfoRetryableFailure("UnknownError", "Pod failed", nil)

	ClassifyFailure(phase, []*core.GpuFault{gpuFault(79, SeverityCritical)})

	assert.Equal(t, "UnknownError", phase.Err().GetCode())
	assert.Equal(t, "Pod failed", phase.Err().GetMessage())
	assert.Nil(t, phase.Err().GetGpuFault())
}
