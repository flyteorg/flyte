# Flyte 1.16.0-b3
Primarily, changes around sys_ptrace and how container tasks work

## What's Changed
* Cleanup workflow store interface by @Sovietaced in https://github.com/flyteorg/flyte/pull/6485
* docs: Add clarification on required Kafka version field in cloud_even… by @daadc in https://github.com/flyteorg/flyte/pull/6502
* Add SignalWatcher in Copilot by @machichima in https://github.com/flyteorg/flyte/pull/6501
* add: nats as cloud events sender by @mattiadevivo in https://github.com/flyteorg/flyte/pull/6507
* [FIX] remove SYS_PTRACE and sharenamespace settings by @machichima in https://github.com/flyteorg/flyte/pull/6509
* [Core feature] Task retry support in Flyte Connectors by @machichima in https://github.com/flyteorg/flyte/pull/6486
* Bump github.com/go-viper/mapstructure/v2 from 2.1.0 to 2.3.0 in /boilerplate/flyte/golang_support_tools by @dependabot[bot] in https://github.com/flyteorg/flyte/pull/6512
* chore: fix some minor issues in the comments by @gopherorg in https://github.com/flyteorg/flyte/pull/6511
* Do not override compile-time podTemplate resources with config defaults if they're set by @punkerpunker in https://github.com/flyteorg/flyte/pull/6483
* Set service account in base ray pod spec, allow it to be overriden by pod template by @Sovietaced in https://github.com/flyteorg/flyte/pull/6514
* refactor: Adjust CoPilot init and Pod status logic by @pingsutw in https://github.com/flyteorg/flyte/pull/6523
* chore: fix inconsistent function name in comment by @jingchanglu in https://github.com/flyteorg/flyte/pull/6528
* Fix PodTemplate resource state pollution between workflow runs by @punkerpunker in https://github.com/flyteorg/flyte/pull/6530
* allow using container name in kubeflow plugin log links by @samhita-alla in https://github.com/flyteorg/flyte/pull/6524
* Make plugin metric registration thread safe by @Sovietaced in https://github.com/flyteorg/flyte/pull/6532
* fix: flyteadmin doesn't shutdown servers gracefully by @lowc1012 in https://github.com/flyteorg/flyte/pull/6289
* Enqueue correct work item in node execution context by @Sovietaced in https://github.com/flyteorg/flyte/pull/6526
* Fix workflow equality check by @Sovietaced in https://github.com/flyteorg/flyte/pull/6521

**Full Changelog**: https://github.com/flyteorg/flyte/compare/v1.16.0-b2...v1.16.0-b3

