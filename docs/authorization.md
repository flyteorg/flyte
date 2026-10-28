# Runs and data proxy authorization

Flyte 2 can authorize runs, actions, tasks, triggers, projects, and settings using
configured roles derived from trusted identity claims. This follows the
[context and repository isolation approach in PR #6190](https://github.com/flyteorg/flyte/pull/6190),
adapted to SQLx. Organization is not part of this implementation.

Authorization is disabled by default. Configure it in the runs configuration file:

```yaml
runs:
  trustForwardedIdentityHeaders: true
  authorization:
    enabled: true
    claimsHeader: X-Amzn-Oidc-Data
    roles:
      - name: payments-reader
        claim: groups
        values: [payments-readers]
        procedures:
          - /flyteidl2.workflow.RunService/ListRuns
          - /flyteidl2.workflow.RunService/GetRunDetails
          - /flyteidl2.workflow.RunService/GetActionDetails
          - /flyteidl2.workflow.RunService/WatchRuns
          - /flyteidl2.workflow.RunService/WatchActions
          - /flyteidl2.task.TaskService/ListTasks
          - /flyteidl2.project.ProjectService/GetProject
          - /flyteidl2.project.ProjectService/ListProjects
        scopes: ["production:payments"]
      - name: development-author
        claim: groups
        values: [developers]
        procedures:
          - /flyteidl2.workflow.RunService/CreateRun
          - /flyteidl2.task.TaskService/DeployTask
          - /flyteidl2.trigger.TriggerService/*
        scopes: ["development:*"]
      - name: administrator
        claim: groups
        values: [flyte-admins]
        procedures: ["*"]
        scopes: ["*:*"]
```

The policy is file-only, without command-line flags. Claim names are literal
(top-level JSON keys); values match exactly against a string or array of strings.
Multiple matching roles are additive. Procedure patterns support exact RPC paths,
`/package.Service/*`, and `*`. Grants are selected **per RPC**: a broad read role
does not widen the scopes of a narrow write role. Unmatched procedures are denied.

## Authentication boundary

The selected JWT must already have been authenticated by a trusted proxy that
validates the token and overwrites/removes client-supplied copies of the claims
header. The service decodes that one header; it does not verify JWT signatures.
There is no fallback to another token or identity header. Use `claimsHeader:
Authorization` only when the proxy validates and forwards Bearer JWTs. Plain
subject/email headers cannot supply role claims. Setting
`trustForwardedIdentityHeaders: false` with authorization enabled fails startup.

This setting protects the RPC handlers mounted by the **runs service**, including
InternalRunService, ProjectService, and SettingsService. OAuth discovery and
UserInfo retain their existing behavior. It does not authorize the separate
Actions, cache, app, secret, or executor services. Data proxy has its own
configuration described below. The other endpoints must
remain protected by their deployment's authentication/network boundaries; this
feature alone is not installation-wide isolation.

InternalRunService callers also need proxy-authenticated claims and explicit role
grants for their RPCs and scopes. Existing direct unauthenticated service calls
will be rejected when authorization is enabled. Configure service identities at
the trusted ingress before enabling it. In-process seed, scheduler, and abort
workers use an explicit trusted system context, never an HTTP header bypass.

## Scope semantics

| Scope | Access |
| --- | --- |
| `production:payments` | One project within a domain |
| `production:*` | All projects within a domain |
| `*:payments` | One project across all domains |
| `*:*` | All domains and projects |

Only an entire segment may be `*`; empty segments and partial globs are rejected.
Scopes are paired: `production:payments` plus `development:reports` never grants
`production:reports`. Parent grants include their descendants.

An immutable copy of the effective scopes lives in the Go context. Repositories
inject a parameterized predicate independent of request filters before pagination,
aggregation, and task counts. Complete scopes are ORed together; isolation and
request filters are ANDed with explicit parentheses. Keyed reads and mutations
validate the complete domain/project key before querying. Batch writes validate
all target scopes before starting. Watches retain their original context and
recheck scope on notification reads. Roles are resolved when a stream opens;
policy changes require server restart/reconnection.

When enabled on a repository, absent authorization context is denied. Explicitly
empty grants yield empty list results and deny keyed operations. When disabled,
legacy callers without scope context retain existing behavior.

Project metadata is shared across domains in today's schema. A project is visible
when any effective scope permits it, and only permitted domains appear in the
response. Creating/updating shared project metadata requires a `*:project` or
`*:*` grant as well as permission for that RPC.

Settings may inherit from instance and domain settings. Reading ancestor settings
is allowed for a permitted descendant. Mutating a domain's settings requires
`domain:*`; mutating instance settings requires `*:*`. Organization in settings
keys continues to be ignored, as before.

Task specs are content-addressed and shared. Reading a spec requires first
resolving its digest from an authorized action; knowing a digest alone is not a
resource grant. There is no public digest-only RPC.

## Validation

Tests cover scope pairing, method-specific grants, invalid/missing claims,
fail-closed context, real Connect unary and streaming interception, SQL OR-filter
escape attempts, authorized counts, keyed reads, shared metadata, settings, and
mixed-scope batches. Run `go test ./flytestdlib/authorization ./dataproxy/... ./runs/...` (repository and service tests start
embedded PostgreSQL).


## Data proxy

Data proxy uses the same shared authorization package and role format. It is
configured independently so split-service deployments do not depend on Runs
configuration being loaded into the data-proxy process. Authorization is disabled
by default; enabling Runs authorization does not implicitly enable data proxy.

```yaml
dataproxy:
  trustForwardedIdentityHeaders: true
  authorization:
    enabled: true
    claimsHeader: X-Amzn-Oidc-Data
    roles:
      - name: payments-data-reader
        claim: groups
        values: [payments-readers]
        procedures:
          - /flyteidl2.dataproxy.DataProxyService/GetActionData
          - /flyteidl2.dataproxy.DataProxyService/CreateDownloadLink
          - /flyteidl2.dataproxy.DataProxyService/TailLogs
          - /flyteidl2.workflow.TranslatorService/LiteralsToLaunchFormJson
          - /flyteidl2.cluster.ClusterService/SelectCluster
        scopes: ["production:payments"]
  runService:
    url: http://flyte-runs:8090
    auth:
      type: ""
```

All DataProxyService, TranslatorService, and ClusterService RPCs are gated by
procedure permission. Data proxy checks the target domain/project before issuing
upload/download URLs, writing inputs, reading action data, or opening log streams.
UploadInputs also checks any referenced task/trigger. Translator's storage-backed
operations check their action/trigger scope and retain URI-to-action validation.
Inline-only translation and SelectCluster (which returns this deployment's own
endpoint without looking up a resource) require RPC permission.

Data proxy forwards **only the selected authenticated claims header** to its
configured Runs endpoint, never a client-provided list of scopes. Runs must use
the same claims header and re-evaluates its own roles for the downstream RPC.
The internal network route must preserve this trusted header; do not expose an
unvalidated direct route to either service. Use TLS when forwarding claims across
an untrusted network. Redirects are disabled on these downstream clients.

With authorization enabled, `runService.auth` service credentials are rejected
at startup: they must not replace the caller's identity with a privileged service
identity. Unauthenticated requests have no claims to forward and fail closed.
Both the data-proxy permission and downstream permission must be granted:

| Data-proxy operation | Required Runs-hosted RPCs |
| --- | --- |
| CreateUploadLocation | ProjectService/GetProject |
| UploadInputs | ProjectService/GetProject; TaskService/GetTaskDetails and/or TriggerService/GetTriggerDetails when referencing a task/trigger |
| CreateDownloadLink | RunService/GetActionDetails |
| GetActionData | RunService/GetActionDataURIs |
| TailLogs | RunService/GetActionLogContext |
| Translator action URI | RunService/GetActionDataURIs |
| Translator trigger inputs | TriggerService/GetTriggerRevisionDetails |

Add the corresponding full procedure paths to the caller's roles in the Runs
policy with the same intended scopes. A data-proxy grant alone does not confer
permission to call Runs. Downstream authentication and permission denials remain
Unauthenticated/PermissionDenied rather than being converted to NotFound.

When authorization is enabled, upload path components reject traversal, absolute
paths, URI escapes, and separators that could escape the scoped storage prefix.
Nested relative filenames remain supported. UploadInputs with an explicit
`base_dir` additionally requires an effective `*:*` grant for that RPC because it
can select an arbitrary bucket/base. Ordinary scoped uploads use the configured
storage prefix. Existing custom-base behavior is preserved when authorization is
disabled. Signed URLs remain usable until they expire; authorization is checked
when they are issued.
