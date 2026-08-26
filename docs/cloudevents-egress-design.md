# CloudEvents egress — design

Status: proposal
Scope: `runs/`
Tracks: flyteorg/flyte#7829

## What is missing

Flyte 1 pushed every execution event to a message broker. Systems outside Flyte — lineage
catalogs, alerting, cost accounting, pipelines in other orchestrators — subscribed to a topic
and reacted. They never talked to Flyte, and Flyte never knew they existed.

Flyte 2 records more events than v1 did, in a queryable table, with a live stream on top. What
it does not do is push them anywhere. Every consumer must now hold a long-lived gRPC stream
against the control plane, which makes each of them a client Flyte has to serve, keep
connected, and survive restarts with.

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

v2 proposed          [NEW] = added by this document; everything else exists today

+----------------------------------------------------------------------+
| InsertEvents()                      runs/repository/impl/action.go   |
|   the only writer of action_events                                   |
+-----------+--------------------+-------------------------+-----------+
            |                    |                         |
            | 1. INSERT          | 2. hand off       [NEW] | 3. notifyActionUpdate()
            |    (commit)        |    events in hand       |    id only, per-action
            v                    v                         v
+----------------------+ +--------------------+ +----------------------------+
| action_events        | | publisher    [NEW] | | pg_notify action_updates   |
|   the durable record | |   Flyte event      | |   -> actionSubscribers     |
|   no cursor column   | |   -> CloudEvent    | |      channel full: DROPPED |
+----------+-----------+ +---------+----------+ +-------------+--------------+
           ^                       |                          |
           |                       |                          v
           |                       |             +-------------------------+
           |                       |             | watch streams           |
           |                       |             |   console / CLI         |
           |                       |             |   existing subscribers  |
           |                       |             +-------------------------+
           |                       |
           +-----------------------+  [NEW] on restart, publisher re-reads
                                   |         SELECT ... WHERE > cursor, to
                                   |         cover the crash window: rows
                                   v         committed but never published
                +--------------------------------------+
                | sender                         [NEW] |
                |   one per transport                  |
                +------------------+-------------------+
                                   v
                   +---------------------+     +---------------------+
                   | broker        [NEW] |---->| lineage / catalog   |
                   |                     |     +---------------------+
                   | operator-run,       |     +---------------------+
                   | not shipped by      |---->| alerting            |
                   | Flyte               |     +---------------------+
                   |                     |     +---------------------+
                   |                     |---->| downstream pipeline |
                   +---------------------+     +---------------------+
                                               new consumers land here
```

### Protobuf

The wiring above has nowhere to land, because the payload type does not exist. v2 needs a
`CloudEventActionExecution` to carry an action's events, and nothing in `flyteidl2` defines one
— the directory contains no cloudevent proto at all.

v1 defines four, in `flyteidl/protos/flyteidl/event/cloudevents.proto`:

| message | wraps | adds |
|---|---|---|
| `CloudEventWorkflowExecution` | `WorkflowExecutionEvent` | output_interface, artifact_ids, reference_execution, principal, launch_plan_id, labels |
| `CloudEventNodeExecution` | `NodeExecutionEvent` | task_exec_id, output_interface, artifact_ids, principal, launch_plan_id, labels |
| `CloudEventTaskExecution` | `TaskExecutionEvent` | labels |
| `CloudEventExecutionStart` | — (no nested event) | execution_id, launch_plan_id, workflow_id, artifact_ids, artifact_trackers, principal |

**v2 should collapse the three execution messages into one.** v1 split them because workflow,
node and task executions are three different types. v2 has one recursive type: an action, whose
root action is the run. A single `CloudEventActionExecution` covers what took three messages,
and the fields it needs are close to what `action_events` already stores:

```
CloudEventActionExecution
  project, domain, run_name, name     the action identity
  attempt, phase, version             the event's position in the action's history
  info, error_kind                    the event payload as recorded
  <control-plane context>             the v1 pattern: whatever a broker consumer
                                      cannot ask Flyte for after the fact
```

## Where the publisher attaches

`InsertEvents` (`runs/repository/impl/action.go:117`) is the only writer of `action_events`,
and it already has the `[]*models.ActionEvent` batch in hand. That is the hook: the publisher
takes the same slice the INSERT just committed. No re-read on the happy path.

The obvious alternative — subscribe to `action_updates` alongside the watch streams — is worse
on three counts:

- **The payload is an id.** `notifyActionUpdate` sends `project/domain/run/name` and nothing
  else (`action.go:1120`). A subscriber has to read the table anyway.
- **It is deduped per action.** A batch of N events for one action produces one notification
  (`action.go:160`). The signal does not tell you how many events to go fetch.
- **Most of it is not events.** Six of the seven `notifyActionUpdate` call sites write no
  `action_events` row at all — `CreateAction`, `UpdateActionPhase`, `AbortAction`, `AbortRun`,
  `UpdateActionState`, `NotifyStateUpdate`. `action_test.go:1008` asserts the phase update must
  *not* synthesize one. A publisher on that channel would mostly wake up and find nothing.

**Hand off, do not send inline.** `InsertEvents` sits on the executor's write path; a broker
round trip must not extend a commit. The publisher takes the batch onto its own queue and the
caller returns.

## The crash window, and what it costs

Handing off in-process means an event can be committed and never published: the queue is
memory, and a restart between the commit and the drain loses whatever it held. Closing that is
the only reason the publisher ever reads `action_events` back — not because the notification
lacks content, but because nothing else survives the restart.

That reconcile needs a cursor, and **the table cannot currently support one**:

```sql
-- runs/migrations/sql/20260408110000_init_schema.sql:56
PRIMARY KEY (project, domain, run_name, name, attempt, phase, version)
-- no serial column; created_at is unindexed
```

The composite key is not globally monotonic, so there is no "everything after X" to scan.
`created_at` cannot stand in: it defaults to `CURRENT_TIMESTAMP`, which is transaction start
time, so a long transaction commits rows *behind* a high-water mark already advanced past them
and the reconcile skips them silently.

So this design has a prerequisite: either a monotonic `BIGSERIAL` column on `action_events`
with an index, or a separate outbox table written in the same transaction as the INSERT — which
`InsertEvents` would have to become, since it commits per chunk today (`action.go:126`). The
outbox costs a write per event and a reaper; the serial column costs a migration on a hot table
and still leaves a small ordering gap between sequence assignment and commit visibility, which
a "re-scan the last N seconds" overlap covers. Pick one before implementing — the rest of the
publisher design does not depend on which.

## What carries over from v1

v1 split the feature along a seam worth keeping: a **publisher** that turns a Flyte event into
a CloudEvent, and a **sender** that puts a CloudEvent on a wire. The publisher knows the Flyte
domain and nothing about brokers; the sender knows brokers and nothing about Flyte.

```
+--------------------------------------+
| PUBLISHER                            |
|   knows Flyte events                 |
|   decides what an event looks like   |
|   on the wire: type, id, time, data  |
+------------------+-------------------+
                   |
                   v
+--------------------------------------+
|          one narrow interface        |  <-- the seam
+------------------+-------------------+
                   |
      +------------+------------+------------+
      v            v            v            v
+-----------+ +---------+ +-----------+ +---------+
| Kafka     | | NATS    | | cloud     | | no-op   |
| sender    | | sender  | | pub/sub   | |         |
+-----+-----+ +----+----+ +-----+-----+ +---------+
      v            v            v
+-----------+ +---------+ +-----------+
| operator's brokers -- not shipped or run by Flyte  |
+---------------------------------------------------+
```

**The sender half transfers almost unchanged.** Four transports, one interface, and the
envelope conventions that go with them — a stable id that doubles as the consumer's
deduplication key, a payload encoding that survives protobuf `oneof` fields, a schema
reference that lets a consumer validate without asking Flyte. None of that depends on the
Flyte data model, and it was working in production for years.

**The publisher half does not transfer.** v1's is written against workflow, node and task
executions. v2 has runs and actions. Mapping one onto the other is the substantive design work
here, and it is what #7829 asks to be documented.

## Turning it on

Config only, resolved once at startup. No per-run parameter, no launch-time opt-in, nothing in
the SDK surface: an operator enables egress for a deployment, and every event flows.

This mirrors v1, and the shape is worth repeating for two reasons. Operators already know it.
And the alternative — letting individual runs choose — makes the event stream unreliable as a
source of truth, because a consumer can no longer assume that silence means nothing happened.

Two properties the config gate should preserve:

- **Off by default.** A deployment that says nothing about egress publishes nothing.
- **Filterable.** An operator who wants only terminal events should be able to say so without
  filtering client-side, because the cost of the events they do not want is paid on the wire.
