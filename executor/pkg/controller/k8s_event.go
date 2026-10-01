package controller

import (
	"strings"

	actionsk8s "github.com/flyteorg/flyte/v2/actions/k8s"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

// EventLevel controls which action events are also emitted as Kubernetes Events.
// Levels are ordered: an event is emitted when its level is at or below the
// configured level.
type EventLevel uint8

const (
	EventLevelOff EventLevel = iota
	EventLevelTerminal
	EventLevelInfo
	EventLevelDebug
)

// eventLevelOf classifies an action event. Terminal phases and system retries are what an
// external consumer acts on; a version bump is detail.
func eventLevelOf(event *workflow.ActionEvent, prevPhase common.ActionPhase) EventLevel {
	switch {
	case actionsk8s.IsTerminalPhase(event.GetPhase()), isSystemRetryEvent(event):
		return EventLevelTerminal
	case event.GetPhase() != prevPhase:
		return EventLevelInfo
	default:
		return EventLevelDebug
	}
}

// isSystemRetryEvent reports whether event is the Queued event recordSystemRetry publishes.
// ActionEvent has no reason field, so the reserved version range identifies it.
func isSystemRetryEvent(event *workflow.ActionEvent) bool {
	return event.GetPhase() == common.ActionPhase_ACTION_PHASE_QUEUED &&
		event.GetVersion() >= systemRetryEventVersionBase
}

// truncateUTF8 cuts s to at most limit bytes, with any invalid UTF-8
// removed. limit must be non-negative.
func truncateUTF8(s string, limit int) string {
	if len(s) > limit {
		s = s[:limit]
	}
	return strings.ToValidUTF8(s, "")
}
