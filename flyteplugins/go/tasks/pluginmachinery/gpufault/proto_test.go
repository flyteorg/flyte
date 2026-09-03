package gpufault

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

func TestToProto(t *testing.T) {
	fault := Fault{
		Kind: KindXid, Code: 79, Name: NameFor(KindXid, 79), Severity: SeverityCritical,
		PCI: "0000:3b:00.0", PID: 1234, Process: "python",
	}
	attribution := Attribution{NodeName: "ip-10-0-0-1", GPUUUID: testUUID, GPUIndex: 0}

	got := ToProto(fault, attribution)

	assert.Equal(t, core.GpuFault_KIND_XID, got.GetKind())
	assert.Equal(t, uint32(79), got.GetCode())
	assert.Equal(t, "GPU has fallen off the bus", got.GetName())
	assert.Equal(t, core.GpuFault_SEVERITY_CRITICAL, got.GetSeverity())
	assert.Equal(t, testUUID, got.GetGpuUuid())
	require.NotNil(t, got.GpuIndex, "index 0 must be distinguishable from an unknown index")
	assert.Equal(t, uint32(0), got.GetGpuIndex())
	assert.Equal(t, "0000:3b:00.0", got.GetPciBusId())
	assert.Equal(t, "ip-10-0-0-1", got.GetNode())
	assert.Equal(t, uint32(1234), got.GetPid())
	assert.Equal(t, "python", got.GetProcess())
}

func TestProtoRoundTrip(t *testing.T) {
	tests := []struct {
		name        string
		fault       Fault
		attribution Attribution
	}{
		{
			name: "xid with a resolved gpu",
			fault: Fault{
				Kind: KindXid, Code: 31, Name: NameFor(KindXid, 31), Severity: SeverityUser,
				PCI: "0000:3b:00.0", PID: 12, Process: "python",
			},
			attribution: Attribution{NodeName: "n1", GPUUUID: testUUID, GPUIndex: 2},
		},
		{
			name: "sxid with an unknown gpu index",
			fault: Fault{
				Kind: KindSXid, Code: 12028, Name: NameFor(KindSXid, 12028), Severity: SeverityCritical,
				PCI: "0000:05:00.0",
			},
			attribution: Attribution{NodeName: "n1", GPUIndex: UnknownGPUIndex},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fault, attribution := FromProto(ToProto(tt.fault, tt.attribution))
			assert.Equal(t, tt.fault, fault)
			assert.Equal(t, tt.attribution, attribution)
		})
	}
}

// A producer that only filled in the numbers still yields a usable fault.
func TestFromProtoFillsInFromTheCode(t *testing.T) {
	fault, attribution := FromProto(&core.GpuFault{Code: 79})

	assert.Equal(t, KindXid, fault.Kind)
	assert.Equal(t, "GPU has fallen off the bus", fault.Name)
	assert.Equal(t, SeverityCritical, fault.Severity)
	assert.Equal(t, UnknownGPUIndex, attribution.GPUIndex)
}

func TestFromProtoNil(t *testing.T) {
	fault, attribution := FromProto(nil)
	assert.Equal(t, Fault{}, fault)
	assert.Equal(t, Attribution{GPUIndex: UnknownGPUIndex}, attribution)
}

func TestFromEventMessage(t *testing.T) {
	fault := Fault{
		Kind: KindXid, Code: 63, Name: NameFor(KindXid, 63), Severity: SeverityCritical,
		PCI: "0000:3b:00.0",
	}
	attribution := Attribution{NodeName: "n1", GPUUUID: testUUID, GPUIndex: 1}

	got := FromEventMessage(FormatEventMessage(fault, attribution))
	require.NotNil(t, got)
	assert.Equal(t, uint32(63), got.GetCode())
	assert.Equal(t, core.GpuFault_SEVERITY_CRITICAL, got.GetSeverity())
	assert.Equal(t, testUUID, got.GetGpuUuid())
}

func TestFromEventMessageIgnoresOtherEvents(t *testing.T) {
	messages := []string{
		"",
		"Back-off restarting failed container",
		"Successfully assigned flytesnacks-development/pod to ip-10-0-0-1",
	}

	for _, message := range messages {
		t.Run(message, func(t *testing.T) {
			assert.Nil(t, FromEventMessage(message))
		})
	}
}

func TestFromEventMessageSeverityIsCappedByTheTable(t *testing.T) {
	// The tail may lower the table's verdict but never raise it.
	up := FromEventMessage("[gpu-health] [CRITICAL] Xid 31 (GPU memory page fault) on GPU 0. xid=31 severity=critical gpu_index=0")
	require.NotNil(t, up)
	assert.Equal(t, core.GpuFault_SEVERITY_USER, up.GetSeverity())

	down := FromEventMessage("[gpu-health] [WARN] SXid 12028 on NVSwitch 0000:05:00.0. sxid=12028 severity=warn pci=0000:05:00.0")
	require.NotNil(t, down)
	assert.Equal(t, core.GpuFault_SEVERITY_WARN, down.GetSeverity())
}

func TestFromEventMessageUnknownSeverityFallsBackToTable(t *testing.T) {
	got := FromEventMessage("[gpu-health] [CRITICAL] Xid 79 (GPU has fallen off the bus) on GPU 0. xid=79 severity=fatal gpu_index=0")
	require.NotNil(t, got)
	assert.Equal(t, core.GpuFault_SEVERITY_CRITICAL, got.GetSeverity())
}
