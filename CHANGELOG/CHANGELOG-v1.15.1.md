## What's Changed
* [Docs] Streaming Decks by @Future-Outlier in https://github.com/flyteorg/flyte/pull/6251
* Upgrade gh actions/cache to v4 by @eapolinario in https://github.com/flyteorg/flyte/pull/6256
* auto-update contributors by @flyte-bot in https://github.com/flyteorg/flyte/pull/6252
* Use go version from `go.mod` files in CI by @eapolinario in https://github.com/flyteorg/flyte/pull/5736
* Add support for merging complete head & worker node pod templates by @Sovietaced in https://github.com/flyteorg/flyte/pull/6232
* Upstream: Ensure that k8s plugins (dask, kfoperator) have resource requests and limits set by @katrogan in https://github.com/flyteorg/flyte/pull/6264
* Override literal in the case of attribute access of primitive values by @eapolinario in https://github.com/flyteorg/flyte/pull/6194
* [Docs] Add Slurm agent setup by @JiangJiaWei1103 in https://github.com/flyteorg/flyte/pull/6231
* Rework MergePodSpecs logic by @Sovietaced in https://github.com/flyteorg/flyte/pull/6262
* fix: Correct error handling in delete function by @pingsutw in https://github.com/flyteorg/flyte/pull/6269
* [Docs] Streaming Decks in flyte fundamentals by @Future-Outlier in https://github.com/flyteorg/flyte/pull/6272
* [Docs] Slurm GPU cluster by @Future-Outlier in https://github.com/flyteorg/flyte/pull/6273
* Set the max parallelism on the execution model correctly when triggered by launch plan by @Sovietaced in https://github.com/flyteorg/flyte/pull/6266
* add bound inputs to array node idl by @pvditt in https://github.com/flyteorg/flyte/pull/6276
* Union/upstream pod helper node preemption handling by @pvditt in https://github.com/flyteorg/flyte/pull/6259
* Remove mockery fork by @eapolinario in https://github.com/flyteorg/flyte/pull/6280
* auto-update contributors by @flyte-bot in https://github.com/flyteorg/flyte/pull/6268
* Fix: Enable kubeflow plugins to use dynamic log links by @fg91 in https://github.com/flyteorg/flyte/pull/6284
* Doc: Document usage of dynamic log links by @fg91 in https://github.com/flyteorg/flyte/pull/6285
* Remove mockery fork logic by @Sovietaced in https://github.com/flyteorg/flyte/pull/6288
* [DOCS] update contribution docs for flytectl by @machichima in https://github.com/flyteorg/flyte/pull/6290
* Don't error out for getting metrics of array node by @troychiu in https://github.com/flyteorg/flyte/pull/6291
* [Housekeeping] upgrade go_generate ci by @taieeuu in https://github.com/flyteorg/flyte/pull/6279
* [flyteagent] Add Output Prefix to GetTaskRequest by @Future-Outlier in https://github.com/flyteorg/flyte/pull/6265
* fix: return nil instead of error directly by @rustco in https://github.com/flyteorg/flyte/pull/6298
* Add allowedAudience to flyte-core external auth deployment documentation by @mwaylonis in https://github.com/flyteorg/flyte/pull/5124
* Feat: Support per-launch plan notification template by @peterxcli in https://github.com/flyteorg/flyte/pull/6064
* Bump jinja2 from 3.1.4 to 3.1.5 in /flytectl/docs by @dependabot in https://github.com/flyteorg/flyte/pull/6126
* chore: fix some function names in comment by @sjtucoder in https://github.com/flyteorg/flyte/pull/6230
* Ignore generated code from codecov reports by @eapolinario in https://github.com/flyteorg/flyte/pull/6299
* [flyteagent] add reserved field in Get Request Payload by @Future-Outlier in https://github.com/flyteorg/flyte/pull/6296
* Bump github.com/go-jose/go-jose/v3 from 3.0.3 to 3.0.4 by @dependabot in https://github.com/flyteorg/flyte/pull/6281
* Update glog 1.2.0 -> 1.2.4 - CVE-2024-45339 by @ddl-ebrown in https://github.com/flyteorg/flyte/pull/6301
* Bump golang.org/x/net and  github.com/mattn/go-sqlite3 by @eapolinario in https://github.com/flyteorg/flyte/pull/6315
* fix: Ensure $(go env GOPATH)/bin exists by @honnix in https://github.com/flyteorg/flyte/pull/6314
* Bump jinja2 from 3.1.5 to 3.1.6 in /flytectl/docs by @dependabot in https://github.com/flyteorg/flyte/pull/6310
* RFC: Add workflow execution concurrency by @katrogan in https://github.com/flyteorg/flyte/pull/5659
* Add aws secret manager support to env_var by @thomasjpfan in https://github.com/flyteorg/flyte/pull/6316
* Document AWS SM by @bra-fsn in https://github.com/flyteorg/flyte/pull/6287
* Pin flytekit to 1.15.0 to unblock functional tests by @eapolinario in https://github.com/flyteorg/flyte/pull/6330
* Fix: fastcache should not cache lookup on failed node by @fg91 in https://github.com/flyteorg/flyte/pull/6318
* Reserve fields in execution protos by @eapolinario in https://github.com/flyteorg/flyte/pull/6321
* upstream bound inputs support by @pvditt in https://github.com/flyteorg/flyte/pull/6322
* fix: Make datacatalog gRPC server MaxRecvMsgSize configurable by @honnix in https://github.com/flyteorg/flyte/pull/6313
* Use child dir for branch taken by @andrewwdye in https://github.com/flyteorg/flyte/pull/6327
* Update Event Docs by @CtfChan in https://github.com/flyteorg/flyte/pull/6334
* [flytepropeller][flyteagent] Set ListAgent Timeout to unblock propeller launch execution by @Future-Outlier in https://github.com/flyteorg/flyte/pull/6312
* fix: really assign string scp claim by @vlada-dudr in https://github.com/flyteorg/flyte/pull/6336
* Add Connector plugin by @Future-Outlier in https://github.com/flyteorg/flyte/pull/6332
* Introduce structured log lines in the agent proto by @pingsutw in https://github.com/flyteorg/flyte/pull/6344
* Bump github.com/golang-jwt/jwt/v4 from 4.5.1 to 4.5.2 by @dependabot in https://github.com/flyteorg/flyte/pull/6362
* Bump github.com/golang-jwt/jwt/v4 from 4.5.1 to 4.5.2 in /flyteadmin by @dependabot in https://github.com/flyteorg/flyte/pull/6365
* fix typo: specify -> specific by @jamestwebber in https://github.com/flyteorg/flyte/pull/6342
* Bump github.com/go-jose/go-jose/v3 from 3.0.3 to 3.0.4 in /flyteadmin by @dependabot in https://github.com/flyteorg/flyte/pull/6366
* Bump github.com/golang-jwt/jwt/v5 from 5.2.1 to 5.2.2 in /flyteidl by @dependabot in https://github.com/flyteorg/flyte/pull/6361
* Doc: Explain how to chain launch plans by @fg91 in https://github.com/flyteorg/flyte/pull/6317
* Bump github.com/golang-jwt/jwt/v5 from 5.2.1 to 5.2.2 in /flytestdlib by @dependabot in https://github.com/flyteorg/flyte/pull/6370
* Bump github.com/golang-jwt/jwt/v5 from 5.2.1 to 5.2.2 in /boilerplate/flyte/golang_support_tools by @dependabot in https://github.com/flyteorg/flyte/pull/6371
* [Docs] Fixing the broken link to Architecture Overview in Deployment Guide by @0yukali0 in https://github.com/flyteorg/flyte/pull/6373
* Dep: Upgrade kubeflow training operator by @fg91 in https://github.com/flyteorg/flyte/pull/6294
* DeepCopy structs in TaskExecMetadata to avoid two routines modifying … by @EngHabu in https://github.com/flyteorg/flyte/pull/6382

## New Contributors
* @rustco made their first contribution in https://github.com/flyteorg/flyte/pull/6298
* @mwaylonis made their first contribution in https://github.com/flyteorg/flyte/pull/5124
* @sjtucoder made their first contribution in https://github.com/flyteorg/flyte/pull/6230
* @bra-fsn made their first contribution in https://github.com/flyteorg/flyte/pull/6287
* @CtfChan made their first contribution in https://github.com/flyteorg/flyte/pull/6334
* @vlada-dudr made their first contribution in https://github.com/flyteorg/flyte/pull/6336
* @jamestwebber made their first contribution in https://github.com/flyteorg/flyte/pull/6342
* @0yukali0 made their first contribution in https://github.com/flyteorg/flyte/pull/6373

**Full Changelog**: https://github.com/flyteorg/flyte/compare/v1.15.0...v1.15.1