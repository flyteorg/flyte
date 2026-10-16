# Cedar authorization (experimental, model v1)

This opt-in first pass enforces permissions on the Connect handlers registered by
Runs and DataProxy. It uses cedar-go v1.8.0 locally; no external policy service is
required. Existing behavior is unchanged when disabled.

## Enable

Mount a reviewed Cedar policy file into both services. For unified mode, the two
configuration blocks must be identical. For separate services, distribute the same
policy contents and identity configuration to both.

```yaml
runs:
  authorization:
    enabled: true
    trustForwardedClaims: true
    claimsHeader: Authorization
    issuer: https://identity.example.com
    policyFile: /etc/flyte/authorization/policy.cedar
  triggerScheduler:
    enabled: false

dataproxy:
  authorization:
    enabled: true
    trustForwardedClaims: true
    claimsHeader: Authorization
    issuer: https://identity.example.com
    policyFile: /etc/flyte/authorization/policy.cedar
  runService:
    url: http://runs:8090
    auth:
      type: ""
```

Configuration is file-only. Policies are parsed at startup; invalid syntax or a
missing file stops startup. An empty file denies everything. Updates require a
restart; restart both services when deploying a new policy version. The SHA-256
of the loaded policy file is included in decision logs, alongside principal,
action, resource, decision, and evaluation-error count. Tokens are never logged.

Cedar policy evaluation errors cause Flyte to deny the request even if another
policy permits it. This is stricter than Cedar's normal skip-on-error behavior.

## Authentication boundary

The listener must only accept requests from a trusted authenticating proxy and
trusted internal services. The proxy must verify JWT signatures, issuer, audience,
and expiry, and strip/overwrite client-supplied trusted identity headers. This
implementation decodes forwarded claims; **it does not verify JWT signatures**.
`trustForwardedClaims` is an explicit operator acknowledgement of that boundary,
independent of the existing run-attribution setting.

Exactly one configured header is accepted, without identity-source fallback. For
`Authorization`, its value must be a Bearer JWT. For a custom claims header, its
value must be the JWT itself. Claims must contain `sub` and the configured `iss`;
`groups`, if present, must be an array of nonempty strings. Users and service
accounts are both `Flyte::Principal` entities, identified by subject within the
single configured issuer. Group memberships come only from those trusted claims.

DataProxy forwards the selected caller claims to Runs. Runs forwards them for
its ProjectService calls. Receiving services evaluate their own permissions.
DataProxy refuses service-credential configuration while Cedar is enabled, since
that could replace the caller with a broader identity. Internal network access
must prevent callers from forging these forwarded headers.

OAuth discovery and IdentityService remain outside the Cedar resource-policy
checks. IdentityService still relies on its existing upstream authentication.
Other components, including Actions, Events, Cache, Secrets, and Apps, are not
protected by this feature. Ingress must restrict their routes to the appropriate
trusted callers, including in unified mode. This is not a replacement for workload
IAM, Kubernetes RBAC, storage isolation, or control-plane network restrictions.

## Resource model and permissions

See [schema.cedarschema](schema.cedarschema) and [example.cedar](example.cedar).
The schema is an authoring contract for offline validation with Cedar tooling.
This version parses policy syntax at startup but does not perform schema-based
policy validation in-process; cedar-go v1.8.0 does not provide that validator.
Review policies and validate them against the supplied schema before deployment.

| Entity | ID / relationship |
| --- | --- |
| `Flyte::Principal` | JWT subject; parent groups; `issuer` attribute |
| `Flyte::Group` | Group string from JWT |
| `Flyte::Deployment` | `default`, the entire local deployment |
| `Flyte::Project` | Project name; parent deployment; `project` attribute |
| `Flyte::Scope` | `project/domain`; parents project and deployment; `project`, `domain` attributes |

The request context is empty. The model deliberately does not use request `org`
as a tenancy boundary: several current database paths do not enforce organization
isolation. Deployments requiring multi-organization isolation need a separate
implementation of that boundary before enabling shared access.

| Permission | Operations |
| --- | --- |
| `RunCreate` | CreateRun on its destination scope |
| `RunRead` | Run/action details, lists, watches, groups |
| `RunAbort`, `RunSignal` | Abort run/action; signal event |
| `TaskRead`, `TaskDeploy` | Task details/lists/versions; deploy |
| `TaskExecute`, `TriggerExecute` | Referenced task or trigger in CreateRun |
| `TriggerRead`, `TriggerWrite` | Trigger reads; deploy/update/delete |
| `DataRead`, `DataWrite` | Data, URIs, reports; upload location/input upload |
| `LogsRead` | Log context, tailing, cluster events |
| `ProjectRead`, `ProjectWrite` | Shared project metadata; create/update |
| `SettingsRead`, `SettingsWrite` | Deployment-wide settings access |
| `Translate`, `ClusterRead` | Deployment-wide conversion and cluster selection |
| `StorageOverride` | Deployment-wide raw/offloaded storage-location overrides |
| `RuntimeWrite` | All InternalRunService reporting APIs, deployment-wide |

All requirements must be allowed. Bulk trigger changes check every target before
any mutation. DeployTask with embedded triggers additionally requires
`TriggerWrite`. CreateRun checks its destination and referenced task/trigger;
recovery/provenance references require `DataRead` on the referenced run's scope.
Inline task execution is authorized by `RunCreate`; it does not refer to a
registered task. Input upload checks referenced task/trigger reads separately.
Offloaded-literal translation checks the owner as well as `Translate`.

Creating runs, deploying tasks, and uploading data can involve ProjectService
validation calls; grant `ProjectRead` on the shared project too. Reading project
metadata exposes the configured domain names, not permission to use those domains.

## Lists and streams

A list/watch scoped to a project and domain requires permission on that scope.
Existing database predicates restrict results **before pagination and counts**.
There is no post-filtering and no arbitrary Cedar-to-SQL translation.

Broad listings require permission on `Deployment::"default"`. This includes
ListProjects, organization-wide lists, ListRuns filtered by task (which selects
the task's scope rather than the run's), and WatchRuns targets other than project
(the current implementation only enforces its project filter). A scoped reader
cannot use these as a way to enumerate other scopes. Use a project-scoped request
instead. Settings remain deployment-wide because responses can expose inherited
parent records.

New/unmapped procedures fail closed. Every incoming stream message is checked
before handler processing. Server-streaming permission is a snapshot at stream
creation, not continuous group/expiry revalidation. Policy changes require
service restart, which also closes existing streams. There is no live revocation
or dynamic policy reload in this pass.

## Runtime and scheduler

InternalRunService requires an authenticated principal explicitly granted
`RuntimeWrite`. Configure runtime service clients to use their existing
OAuth2ClientCredentials support through the trusted authentication ingress.
There is no magic header or implicit system bypass.

When Cedar is enabled, the trigger scheduler must either be disabled or have an
explicit authenticated `runs.schedulerService` configuration:

```yaml
runs:
  schedulerService:
    url: https://flyte-authenticated.example.com
    auth:
      type: OAuth2ClientCredentials
      issuerUrl: https://identity.example.com
      clientId: flyte-scheduler
      clientSecretFile: /var/run/secrets/scheduler/client-secret
```

Grant the scheduler's actual token subject `RunCreate`, `TriggerExecute`, and
`ProjectRead` for the scopes it schedules. Scheduled launches use the scheduler's
authority, not a stored user token. In-process seed and reconciliation work
continues using trusted repository access; repositories are not independently
protected from arbitrary in-process callers.

## Storage limitations

Custom input `base_dir`, `run_base_dir`, and supplied `offloaded_input_data` require
`StorageOverride` on the deployment. This is intentionally conservative: a scope
on the request does not prove ownership of a supplied URI. It also means SDK flows
that offload inputs need that additional grant in this first pass. Scope users
should not receive it unless they are trusted with arbitrary storage references.
Storage ownership validation for these flows is follow-up work.

Upload filename/root traversal is rejected when authorization is enabled. Data
access still relies on trusted stored execution metadata and the workload's cloud
permissions. `DataRead` controls the data APIs; metadata APIs may expose storage
URIs, task specifications, and other metadata. It is not a guarantee of secret
redaction throughout every response.
