// Package gang is the scheduler-agnostic core for gang (all-or-nothing) workloads
// that an external admission gate holds or evicts.
//
// The first gate is Kueue holding JobSets suspended, but nothing in this package
// knows about JobSets, Kubernetes or Kueue: the plugin that observes the gate
// builds an Eviction descriptor, and the executors share the decision, the error
// codes and the budget accounting defined here. Keep it a leaf package that
// imports nothing heavier than the generated protobuf types, so any component that
// schedules or retries gangs can depend on it.
package gang

import (
	"fmt"
	"strings"
	"time"

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
)

// Reason classifies why a gate released or evicted a gang.
type Reason string

const (
	ReasonUnknown          Reason = "Unknown"
	ReasonPreempted        Reason = "Preempted"
	ReasonPodsReadyTimeout Reason = "PodsReadyTimeout"
	ReasonDeactivated      Reason = "Deactivated"
)

// Eviction describes one gate decision about a gang. The edge that observes the
// gate builds it; nothing here is specific to that edge.
type Eviction struct {
	// Source names the gate that made the decision, e.g. "kueue". Informational.
	Source string
	// Reason is the classified cause, ReasonUnknown when the edge cannot tell.
	Reason Reason
	// Message is the gate's own explanation, if it gave one.
	Message string
	// Started is true when the whole gang was up at least once in this attempt.
	// It decides whether anything was lost: a gang that never fully started can
	// simply be held until the gate re-admits it.
	Started bool
	// OccurredAt is when the edge observed the decision.
	OccurredAt time.Time
}

// Policy configures how a post-start eviction is reported.
type Policy struct {
	// AsSystemRetry reports the eviction as a SYSTEM-retryable failure, which the
	// executors turn into a fresh attempt without charging user retries. When
	// false it is reported as a USER-retryable failure instead, which keeps the
	// gate's own in-place requeue semantics for deployments that prefer them.
	AsSystemRetry bool
}

// Action is what the caller should do with an evicted gang.
type Action int

const (
	// Hold means the gang never fully started: keep reporting the current phase
	// and let the gate re-admit it. Nothing was lost and no restart is burned.
	Hold Action = iota
	// SystemRetry means the gang had started: report a SYSTEM-retryable failure so
	// the executor discards the resource and re-runs the attempt.
	SystemRetry
	// UserRetry is SystemRetry charged to the user's retry budget instead.
	UserRetry
)

func (a Action) String() string {
	switch a {
	case Hold:
		return "Hold"
	case SystemRetry:
		return "SystemRetry"
	case UserRetry:
		return "UserRetry"
	default:
		return fmt.Sprintf("Action(%d)", int(a))
	}
}

// Decide maps an Eviction to an Action. The returned error is nil for Hold and
// otherwise carries CodeGangEvicted with the kind implied by the policy.
func Decide(e Eviction, p Policy) (Action, *core.ExecutionError) {
	if !e.Started {
		return Hold, nil
	}
	action, kind := UserRetry, core.ExecutionError_USER
	if p.AsSystemRetry {
		action, kind = SystemRetry, core.ExecutionError_SYSTEM
	}
	return action, &core.ExecutionError{
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

// Budget bounds how many times one attempt may be evicted before the executor
// gives up on it. Max == 0 means unlimited.
type Budget struct {
	Max uint32
}

// Exhausted reports whether evictions has reached the budget.
func (b Budget) Exhausted(evictions uint32) bool {
	return b.Max > 0 && evictions >= b.Max
}
