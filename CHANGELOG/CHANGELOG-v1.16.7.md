## What's Changed
* Flyte 2: Cut over to Flyte 2 new backend implementation by @EngHabu in https://github.com/flyteorg/flyte/pull/6583
* Allow HTTP registry for Knative Serving in devbox by @pingsutw in https://github.com/flyteorg/flyte/pull/7302
* Move top-level docs into docs/ and tidy repo root by @pingsutw in https://github.com/flyteorg/flyte/pull/7303
* ci: add PR labeler workflow for main/master branches by @pingsutw in https://github.com/flyteorg/flyte/pull/7304
* docs: add devbox-based local development guide by @pingsutw in https://github.com/flyteorg/flyte/pull/7294
* App events new by @machichima in https://github.com/flyteorg/flyte/pull/7309
* Build on depot.dev instead by @EngHabu in https://github.com/flyteorg/flyte/pull/7306
* Move translatorservice to the dataplane as well by @katrogan in https://github.com/flyteorg/flyte/pull/7318
* ci(flyte-binary-v2): use pull_request_target so fork PRs can access DEPOT_PROJECT_ID by @pingsutw in https://github.com/flyteorg/flyte/pull/7320
* Add dependabot config and exclude generated directory by @Sovietaced in https://github.com/flyteorg/flyte/pull/7325
* ci(build-ci-image): use pull_request_target so fork PRs can access DEPOT_PROJECT_ID by @pingsutw in https://github.com/flyteorg/flyte/pull/7327
* docs: note LF AI & Data Foundation Graduated status in README by @EngHabu in https://github.com/flyteorg/flyte/pull/7324
* RetryStrategy and TimeoutStrategy by @kumare3 in https://github.com/flyteorg/flyte/pull/7308
* ci: switch ci trigger to pull_request_target for check-generate workflow by @SZL741023 in https://github.com/flyteorg/flyte/pull/7336
* feat(flytestdlib): add namespaceutils package to v2 by @SVilgelm in https://github.com/flyteorg/flyte/pull/7312
* remove unused runsettings by @afbrock in https://github.com/flyteorg/flyte/pull/7335
* fix(dataproxy): GetActionData returns empty inputs for sub-actions by @pingsutw in https://github.com/flyteorg/flyte/pull/7337
* handle azure Gen2 storage path when parsing ref by @SVilgelm in https://github.com/flyteorg/flyte/pull/7339
* Bump github.com/redis/go-redis/v9 from 9.17.2 to 9.19.0 by @dependabot[bot] in https://github.com/flyteorg/flyte/pull/7331
* Bump github.com/aws/aws-sdk-go-v2 from 1.41.1 to 1.41.7 by @dependabot[bot] in https://github.com/flyteorg/flyte/pull/7330
* change settings label type from list to map to align with k8s by @afbrock in https://github.com/flyteorg/flyte/pull/7342
* Inject INTERNAL_APP_ENDPOINT_PATTERN into KService pods for in-cluster app discovery by @AdilFayyaz in https://github.com/flyteorg/flyte/pull/7323
* Bump github.com/magiconair/properties from 1.8.6 to 1.8.10 by @dependabot[bot] in https://github.com/flyteorg/flyte/pull/7345
* fix: stop Knative apps privately + keep UI state accurate after stop/start by @AdilFayyaz in https://github.com/flyteorg/flyte/pull/7343
* [Proto] Export actions in rust proto by @machichima in https://github.com/flyteorg/flyte/pull/7349
* use Helm to init rustfs in sandbox based on main branch by @BarryWu0812 in https://github.com/flyteorg/flyte/pull/7310
* feat(connector): record connector endpoint on LogContext by @pingsutw in https://github.com/flyteorg/flyte/pull/7317
* fix: make copilot task metadata parsing nil-safe by @xjerod in https://github.com/flyteorg/flyte/pull/7351
* Cache image-pull-secret reads in webhook via controller-runtime informer by @EngHabu in https://github.com/flyteorg/flyte/pull/7353
* feat(tasklog): add runName and actionName template vars by @pingsutw in https://github.com/flyteorg/flyte/pull/7367
* Allow routing secrets endpoints via SelectCluster by @katrogan in https://github.com/flyteorg/flyte/pull/7245
* chore(webapi): promote pluginmachinery State + Phase to public package by @pingsutw in https://github.com/flyteorg/flyte/pull/7370
* fix(executor): wire previous checkpoint path for in-place retries by @AdilFayyaz in https://github.com/flyteorg/flyte/pull/7371
* [Fix] make CI only runs on main branch by @machichima in https://github.com/flyteorg/flyte/pull/7372
* chore(deps): add 5-day cooldown to dependabot updates by @SVilgelm in https://github.com/flyteorg/flyte/pull/7379
* add enable runstore trigger by @popojk in https://github.com/flyteorg/flyte/pull/7381
* Run dependabot for Flyte v1 too by @Sovietaced in https://github.com/flyteorg/flyte/pull/7376
* build(deps): bump github.com/spf13/pflag from 1.0.6 to 1.0.10 by @dependabot[bot] in https://github.com/flyteorg/flyte/pull/7392
* Update Go to 1.25 by @Sovietaced in https://github.com/flyteorg/flyte/pull/7326
* feat(connector): look up connector by project and domain by @pingsutw in https://github.com/flyteorg/flyte/pull/7378
* Fix main build by @Sovietaced in https://github.com/flyteorg/flyte/pull/7395
* build(deps): bump golang.org/x/net from 0.48.0 to 0.54.0 by @dependabot[bot] in https://github.com/flyteorg/flyte/pull/7390
* build(deps): bump golang.org/x/oauth2 from 0.32.0 to 0.36.0 by @dependabot[bot] in https://github.com/flyteorg/flyte/pull/7387
* proto: add package_name field to PythonWheels by @machichima in https://github.com/flyteorg/flyte/pull/7384
* feat: add ClusteredTaskSpec proto definition for JobSet-based distributed training by @AdilFayyaz in https://github.com/flyteorg/flyte/pull/7383
* fix(flytek8s): classify ImagePullBackOff caused by registry 429 as system error by @pingsutw in https://github.com/flyteorg/flyte/pull/7362
* build(deps): bump golang.org/x/time from 0.12.0 to 0.15.0 by @dependabot[bot] in https://github.com/flyteorg/flyte/pull/7407
* build(deps): bump github.com/eko/gocache/lib/v4 from 4.2.0 to 4.2.3 by @dependabot[bot] in https://github.com/flyteorg/flyte/pull/7406
* build(deps): bump buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go from 1.36.10-20250912141014-52f32327d4b0.1 to 1.36.11-20260415201107-50325440f8f2.1 by @dependabot[bot] in https://github.com/flyteorg/flyte/pull/7405
* Update pgx to fix vuln by @Sovietaced in https://github.com/flyteorg/flyte/pull/7400
* feat(connector): handle TaskExecution_RETRYABLE_FAILED phase by @pingsutw in https://github.com/flyteorg/flyte/pull/7409
* propagate ContainerError.Kind to ExecutionError.Recoverability in ReadError by @kumare3 in https://github.com/flyteorg/flyte/pull/7411
* fix: make pod_name and container_name optional in LoggingContext; add ContainerIdentifier to TailLogsResponse by @pingsutw in https://github.com/flyteorg/flyte/pull/7341
* build(deps): bump github.com/flyteorg/stow from 0.3.12 to 0.3.13 by @dependabot[bot] in https://github.com/flyteorg/flyte/pull/7415
* build(deps): bump github.com/eko/gocache/store/freecache/v4 from 4.2.0 to 4.2.4 by @dependabot[bot] in https://github.com/flyteorg/flyte/pull/7414
* Get rid of pgconn by @Sovietaced in https://github.com/flyteorg/flyte/pull/7412
* fix: Get rid of deprecated x/net/http2/h2c by @alexandear in https://github.com/flyteorg/flyte/pull/7424
* Update msgpack to fix vuln by @Sovietaced in https://github.com/flyteorg/flyte/pull/7421
* feat: add clustered-task plugin for JobSet-based distributed training by @AdilFayyaz in https://github.com/flyteorg/flyte/pull/7417
* feat(workflow): add RunService.SignalEvent + EventPayload by @SVilgelm in https://github.com/flyteorg/flyte/pull/7425
* Update grpc-go by @Sovietaced in https://github.com/flyteorg/flyte/pull/7427
* build(deps): bump google.golang.org/api from 0.247.0 to 0.280.0 by @dependabot[bot] in https://github.com/flyteorg/flyte/pull/7422
* Allow prefixes in dask dashboard links by @katrogan in https://github.com/flyteorg/flyte/pull/7408
* build(deps): bump github.com/jackc/pgx/v5 from 5.9.0 to 5.9.2 by @dependabot[bot] in https://github.com/flyteorg/flyte/pull/7429
* build(deps): bump github.com/fsnotify/fsnotify from 1.9.0 to 1.10.1 by @dependabot[bot] in https://github.com/flyteorg/flyte/pull/7428
* Enable rustfs console by @BarryWu0812 in https://github.com/flyteorg/flyte/pull/7347
* Unpin k8s libs by @Sovietaced in https://github.com/flyteorg/flyte/pull/7413
* feat: add worker-0 fast-fail and host maintenance retry to clustered plugin by @AdilFayyaz in https://github.com/flyteorg/flyte/pull/7418
* fix(otelutils): avoid resource.Merge schema URL conflict panic by @pingsutw in https://github.com/flyteorg/flyte/pull/7431
* build(deps): bump github.com/grpc-ecosystem/grpc-gateway/v2 from 2.27.7 to 2.29.0 by @dependabot[bot] in https://github.com/flyteorg/flyte/pull/7423
* build(deps): bump k8s.io/klog/v2 from 2.130.1 to 2.140.0 by @dependabot[bot] in https://github.com/flyteorg/flyte/pull/7437
* build(deps): bump github.com/coocood/freecache from 1.2.4 to 1.2.7 by @dependabot[bot] in https://github.com/flyteorg/flyte/pull/7436
* build(deps): bump github.com/samber/lo from 1.39.0 to 1.53.0 by @dependabot[bot] in https://github.com/flyteorg/flyte/pull/7434
* ci: add flyteidl-release workflow to main by @pingsutw in https://github.com/flyteorg/flyte/pull/7438
* ci: port release workflows from master to main by @pingsutw in https://github.com/flyteorg/flyte/pull/7439

## New Contributors
* @xjerod made their first contribution in https://github.com/flyteorg/flyte/pull/7351
* @alexandear made their first contribution in https://github.com/flyteorg/flyte/pull/7424

**Full Changelog**: https://github.com/flyteorg/flyte/compare/v1.16.6...v1.16.7
