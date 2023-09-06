package core

import "fmt"

//go:generate enumer --type=TransitionType

// Type of Transition, refer to Transition to understand what transition means
type TransitionType int

const (
	// The transition is eventually consistent. For all the state written may not be visible in the next call, but eventually will persist
	// Best to use when the plugin logic is completely idempotent. This is also the most performant option.
	TransitionTypeEphemeral TransitionType = iota
	// @deprecated support for Barrier type transitions has been deprecated
	// This transition tries its best to make the latest state visible for every consecutive read. But, it is possible
	// to go back in time, i.e. monotonic consistency is violated (in rare cases).
	TransitionTypeBarrier
)

// A Plugin Handle method returns a Transition. This transition indicates to the Flyte framework that if the plugin wants to continue "Handle"ing this task,
// or if wants to move the task to success, attempt a retry or fail. The transition automatically sends an event to Admin service which shows the plugin
// provided information in the Console/cli etc
// The information to be published is in the PhaseInfo structure. Transition Type indicates the type of consistency for subsequent handle calls in case the phase info results in a non terminal state.
// the PhaseInfo structure is very important and is used to record events in Admin. Only if the Phase + PhaseVersion was not previously observed, will an event be published to Admin
// there are only a configurable number of phase-versions usable. Usually it is preferred to be a monotonically increasing sequence
type Transition struct {
	ttype TransitionType
	info  PhaseInfo
}

func (t Transition) Type() TransitionType {
	return t.ttype
}

func (t Transition) Info() PhaseInfo {
	return t.info
}

func (t Transition) String() string {
	return fmt.Sprintf("%s,%s", t.ttype, t.info)
}

// UnknownTransition is synonymous to UndefinedTransition. To be returned when an error is observed
var UnknownTransition = Transition{TransitionTypeEphemeral, PhaseInfoUndefined}

// Creates and returns a new Transition based on the PhaseInfo.Phase
// Phases: PhaseNotReady, PhaseQueued, PhaseInitializing, PhaseRunning will cause the system to continue invoking Handle
func DoTransitionType(ttype TransitionType, info PhaseInfo) Transition {
	return Transition{ttype: ttype, info: info}
}

// Same as DoTransition, but TransitionTime is always Ephemeral
func DoTransition(info PhaseInfo) Transition {
	return DoTransitionType(TransitionTypeEphemeral, info)
}
