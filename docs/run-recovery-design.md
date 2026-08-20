# Run Recovery — design

Status: proposal
Scope: `actions/`, `executor/`, `runs/`, plus the client surface in `flyte-sdk`

## What recovery is

Recovery creates a **new run that reuses a prior run's successful actions and re-executes only
what did not succeed**.
It runs the source run's code unchanged: recovery provides durability against intermittent
system- and network-level failures, and is not a way to patch a run and resume it.

Its only effect on the wire is to set `RunSpec.relation = Relation{related_to: <RunIdentifier>,
relation_type: RELATION_TYPE_RECOVER}` (plus an optional `RunSpec.recover`) on
`CreateRunRequest`. The IDL is already generated into this repo
(`gen/go/flyteidl2/common/run.pb.go`, `gen/go/flyteidl2/task/run.pb.go`).

**Nothing in this repo reads that field today.** A `recover=` run is silently an ordinary run:
the SDK's feature gate only checks that `RunSpec.relation` exists in the linked `flyteidl2`
build, which it does, so no error surfaces. Closing that gap is what this document covers.

## Concept

The backend's whole job is: when a sub-action is enqueued on a recovery run and the source
run already has a successful action of the same name, record it as terminal with the source's
output URI and never execute it. Everything else is machinery in service of that.

## Data flow

**1 — Run creation. The only enqueue that carries a RunSpec.**

```
   ┌──────────────┐
   │  SDK client  │
   └──────┬───────┘
          │  CreateRun { RunSpec.relation = RECOVER → r1 }
          ▼
   ┌──────────────────┐
   │   runs service   │   validate: source run exists, is terminal,
   └──────┬───────────┘             same project/domain
          │  Enqueue(root action, runSpec)
          ▼
   ┌──────────────────┐
   │ actions service  │   RunSpec.relation + .recover
   └──────┬───────────┘        └─► Spec.RecoveryContext          (D1)
          │  Create
          ▼
   ┌──────────────────────┐
   │ etcd: root TaskAction│
   └──────────────────────┘
```

**2 — One sub-action enqueue. No RunSpec on this path; the context is inherited.**

```
   ┌────────────────────────┐
   │ SDK controller         │  runs inside the parent's task pod
   └──────────┬─────────────┘
              │  Enqueue(action)          ← carries no RunSpec
              ▼
   ┌────────────────────────────────────────────────────────┐
   │ actions service                                        │
   │                                                        │
   │   1. Get parent TaskAction  ──►  RecoveryContext   (D1) │
   │   2. gate ladder                                       │
   │   3. still a candidate?  look the source action up     │
   └───────┬────────────────────────────────┬───────────────┘
           │ LookupAction(                  │  Create TaskAction
           │   source run, action name)     │
           ▼                                ▼
   ┌──────────────────┐         ┌───────────────────────────────┐
   │   runs service   │         │ etcd: TaskAction CR           │
   │   └─ Postgres    │   (D2)  │   hit  → Spec.RecoveredFrom   │
   └──────────────────┘         │   miss → exactly as today     │
                                └───────────────────────────────┘
```


## SDK changes

The wire format needs nothing new — the SDK already sends `RunSpec.relation` and
`RunSpec.recover`, and already treats `ACTION_PHASE_RECOVERED` as terminal. What changes is the
surface: recovery becomes its own command with narrower semantics than the flag it replaces.
Paths below are in the `flyte-sdk` repo.

**S1 — `flyte recover <run>` as a first-class command.** New `src/flyte/cli/_recover.py`,
registered in `cli/main.py` (import, the command-group list, `add_command`). Surface:

```
flyte recover <run> [-p/-d] [--name NAME] [--follow] [--force-rerun-action NAME]...
```

It calls the existing no-substitution rerun path — `with_runcontext(mode="remote",
recover=True, ...).rerun(run)` — which fetches the source run's task spec from the platform
rather than bundling local code. That is what makes the recovery run reuse the source's
`code_bundle_uri` (carried on `TaskMetadata`) and container image verbatim, with no repackaging
step; the guarantee falls out of the path chosen, so it needs no new machinery. Deliberately
**not** offered in v1: `--env`, `--label`, and every other `RunSpec` override. They are easy to
add later and impossible to withdraw.

**S2 — delete `flyte rerun --recover`.** `cli/_rerun.py:46-59` (both options), `:98-99` (the
`--force-rerun-action requires --recover` guard, which is vacuous once recovery is its own
command), `:129` (the passthrough). No deprecation alias: nothing in this repo has ever read
`RunSpec.relation`, so the flag has never done anything and has no users to migrate. The twelve
recover assertions in `tests/cli/test_rerun.py` move to a new `tests/cli/test_recover.py`.

**S3 — reject substituted code.** `Runner.rerun(task_template=...)` composes with `recover=True`
today (`_run.py:1406-1412`; `_resolve_recover_ref` returns the source name regardless), which is
exactly the combination the new semantics forbid — new code against reused actions. Guard it in
`Runner.rerun`.

**S4 — no escape hatch for missing source outputs.** `--allow-missing-outputs` must not appear on
`flyte recover`, and `allow_missing_source_outputs` must be rejected alongside `recover` in the
runner. The flag exists because a plain rerun needs only the source run's *inputs*, so vanished
outputs are harmless; under recovery the source outputs **are** the deliverable, and proceeding
guarantees that every recovered action fails when its output is consumed. Consequences in
`_run.py:1440-1487`: on the recovery path the `GetActionData` 404 is unconditionally fatal, so
the `if "inputs" in str(e.message)` split — which the code itself notes is a race — never has to
be resolved there; and the two error messages that point users at `flyte run ... --recover-from`
(`:1457`, `:1470`) must stop doing so.

This has a payoff for the backend: a recovery run reaches `CreateRun` only if the source run's
inputs and outputs were both readable moments earlier. Not a guarantee they survive until a
recovered output is consumed — nothing checks that, and nothing should — but the known-missing
case is off the table.

**S5 — chained recovery keeps pointing at the immediate parent.** `_apply_overrides` clears
inherited provenance and sets `related_to` to the run being recovered (`_run.py:620-625`), so
recovering a recovery run points at the previous recovery, not at the original. Keep it: each hop
already carries a fully-resolved output URI, so flattening the chain client-side would buy
nothing. This is the main path rather than an edge case — durability means a run that failed on
an intermittent fault gets recovered, and if it fails again it gets recovered again — and it is
why the lookup must accept a source action whose phase is `RECOVERED` (gate 5 below).

**S6 — open: the client cannot tell whether the backend implements recovery.** The only gate
today checks that `RunSpec.relation` exists in the linked `flyteidl2` build (`_run.py:660-665`),
which says nothing about the server. Against a backend without this feature, `flyte recover`
silently degrades into a full rerun — the worst failure mode available, because it looks like it
worked. A flag that never did anything could get away with that; a command named `recover`
cannot. Needs a real signal, and the cheapest one is on this side of the wire: have `CreateRun`
reject `RELATION_TYPE_RECOVER` until it is implemented, rather than accepting and ignoring it.

**Not changing, but load-bearing.** Two existing SDK behaviours the backend design depends on;
both are contracts now, and breaking either silently breaks recovery:

- `ACTION_PHASE_RECOVERED` is terminal and success-equivalent in `ActionPhase`
  (`src/flyte/models.py:779-800`) and in the in-task controller's `Action.is_terminal`
  (`src/flyte/_internal/controllers/remote/_action.py:63-75`), which consumes the update's
  `output_uri` as the action's realized output. D3 states the obligations this puts on the
  executor.
- The exclusions in `generate_task_identity_hash` (`_internal/runtime/convert.py:793-812`) exist
  specifically so recovery can match; ENG26-831 closed the three known name instabilities
  (`tests/flyte/internal/runtime/test_action_name_stability_gaps.py`), leaving sets of unsortable
  elements as a documented limitation. Any future change to the hash's inputs is an
  API-compatibility change for recovery.

## Constraints this repo imposes

Three properties of the current architecture shape the whole design.

**The actions service has no database.** Its state store is the `TaskAction` CR in etcd
(`actions/k8s/client.go`). Anything recovery needs to know about a *past* run has to come from
somewhere else.

**Terminal `TaskAction` CRs are short-lived.** They are garbage collected on a timer — default
`Interval: 30m`, `MaxTTL: 1h` (`executor/pkg/config/config.go:30-33`) — and OwnerReferences
cascade the deletion from the root down. The CRs are a window of about an hour, not a history.

**`TaskAction` status is a subresource** (`executor/api/v1/taskaction_types.go:351`), so
`Create` drops it. Producing an action that is terminal from birth takes two writes, and the
reconciler fires between them.

The three decisions below each answer one of these.

---

## D1 — Recovery context propagates through the TaskAction spec

**Problem.** The recovery decision runs per sub-action at enqueue time, so every enqueue must
first answer "is this a recovery run?". The answer lives in the RunSpec — which sub-action
enqueues do not have.

`ActionsClient.Enqueue(ctx, action, runSpec)` has two callers, and only one of them passes a
RunSpec:

- **root** — `runs/service/run_service.go:366`, on the `CreateRun` path, passes it.
- **sub-actions** — the SDK controller, over Connect from inside the task pod. Its
  `EnqueueRequest` carries only `action`; there is no `run_spec` field set on it.

So by the time a sub-action reaches `Enqueue`, the relation exists but is unreachable. Something
has to carry it down the tree.

**Decision.** Stamp the recovery context onto the root `TaskAction` spec and let the existing
inheritance carry it to children. `Enqueue` already fetches the parent TaskAction to wire the
OwnerReference:

```go
// actions/k8s/client.go:143-148
parentTaskAction, err := c.setParentOwnership(ctx, taskAction, actionID.Run, *action.ParentActionName)
…
inheritRunContextFromParentTaskAction(taskAction, parentTaskAction)
```

`setParentOwnership` (`client.go:946-969`) performs the `Get` on the parent regardless — it
needs the parent's UID for the OwnerReference — so reading one more field off the returned
object costs nothing. `applyRunSpecToTaskAction` (`client.go:1040`) already projects RunSpec
fields onto the root, and `inheritRunContextFromParentTaskAction` (`client.go:1071`) already
copies context to children; this extends both.

**What the field carries.** Not the relation alone. Gate 3 needs
`RunSpec.recover.force_rerun_actions`, which is a *sibling* of `RunSpec.relation`, not part of
it — `RunSpec` has separate `GetRelation()` and `GetRecover()` accessors
(`gen/go/flyteidl2/task/run.pb.go:694`, `:701`), and the `Recover` message deliberately carries
no run reference of its own. A field holding only the serialized `Relation` would leave gate 3
unimplementable.

```go
// executor/api/v1/taskaction_types.go — TaskActionSpec
// RecoveryContext carries the parts of RunSpec that the enqueue-time recovery
// decision needs. Stamped on the root and inherited by every descendant,
// because RunSpec itself reaches only the root enqueue.
// +optional
RecoveryContext *RecoveryContext `json:"recoveryContext,omitempty"`

type RecoveryContext struct {
    // Relation is the proto-serialized common.Relation.
    Relation []byte `json:"relation,omitempty"`
    // ForceRerunActions mirrors RunSpec.recover.force_rerun_actions.
    ForceRerunActions []string `json:"forceRerunActions,omitempty"`
}
```

Serializing the whole RunSpec instead would future-proof this against needing more of it later,
but RunSpec carries env vars, labels and annotations, and the copy would land on *every*
descendant CR. The controller already tracks CR size as a metric (`observeCRDSize`,
`executor/pkg/controller/taskaction_controller.go:292`), so inflating every CR for fields
nothing reads is the wrong trade. Carry the two fields; widen it if a third is ever needed.

**Condition actions need this too.** `Enqueue` has two branches, and the `Action_Condition`
branch (`client.go:181-186`) calls neither `applyRunSpecToTaskAction` nor
`inheritRunContextFromParentTaskAction` — it discards `setParentOwnership`'s return value
entirely. As written, condition actions inherit no run context at all, so a subtree beneath a
condition would silently lose the recovery context and never recover. Either capture the return
value and inherit there too, or factor the root/child split into one helper both branches use.
The second is preferable: that logic should not exist twice.

**Consequence.** Non-recovery runs pay exactly zero additional cost, and "is this a recovery
run?" is answered from the object being constructed rather than from remote state.

## D2 — The runs service Postgres is the history; the CR is not consulted

**Decision.** Look the source action up in the runs service `actions` table. Do not read the
source run's `TaskAction` CRs at all.

Two reasons, in order of importance.

**The CR does not carry the output URI.** `TaskActionStatus`
(`executor/api/v1/taskaction_types.go:266-348`) has `StateJSON` (documented as "for
observability"), `PluginState`, `CacheStatus`, `Conditions`, `PhaseHistory`, `ErrorState`,
`SignalValue` — and no output reference. The URI is *computed* when a status event is emitted
and never stored on the object:

```go
// executor/pkg/controller/taskaction_controller.go:715-728
prefix, err := plugin.ComputeActionOutputPath(ctx, taskAction.Namespace, taskAction.Name,
    taskAction.Spec.RunOutputBase, taskAction.Spec.ActionName, attempt)
return &task.OutputReferences{OutputUri: base + "/outputs.pb", …}
```

A CR-first lookup would therefore *hit* while returning nothing usable — worse than a miss,
because it looks like a successful reuse and hands the caller an empty URI.

**The CR is short-lived**, per the constraints above. An hour-wide window is not a history to
recover from.

**What Postgres has.** The `actions` table
(`runs/migrations/sql/20260408110000_init_schema.sql:14-49`) is keyed exactly the way this
lookup wants:

```sql
PRIMARY KEY (project, domain, run_name, name)
```

so the lookup is a single primary-key point-read. The table also carries `phase`, `attempts`,
`cache_status`, and `action_details`. Every action — including nested sub-actions — reports an
`ActionEvent` carrying `Outputs`, `Attempt`, `Phase` and `CacheStatus`
(`executor/pkg/controller/taskaction_controller.go:671-688`), so the data is there for
sub-actions, not just roots.

**How the actions service reaches it.** It does not need a database connection of its own.
`NewActionsClient` already takes a `workflowconnect.InternalRunServiceClient`
(`actions/k8s/client.go:94`), so the lookup is a new RPC on an existing client, not new
infrastructure.

**Cost.** One point-read per sub-action, on recovery runs only, on a path that already performs
a `TaskAction` CR create against the K8s API. The read is noise next to the write it
accompanies.

Loading the whole source run into memory on the first lookup is possible — `idx_actions_run_lookup`
supports it — but is **out of scope for the first version**. It adds an invalidation boundary
and a memory bound to reason about, in exchange for an optimisation nobody has measured a need
for. Revisit with a profile.

## D3 — Recovered actions are marked in the spec, not the status

**Problem.** For a recovered action to never execute, the executor must see "already done"
before it does anything. But status is a subresource: `Create` drops it, a second
`Status().Update()` is required, and the reconciler fires on the create event in between — so
it starts running the task.

The reconciler's terminal short-circuit does not help, because at that moment the status is
still empty:

```go
// executor/pkg/controller/taskaction_controller.go:304
if isTerminal(taskAction) { … return ctrl.Result{}, nil }
```

**Decision.** Carry the recovery outcome in `TaskActionSpec`, which survives `Create`:

```go
// executor/api/v1/taskaction_types.go
type RecoveredFrom struct {
    SourceRunName string
    OutputUri     string
    Phase         common.ActionPhase
    CacheStatus   core.CatalogCacheStatus
    Attempt       uint32
}
```

and short-circuit in `Reconcile` **before** any plugin dispatch, writing the terminal status
from the spec. No pod is ever created.

**The status this writes is not the status the normal path computes.** Two fields must come from
`RecoveredFrom` rather than from the usual derivation, and both are load-bearing for the client
(see "What the client sends, and what it expects back"):

- **Phase** is `ACTION_PHASE_RECOVERED`, not `SUCCEEDED`.
- **Output URI** is `RecoveredFrom.OutputUri`. The normal path calls `outputRefs`
  (`executor/pkg/controller/taskaction_controller.go:715-728`), which derives the URI from *this*
  run's `RunOutputBase` and action name — for an action that never executed, that points at a
  location nothing ever wrote.

Placing it in the spec rather than the status is not a style choice — it is the only field a
single `Create` can carry, and a single `Create` is what makes "never executes" true rather
than "usually doesn't execute".

---

## The recovery decision

Evaluated per sub-action at enqueue, in order. Any gate that does not pass falls through to the
normal fresh path — **fail-open, never fail the enqueue**. Recovery is an optimisation; a
recovery that cannot happen must degrade into an ordinary execution, not an error.

| # | gate | notes |
|---|---|---|
| 1 | is this the root action? | The root always runs fresh |
| 2 | is `relation_type == RELATION_TYPE_RECOVER` with a non-nil `related_to`? | Read from the inherited spec (D1). `rerun` shares the field with a different type — the type is what makes it a recovery. |
| 3 | is the action name in `RunSpec.recover.force_rerun_actions`? | Escape hatch. Evaluated **before** any lookup. Listing a parent forces only that parent; children are decided independently, so forcing a subtree means listing all of it. |
| 4 | do project/domain match the source run? | Scope gate. The lookup is keyed by run identity, so a cross-scope relation would read another tenant's rows and leak output URIs. `CreateRun` validation should make this unreachable; keep it as defence in depth. |
| 5 | is the source action `SUCCEEDED` or `RECOVERED`? | Evaluated after the lookup, and the only gate that can fail on a run that is otherwise a valid recovery. `RECOVERED` counts — chained recovery is the common path, and its URI is already fully resolved. Every other phase, and a missing row, is a miss. |

On all five passing: populate `Spec.RecoveredFrom` and create the CR (D3). Otherwise create the
CR as it is created today.

**Matching is action-name string equality and nothing else.** No hash comparison, no
task-version check, no spec diff. "Changed" is an emergent property of the SDK's action name,
which composes the parent action name, an inputs hash, a task identity hash and the call sequence
(`flyte-sdk: src/flyte/models.py:71-78`,
`src/flyte/_internal/runtime/convert.py:838-851`). Two consequences worth stating because
neither is visible at runtime:

- The task identity hash covers the task's fully-qualified name, its interface and its
  `discovery_version` (the function-body AST hash). It **deliberately excludes** the container
  image, code-bundle version, resources, env vars and plugin config
  (`convert.py:793-812`) — the SDK maintains that exclusion specifically so recovery can match.
  So a name shifts only when that task's own body or interface changed, or its inputs changed —
  plus everything beneath it, since the parent name is a name component. Sibling subtrees are
  unaffected, and editing the root task's body does not shift its children, whose parent
  component is the fixed name `a0`. Under `flyte recover`, which pins the source run's code
  bundle and does not accept new inputs, no shift should occur at all: a miss means the source
  action genuinely never succeeded.
- A mismatch is **miss-only** — a spurious re-run, never a wrong result silently substituted.

## Observability

Recovery is fail-open and silent: a run that recovers zero actions is indistinguishable from
one that recovers everything, from the outside. Metrics are the only signal and must ship with
the feature, not after it:

- `recovery_recovered_total{type}`
- `recovery_rerun_total{type,reason}` — reasons: `forced`, `missing`, `non_final`, `scope`
- `recovery_lookup_errors_total`

Provenance should also surface on the action so the UI can render "recovered from &lt;run&gt;",
via `ActionMetadata.relation` / a `recovered_from` field.

## Work items

| area | change |
|---|---|
| `runs/service/run_service.go` | Validate the relation on `CreateRun`: `related_to` non-empty, same project/domain, source run exists and is terminal. Reject `RunSpec.recover` when the type is not `RECOVER`. Persist the relation. Until the rest of this lands, reject `RELATION_TYPE_RECOVER` outright rather than accepting and ignoring it (see S6). |
| `runs/` | Add an `InternalRunService` RPC looking an action up by `(project, domain, run_name, name)`, returning phase, attempts, cache status and the output URI, plus the repository query behind it. The output URI is not a column on `actions`: a trace action carries it in `detailed_info.outputs_uri`, a task action in the last attempt's `action_events.info`, and a `RECOVERED` action has no attempt rows at all — it must be readable for all three. |
| `executor/api/v1/taskaction_types.go` | Add `RecoveryContext` and `RecoveredFrom` to `TaskActionSpec`; regenerate CRDs. |
| `executor/pkg/controller/taskaction_controller.go` | Short-circuit on `Spec.RecoveredFrom` ahead of plugin dispatch; write the terminal status with phase `RECOVERED` and the source output URI, bypassing `outputRefs` (see D3). |
| `actions/k8s/client.go` | Propagate `RecoveryContext` in `applyRunSpecToTaskAction` and `inheritRunContextFromParentTaskAction`; make the `Action_Condition` branch inherit it too (see D1); run the gate ladder in `Enqueue`; populate `RecoveredFrom` via the `InternalRunServiceClient` the struct already holds. |
| metrics | The three counters above. |
| `flyte-sdk` | S1–S4: add `flyte recover`, delete `flyte rerun --recover`, reject substituted code and `allow_missing_source_outputs` under recovery. Independent of everything above — the flag it replaces has never done anything, so the command can ship against a backend that still rejects `RELATION_TYPE_RECOVER`. |

Suggested order: D1 first — it is independently verifiable (stamp the relation, assert it
appears on every child CR) and unblocks everything else.
