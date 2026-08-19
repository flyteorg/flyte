package gpufault

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testUUID = "GPU-6f3c1234-5678-90ab-cdef-1234567890ab"

func TestFormatEventMessage(t *testing.T) {
	tests := []struct {
		name        string
		fault       Fault
		attribution Attribution
		want        string
	}{
		{
			name: "critical fault with a process",
			fault: Fault{
				Kind: KindXid, Code: 79, Name: NameFor(KindXid, 79), Severity: SeverityCritical,
				PCI: "0000:3b:00.0", PID: 1234, Process: "python",
			},
			attribution: Attribution{NodeName: "ip-10-0-0-1", GPUUUID: testUUID, GPUIndex: 0},
			want: "[gpu-health] [CRITICAL] Xid 79 (GPU has fallen off the bus) on GPU 0 " + testUUID +
				". xid=79 severity=critical gpu_uuid=" + testUUID + " gpu_index=0 pci=0000:3b:00.0 node=ip-10-0-0-1 pid=1234 process=python",
		},
		{
			name: "user fault without a process",
			fault: Fault{
				Kind: KindXid, Code: 13, Name: NameFor(KindXid, 13), Severity: SeverityUser,
				PCI: "0000:3b:00.0",
			},
			attribution: Attribution{NodeName: "ip-10-0-0-1", GPUUUID: testUUID, GPUIndex: 3},
			want: "[gpu-health] [USER] Xid 13 (Graphics Engine Exception) on GPU 3 " + testUUID +
				". xid=13 severity=user gpu_uuid=" + testUUID + " gpu_index=3 pci=0000:3b:00.0 node=ip-10-0-0-1",
		},
		{
			name: "gpu could not be resolved",
			fault: Fault{
				Kind: KindXid, Code: 48, Name: NameFor(KindXid, 48), Severity: SeverityCritical,
				PCI: "0000:af:00.0",
			},
			attribution: Attribution{NodeName: "gke-node-1", GPUIndex: UnknownGPUIndex},
			want: "[gpu-health] [CRITICAL] Xid 48 (Double Bit ECC Error) on GPU at PCI 0000:af:00.0." +
				" xid=48 severity=critical pci=0000:af:00.0 node=gke-node-1",
		},
		{
			name: "nvswitch fault",
			fault: Fault{
				Kind: KindSXid, Code: 12028, Name: NameFor(KindSXid, 12028), Severity: SeverityCritical,
				PCI: "0000:05:00.0",
			},
			attribution: Attribution{NodeName: "ip-10-0-0-1", GPUIndex: UnknownGPUIndex},
			want: "[gpu-health] [CRITICAL] SXid 12028 on NVSwitch 0000:05:00.0." +
				" sxid=12028 severity=critical pci=0000:05:00.0 node=ip-10-0-0-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, FormatEventMessage(tt.fault, tt.attribution))
		})
	}
}

// Sentence is the message up to the full stop. Classification prepends it to failure
// messages, so it must never drag the k=v tail along with it.
func TestSentence(t *testing.T) {
	tests := []struct {
		name        string
		fault       Fault
		attribution Attribution
		want        string
	}{
		{
			name: "xid",
			fault: Fault{
				Kind: KindXid, Code: 79, Name: NameFor(KindXid, 79), Severity: SeverityCritical,
				PCI: "0000:3b:00.0",
			},
			attribution: Attribution{NodeName: "ip-10-0-0-1", GPUUUID: testUUID, GPUIndex: 0},
			want:        "[gpu-health] [CRITICAL] Xid 79 (GPU has fallen off the bus) on GPU 0 " + testUUID + ".",
		},
		{
			name: "sxid",
			fault: Fault{
				Kind: KindSXid, Code: 12028, Name: NameFor(KindSXid, 12028), Severity: SeverityCritical,
				PCI: "0000:05:00.0",
			},
			attribution: Attribution{NodeName: "ip-10-0-0-1", GPUIndex: UnknownGPUIndex},
			want:        "[gpu-health] [CRITICAL] SXid 12028 on NVSwitch 0000:05:00.0.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sentence := Sentence(tt.fault, tt.attribution)
			assert.Equal(t, tt.want, sentence)
			assert.True(t, len(FormatEventMessage(tt.fault, tt.attribution)) > len(sentence))
			assert.NotContains(t, sentence, "=")
		})
	}
}

func TestEventMessageRoundTrip(t *testing.T) {
	tests := []struct {
		name        string
		fault       Fault
		attribution Attribution
	}{
		{
			name: "xid with process",
			fault: Fault{
				Kind: KindXid, Code: 79, Name: NameFor(KindXid, 79), Severity: SeverityCritical,
				PCI: "0000:3b:00.0", PID: 1234, Process: "python",
			},
			attribution: Attribution{NodeName: "ip-10-0-0-1", GPUUUID: testUUID, GPUIndex: 0},
		},
		{
			name: "xid without process or gpu",
			fault: Fault{
				Kind: KindXid, Code: 92, Name: NameFor(KindXid, 92), Severity: SeverityWarn,
				PCI: "0000:af:00.0",
			},
			attribution: Attribution{NodeName: "gke-node-1", GPUIndex: UnknownGPUIndex},
		},
		{
			name: "unknown code keeps its generic name",
			fault: Fault{
				Kind: KindXid, Code: 4242, Name: NameFor(KindXid, 4242), Severity: SeverityWarn,
				PCI: "0000:af:00.0",
			},
			attribution: Attribution{NodeName: "gke-node-1", GPUIndex: UnknownGPUIndex},
		},
		{
			name: "sxid",
			fault: Fault{
				Kind: KindSXid, Code: 24001, Name: NameFor(KindSXid, 24001), Severity: SeverityCritical,
				PCI: "0000:0e:00.0",
			},
			attribution: Attribution{NodeName: "ip-10-0-0-1", GPUIndex: UnknownGPUIndex},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := FormatEventMessage(tt.fault, tt.attribution)

			fault, attribution, ok := ParseEventMessage(message)
			require.True(t, ok, "expected %q to parse", message)

			assert.Equal(t, tt.fault, fault)
			assert.Equal(t, tt.attribution, attribution)
		})
	}
}

func TestParseEventMessageRejects(t *testing.T) {
	messages := []string{
		"",
		"BackOff restarting failed container",
		"[gpu-health] something without a machine tail",
		"[gpu-health] [WARN] Xid ?? xid=notanumber severity=warn",
	}

	for _, message := range messages {
		t.Run(message, func(t *testing.T) {
			_, _, ok := ParseEventMessage(message)
			assert.False(t, ok)
		})
	}
}

func TestFormatEventMessageSanitizesProcessName(t *testing.T) {
	fault := Fault{
		Kind: KindXid, Code: 13, Name: NameFor(KindXid, 13), Severity: SeverityUser,
		PCI: "0000:3b:00.0", PID: 7, Process: "my train job",
	}

	message := FormatEventMessage(fault, Attribution{NodeName: "n1", GPUIndex: UnknownGPUIndex})
	assert.Contains(t, message, "process=my_train_job")

	parsed, _, ok := ParseEventMessage(message)
	require.True(t, ok)
	assert.Equal(t, "my_train_job", parsed.Process)
}
