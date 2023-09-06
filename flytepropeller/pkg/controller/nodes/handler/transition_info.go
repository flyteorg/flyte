package handler

import (
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flytestdlib/storage"
)

//go:generate enumer --type=EPhase --trimprefix=EPhase

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
	EPhaseFailing
	EPhaseDynamicRunning
	EPhaseRecovered
)

func (p EPhase) IsTerminal() bool {
	if p == EPhaseFailed || p == EPhaseSuccess || p == EPhaseSkip || p == EPhaseTimedout || p == EPhaseRecovered {
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

// Carries any information that should be sent as part of NodeEvents
type TaskNodeInfo struct {
	TaskNodeMetadata *event.TaskNodeMetadata
}

type GateNodeInfo struct {
}

type ArrayNodeInfo struct {
}

type OutputInfo struct {
	OutputURI storage.DataReference
	DeckURI   *storage.DataReference
}

type ExecutionInfo struct {
	DynamicNodeInfo  *DynamicNodeInfo
	WorkflowNodeInfo *WorkflowNodeInfo
	BranchNodeInfo   *BranchNodeInfo
	Inputs           *core.LiteralMap
	OutputInfo       *OutputInfo
	TaskNodeInfo     *TaskNodeInfo
	GateNodeInfo     *GateNodeInfo
	ArrayNodeInfo    *ArrayNodeInfo
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

func (p PhaseInfo) WithInfo(i *ExecutionInfo) PhaseInfo {
	return PhaseInfo{
		p:          p.p,
		occurredAt: p.occurredAt,
		err:        p.err,
		info:       i,
		reason:     p.reason,
	}
}

func (p PhaseInfo) WithOccuredAt(t time.Time) PhaseInfo {
	return PhaseInfo{
		p:          p.p,
		occurredAt: t,
		err:        p.err,
		info:       p.info,
		reason:     p.reason,
	}
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

func PhaseInfoQueued(reason string, inputs *core.LiteralMap) PhaseInfo {
	return phaseInfo(EPhaseQueued, nil, &ExecutionInfo{
		Inputs: inputs,
	}, reason)
}

func PhaseInfoRunning(info *ExecutionInfo) PhaseInfo {
	return phaseInfo(EPhaseRunning, nil, info, "running")
}

func PhaseInfoDynamicRunning(info *ExecutionInfo) PhaseInfo {
	return phaseInfo(EPhaseDynamicRunning, nil, info, "dynamic workflow running")
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

func PhaseInfoRecovered(info *ExecutionInfo) PhaseInfo {
	return phaseInfo(EPhaseRecovered, nil, info, "successfully recovered")
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

func PhaseInfoFailure(kind core.ExecutionError_ErrorKind, code, reason string, info *ExecutionInfo) PhaseInfo {
	return phaseInfoFailed(EPhaseFailed, &core.ExecutionError{Kind: kind, Code: code, Message: reason}, info)
}

func PhaseInfoFailureErr(err *core.ExecutionError, info *ExecutionInfo) PhaseInfo {
	return phaseInfoFailed(EPhaseFailed, err, info)
}

func PhaseInfoFailingErr(err *core.ExecutionError, info *ExecutionInfo) PhaseInfo {
	return phaseInfoFailed(EPhaseFailing, err, info)
}

func PhaseInfoRetryableFailure(kind core.ExecutionError_ErrorKind, code, reason string, info *ExecutionInfo) PhaseInfo {
	return phaseInfoFailed(EPhaseRetryableFailure, &core.ExecutionError{Kind: kind, Code: code, Message: reason}, info)
}

func PhaseInfoRetryableFailureErr(err *core.ExecutionError, info *ExecutionInfo) PhaseInfo {
	return phaseInfoFailed(EPhaseRetryableFailure, err, info)
}
