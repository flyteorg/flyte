package gpufault

import "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"

// PodEvent is what a consumer read off a Kubernetes Event recorded against a pod. The
// fields are plain strings so the package stays free of Kubernetes imports.
type PodEvent struct {
	// Reason is the event's machine-readable reason.
	Reason string
	// Message is the event's note, which for a fault is FormatEventMessage's output.
	Message string
	// RegardingUID is the UID of the object the event was recorded against.
	RegardingUID string
	// ReportingNode is the node the event was reported from: the event's reporting
	// instance, or the deprecated source host when only an older recorder filled it in.
	ReportingNode string
}

// Pod is what a consumer knows about the pod a fault would be credited to. Either field
// may be empty when the pod is gone and only its name is left.
type Pod struct {
	UID      string
	NodeName string
}

// FromPodEvent turns an event recorded against a pod into a typed fault, returning nil
// unless the event can be trusted to report a fault on that pod. It is the one place
// that decides whether a fault event is believed; whether a believed fault explains a
// failure is a separate question, see RelevantToFailure.
//
// An event is trusted when three things hold.
//
// Its reason is one the emitter uses. The message prefix alone is free text anyone who
// can record an event can write.
//
// It was recorded against this very pod. Events are found by the pod's namespace and
// name, which a recreated pod reuses, so the event's regarding UID has to be the pod's.
// An event without one is rejected: the API server does not fill that field, so its
// absence is a client that did not say which object it meant. When the pod's own UID is
// unknown, because it was deleted and only its name is left, the name match is all there
// is and it is accepted knowingly: a same-name replacement pod's faults could be
// credited here, a deliberate trade against losing every fault on the path where the
// hardware most clearly failed.
//
// It was reported from the pod's own node. The emitter's credentials only let it record
// events for the node it runs on, so an event reported from any other node was not
// written by the emitter that watched this pod's GPUs. An event that names no reporting
// node is rejected when the pod's node is known. When the pod's node is unknown, as for
// a deleted pod, the same trade as for the UID is made and the event is accepted.
//
// A pod that has a UID but no node was never scheduled. It never held a GPU, so it has
// no fault to trust.
func FromPodEvent(ev PodEvent, pod Pod) *core.GpuFault {
	if !IsFaultReason(ev.Reason) || !recordedAgainst(ev, pod) || !reportedFromNodeOf(ev, pod) {
		return nil
	}
	return FromEventMessage(ev.Message)
}

// IsFaultReason reports whether an event reason is one the GPU fault emitter writes.
func IsFaultReason(reason string) bool {
	return reason == EventReasonXid || reason == EventReasonSXid
}

func recordedAgainst(ev PodEvent, pod Pod) bool {
	if ev.RegardingUID == "" {
		return false
	}
	return pod.UID == "" || ev.RegardingUID == pod.UID
}

func reportedFromNodeOf(ev PodEvent, pod Pod) bool {
	if pod.NodeName == "" {
		// A pod with a UID but no node was never scheduled. Without either, the pod is
		// gone and the trade described on FromPodEvent applies.
		return pod.UID == ""
	}
	return ev.ReportingNode == pod.NodeName
}
