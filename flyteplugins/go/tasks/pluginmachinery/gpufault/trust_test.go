package gpufault

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

func TestFromPodEvent(t *testing.T) {
	const (
		podUID   = "pod-uid"
		podNode  = "ip-10-0-0-1"
		elseNode = "ip-10-0-0-2"
	)
	xidMessage := FormatEventMessage(FromProto(gpuFault(79, SeverityCritical)))
	sxidMessage := FormatEventMessage(FromProto(sxidFault(12028)))
	trusted := PodEvent{Reason: EventReasonXid, Message: xidMessage, RegardingUID: podUID, ReportingNode: podNode}
	scheduled := Pod{UID: podUID, NodeName: podNode}

	with := func(change func(*PodEvent)) PodEvent {
		ev := trusted
		change(&ev)
		return ev
	}

	tests := []struct {
		name     string
		event    PodEvent
		pod      Pod
		wantKind core.GpuFault_Kind
	}{
		{
			name:     "an xid recorded against the pod from its own node",
			event:    trusted,
			pod:      scheduled,
			wantKind: core.GpuFault_KIND_XID,
		},
		{
			name: "an sxid recorded against the pod from its own node",
			event: with(func(ev *PodEvent) {
				ev.Reason = EventReasonSXid
				ev.Message = sxidMessage
			}),
			pod:      scheduled,
			wantKind: core.GpuFault_KIND_SXID,
		},
		{
			// The prefix and the tail are free text; only the reason says the emitter
			// wrote the event.
			name:  "a valid looking message under some other reason",
			event: with(func(ev *PodEvent) { ev.Reason = "BackOff" }),
			pod:   scheduled,
		},
		{
			name:  "an event without a reason",
			event: with(func(ev *PodEvent) { ev.Reason = "" }),
			pod:   scheduled,
		},
		{
			name:  "a fault reason on a message that is not a fault",
			event: with(func(ev *PodEvent) { ev.Message = "Back-off restarting failed container" }),
			pod:   scheduled,
		},
		{
			name:  "an event recorded against an earlier pod of the same name",
			event: with(func(ev *PodEvent) { ev.RegardingUID = "some-other-pod-uid" }),
			pod:   scheduled,
		},
		{
			name:  "an event that does not say which object it means",
			event: with(func(ev *PodEvent) { ev.RegardingUID = "" }),
			pod:   scheduled,
		},
		{
			// Losing the UID does not excuse an event that names no object at all.
			name:  "an event without a regarding UID on a pod whose UID is unknown",
			event: with(func(ev *PodEvent) { ev.RegardingUID = "" }),
			pod:   Pod{},
		},
		{
			// The pod was deleted and only its name is left, so any incarnation's event
			// is credited to it, knowingly.
			name:     "a pod whose UID is unknown is matched by name",
			event:    with(func(ev *PodEvent) { ev.RegardingUID = "some-other-pod-uid" }),
			pod:      Pod{NodeName: podNode},
			wantKind: core.GpuFault_KIND_XID,
		},
		{
			name:  "an event reported from another node",
			event: with(func(ev *PodEvent) { ev.ReportingNode = elseNode }),
			pod:   scheduled,
		},
		{
			name:  "an event that names no reporting node when the pod's node is known",
			event: with(func(ev *PodEvent) { ev.ReportingNode = "" }),
			pod:   scheduled,
		},
		{
			// The identity object of a deleted pod carries neither field, so the event
			// is taken on its name, from whichever node reported it.
			name:     "a pod whose UID and node are both unknown",
			event:    with(func(ev *PodEvent) { ev.ReportingNode = elseNode }),
			pod:      Pod{},
			wantKind: core.GpuFault_KIND_XID,
		},
		{
			name:     "a pod whose UID and node are both unknown and an event with no reporting node",
			event:    with(func(ev *PodEvent) { ev.ReportingNode = "" }),
			pod:      Pod{},
			wantKind: core.GpuFault_KIND_XID,
		},
		{
			name:  "a pod whose UID is unknown but whose node is known still checks the node",
			event: with(func(ev *PodEvent) { ev.ReportingNode = elseNode }),
			pod:   Pod{NodeName: podNode},
		},
		{
			// A pod with a UID and no node was never scheduled and never held a GPU,
			// whatever any event says about it.
			name:  "a pod that was never scheduled",
			event: trusted,
			pod:   Pod{UID: podUID},
		},
		{
			name:  "a pod that was never scheduled and an event with no reporting node",
			event: with(func(ev *PodEvent) { ev.ReportingNode = "" }),
			pod:   Pod{UID: podUID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FromPodEvent(tt.event, tt.pod)
			if tt.wantKind == core.GpuFault_KIND_UNSPECIFIED {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.Equal(t, tt.wantKind, got.GetKind())
			assert.True(t, proto.Equal(FromEventMessage(tt.event.Message), got))
		})
	}
}

func TestIsFaultReason(t *testing.T) {
	assert.True(t, IsFaultReason(EventReasonXid))
	assert.True(t, IsFaultReason(EventReasonSXid))
	assert.False(t, IsFaultReason(""))
	assert.False(t, IsFaultReason("BackOff"))
	assert.False(t, IsFaultReason("gpuxiderror"))
}
