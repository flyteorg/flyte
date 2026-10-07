// Package gang is the scheduler-agnostic core for gang (all-or-nothing) workloads
// that an external admission gate holds or evicts.
//
// The first gate is Kueue holding JobSets suspended, but nothing in this package
// knows about JobSets, Kubernetes or Kueue: the plugin that observes the gate
// decides that a started gang was evicted and describes it with an Eviction, and
// the executors share the error codes and the budget accounting defined here.
// Keep it a leaf package that imports nothing heavier than the generated protobuf
// types, so any component that schedules or retries gangs can depend on it.
package gang

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

const (
	// CodeGangEvicted is the SYSTEM error code reported when an admission gate revokes
	// a gang that had already fully started. Executors treat it as a system-retryable
	// failure charged against a dedicated eviction budget, never against the user's
	// attempts.
	CodeGangEvicted = "GangEvicted"

	// CodeGangEvictionsExceeded is the SYSTEM error code reported when a gang has been
	// evicted more times than the eviction budget allows.
	CodeGangEvictionsExceeded = "GangEvictionsExceeded"

	// CodeGangAdmissionTimeout is the SYSTEM error code reported when a gate held a gang
	// that never started for longer than the configured admission timeout. It is not an
	// eviction (nothing had run), but it is relaunched and charged to the same budget, so
	// a gang that can never be admitted does not retry forever.
	CodeGangAdmissionTimeout = "GangAdmissionTimeout"
)

// Reason classifies why a gate released or evicted a gang.
type Reason string

const (
	ReasonUnknown          Reason = "Unknown"
	ReasonPreempted        Reason = "Preempted"
	ReasonPodsReadyTimeout Reason = "PodsReadyTimeout"
	ReasonDeactivated      Reason = "Deactivated"
)

// Eviction describes a gate revoking a gang that had already fully started. A gang
// the gate holds before it ever started is not an eviction: nothing was lost, so
// the caller keeps waiting instead of building one.
type Eviction struct {
	// Source names the gate that made the decision, e.g. "kueue". Informational.
	Source string
	// Reason is the classified cause, ReasonUnknown when the edge cannot tell.
	Reason Reason
	// Message is the gate's own explanation, if it gave one.
	Message string
}

// Policy configures how an eviction is reported.
type Policy struct {
	// AsSystemRetry reports the eviction as a SYSTEM-retryable failure: the executors
	// relaunch the attempt in place without charging user retries, and count it against
	// the eviction budget. When false it is reported as a USER-retryable failure: it
	// starts a new attempt and is charged to the task's own retries instead. Either way
	// the gang is relaunched as a new resource.
	AsSystemRetry bool
}

// ExecutionError returns the GangEvicted error for e, with the kind implied by the policy.
func (e Eviction) ExecutionError(p Policy) *core.ExecutionError {
	kind := core.ExecutionError_USER
	if p.AsSystemRetry {
		kind = core.ExecutionError_SYSTEM
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
	if e.Reason != "" && e.Reason != ReasonUnknown {
		fmt.Fprintf(&b, " (%s)", e.Reason)
	}
	if msg := strings.TrimSpace(e.Message); msg != "" {
		b.WriteString(": ")
		b.WriteString(msg)
	}
	return b.String()
}

// IsEviction reports whether err is a system-retryable gang eviction, i.e. one
// that executors must charge to the eviction budget rather than the ordinary
// system-retry budget.
func IsEviction(err *core.ExecutionError) bool {
	return err != nil && err.GetKind() == core.ExecutionError_SYSTEM && err.GetCode() == CodeGangEvicted
}

// UsesEvictionBudget reports whether err is relaunched against the eviction budget
// rather than the ordinary system-failure count: a system-retryable gang eviction, or a
// gang that timed out waiting for admission.
func UsesEvictionBudget(err *core.ExecutionError) bool {
	return IsEviction(err) ||
		(err != nil && err.GetKind() == core.ExecutionError_SYSTEM && err.GetCode() == CodeGangAdmissionTimeout)
}

// ReasonFromKueueMessage classifies the message Kueue records on the job's
// "Stopped" event (v0.19) when it evicts a workload.
func ReasonFromKueueMessage(msg string) Reason {
	m := strings.ToLower(msg)
	switch {
	case strings.Contains(m, "podsready"):
		return ReasonPodsReadyTimeout
	case strings.Contains(m, "preempted"):
		return ReasonPreempted
	case strings.Contains(m, "deactivated"):
		return ReasonDeactivated
	default:
		return ReasonUnknown
	}
}

// WithMessage returns a copy of err with msg appended as the gate's last word on
// the eviction. It is a no-op for a nil err or a blank msg.
func WithMessage(err *core.ExecutionError, msg string) *core.ExecutionError {
	msg = strings.TrimSpace(msg)
	if err == nil || msg == "" {
		return err
	}
	out := proto.Clone(err).(*core.ExecutionError)
	if out.GetMessage() == "" {
		out.Message = msg
	} else {
		out.Message = out.GetMessage() + "; gate: " + msg
	}
	return out
}

// Budget bounds how many times a gang may be relaunched after an eviction or an
// admission timeout before the executor gives up on it. Max == 0 means unlimited.
type Budget struct {
	Max uint32
}

// Exhausted reports whether evictions, the relaunches already spent, has reached the
// budget, so the next eviction must fail instead of being relaunched.
func (b Budget) Exhausted(evictions uint32) bool {
	return b.Max > 0 && evictions >= b.Max
}
