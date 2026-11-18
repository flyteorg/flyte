FlyteAdmin
==========

[![Current Release](https://img.shields.io/github/release/flyteorg/flyteadmin.svg)](https://github.com/flyteorg/flyteadmin/releases/latest)
![Master](https://github.com/flyteorg/flyteadmin/workflows/Master/badge.svg)
[![GoDoc](https://godoc.org/github.com/flyteorg/flyteadmin?status.svg)](https://pkg.go.dev/mod/github.com/flyteorg/flyteadmin)
[![License](https://img.shields.io/badge/LICENSE-Apache2.0-ff69b4.svg)](http://www.apache.org/licenses/LICENSE-2.0.html)
[![CodeCoverage](https://img.shields.io/codecov/c/github/flyteorg/flyteadmin.svg)](https://codecov.io/gh/flyteorg/flyteadmin)
[![Go Report Card](https://goreportcard.com/badge/github.com/flyteorg/flyteadmin)](https://goreportcard.com/report/github.com/flyteorg/flyteadmin)
![Commit activity](https://img.shields.io/github/commit-activity/w/flyteorg/flyteadmin.svg?style=plastic)
![Commit since last release](https://img.shields.io/github/commits-since/flyteorg/flyteadmin/latest.svg?style=plastic)
[![Slack](https://img.shields.io/badge/slack-join_chat-white.svg?logo=slack&style=social)](https://slack.flyte.org)

FlyteAdmin is the control plane for Flyte responsible for managing entities (task, workflows, launch plans) and
administering workflow executions. FlyteAdmin implements the
[AdminService](https://github.com/flyteorg/flyteidl/blob/master/protos/flyteidl/service/admin.proto) which
defines a stateless REST/gRPC service for interacting with registered Flyte entities and executions.
FlyteAdmin uses a relational style Metadata Store abstracted by [GORM](http://gorm.io/) ORM library.

For more background on Flyte, check out the official [website](https://flyte.org/) and the [docs](https://docs.flyte.org/en/latest/index.html)

Before Check-In
---------------

Flyte Admin has a few useful make targets for linting and testing. Please use these before checking in to help suss out
minor bugs and linting errors.

```
  # Please make sure you have all the dependencies installed:
  $ make install
  
  # In case you are only missing goimports:
  $ go install golang.org/x/tools/cmd/goimports@latest
  $ make goimports
```

```
  $ make test_unit
```

```
  $ make lint
```

## Database execution retention

`flyteadmin-retention` is a separate one-shot binary, packaged in the FlyteAdmin
and combined Flyte images. It deletes execution history from an existing,
already migrated PostgreSQL database without starting the API or scheduler.
Build with `make -C flyteadmin compile_retention`, then preview cleanup:

```sh
flyteadmin/bin/flyteadmin-retention --config flyteadmin/flyteadmin_config.yaml \
  --retention-age 2160h --batch-size 100 --max-runtime 20m --dry-run
```

Pass `--dry-run=false` to delete. Defaults are 90 days, 100 candidates per page,
a 20-minute overall deadline, and dry-run enabled. Durations use Go syntax
(`2160h`, not `90d`). Age, batch size, and runtime must be positive. The job uses
the top-level `database` settings, including password files; SQLite is not
supported. It never creates a database or runs migrations. Deploy FlyteAdmin
first to apply the retention-index migration before scheduling cleanup.

V1 uses `execution_updated_at` for the last workflow event timestamp. Retention
selects executions in SUCCEEDED, FAILED, ABORTED, or TIMED_OUT phase whose
last event is older than the cutoff fixed at invocation start. Missing timestamps
and executions with nonterminal or unknown-phase nodes/tasks are preserved.
Nodes may also be terminal in SKIPPED or RECOVERED phase; task phases follow V1's
SUCCEEDED, FAILED, ABORTED terminal set.

An execution remains protected while another execution references its source ID,
while its nodes own launched workflow executions, or while other executions'
nodes reference its node/task parents. Eligible descendants can be deleted first;
a skipped parent becomes eligible on a later invocation. This preserves retained
recovery and launched-workflow history. Cycles or incomplete child history can
keep an execution ineligible indefinitely.

Each execution and its related `execution_events`, `node_execution_events`,
`node_executions`, `task_executions`, `signals`, and `execution_tags` are physically
deleted in one transaction. Existing execution/node/task rows are locked and
eligibility is rechecked. Dynamic-node parent links within the deleted execution
are cleared transactionally to allow foreign-key-safe deletion. Shared workflows,
tasks, launch plans, projects, schedules, and Datacatalog entries are preserved.
Blob storage and Redis metadata are not cleaned up.

The only advisory lock prevents overlapping retention invocations. It reserves
one connection while cleanup uses the pool separately. There are no per-execution
advisory locks, tombstones, insert guards, or transaction-timeout settings.
Late/concurrent writes can recreate rows after deletion. Batch size bounds
candidate executions per page, not rows; one large execution may consume the
remaining runtime budget. Timeout, cancellation, or failure rolls back the active
transaction and exits nonzero while preserving earlier commits. JSON output
reports the cutoff, dry-run mode, lock contention, and counts by table.

### Scheduling

The `flyte-core` chart has an optional CronJob, disabled by default:

```yaml
flyteadmin:
  retention:
    enabled: true
    schedule: "0 3 * * *"
    timeZone: "Etc/UTC"
    age: 2160h
    batchSize: 100
    maxRuntime: 20m
    dryRun: true
    activeDeadlineSeconds: 1500
    backoffLimit: 1
    resources: {}
```

Use a newly built FlyteAdmin image containing `flyteadmin-retention`. The job
shares the Admin configuration, database secret mounts, environment, service
account, extra volumes, security context, and node placement. Its pod label is
separate from the Admin service selector. Review preview counts before setting
`dryRun: false`. `suspend: true` pauses scheduling. The chart sets
`concurrencyPolicy: Forbid` and leaves starting deadlines and job history limits
to Kubernetes defaults. No in-cluster CronJob is created unless enabled.

### OTEL metrics

The binary reads the existing `otel` configuration. The default `noop` exporter
disables export; `file`, `otlpgrpc`, and `otlphttp` are supported. V1's shared
telemetry factory only creates tracers, so the binary owns its metric provider.
For OTLP HTTP, a configured `/v1/traces` path is mapped to `/v1/metrics`.
It exports under service name `flyteadmin-retention` and flushes on exit with a
separate 10-second budget, including on timeout/failure.

| Metric | Type | Labels | Meaning |
| --- | --- | --- | --- |
| `flyteadmin.retention.duration` | Histogram, seconds | `result`, `dry_run` | Invocation duration including connection setup; result is success, failure, or skipped when the lock is held |
| `flyteadmin.retention.deleted` | Counter | `entity` | Physically deleted rows, labeled by table |

The histogram's observation count supplies invocation totals. Deletion counts
are recorded after each transaction commits; rolled-back rows and dry-run
estimates are excluded. Partial progress is counted even if a later transaction
fails. No execution identifiers appear in metric labels. Initialization failures
before metric setup cannot be recorded, and exporter failures are logged.
