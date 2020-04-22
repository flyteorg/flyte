package handler

import (
	"time"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytestdlib/storage"
)

type EPhase uint8

const (
	EPhaseUndefined EPhase = iota
	EPhaseNotReady
	EPhaseQueued
	EPhaseRunning
	EPhaseSkip
	EPhaseFailed
	EPhaseRetryableFailure
	EPhaseSuccess
	EPhaseTimedout
)

func (p EPhase) String() string {
	switch p {
	case EPhaseNotReady:
		return "not-ready"
	case EPhaseQueued:
		return "queued"
	case EPhaseRunning:
		return "running"
	case EPhaseSkip:
		return "skip"
	case EPhaseFailed:
		return "failed"
	case EPhaseRetryableFailure:
		return "retryable-fail"
	case EPhaseSuccess:
		return "success"
	case EPhaseTimedout:
		return "timedout"
	}
	return "undefined"
}

func (p EPhase) IsTerminal() bool {
	if p == EPhaseFailed || p == EPhaseSuccess || p == EPhaseSkip || p == EPhaseTimedout {
		return true
	}
	return false
}

type DynamicNodeInfo struct {
}

type WorkflowNodeInfo struct {
	LaunchedWorkflowID *core.WorkflowExecutionIdentifier
}

type BranchNodeInfo struct {
}

type TaskNodeInfo struct {
	CacheHit bool
	// TaskPhase etc
}

type OutputInfo struct {
	OutputURI storage.DataReference
}

type ExecutionInfo struct {
	DynamicNodeInfo  *DynamicNodeInfo
	WorkflowNodeInfo *WorkflowNodeInfo
	BranchNodeInfo   *BranchNodeInfo
	OutputInfo       *OutputInfo
	TaskNodeInfo     *TaskNodeInfo
}

type PhaseInfo struct {
	p          EPhase
	occurredAt time.Time
	err        *core.ExecutionError
	info       *ExecutionInfo
	reason     string
}

func (p PhaseInfo) GetPhase() EPhase {
	return p.p
}

func (p PhaseInfo) GetOccurredAt() time.Time {
	return p.occurredAt
}

func (p PhaseInfo) GetErr() *core.ExecutionError {
	return p.err
}

func (p PhaseInfo) GetInfo() *ExecutionInfo {
	return p.info
}

func (p PhaseInfo) GetReason() string {
	return p.reason
}

var PhaseInfoUndefined = PhaseInfo{p: EPhaseUndefined}

func phaseInfo(p EPhase, err *core.ExecutionError, info *ExecutionInfo, reason string) PhaseInfo {
	return PhaseInfo{
		p:          p,
		err:        err,
		occurredAt: time.Now(),
		info:       info,
		reason:     reason,
	}
}

func PhaseInfoNotReady(reason string) PhaseInfo {
	return phaseInfo(EPhaseNotReady, nil, nil, reason)
}

func PhaseInfoQueued(reason string) PhaseInfo {
	return phaseInfo(EPhaseQueued, nil, nil, reason)
}

func PhaseInfoRunning(info *ExecutionInfo) PhaseInfo {
	return phaseInfo(EPhaseRunning, nil, info, "running")
}

func PhaseInfoSuccess(info *ExecutionInfo) PhaseInfo {
	return phaseInfo(EPhaseSuccess, nil, info, "successfully completed")
}

func PhaseInfoSkip(info *ExecutionInfo, reason string) PhaseInfo {
	return phaseInfo(EPhaseSkip, nil, info, reason)
}

func PhaseInfoTimedOut(info *ExecutionInfo, reason string) PhaseInfo {
	return phaseInfo(EPhaseTimedout, nil, info, reason)
}

func phaseInfoFailed(p EPhase, err *core.ExecutionError, info *ExecutionInfo) PhaseInfo {
	if err == nil {
		err = &core.ExecutionError{
			Code:    "Unknown",
			Message: "Unknown error message",
		}
	}
	return phaseInfo(p, err, info, err.Message)
}

func PhaseInfoFailure(code, reason string, info *ExecutionInfo) PhaseInfo {
	return phaseInfoFailed(EPhaseFailed, &core.ExecutionError{Code: code, Message: reason}, info)
}

func PhaseInfoFailureErr(err *core.ExecutionError, info *ExecutionInfo) PhaseInfo {
	return phaseInfoFailed(EPhaseFailed, err, info)
}

func PhaseInfoRetryableFailure(code, reason string, info *ExecutionInfo) PhaseInfo {
	return phaseInfoFailed(EPhaseRetryableFailure, &core.ExecutionError{Code: code, Message: reason}, info)
}

func PhaseInfoRetryableFailureErr(err *core.ExecutionError, info *ExecutionInfo) PhaseInfo {
	return phaseInfoFailed(EPhaseRetryableFailure, err, info)
}
