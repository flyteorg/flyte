// Package gang describes how a gang (all-or-nothing) workload that an external admission
// gate took back after it had started is reported to the executors.
//
// Nothing in this package knows about JobSets, Kubernetes or a particular gate: the plugin
// that observes the gate decides that a started gang was evicted, works out whose cause it
// was, and describes it with an Eviction. Keep it a leaf package that imports nothing
// heavier than the generated protobuf types.
package gang

import (
	"fmt"
	"strings"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

// CodeGangEvicted is the error code reported when an admission gate takes back a gang
// that had already fully started. Its kind says who pays for the retry: USER when the
// cause was the user's (the retry uses one of the task's retries), SYSTEM otherwise.
const CodeGangEvicted = "GangEvicted"

// Eviction describes a gate taking back a gang that had already fully started. A gang the
// gate holds before it ever started is not an eviction: nothing was lost, so the caller
// keeps waiting instead of building one.
type Eviction struct {
	// Source names the gate that made the decision, e.g. "kueue".
	Source string
	// Reason is the gate's own reason for the eviction, e.g. "Preempted". Empty when the
	// reason could not be read.
	Reason string
	// Detail refines Reason when the gate gives more, e.g. which kind of preemption.
	Detail string
	// Message is the gate's own explanation, if it gave one.
	Message string
	// UserCaused reports the eviction as a USER-retryable failure: the cause was the
	// user's (for example their gang lost a priority contest in the queue they chose), so
	// the relaunch uses one of the task's retries. Otherwise it is SYSTEM-retryable and
	// does not.
	UserCaused bool
	// RetriesLeft is how many of the task's retries remain before this eviction uses one:
	// 0 means there is no retry and the task fails. It is only reported for a user-caused
	// eviction.
	RetriesLeft uint32
}

// ExecutionError returns the GangEvicted error for e. The message says why the gang was
// taken back and whether that counts against the task's retries.
func (e Eviction) ExecutionError() *core.ExecutionError {
	kind := core.ExecutionError_SYSTEM
	if e.UserCaused {
		kind = core.ExecutionError_USER
	}
	return &core.ExecutionError{
		Code:    CodeGangEvicted,
		Message: e.describe(),
		Kind:    kind,
	}
}

func (e Eviction) describe() string {
	var b strings.Builder
	b.WriteString("gang evicted")
	if e.Source != "" {
		b.WriteString(" by ")
		b.WriteString(e.Source)
	}
	switch {
	case e.Reason != "" && e.Detail != "":
		fmt.Fprintf(&b, " (%s, %s)", e.Reason, e.Detail)
	case e.Reason != "":
		fmt.Fprintf(&b, " (%s)", e.Reason)
	}
	if msg := strings.TrimSpace(e.Message); msg != "" {
		b.WriteString(": ")
		b.WriteString(msg)
	}
	switch {
	case !e.UserCaused:
		b.WriteString("; this does not count against the task's retries")
	case e.RetriesLeft == 0:
		b.WriteString("; this counts against the task's retries and none are left, so the task fails")
	case e.RetriesLeft == 1:
		b.WriteString("; this uses the task's last retry")
	default:
		fmt.Fprintf(&b, "; this uses one of the task's retries (%d left after it)", e.RetriesLeft-1)
	}
	return b.String()
}
