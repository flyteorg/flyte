# CloudEvents V2 design

Status: proposal
Tracks: flyteorg/flyte#7829

## What is missing

Flyte 1 pushed every execution event to a message broker. Systems outside Flyte — lineage
catalogs, alerting, cost accounting, pipelines in other orchestrators — subscribed to a topic
and reacted. They never talked to Flyte, and Flyte never knew they existed.

Flyte 2 record the event into k8s event directly. The user can use 3rd party exporter to watch
the events and store into their own down stream pipeline(e.g. Kafka, LOKI). This approach will
decouple flyte's event system and users pipeline implementation. 

```
v1
+------------+       +--------------+       +---------------------+
| flyteadmin |------>| broker       |------>| lineage / catalog   |
+------------+ push  |              |       +---------------------+
                     | durable      |       +---------------------+
                     | replayable   |------>| alerting            |
                     | fans out     |       +---------------------+
                     |              |       +---------------------+
                     |              |------>| downstream pipeline |
                     +--------------+       +---------------------+
                                            consumers are decoupled;
                                            admin never knows they exist

v2 proposed

+---------------------------------------+
| TaskActionReconciler                  |
| recordEvent :127                      |
+------+--------------------------------+
       |
       |  4 call sites: :454 timeout  :567 system retry
       |               :1053 abort   :1086 status changed
       v
+---------------------------------------+
| level filter (new)                    |
| terminal / info / debug               |
+------+--------------------------------+
       |
       |  Eventf(taskAction, ...)  regarding = TaskAction
       v
+---------------------------------------+
| apiserver  events.k8s.io/v1           |
+------+--------------------------------+
  note <= 1024 B, TTL 1h
       |
       |  core/v1 informer, OnAdd
       v
+---------------------------------------+
| kubernetes-event-exporter             |
+------+--------------------------------+
  route match: involvedObject.kind=TaskAction
       |
       +--------------+--------------+
       v              v              v
  +---------+   +---------+   +-----------+
  | Kafka   |   | Loki    |   | webhook   |
  +---------+   +---------+   +-----------+
```

### Envelope action event into k8s event

v1 needed a proto because it owned the wire. This design does not: the envelope is the
Kubernetes Event object, and the exporter serializes it. 

v1 defined four messages in `flyteidl/protos/flyteidl/event/cloudevents.proto`
(`CloudEventWorkflowExecution`, `CloudEventNodeExecution`, `CloudEventTaskExecution`,
`CloudEventExecutionStart`), because workflow, node and task executions were three different
types. v2 has one recursive type — an action, whose root action is the run — so the three
collapse into one payload shape, close to what `action_events` already stores:

```go
// Annotation keys are the external contract: exporter templates read them by name.
const (
	annPrefix      = "flyte.org/"
	annProject     = annPrefix + "project"
	annDomain      = annPrefix + "domain"
	annRunName     = annPrefix + "run-name"
	annActionName  = annPrefix + "action-name"
	annAttempt     = annPrefix + "attempt"
	annPhase       = annPrefix + "phase"
	annVersion     = annPrefix + "version"
	annErrorKind   = annPrefix + "error-kind"
	annErrorCode   = annPrefix + "error-code"
	annInfo        = annPrefix + "info"     // jsonpb of ActionEvent
	annLaunchPlan  = annPrefix + "launch-plan"
	annPrincipal   = annPrefix + "principal"
	annCluster     = annPrefix + "cluster"
)

// noteLimit is the apiserver's NoteLengthLimit. Exceeding it rejects the event.
const noteLimit = 1024

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
		annInfo:       string(info),
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
			GenerateName: taskAction.Name + ".",
			Namespace:    taskAction.Namespace, // must equal Regarding.Namespace
			Annotations:  ann,
		},
		EventTime:           metav1.NewMicroTime(event.GetReportedTime().AsTime()), // required
		ReportingController: "taskaction-controller",
		ReportingInstance:   instance, // required, <= 128 chars
		Type:                eventType(event),   // Normal | Warning
		Reason:              eventReason(event), // ActionSucceeded | ActionFailed | SystemRetry | ...
		Action:              "Reconciling",
		Note:                truncateRunes(humanSummary(event), noteLimit),
		Regarding: corev1.ObjectReference{
			APIVersion: flyteorgv1.GroupVersion.String(),
			Kind:       "TaskAction",
			Namespace:  taskAction.Namespace,
			Name:       taskAction.Name,
			UID:        taskAction.UID,
		},
	}, nil
}

// truncateRunes cuts on a rune boundary; a byte cut can emit invalid UTF-8 and the
// apiserver rejects the whole event.
func truncateRunes(s string, max int) string {
	if len(s) <= max {
		return s
	}
	for max > 0 && !utf8.RuneStart(s[max]) {
		max--
	}
	return s[:max]
}
```

`Eventf` cannot set annotations — `events.EventRecorder` exposes only `Eventf` — so an event
carrying the payload is created directly with the client, which gives up the EventCorrelator's
aggregation and spam filter. Terminal events take that trade; `info` and `debug` levels stay on
`Eventf` with `note` alone.

The exporter watches `core/v1`, so route matchers use that spelling: `note` ↔ `message`,
`regarding` ↔ `involvedObject`, `reportingController` ↔ `source.component`. One store, two
views, translated by the apiserver.

## How to export and watch events

Flyte ships no exporter and no broker client. Once the events are on the apiserver, getting
them to a queue is an off-the-shelf problem: an operator deploys a third-party event exporter,
points it at the events Flyte emits, and routes them wherever they want.
[resmoio/kubernetes-event-exporter](https://github.com/resmoio/kubernetes-event-exporter) is the
reference — it watches cluster events, filters them with a routing tree, and ships them to
Kafka, Loki, Elasticsearch, SNS/SQS, webhooks and a dozen other sinks. Config only:

```yaml
route:
  routes:
    - match:
        - receiver: flyte-kafka
          kind: TaskAction                        # core/v1 spelling of regarding.kind
          reportingController: taskaction-controller
receivers:
  - name: flyte-kafka
    kafka:
      topic: flyte-action-events
      brokers: ["kafka-0:9092"]
      layout:
        reason: "{{ .Reason }}"
        action: "{{ index .ObjectMeta.Annotations \"flyte.org/action-name\" }}"
        phase: "{{ index .ObjectMeta.Annotations \"flyte.org/phase\" }}"
        payload: "{{ index .ObjectMeta.Annotations \"flyte.org/info\" | fromJson }}"
```

The annotation keys from the previous section are the whole interface. A consumer subscribes to
the topic and never touches the Kubernetes API.

**This is a platform-operator component, not a user one.** The exporter watches events across
the cluster and needs RBAC of its own. Users consume the queue; they are not expected to hold a
kubeconfig for an execution cluster.

## Log level filter

Not every action event is worth an object in etcd. An operator sets one level in the executor's
config, and the executor emits a Kubernetes Event only for events at or above it. Everything
below is still written to `action_events` as it is today — the filter decides what gets a
Kubernetes Event, never what gets recorded.

```yaml
executor:
  events:
    # off | terminal | info | debug   (default: off)
    k8sEventLevel: terminal
```

| Level | Emits | Volume per attempt |
|---|---|---|
| `off` | nothing | 0 |
| `terminal` | succeeded, failed, aborted, timed out, system retry | ~1–2 |
| `info` | the above plus phase transitions (queued, initializing, running) | ~4–6 |
| `debug` | the above plus every phase-version bump, which is one per batch of cluster events picked up by `attachRecentObjectEvents` | 10+, unbounded for a task that keeps producing cluster events |

```go
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
	case isTerminalPhase(event.GetPhase()), event.GetReason() == systemRetryReason:
		return EventLevelTerminal
	case event.GetPhase() != prevPhase:
		return EventLevelInfo
	default:
		return EventLevelDebug // same phase, higher version
	}
}

func (r *TaskActionReconciler) emitK8sEvent(
	ctx context.Context,
	taskAction *flyteorgv1.TaskAction,
	event *workflow.ActionEvent,
	prevPhase common.ActionPhase,
) {
	if r.K8sEventLevel == EventLevelOff || eventLevelOf(event, prevPhase) > r.K8sEventLevel {
		return
	}
	// Terminal events carry the payload in annotations, so they bypass the recorder.
	if eventLevelOf(event, prevPhase) == EventLevelTerminal {
		ev, err := buildActionEventK8s(taskAction, event, r.reportingInstance)
		if err == nil {
			err = r.Create(ctx, ev)
		}
		if err != nil {
			log.FromContext(ctx).Error(err, "failed to emit action k8s event")
		}
		return
	}
	r.Recorder.Eventf(taskAction, nil, eventType(event), eventReason(event),
		"Reconciling", "%s", humanSummary(event))
}
```

The hook is `recordEvent` (`taskaction_controller.go:127`). All four emitters funnel through it
— timeout (`:454`), system retry (`:567`), abort (`:1053`) and the status-change path (`:1086`)
— so the filter belongs there and nowhere else. Placing it at the call sites guarantees one gets
missed.


