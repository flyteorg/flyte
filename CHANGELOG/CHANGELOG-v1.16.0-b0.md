# Flyte 1.16.0-b0 release notes

## What's Changed
* [propeller] tasklog bug fix by @Future-Outlier in https://github.com/flyteorg/flyte/pull/6399
* Make cloud events config type lowercase by @Sovietaced in https://github.com/flyteorg/flyte/pull/6408
* Allow setting `s3.endpoint` in flyte-core by @dwo in https://github.com/flyteorg/flyte/pull/6401
* auto-update contributors by @flyte-bot in https://github.com/flyteorg/flyte/pull/6297
* OIDC config to allow mismatched discovery / issuer by @ddl-ebrown in https://github.com/flyteorg/flyte/pull/5712
* Update Flyte components by @flyte-bot in https://github.com/flyteorg/flyte/pull/6423
* Update Flyte components by @flyte-bot in https://github.com/flyteorg/flyte/pull/6424
* [agent][flytectl][deployment] Rename agent to connector by @Future-Outlier in https://github.com/flyteorg/flyte/pull/6400
* [Docs] How to release Flyte by @Future-Outlier in https://github.com/flyteorg/flyte/pull/6425
* ci: Remove docs job from GitHub actions workflow by @pingsutw in https://github.com/flyteorg/flyte/pull/6432
* Allow configuring deployment annotations in Helm chart by @ppeerttu in https://github.com/flyteorg/flyte/pull/6385
* Add support for configuring storage from a secret to avoid exposing secrets by @Sovietaced in https://github.com/flyteorg/flyte/pull/6419
* Feat: Apply labels and annotations set via task decorator to pods and CR objects by @fg91 in https://github.com/flyteorg/flyte/pull/6421
* flyte-core: Default propeller `rawoutput-prefix` to use `storage.bucketName` by @dwo in https://github.com/flyteorg/flyte/pull/6433
* Add support for affinity & topologySpreadConstraints to webhook component by @Sovietaced in https://github.com/flyteorg/flyte/pull/6431
* [Docs] Rename container_image to image for improved UX by @JiangJiaWei1103 in https://github.com/flyteorg/flyte/pull/6211
* update idl w/ log context by @pvditt in https://github.com/flyteorg/flyte/pull/6443
* auto-update contributors by @flyte-bot in https://github.com/flyteorg/flyte/pull/6439
* Set MaxConcurrentStreams for flyteadmin grpc server by @pmahindrakar-oss in https://github.com/flyteorg/flyte/pull/6448
* Update "stale" GH Action by @davidmirror-ops in https://github.com/flyteorg/flyte/pull/6449
* Update `DemystifyFailure` to respect `NodeShutdown` (#724) by @hamersaw in https://github.com/flyteorg/flyte/pull/6452
* Add max-delta-timestamp configuration option to FlytePropeller ArrayNode by @hamersaw in https://github.com/flyteorg/flyte/pull/6453
* refactor: replace HasPrefix+TrimPrefix with CutPrefix by @studystill in https://github.com/flyteorg/flyte/pull/6456
* Allow rollout strategy to be configured in Helm charts by @Sovietaced in https://github.com/flyteorg/flyte/pull/6447
* Fix PR stale comment by @davidmirror-ops in https://github.com/flyteorg/flyte/pull/6458
* Treat kubelet NodeAffinity status.reason as retryable system error by @mhotan in https://github.com/flyteorg/flyte/pull/6461
* [flyteadmin][refactor] Make createMsgChan to private function by @Future-Outlier in https://github.com/flyteorg/flyte/pull/6467
* auto-update contributors by @flyte-bot in https://github.com/flyteorg/flyte/pull/6445
* Cleanup helm template since metrics path is not configurable by @Sovietaced in https://github.com/flyteorg/flyte/pull/6444
* Agent auto discovery conflicts with explicit task type mapping by @popojk in https://github.com/flyteorg/flyte/pull/6464
* Provide overrides for flyteadmin readiness and liveness probes by @pmahindrakar-oss in https://github.com/flyteorg/flyte/pull/6484
* Add version details to bug report template by @Sovietaced in https://github.com/flyteorg/flyte/pull/6479

## New Contributors
* @dwo made their first contribution in https://github.com/flyteorg/flyte/pull/6401
* @ppeerttu made their first contribution in https://github.com/flyteorg/flyte/pull/6385
* @studystill made their first contribution in https://github.com/flyteorg/flyte/pull/6456