package controller

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	actionsk8s "github.com/flyteorg/flyte/v2/actions/k8s"
	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
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

// ParseEventLevel parses the executor's k8sEventLevel config value.
func ParseEventLevel(s string) (EventLevel, error) {
	switch s {
	case "off":
		return EventLevelOff, nil
	case "terminal":
		return EventLevelTerminal, nil
	case "info":
		return EventLevelInfo, nil
	case "debug":
		return EventLevelDebug, nil
	default:
		return EventLevelOff, fmt.Errorf("unknown k8s event level %q, want off, terminal, info or debug", s)
	}
}

const (
	annPrefix     = "flyte.org/"
	annProject    = annPrefix + "project"
	annDomain     = annPrefix + "domain"
	annRunName    = annPrefix + "run-name"
	annActionName = annPrefix + "action-name"
	annAttempt    = annPrefix + "attempt"
	annPhase      = annPrefix + "phase"
	annVersion    = annPrefix + "version"
	annErrorKind  = annPrefix + "error-kind"
	annErrorCode  = annPrefix + "error-code"
	annInfo       = annPrefix + "info" // ActionEvent as protojson, default options
	annCluster    = annPrefix + "cluster"

	// Control plane context: what the runs service wrote on the TaskAction and the
	// action event does not carry.
	annParentActionName = annPrefix + "parent-action-name"
	annTaskType         = annPrefix + "task-type"
	annGroup            = annPrefix + "group"
	annLabels           = annPrefix + "labels" // user labels from the RunSpec, as a JSON object
)

// noteLimit is the apiserver's NoteLengthLimit. Exceeding it rejects the event.
const noteLimit = 1024

const (
	ReasonActionQueued              K8sEventType = "ActionQueued"
	ReasonActionWaitingForResources K8sEventType = "ActionWaitingForResources"
	ReasonActionInitializing        K8sEventType = "ActionInitializing"
	ReasonActionRunning             K8sEventType = "ActionRunning"
	ReasonActionSucceeded           K8sEventType = "ActionSucceeded"
	ReasonActionFailed              K8sEventType = "ActionFailed"
	ReasonActionAborted             K8sEventType = "ActionAborted"
	ReasonActionTimedOut            K8sEventType = "ActionTimedOut"
	ReasonActionPaused              K8sEventType = "ActionPaused"
	ReasonActionRecovered           K8sEventType = "ActionRecovered"
	ReasonActionSystemRetry         K8sEventType = "ActionSystemRetry"
	ReasonActionPhaseUnknown        K8sEventType = "ActionPhaseUnknown"
)

func buildActionEventK8s(
	taskAction *flyteorgv1.TaskAction,
	event *workflow.ActionEvent,
	instance string,
) (*eventsv1.Event, error) {
	info, err := protojson.Marshal(event)
	if err != nil {
		return nil, err
	}

	ann := map[string]string{
		annProject:    event.GetId().GetRun().GetProject(),
		annDomain:     event.GetId().GetRun().GetDomain(),
		annRunName:    event.GetId().GetRun().GetName(),
		annActionName: event.GetId().GetName(),
		annAttempt:    strconv.FormatUint(uint64(event.GetAttempt()), 10),
		annPhase:      event.GetPhase().String(),
		annVersion:    strconv.FormatUint(uint64(event.GetVersion()), 10),
		annCluster:    event.GetCluster(),
		// TODO: should we limit the size of this? apiserver limits all annotations
		// on one object to 256 kiB.
		annInfo: string(info),
	}
	if e := event.GetErrorInfo(); e != nil {
		ann[annErrorKind] = e.GetKind().String()
		ann[annErrorCode] = e.GetCode()
	}
	for k, v := range controlPlaneContext(taskAction) {
		ann[k] = v
	}

	return &eventsv1.Event{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: taskAction.Name + "-",
			Namespace:    taskAction.Namespace, // must equal Regarding.Namespace
			Annotations:  ann,
		},
		// TODO: need to fallback when event.GetReportedTime() is nil
		EventTime:           metav1.NewMicroTime(event.GetReportedTime().AsTime()), // required
		ReportingController: "taskaction-controller",
		ReportingInstance:   instance,           // required, <= 128 chars
		Type:                eventType(event),   // Normal | Warning
		Reason:              eventReason(event), // ActionSucceeded | ActionFailed | SystemRetry | ...
		Action:              "Reconciling",
		Note:                humanSummary(event),
		Regarding: corev1.ObjectReference{
			APIVersion: flyteorgv1.GroupVersion.String(),
			Kind:       "TaskAction",
			Namespace:  taskAction.Namespace,
			Name:       taskAction.Name,
			UID:        taskAction.UID,
		},
	}, nil
}

// controlPlaneContext returns the annotations that carry what the runs service wrote on
// taskAction: its place in the action tree, its plugin type, its group, and the user
// labels from the RunSpec. A key is omitted when the TaskAction has no value for it.
func controlPlaneContext(taskAction *flyteorgv1.TaskAction) map[string]string {
	ann := map[string]string{}
	if parent := taskAction.Spec.ParentActionName; parent != nil && *parent != "" {
		ann[annParentActionName] = *parent
	}
	if taskAction.Spec.TaskType != "" {
		ann[annTaskType] = taskAction.Spec.TaskType
	}
	if taskAction.Spec.Group != "" {
		ann[annGroup] = taskAction.Spec.Group
	}

	userLabels := map[string]string{}
	for k, v := range taskAction.Labels {
		if !strings.HasPrefix(k, annPrefix) {
			userLabels[k] = v
		}
	}
	if len(userLabels) > 0 {
		if b, err := json.Marshal(userLabels); err == nil {
			ann[annLabels] = string(b)
		}
	}
	return ann
}

// eventType returns Warning for events that need attention and Normal for the rest.
func eventType(event *workflow.ActionEvent) string {
	if isSystemRetryEvent(event) {
		return corev1.EventTypeWarning
	}
	switch event.GetPhase() {
	case common.ActionPhase_ACTION_PHASE_FAILED, common.ActionPhase_ACTION_PHASE_TIMED_OUT:
		return corev1.EventTypeWarning
	default:
		return corev1.EventTypeNormal
	}
}

// eventReason returns the reason of the k8s event for an action event.
func eventReason(event *workflow.ActionEvent) string {
	if isSystemRetryEvent(event) {
		return string(ReasonActionSystemRetry)
	}
	switch event.GetPhase() {
	case common.ActionPhase_ACTION_PHASE_QUEUED:
		return string(ReasonActionQueued)
	case common.ActionPhase_ACTION_PHASE_WAITING_FOR_RESOURCES:
		return string(ReasonActionWaitingForResources)
	case common.ActionPhase_ACTION_PHASE_INITIALIZING:
		return string(ReasonActionInitializing)
	case common.ActionPhase_ACTION_PHASE_RUNNING:
		return string(ReasonActionRunning)
	case common.ActionPhase_ACTION_PHASE_SUCCEEDED:
		return string(ReasonActionSucceeded)
	case common.ActionPhase_ACTION_PHASE_FAILED:
		return string(ReasonActionFailed)
	case common.ActionPhase_ACTION_PHASE_ABORTED:
		return string(ReasonActionAborted)
	case common.ActionPhase_ACTION_PHASE_TIMED_OUT:
		return string(ReasonActionTimedOut)
	case common.ActionPhase_ACTION_PHASE_PAUSED:
		return string(ReasonActionPaused)
	case common.ActionPhase_ACTION_PHASE_RECOVERED:
		return string(ReasonActionRecovered)
	default:
		return string(ReasonActionPhaseUnknown)
	}
}

// humanSummary returns a one-line description of event for human reading the k8s event.
// the result fits in the note of a k8s event
func humanSummary(event *workflow.ActionEvent) string {
	var state string
	if isSystemRetryEvent(event) {
		state = "restarted after a system failure"
	} else {
		phase := strings.TrimPrefix(event.GetPhase().String(), "ACTION_PHASE_")
		state = strings.ToLower(strings.ReplaceAll(phase, "_", " "))
	}
	summary := fmt.Sprintf("action %s attempt %d %s", event.GetId().GetName(), event.GetAttempt(), state)
	if e := event.GetErrorInfo(); e != nil {
		kind := strings.TrimPrefix(e.GetKind().String(), "KIND_")
		summary += fmt.Sprintf(": %s error %q: %s", kind, e.GetCode(), e.GetMessage())
	}
	return truncateUTF8(summary, noteLimit)
}

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
