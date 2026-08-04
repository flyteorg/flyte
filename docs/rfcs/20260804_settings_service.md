# [RFC] Settings Service — hierarchical runtime configuration for Flyte 2

**Authors:**

- @pingsutw

## 1 Executive Summary

Flyte 2 already ships a complete, strongly-typed **settings API** in its IDL
(`flyteidl2/settings/settings_service.proto` and
`flyteidl2/settings/settings_definition.proto`, added in #7127, with generated
clients for Go, TypeScript, Python, and Rust) — but **no server implements it
yet**.

This RFC proposes implementing the `SettingsService` so platform operators can
set runtime defaults (default queue, task resource bounds, environment
variables, labels/annotations, service account, storage paths) at **org,
domain, and project scope**, with well-defined inheritance and override
semantics, instead of baking them into static server config or repeating them
in every task decorator.

## 2 Motivation — why do we need it?

Today, every knob that shapes *how* a task runs comes from exactly two places:

1. **Static server configuration** (e.g. the `runs.*` config section, executor
   plugin config) — global, requires a redeploy to change, and cannot differ
   per project or domain.
2. **User code** — `@task(resources=..., env_vars=...)` repeated in every
   repository, by every team, with no way for a platform team to enforce or
   default anything centrally.

This leaves real gaps:

- **No per-environment defaults.** There is no way to say "everything in the
  `production` domain runs with `LOG_LEVEL=info` and a max of 64 concurrent
  actions, while `development` runs with `LOG_LEVEL=debug` and no cap."
- **No platform guardrails.** An admin cannot enforce "no task in this project
  may request more than 16 CPUs / 64Gi memory" — resource limits are whatever
  user code asks for.
- **Config drift and repetition.** Teams copy the same labels, annotations,
  service accounts, and env vars into every task. When the value changes
  (a new cost-attribution label, a rotated service account), it changes in N
  repositories instead of one place.
- **Redeploys for config changes.** Changing a default today means editing
  server YAML and rolling the deployment. Settings should be data, not code.

Flyte 1 addressed part of this with *matchable attributes*
(`project_domain_attributes`, etc.), which proved the need but had known
usability problems: free-form structure, unclear precedence, and a confusing
CRUD surface. The Flyte 2 settings IDL was designed as its replacement —
**typed schema, explicit inheritance states, and scope-annotated reads** — and
merged after design review. Implementing the service is the missing step.

## 3 When do we need it? — concrete examples

Settings are resolved along the scope chain **org → domain → project**, where
the most specific level wins for scalars and maps merge additively
(child overrides parent on key conflict). Every leaf carries a
`SettingState`: `INHERIT` (defer to parent, the default), `VALUE` (set here),
or `UNSET` (explicitly blank, blocks inheritance).

A full worked walkthrough of every RPC with exact request/response payloads
already lives in [`flyteidl2/settings/settings_customer_flow.md`](../../flyteidl2/settings/settings_customer_flow.md).
Some motivating scenarios:

### Example 1 — environment defaults per domain

An admin wants sane logging everywhere, verbose logging in development:

| Scope | `environment_variables` |
|---|---|
| org | `{LOG_LEVEL: info}` |
| domain `development` | `{LOG_LEVEL: debug}` |
| project `recsys` (in `development`) | `{TEAM: ml}` |

A run in `recsys`/`development` resolves to
`{LOG_LEVEL: debug, TEAM: ml}` — maps merge parent-first, child wins on
conflict. `GetSettings` returns each value annotated with the `scope_level` it
came from, so a UI can show "inherited from domain".

### Example 2 — resource guardrails

Platform team enforces ceilings once, at the org level:

```
task_resource.max.cpu    = "16"
task_resource.max.memory = "64Gi"
```

At run submission the resolved settings are applied to every task: requests
above the max are capped, and (optionally, via
`task_resource.mirror_limits_request`) requests are mirrored into missing
limits. A GPU max is only ever used to cap, never to inject a GPU into a task
that didn't ask for one.

### Example 3 — routing runs to the right queue

```
run.default_queue = "gpu-pool"        (set on domain `production`)
run.max_action_concurrency = 64      (set on org)
```

Any run created in `production` without an explicit queue lands on
`gpu-pool`; every run is capped at 64 concurrently-executing actions unless a
more specific scope overrides it.

### Example 4 — explicitly blanking an inherited value

The org sets `security.service_account = "default-runner"`. One untrusted
project must *not* inherit it. Setting the project-level value to state
`UNSET` blocks inheritance without providing a replacement — a tri-state
(`INHERIT`/`UNSET`/`VALUE`) that plain key-value systems cannot express.

### The full settings schema (already in the IDL)

| Setting | Type | Controls |
|---|---|---|
| `run.default_queue` | string | Default queue for runs |
| `run.max_action_concurrency` | int64 | Max concurrently-executing actions per run (0 = unlimited) |
| `run.run_base_dir` | string | Base dir for code bundles / offloaded run metadata |
| `security.service_account` | string | Kubernetes service account for task pods |
| `storage.raw_data_path` | string | Base path for raw data, e.g. `s3://bucket/prefix` |
| `task_resource.min.{cpu,gpu,memory,storage}` | quantity | Minimum resource requests injected into tasks |
| `task_resource.max.{cpu,gpu,memory,storage}` | quantity | Maximum limits; existing values are capped |
| `task_resource.mirror_limits_request` | bool | Copy requests into missing limits |
| `labels`, `annotations` | string map | Applied to task pods, additive across scopes |
| `environment_variables` | string map | Injected into task pods, additive across scopes |
| `pod_template_name` | string | *(proposed addition, see below)* Name of a `PodTemplate` resource used as the base for task pods |

### Example 5 — per-project/domain pod templates

Flyte 1 supported per-project-domain `PodTemplate`s implicitly: task pods ran
in per-project-domain namespaces, and the namespace-aware `PodTemplateStore`
picked up a `PodTemplate` resource created in each namespace. Flyte 2 runs all
task pods in a **single namespace**, so that dimension collapses — today only
per-task (`pod_template_name` in `TaskMetadata`) and cluster-wide
(`default-pod-template-name` plugin config) selection work.

A `pod_template_name` setting restores the per-scope capability with no new
machinery: an admin sets it at org, domain, or project scope; at run creation
the resolved name is stamped onto tasks that don't specify their own; the
existing `PodTemplateStore` lookup does the rest. This covers everything the
typed settings don't (tolerations, node selectors, sidecars, init
containers, …) by referencing the full Kubernetes API instead of duplicating
it in the settings schema.

## 4 Proposed Implementation

### API (one proto field to add, everything else is done)

The only schema change this RFC proposes is a top-level
`StringSetting pod_template_name` field on `Settings` (alongside the other
pod-level settings: `labels`, `annotations`, `environment_variables`) plus
`make buf` regeneration. Everything else is already defined and
code-generated. Four RPCs:

```protobuf
service SettingsService {
  rpc GetSettings(GetSettingsRequest) returns (GetSettingsResponse);            // merged, effective values
  rpc GetSettingsForEdit(GetSettingsForEditRequest) returns (GetSettingsForEditResponse); // unmerged, one record per scope level
  rpc CreateSettings(CreateSettingsRequest) returns (CreateSettingsResponse);
  rpc UpdateSettings(UpdateSettingsRequest) returns (UpdateSettingsResponse);   // optimistic locking via version
}
```

`SettingsKey{org, domain, project}` determines scope by which fields are
populated: `{org}` = org-level, `{org, domain}` = domain-level,
`{org, domain, project}` = project-level.

**Org handling in OSS.** OSS deployments have no organization concept, so an
empty `org` is normalized server-side to the existing placeholder
`DefaultOrganization = "flyte"`
(`flyteplugins/go/tasks/pluginmachinery/secret/embedded_secret_manager.go`) —
the same convention the secret and app services already use. Org-level
settings therefore act as **instance-wide defaults**; clients never need to
send an org.

### Storage

One Postgres row **per scope level**, following the existing sqlx +
hand-written-SQL pattern in `runs/repository/`:

```sql
CREATE TABLE IF NOT EXISTS settings (
    id         BIGSERIAL   PRIMARY KEY,
    key        TEXT        NOT NULL UNIQUE,   -- "v1:{org}:{domain}:{project}"
    data       JSONB       NOT NULL DEFAULT '{}',
    version    BIGINT      NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

- `data` is the `Settings` proto serialized as protojson, **sparse**: leaves in
  state `INHERIT` are pruned before write (inherited values are never
  persisted), and hydrated back to explicit `INHERIT` on read.
- `version` implements optimistic locking: `UPDATE ... SET version = version + 1
  WHERE key = $1 AND version = $2`; zero rows affected ⇒ the caller loses the
  race and must re-read (same pattern as the existing trigger repository).
- Resolving a key costs **one query**: build the ≤3 ancestor keys and fetch
  them with `WHERE key = ANY($1)`. Missing rows simply mean all-`INHERIT`.

### Resolution engine

Pure functions, independently testable without a database:

- **Scalars** (`StringSetting`, `Int64Setting`, `BoolSetting`,
  `QuantitySetting`): iterate org → domain → project; the last level whose
  state is not `INHERIT` wins; annotate the winner with its `scope_level`.
- **String maps**: merge additively parent-first, child overriding on key
  conflict; a level with state `UNSET` clears everything accumulated so far.
- **Task resources**: per-field scalar merge over the min/max quantity bounds.

### Service placement and wiring

Implement `settingsconnect.SettingsServiceHandler` directly (Connect-RPC over
the shared `http.ServeMux`, mirroring `ProjectService` in
`runs/service/project_service.go`), and register it in `runs/setup.go` with
the existing OTel interceptor. Living in `runs/` shares the database,
migrations, `SetupContext`, and the embedded-Postgres test harness.

Validation is hand-written in the service (the generated `Validate()` from
protoc-gen-validate does not enforce the `buf.validate` annotations used in
the settings protos): key shape (empty org defaults to `"flyte"`; project
requires domain),
quantities parse via `resource.ParseQuantity`, and
`max_action_concurrency` ∈ {0} ∪ [2, MaxUint32] (a cap of 1 would deadlock
any run with more than one action).

## 5 Implementation plan — phased for parallel contribution

Each task below is a self-contained PR with an existing in-repo pattern to
copy from. Tasks marked **[independent]** have no dependency on other tasks in
the same phase.

### Phase 1 — persistence layer (good first issues)

| # | Task | Pattern to copy | Deliverable |
|---|---|---|---|
| 1.1 | **[independent]** Migration: `settings` table | `runs/migrations/sql/*.sql` | one SQL file |
| 1.2 | **[independent]** Model + key encoder (`"v1:{org}:{domain}:{project}"`, empty org normalized to `"flyte"`) | `runs/repository/models/project.go` | `models/settings.go` + unit test |
| 1.3 | Repo interface (`Create/Get/GetByKeys/Update`) + sentinel errors (`ErrSettingsNotFound`, `ErrSettingsVersionConflict`) | `runs/repository/interfaces/project.go` | `interfaces/settings.go` (mockery picks it up automatically) |
| 1.4 | sqlx implementation incl. optimistic-locking `Update` | `runs/repository/impl/project.go`, locking: `impl/trigger.go` | `impl/settings.go` + embedded-Postgres tests |
| 1.5 | **[independent]** Proto: add `pod_template_name` `StringSetting` to `Settings` + regen (`make buf`) | existing fields in `settings_definition.proto` | one proto field + generated code |

### Phase 2 — resolution engine + service (the core)

| # | Task | Depends on | Deliverable |
|---|---|---|---|
| 2.1 | **[independent]** Transformers: `Settings` proto ⇄ sparse protojson (`PruneInherited` / `Hydrate`) | — | `repository/transformers/settings.go` + table tests |
| 2.2 | **[independent]** Merge engine: scalar merge, map merge, task-resource merge, scope annotation | — | pure functions + table tests covering every `SettingState` combination |
| 2.3 | **[independent]** Validators: key shape, quantities, concurrency bounds | — | pure functions + tests |
| 2.4 | Connect handlers: `CreateSettings`, `UpdateSettings` | 1.x, 2.1, 2.3 | `runs/service/settings_service.go` (write half) |
| 2.5 | Connect handlers: `GetSettings`, `GetSettingsForEdit` (single `GetByKeys` fetch + merge) | 1.x, 2.1, 2.2 | read half + service tests |
| 2.6 | Registration in `runs/setup.go` + readiness | 2.4, 2.5 | 3-line wiring + smoke test in `runs/test/api` |

Phase 2 exit criteria: every request/response example in
`settings_customer_flow.md` reproduces byte-for-byte against a running server —
that document doubles as the acceptance test spec.

### Phase 3 — make settings take effect

Nothing reads settings yet; today the equivalent knobs come from static
config. This phase wires resolution into the run-creation path:

| # | Task | Applies |
|---|---|---|
| 3.1 | Settings applier scaffold: resolve once per run creation, apply to run/task specs | `run.default_queue`, `run.max_action_concurrency` |
| 3.2 | **[independent]** Pod-level settings | `environment_variables`, `labels`, `annotations`, `security.service_account` |
| 3.2b | **[independent]** Pod template selection: stamp resolved name onto tasks without an explicit `pod_template_name`; `PodTemplateStore` handles the rest | `pod_template_name` |
| 3.3 | **[independent]** Task resource bounds | `task_resource.min/max`, `mirror_limits_request` |
| 3.4 | **[independent]** Storage settings | `storage.raw_data_path`, `run.run_base_dir` |
| 3.5 | Precedence rule + docs: explicit user-provided spec values always win over settings-derived defaults | all of the above |

### Phase 4 — surfaces (all independent, all parallelizable)

| # | Task |
|---|---|
| 4.1 | CLI: `get settings` / `update settings` (clients already generated) |
| 4.2 | UI: settings page using `GetSettingsForEdit` (per-level editing with version-based conflict detection; the `desc` proto field option carries per-field help text readable via reflection) |
| 4.3 | User-facing docs: concepts page for scopes, inheritance, `UNSET` semantics |
| 4.4 | `flytekit` helpers, if any ergonomics gaps surface |

## 6 Drawbacks

- One more table and service to operate — mitigated by living inside the
  existing `runs` deployable, no new binary.
- A second source of truth next to static config until Phase 3 defines
  precedence clearly (settings override static defaults; explicit user specs
  override both).

## 7 Alternatives

- **Port Flyte 1 matchable attributes.** Rejected during the IDL design
  (#7127): free-form structure, no inheritance states, confusing CRUD.
- **Static config only.** Cannot express per-project/domain values without
  redeploys; no guardrails.
- **Per-workflow config in user code.** Exactly the repetition and lack of
  central enforcement this proposal removes.

## 8 Unresolved questions

- Should resolved settings be cached server-side, or is one indexed
  `key = ANY(...)` query per run creation cheap enough? (Proposal: ship
  without a cache, measure.)
- Authorization: the current server performs no authz; settings writes should
  be gated once an authz story exists. Out of scope here.
- Future scopes (e.g. user-level) — the key encoding is versioned (`v1:`) to
  leave room.

## 9 Conclusion

The API is designed, reviewed, merged, and code-generated in four languages;
the storage and service patterns it needs already exist in the repo. What
remains is well-bounded implementation work that slices into ~15 small PRs,
most of them independent — a good on-ramp for new contributors while
delivering a long-requested capability: central, hierarchical, typed runtime
configuration.
