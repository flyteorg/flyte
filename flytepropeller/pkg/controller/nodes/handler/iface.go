package handler

import (
	"context"
	"time"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

//go:generate mockery -all

type Data = core.LiteralMap
type VarName = string

type Phase int

const (
	// Indicates that the handler was unable to Start the Node due to an internal failure
	PhaseNotStarted Phase = iota
	// Incase of retryable failure and should be retried
	PhaseRetryableFailure
	// Indicates that the node is queued because the task is queued
	PhaseQueued
	// Indicates that the node is currently executing and no errors have been observed
	PhaseRunning
	// PhaseFailing is currently used by SubWorkflow Only. It indicates that the Node's primary work has failed,
	// but, either some cleanup or exception handling condition is in progress
	PhaseFailing
	// This is a terminal Status and indicates that the node execution resulted in a Failure
	PhaseFailed
	// This is a pre-terminal state, currently unused and indicates that the Node execution has succeeded barring any cleanup
	PhaseSucceeding
	// This is a terminal state and indicates successful completion of the node execution.
	PhaseSuccess
	// This Phase indicates that the node execution can be skipped, because of either conditional failures or user defined cases
	PhaseSkipped
	// This phase indicates that an error occurred and is always accompanied by `error`. the execution for that node is
	// in an indeterminate state and should be retried
	PhaseUndefined
)

var PhasesToString = map[Phase]string{
	PhaseNotStarted:       "NotStarted",
	PhaseQueued:           "Queued",
	PhaseRunning:          "Running",
	PhaseFailing:          "Failing",
	PhaseFailed:           "Failed",
	PhaseSucceeding:       "Succeeding",
	PhaseSuccess:          "Success",
	PhaseSkipped:          "Skipped",
	PhaseUndefined:        "Undefined",
	PhaseRetryableFailure: "RetryableFailure",
}

func (p Phase) String() string {
	str, found := PhasesToString[p]
	if found {
		return str
	}

	return "Unknown"
}

// This encapsulates the status of the node
type Status struct {
	Phase      Phase
	Err        error
	OccurredAt time.Time
}

var StatusNotStarted = Status{Phase: PhaseNotStarted}
var StatusQueued = Status{Phase: PhaseQueued}
var StatusRunning = Status{Phase: PhaseRunning}
var StatusSucceeding = Status{Phase: PhaseSucceeding}
var StatusSuccess = Status{Phase: PhaseSuccess}
var StatusUndefined = Status{Phase: PhaseUndefined}
var StatusSkipped = Status{Phase: PhaseSkipped}

func (s Status) WithOccurredAt(t time.Time) Status {
	s.OccurredAt = t
	return s
}

func StatusFailed(err error) Status {
	return Status{Phase: PhaseFailed, Err: err}
}

func StatusRetryableFailure(err error) Status {
	return Status{Phase: PhaseRetryableFailure, Err: err}
}

func StatusFailing(err error) Status {
	return Status{Phase: PhaseFailing, Err: err}
}

type OutputResolver interface {
	// Extracts a subset of node outputs to literals.
	ExtractOutput(ctx context.Context, w v1alpha1.ExecutableWorkflow, n v1alpha1.ExecutableNode,
		bindToVar VarName) (values *core.Literal, err error)
}

type PostNodeSuccessHandler interface {
	HandleNodeSuccess(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode) (Status, error)
}

// Interface that should be implemented for a node type.
type IFace interface {
	//OutputResolver

	// Initialize should be called, before invoking any other methods of this handler. Initialize will be called using one thread
	// only
	Initialize(ctx context.Context) error

	// Start node is called for a node only if the recorded state indicates that the node was never started previously.
	// the implementation should handle idempotency, even if the chance of invoking it more than once for an execution is rare.
	StartNode(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode, nodeInputs *Data) (Status, error)

	// For any node that is not in a NEW/READY state in the recording, CheckNodeStatus will be invoked. The implementation should handle
	// idempotency and return the current observed state of the node
	CheckNodeStatus(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode, previousNodeStatus v1alpha1.ExecutableNodeStatus) (Status, error)

	// This is called in the case, a node failure is observed.
	HandleFailingNode(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode) (Status, error)

	// Abort is invoked as a way to clean up failing/aborted workflows
	AbortNode(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode) error
}
