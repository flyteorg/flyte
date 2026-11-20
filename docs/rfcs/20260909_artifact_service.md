# [RFC] Artifact Service — typed, versioned, addressable assets for Flyte 2

**Authors:**

- Alex Wu

## 1 Executive Summary

Flyte 2 ships a complete artifact IDL — `flyteidl2/artifact/artifact.proto`
and `flyteidl2/artifact/artifact_service.proto`, with generated clients for
Go, TypeScript, Python and Rust — and the Python SDK already calls it
(`flyte.remote.Artifact`, `flyte create artifact`, `flyte get artifact`).
**No server implements it.** Every call fails `Unimplemented`.

This RFC proposes implementing `ArtifactService` so a `File`, `Dir` or
`DataFrame` can be published under a name and version, retrieved by that
identity from another run, and carry its own description, metadata, card and
provenance.

It is staged. **Stage 1, the subject of this document, is the service
itself**: the five RPCs and their storage. Later stages add the automatic
registration of task-produced artifacts and artifact-driven triggers.

## 2 Motivation

Blob storage already holds the bytes. What is missing is everything above it:

- **No identity.** A model produced by a run is a URI. Nothing names it
  `sentiment-model`, versions it, or lets the next run ask for `@v3`.
- **No provenance.** Given a file, there is no record of which run, action
  and attempt produced it, or which artifact versions it was derived from.
- **No discovery.** Listing "every model in this project" means listing blob
  prefixes and guessing.
- **No typed reuse.** `flyte.run(main, x=Artifact.get("weights"))` is already
  supported by the SDK — it coerces the stored literal to the task's declared
  input type — but there is nothing to fetch from.

The SDK surface for all of this exists and is tested. Implementing the
backend service is the missing step.

## 3 High-level design

The bytes never pass through the artifact service. Offloaded values are
uploaded to blob storage by the SDK's type engine; a `core.Literal` referencing
them by URI is what gets stored. An artifact is a **named, versioned reference**
to data that already exists.

```
                              ┌───────────────────────────┐
 ┌───────────────────────────►│        Blob storage       │  the bytes live here
 │                            └─────────────┬─────────────┘
 │  File.from_local() uploads               │ uri, carried inside core.Literal
 │  the bytes for both (A) and (B)          ▼
 │
 │   SDK                                                 Backend
 │  ┌────────────────────────────┐                       ┌──────────────────────┐
 ├──┤ (A) Artifact.create(...)   │── CreateArtifact ────►│                      │
 │  │     flyte create artifact  │                       │   ArtifactService    │
 │  ├────────────────────────────┤                       │                      │
 └──┤ (B) @task(produces_        │                       │ ┌────────────────┐   │
    │       artifacts=True)      │                       │ │ artifacts      │   │
    │     return artifacts.new(  │                       │ │ one row per    │   │
    │         file, metadata)    │                       │ │ name@version   │   │
    └─────────────┬──────────────┘                       │ └────────────────┘   │
                  │ outputs.pb, carrying                 │                      │
                  │ a ProducedArtifact decl.             │                      │
                  ▼                                      │                      │
         ┌───────────────────┐                           │                      │
         │  Actions service  │── CreateArtifact ────────►│                      │
         │  extract+register │                           │                      │
         └───────────────────┘                           │                      │
                                                         │                      │
    ┌────────────────────────────┐                       │                      │
    │ (C) Artifact.get           │◄── GetArtifact ───────│                      │
    │     Artifact.listall       │◄── ListArtifacts ─────│                      │
    │     Artifact.list_names    │◄── ListArtifactNames ─│                      │
    └────────────────────────────┘                       └──────────────────────┘
```

### Write path A — direct publication

The caller builds the artifact and publishes it:

```python
await Artifact.create(file, name="sentiment-model", kind="model")
```

The SDK converts the value with the type engine (uploading local data to blob
storage first), then calls `CreateArtifact`. Publishing from inside a running
task stamps `ArtifactSource.task_action` automatically; publishing from a
laptop leaves the source unset, or sets `external_ref` for an import.

### Write path B — declared task output

The task marks an output as an artifact:

```python
@env.task(produces_artifacts=True)
async def train() -> File:
    file = await File.from_local("weights.pt")
    return artifacts.new(file, artifacts.Metadata(name="sentiment-model"))
```

The SDK does **not** call the service here. It attaches a
`task.ProducedArtifact` declaration to the `Outputs` envelope written to
`outputs.pb`. On a successful terminal action whose task metadata sets
`produces_artifacts`, the actions service reads that envelope, pairs each
declaration with its output literal, and registers it via the same
`CreateArtifact` RPC — stamping the producing action as provenance and
defaulting the version from the action's identity.

Registration is deliberately the backend's job, not the task's: an artifact
then exists only if the action that produced it succeeded, and its provenance
is asserted by the control plane rather than self-reported.

### Read path

`GetArtifact` by name and version (or `latest`), `ListArtifacts` for the
versions of a name, `ListArtifactNames` to browse names. Retrieved artifacts
bind directly as run inputs; coercion to the declared parameter type happens
SDK-side and needs nothing from the server.

## 4 Stage 1 — the Artifact API

### 4.1 RPCs

All five are already defined in `artifact_service.proto`; no IDL change is
proposed.

| RPC | Behavior |
|---|---|
| `CreateArtifact` | Insert one version. Immutable; a duplicate identity returns `AlreadyExists`. |
| `GetArtifact` | Fetch by `ArtifactName` + version. Unset version or `"latest"` resolves to the newest by `created_at`. |
| `ListArtifacts` | Versions within a project, newest first. Optional exact `name`; filters `name CONTAINS`, `created_at GREATER_THAN`, and the `parent_name`+`parent_version` pair. Paginated. |
| `ListArtifactNames` | Distinct names, each with its latest version and total version count, ordered by the latest version's `created_at`. Filter `name CONTAINS`. Paginated. |
| `ListArtifactMetadataKeys` | Distinct `user_metadata` keys in a project, for filter suggestions. Keys only, sorted and capped. |

There is no update and no delete: an artifact version is immutable, and the
SDK's `Artifact.delete()` raises `NotImplementedError` today.

Server responsibilities on create, beyond persistence:

1. **Stamp identity onto the value.** Set `spec.value.artifact_id` to the
   final `{org, project, domain, name}` + version before writing. This is the
   only point where the full identity is known, and it travels with the
   literal into downstream run inputs and cache keys — two versions of an
   artifact are distinct inputs by design.
2. **Stamp `org`.** OSS has no organization; clients leave it empty and the
   server normalizes to the existing placeholder
   `secret.DefaultOrganization = "flyte"`, as the secret and app services do.
3. **Stamp `created_by`.** Persist the caller's OIDC subject only, matching
   the `actions.created_by_subject` convention. Stage 1 serves the
   subject back as-is; identity enrichment is additive later.
4. **Reject** an unset `spec.value` or `spec.type`, and duplicate entries in
   `parent_artifacts`.

### 4.2 Validation limits

Every bound is already declared in the IDL as a `buf.validate` constraint, so
the service does not invent limits — it enforces the generated ones by calling
`req.Msg.Validate()` at the top of each RPC, as `ProjectService`,
`TaskService` and `RunService` do.

| Field | Max | Declared in |
|---|---:|---|
| `artifact_id.name.org` | 63 | `artifact.ArtifactName.org` |
| `artifact_id.name.project` | 64 | `artifact.ArtifactName.project` (min 1) |
| `artifact_id.name.domain` | 64 | `artifact.ArtifactName.domain` (min 1) |
| `artifact_id.name.name` | 255 | `artifact.ArtifactName.name` (min 1) |
| `artifact_id.version` | 255 | `artifact.ArtifactIdentifier.version` (min 1) |
| `spec.info.description` | 255 | `core.ArtifactInfo.description` |
| `spec.info.card.uri` | 1024 | `core.ArtifactCard.uri` |
| `spec.info.card.format` / `.type` | 16 | `core.ArtifactCard.format` / `.type` |
| `spec.source.external_ref` | 1024 | `artifact.ArtifactSource.external_ref` |
| `spec.source.task_action.action` run/action names | 30 | `common.RunIdentifier.name`, `common.ActionIdentifier.name` |
| `spec.parent_artifacts` | 32 entries | `artifact.ArtifactSpec.parent_artifacts` |

Two consequences for the implementation:

- **Column sizing.** Identity columns are bounded and short; the storage above
  uses `TEXT` in line with the existing tables, but the bounds are what a
  `VARCHAR(n)` would use if a future migration wants them enforced in SQL too.
- **A scope mismatch to be aware of.** `ArtifactName` allows 64-character
  project and domain, while `common.RunIdentifier` caps both at 63. An
  artifact produced by a task therefore always fits its own bounds, but a
  manually published artifact may carry a project or domain one character
  longer than any run can. Validate against the artifact's own bounds and do
  not assume the two are interchangeable.

`buf.validate` cannot express the rest, so `CreateArtifact` still checks
explicitly: `spec.value` and `spec.type` present, and no duplicate entry in
`parent_artifacts` after scope and name inheritance is applied.

### 4.3 Data model

Placement follows the existing services: model in
`runs/repository/models/artifact.go`, sqlx implementation in
`runs/repository/impl/artifact.go`, service in
`runs/service/artifact_service.go`, mounted in `runs/setup.go` alongside
`SettingsService` and `ProjectService`, migration in `runs/migrations/sql/`.

```sql
CREATE TABLE IF NOT EXISTS artifacts (
    org              TEXT        NOT NULL,
    project          TEXT        NOT NULL,
    domain           TEXT        NOT NULL,
    name             TEXT        NOT NULL,
    version          TEXT        NOT NULL,

    literal_value    BYTEA       NOT NULL,   -- core.Literal
    literal_type     BYTEA       NOT NULL,   -- core.LiteralType
    description      TEXT        NOT NULL DEFAULT '',
    user_metadata    JSONB       NOT NULL DEFAULT '{}',
    card_uri         TEXT,
    card_format      TEXT,
    card_type        TEXT,
    source           BYTEA,                  -- artifact.ArtifactSource
    parent_artifacts BYTEA,                  -- repeated core.ArtifactVersionId

    created_by_subject TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (org, project, domain, name, version)
);

-- versions of one name, and `latest` resolution
CREATE INDEX IF NOT EXISTS idx_artifacts_name_created
    ON artifacts (org, project, domain, name, created_at DESC);

-- project-wide listing
CREATE INDEX IF NOT EXISTS idx_artifacts_project_created
    ON artifacts (org, project, domain, created_at DESC);
```

Notes on the shape:

- **Protos as bytes, queryable fields as columns.** `literal_value` and
  `literal_type` are opaque to SQL and stored serialized, the same way
  `actions.action_spec` is. Anything a filter or a response header needs —
  identity, timestamps, description, metadata, card — is a column.
- **`user_metadata` as JSONB**, so `ListArtifactMetadataKeys` is a
  `jsonb_object_keys` scan and future metadata filters are containment
  queries rather than a decode of every row.
- **Card flattened into three columns** rather than a nested blob: the IDL
  caps each at 1024/16/16 characters and listings display them.
- **The primary key is the identity**, which gives immutability and
  `AlreadyExists` for free via a unique-violation check.

**Lineage.** `parent_artifacts` is stored as given and never resolved —
referenced versions need not exist. The stored pointers only walk *upwards*,
so the `parent_name`+`parent_version` filter (finding a version's children)
needs a side table:

```sql
CREATE TABLE IF NOT EXISTS artifact_parents (
    -- the child version
    org            TEXT    NOT NULL,
    project        TEXT    NOT NULL,
    domain         TEXT    NOT NULL,
    name           TEXT    NOT NULL,
    version        TEXT    NOT NULL,
    -- one declared parent of it
    parent_name    TEXT    NOT NULL,
    parent_version TEXT    NOT NULL,
    ordinal        INTEGER NOT NULL,   -- 0 = primary parent, later = merged in

    PRIMARY KEY (org, project, domain, name, version, parent_name, parent_version),
    FOREIGN KEY (org, project, domain, name, version)
        REFERENCES artifacts (org, project, domain, name, version) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_artifact_parents_lookup
    ON artifact_parents (org, project, domain, parent_name, parent_version);
```

Rows are written in the same transaction as the artifact, with empty scope
fields resolved to the child's own scope and an empty name to the child's own
name, per the IDL. If lineage slips out of Stage 1, the column stays and only
the side table and its filter are deferred.

### 4.4 Out of scope for Stage 1

- Extraction of `ProducedArtifact` declarations from task outputs (Stage 2).
- Artifact-driven triggers — `flyte.OnArtifact` (Stage 3).
- Identity enrichment of `created_by` beyond the raw subject.
- Deletion and retention.

### Stage 2 — task-produced artifacts

A task marks an output as an artifact and declares it on the environment:

```python
@env.task(produces_artifacts=True)
async def train() -> File:
    file = await File.from_local("weights.pt")
    return artifacts.new(file, artifacts.Metadata(name="sentiment-model"))
```

The SDK does not call the service. `artifacts.new` attaches metadata that
`convert_from_native_to_outputs` turns into a `task.ProducedArtifact`
declaration on the `Outputs` envelope; `produces_artifacts` rides along on
`TaskMetadata`. Both land in `outputs.pb` in blob storage. Registration is the
backend's job, so an artifact exists only if the action that produced it
succeeded, and its provenance is asserted rather than self-reported.

#### Where it runs

In the executor's `TaskActionReconciler`, at the point it takes the TaskAction
CR terminal. Everything the step needs is already there:

- `r.DataStore` — and it already reads `outputs.pb` for cache writeback
  (`executor/pkg/controller/taskaction_cache.go`), so a `Discoverable` task can
  share the read;
- the serialized `TaskTemplate` on the CR spec, carrying `produces_artifacts`;
- the output prefix it computes itself;
- controller-runtime, which supplies requeue-with-backoff for free and makes
  the CR itself the durable place to record completion.

That last point is what keeps this cheap: **no scanning, no queue, and no new
database column.** The completion marker is a CR annotation, and a failed
registration is a returned error the framework retries.

#### Flow

```
 task pod
   │ File.from_local()             ┌────────────────────────────┐
   ├──────────────────────────────►│        Blob storage        │
   │ upload_outputs()              │                            │
   └──────────────────────────────►│  <prefix>/outputs.pb       │
                                   │     ├ literals[]           │
                                   │     └ produced_artifacts[] │
                                   └─────────────┬──────────────┘
                                                 │ ① read, gated
 executor · TaskActionReconciler                 │
 ┌───────────────────────────────────────────────┴────────────────────┐
 │ gate: terminal ∧ SUCCEEDED ∧ produces_artifacts ∧ ¬annotation      │
 │  ② pair each declaration with the output literal it names          │
 │  ③ version, if empty ← <run>-<action>-<attempt>   (deterministic)  │
 │     source                ← TaskActionSource{action, attempt}      │
 └───────────────────────────────┬────────────────────────────────────┘
                                 │ ④ CreateArtifact, one per declaration
                                 ▼
                   ┌──────────────────────────┐
                   │      ArtifactService     │──► artifacts (one row/version)
                   └──────────────────────────┘
                                 │
   ⑤ all handled → patch annotation flyte.org/artifacts-registered=true
      transient error → return err, controller-runtime requeues with backoff
```

#### Rules

- **Gate before reading.** Without the `produces_artifacts` check every
  successful action pays a blob read it does not need. Cap the object size and
  skip anything larger.
- **Deterministic default version.** `<run>-<action>-<attempt>` is unique
  within the project and stable across retries, so a replayed registration
  collapses into `AlreadyExists` instead of minting a second version. Treat
  `AlreadyExists` as success — it is what makes "create remotely, then record
  locally" converge.
- **Per-artifact error handling.** A transient failure requeues the whole
  action. A permanent one (`InvalidArgument`, `PermissionDenied`,
  `Unimplemented`, …) drops that artifact with a log and a metric and lets the
  rest proceed: one bad declaration must not pin the CR in reconcile forever.
  That metric is the only signal a user-visible artifact went missing, so it
  needs an alert.
- **Never block terminal.** Registration is off the path the SDK waits on. An
  artifact service that is down delays artifacts; it does not delay actions.

### Stage 3 — artifact triggers

`flyte.OnArtifact` lets a task subscribe to an artifact name: every new version
fires one run, with that exact version bound to an input. It replaces polling
("re-train nightly, in case the dataset moved") and direct invocation
("the producing pipeline calls the consumer"), coupling the two sides through
the artifact's name alone.

The IDL is again already in place — `task.ArtifactTrigger`
(`artifact_name`, `version`, `input_arg`) and
`TriggerAutomationSpecType.TYPE_ARTIFACT` — and again nothing implements it:
`runs/repository/transformers/trigger.go` only ever sets `TYPE_SCHEDULE`, and
`runs/scheduler/` handles schedules exclusively.

Three pieces, in rough order of effort:

**1. Trigger side — make artifact triggers queryable.** Accept and persist
`TYPE_ARTIFACT`. The important part is the storage shape, not the acceptance:
schedule triggers are swept by time, artifact triggers must be looked up *by
the artifact name they watch*. `artifact_name` (and `version`) therefore need
to be indexed columns, not fields buried inside a serialized
`automation_spec` — otherwise every artifact creation degrades into a full
scan and deserialize. Same reasoning as the `artifact_parents` side table
above.

**2. Artifact side — match on write.** `CreateArtifact` looks up the triggers
watching `(org, project, domain, name)` whose version pin is empty or equal to
this version. This step is small.

**3. Firing pipeline — the bulk of the work.** Matched triggers must not be
fired inline: that would couple `CreateArtifact`'s latency and failure modes
to run creation, and leaves no good answer for a half-failed fire. Nor can the
artifact be committed first and the fires enqueued in memory afterwards — a
crash in between loses the event permanently.

Use a transactional outbox:

- a pending-fire table written **in the same transaction** as the artifact, so
  "artifact exists but no fire was recorded" and "a fire was recorded for an
  artifact that never committed" are both impossible;
- each pending row self-contained — the artifact version, the trigger identity
  and its revision at match time — so a later edit of the trigger cannot change
  what an already-matched fire does;
- a worker pool that drains committed rows after commit (the low-latency
  path), starting a run and binding the artifact to `input_arg`;
- a recovery sweep that re-enqueues rows the worker never completed, making
  the table the source of truth and the in-memory queue only an accelerator;
- a dead-letter path for rows that can never succeed, so one poison entry
  cannot wedge the sweep.
