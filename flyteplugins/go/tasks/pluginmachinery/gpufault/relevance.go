package gpufault

import "time"

// RelevanceWindow bounds how long before a failure a fault can still explain it. It only
// separates the fault that explains this failure from one the node saw much earlier;
// which pod a fault belongs to is a question the consumer settles for itself, before it
// gets here. It is measured from the failure's own time rather than from when the check
// runs, so a slow consumer cannot age a fault out.
//
// Thirty minutes spans the slow paths between a fault and the failure it causes: a
// container left wedged after a bus fault until the kubelet gives up on it, and a node
// going NotReady with its pods evicted only after the node-monitor grace period and
// eviction timeout.
//
// A fault that was still firing inside the window counts even if it started before it,
// because what the window bounds is how stale a fault's last sign of life may be, not how
// old the fault is. See RelevantToFailure.
const RelevanceWindow = 30 * time.Minute

// AfterFailureSlack is how far past the failure a fault may first be recorded and still
// count. The kernel line and the container's termination are stamped by different
// processes on the same node and the daemon reads the kernel log with a small lag, so a
// fault can first be recorded moments after the failure it caused; a fault that only
// started later than that cannot have caused it.
//
// It bounds when a fault started, not when it stopped. Hardware that keeps faulting after
// the container died goes on being observed for as long as it goes on faulting, and that
// says nothing about whether it caused the failure. See RelevantToFailure.
const AfterFailureSlack = 2 * time.Minute

// RelevantToFailure reports whether a fault was active close enough to a failure to
// explain it.
//
// A fault that keeps repeating is aggregated into a single event whose last observation
// moves with every repeat, so a fault describes an interval and not a moment: it was first
// recorded at activeFrom and was still firing at activeUntil. The failure has an interval
// of its own, the window before it in which a fault could have caused it and the small
// slack after it in which a fault it caused could still be recorded. The fault counts when
// those two intervals overlap.
//
// Testing the last observation alone drops the fault that matters most: hardware that
// keeps faulting after the container died has a last observation well past the failure, so
// the longer it goes on the more certainly it would be discarded. Testing the first
// recording alone drops the opposite case, a fault that started before the window opened
// and was still firing when the task died. Overlap keeps both and still rejects a fault
// that only started after the failure, or one that had stopped firing before the window
// opened.
//
// Either time may be zero, and the other then stands in for it, which is how a fault
// recorded once rather than aggregated is judged as the moment it is. A fault with neither
// time, or one whose last observation precedes its first recording, explains nothing: no
// honest recorder produces either, and nothing can be concluded from them.
func RelevantToFailure(activeFrom, activeUntil, failureAt time.Time) bool {
	if activeFrom.IsZero() {
		activeFrom = activeUntil
	}
	if activeUntil.IsZero() {
		activeUntil = activeFrom
	}
	if activeFrom.IsZero() || activeUntil.Before(activeFrom) {
		return false
	}

	relevantFrom := failureAt.Add(-RelevanceWindow)
	relevantUntil := failureAt.Add(AfterFailureSlack)

	return !activeFrom.After(relevantUntil) && !activeUntil.Before(relevantFrom)
}
