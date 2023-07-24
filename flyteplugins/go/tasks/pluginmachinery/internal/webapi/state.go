package webapi

import (
	"time"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/webapi"
)

//go:generate enumer -type=Phase -trimprefix=Phase

// Phase represents current phase of the execution
type Phase int

const (
	// PhaseNotStarted the default phase.
	PhaseNotStarted Phase = iota

	// PhaseAllocationTokenAcquired once all required tokens have been acquired. The task is ready to be executed
	// remotely.
	PhaseAllocationTokenAcquired

	// PhaseResourcesCreated indicates the task has been created remotely.
	PhaseResourcesCreated

	// The resource has successfully been executed remotely.
	PhaseSucceeded

	// The resource has failed to be executed.
	PhaseUserFailure

	// The resource has failed to be executed due to a system error.
	PhaseSystemFailure
)

func (i Phase) IsTerminal() bool {
	return i == PhaseSucceeded || i == PhaseUserFailure || i == PhaseSystemFailure
}

// State is the persisted State of the resource.
type State struct {
	// Phase current phase of the resource.
	Phase Phase `json:"phase,omitempty"`

	// ResourceMeta contain metadata about resource this task created. This can be a complex structure or a simple type
	// (e.g. a string). It should contain enough information for the plugin to interact (retrieve, check status, delete)
	// with the resource through the remote service.
	ResourceMeta webapi.ResourceMeta `json:"resourceMeta,omitempty"`

	// This number keeps track of the number of failures within the sync function. Without this, what happens in
	// the sync function is entirely opaque. Note that this field is completely orthogonal to Flyte system/node/task
	// level retries, just errors from hitting API, inside the sync loop
	SyncFailureCount int `json:"syncFailureCount,omitempty"`

	// In creating the resource, this is the number of failures
	CreationFailureCount int `json:"creationFailureCount,omitempty"`

	// The time the execution first requests for an allocation token
	AllocationTokenRequestStartTime time.Time `json:"allocationTokenRequestStartTime,omitempty"`
}
