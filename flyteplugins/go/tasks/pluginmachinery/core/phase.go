package core

import (
	"fmt"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	structpb "github.com/golang/protobuf/ptypes/struct"
)

const DefaultPhaseVersion = uint32(0)
const SystemErrorCode = "SystemError"

//go:generate enumer -type=Phase

type Phase int8

const (
	// Does not mean an error, but simply states that we dont know the state in this round, try again later. But can be used to signal a system error too
	PhaseUndefined Phase = iota
	PhaseNotReady
	// Indicates plugin is not ready to submit the request as it is waiting for resources
	PhaseWaitingForResources
	// Indicates plugin has submitted the execution, but it has not started executing yet
	PhaseQueued
	// The system has started the pre-execution process, like container download, cluster startup etc
	PhaseInitializing
	// Indicates that the task has started executing
	PhaseRunning
	// Indicates that the task has completed successfully
	PhaseSuccess
	// Indicates that the Failure is recoverable, by re-executing the task if retries permit
	PhaseRetryableFailure
	// Indicate that the failure is non recoverable even if retries exist
	PhasePermanentFailure
	// Indicates the task is waiting for the cache to be populated so it can reuse results
	PhaseWaitingForCache
)

var Phases = []Phase{
	PhaseUndefined,
	PhaseNotReady,
	PhaseWaitingForResources,
	PhaseQueued,
	PhaseInitializing,
	PhaseRunning,
	PhaseSuccess,
	PhaseRetryableFailure,
	PhasePermanentFailure,
	PhaseWaitingForCache,
}

// Returns true if the given phase is failure, retryable failure or success
func (p Phase) IsTerminal() bool {
	return p.IsFailure() || p.IsSuccess()
}

func (p Phase) IsFailure() bool {
	return p == PhasePermanentFailure || p == PhaseRetryableFailure
}

func (p Phase) IsSuccess() bool {
	return p == PhaseSuccess
}

func (p Phase) IsWaitingForResources() bool {
	return p == PhaseWaitingForResources
}

type ExternalResource struct {
	// A unique identifier for the external resource
	ExternalID string
	// Captures the status of caching for this external resource
	CacheStatus core.CatalogCacheStatus
	// A unique index for the external resource. Although the ID may change, this will remain the same
	// throughout task event reports and retries.
	Index uint32
	// Log information for the external resource
	Logs []*core.TaskLog
	// The number of times this external resource has been attempted
	RetryAttempt uint32
	// Phase (if exists) associated with the external resource
	Phase Phase
}

type TaskInfo struct {
	// log information for the task execution
	Logs []*core.TaskLog
	// This value represents the time the status occurred at. If not provided, it will be defaulted to the time Flyte
	// checked the task status.
	OccurredAt *time.Time
	// This value represents the time the status was reported at. If not provided, will be defaulted to the current time
	// when Flyte published the event.
	ReportedAt *time.Time
	// Custom Event information that the plugin would like to expose to the front-end
	CustomInfo *structpb.Struct
	// A collection of information about external resources launched by this task
	ExternalResources []*ExternalResource
}

func (t *TaskInfo) String() string {
	return fmt.Sprintf("Info<@%s>", t.OccurredAt.String())
}

// Additional info that should be sent to the front end. The Information is sent to the front-end if it meets certain
// criterion, for example currently, it is sent only if an event was not already sent for
type PhaseInfo struct {
	// Observed Phase of the launched Task execution
	phase Phase
	// Phase version. by default this can be left as empty => 0. This can be used if there is some additional information
	// to be provided to the Control plane. Phase information is immutable in control plane for a given Phase, unless
	// a new version is provided.
	version uint32
	// In case info needs to be provided
	info *TaskInfo
	// If only an error is observed. It is complementary to info
	err *core.ExecutionError
	// reason why the current phase exists.
	reason string
	// cleanupOnFailure indicates that this task should be cleaned up even though the phase indicates a failure. This
	// applies to situations where a task is marked a failure but is still running, for example an ImagePullBackoff in
	// a k8s Pod where the image does not exist will continually reattempt the pull even though it will never succeed.
	cleanupOnFailure bool
}

func (p PhaseInfo) Phase() Phase {
	return p.phase
}

func (p PhaseInfo) Version() uint32 {
	return p.version
}

func (p PhaseInfo) Reason() string {
	return p.reason
}

func (p PhaseInfo) Info() *TaskInfo {
	return p.info
}

func (p PhaseInfo) Err() *core.ExecutionError {
	return p.err
}

func (p PhaseInfo) CleanupOnFailure() bool {
	return p.cleanupOnFailure
}

func (p PhaseInfo) WithVersion(version uint32) PhaseInfo {
	return PhaseInfo{
		phase:   p.phase,
		version: version,
		info:    p.info,
		err:     p.err,
		reason:  p.reason,
	}
}

func (p PhaseInfo) String() string {
	if p.err != nil {
		return fmt.Sprintf("Phase<%s:%d Error:%s>", p.phase, p.version, p.err)
	}
	return fmt.Sprintf("Phase<%s:%d %s Reason:%s>", p.phase, p.version, p.info, p.reason)
}

// PhaseInfoUndefined should be used when the Phase is unknown usually associated with an error
var PhaseInfoUndefined = PhaseInfo{phase: PhaseUndefined}

func phaseInfo(p Phase, v uint32, err *core.ExecutionError, info *TaskInfo, cleanupOnFailure bool) PhaseInfo {
	if info == nil {
		info = &TaskInfo{}
	}
	if info.OccurredAt == nil {
		t := time.Now()
		info.OccurredAt = &t
	}
	return PhaseInfo{
		phase:            p,
		version:          v,
		info:             info,
		err:              err,
		cleanupOnFailure: cleanupOnFailure,
	}
}

// Return in the case the plugin is not ready to start
func PhaseInfoNotReady(t time.Time, version uint32, reason string) PhaseInfo {
	pi := phaseInfo(PhaseNotReady, version, nil, &TaskInfo{OccurredAt: &t}, false)
	pi.reason = reason
	return pi
}

// Deprecated: Please use PhaseInfoWaitingForResourcesInfo instead
func PhaseInfoWaitingForResources(t time.Time, version uint32, reason string) PhaseInfo {
	pi := phaseInfo(PhaseWaitingForResources, version, nil, &TaskInfo{OccurredAt: &t}, false)
	pi.reason = reason
	return pi
}

// Return in the case the plugin is not ready to start
func PhaseInfoWaitingForResourcesInfo(t time.Time, version uint32, reason string, info *TaskInfo) PhaseInfo {
	pi := phaseInfo(PhaseWaitingForResources, version, nil, info, false)
	pi.reason = reason
	return pi
}

func PhaseInfoQueued(t time.Time, version uint32, reason string) PhaseInfo {
	pi := phaseInfo(PhaseQueued, version, nil, &TaskInfo{OccurredAt: &t}, false)
	pi.reason = reason
	return pi
}

func PhaseInfoQueuedWithTaskInfo(version uint32, reason string, info *TaskInfo) PhaseInfo {
	pi := phaseInfo(PhaseQueued, version, nil, info, false)
	pi.reason = reason
	return pi
}

func PhaseInfoInitializing(t time.Time, version uint32, reason string, info *TaskInfo) PhaseInfo {

	pi := phaseInfo(PhaseInitializing, version, nil, info, false)
	pi.reason = reason
	return pi
}

func phaseInfoFailed(p Phase, err *core.ExecutionError, info *TaskInfo, cleanupOnFailure bool) PhaseInfo {
	if err == nil {
		err = &core.ExecutionError{
			Code:    "Unknown",
			Message: "Unknown error message",
		}
	}
	return phaseInfo(p, DefaultPhaseVersion, err, info, cleanupOnFailure)
}

func PhaseInfoFailed(p Phase, err *core.ExecutionError, info *TaskInfo) PhaseInfo {
	return phaseInfo(p, DefaultPhaseVersion, err, info, false)
}

func PhaseInfoRunning(version uint32, info *TaskInfo) PhaseInfo {
	return phaseInfo(PhaseRunning, version, nil, info, false)
}

func PhaseInfoSuccess(info *TaskInfo) PhaseInfo {
	return phaseInfo(PhaseSuccess, DefaultPhaseVersion, nil, info, false)
}

func PhaseInfoSystemFailure(code, reason string, info *TaskInfo) PhaseInfo {
	return PhaseInfoFailed(PhasePermanentFailure, &core.ExecutionError{Code: code, Message: reason, Kind: core.ExecutionError_SYSTEM}, info)
}

func PhaseInfoFailure(code, reason string, info *TaskInfo) PhaseInfo {
	return PhaseInfoFailed(PhasePermanentFailure, &core.ExecutionError{Code: code, Message: reason, Kind: core.ExecutionError_USER}, info)
}

func PhaseInfoRetryableFailure(code, reason string, info *TaskInfo) PhaseInfo {
	return PhaseInfoFailed(PhaseRetryableFailure, &core.ExecutionError{Code: code, Message: reason, Kind: core.ExecutionError_USER}, info)
}

func PhaseInfoRetryableFailureWithCleanup(code, reason string, info *TaskInfo) PhaseInfo {
	return phaseInfoFailed(PhaseRetryableFailure, &core.ExecutionError{Code: code, Message: reason, Kind: core.ExecutionError_USER}, info, true)
}

func PhaseInfoSystemRetryableFailure(code, reason string, info *TaskInfo) PhaseInfo {
	return PhaseInfoFailed(PhaseRetryableFailure, &core.ExecutionError{Code: code, Message: reason, Kind: core.ExecutionError_SYSTEM}, info)
}

// Creates a new PhaseInfo with phase set to PhaseWaitingForCache
func PhaseInfoWaitingForCache(version uint32, info *TaskInfo) PhaseInfo {
	return phaseInfo(PhaseWaitingForCache, version, nil, info, false)
}
