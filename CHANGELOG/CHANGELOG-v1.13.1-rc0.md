# Flyte v1.13.1-rc0 Release Notes

## Changelog
* chore: update runllm widget configuration by @agiron123 in https://github.com/flyteorg/flyte/pull/5530
* Use jsonpb AllowUnknownFields everywhere by @andrewwdye in https://github.com/flyteorg/flyte/pull/5521
* [flyteadmin] Use WithContext in all DB calls by @Sovietaced in https://github.com/flyteorg/flyte/pull/5538
* add new project state: SYSTEM_ARCHIVED by @troychiu in https://github.com/flyteorg/flyte/pull/5544
* Pryce/doc 434 clarify how code is pushed into a given image during pyflyte by @pryce-turner in https://github.com/flyteorg/flyte/pull/5548
* Increase more memory limits in flyteagent by @Future-Outlier in https://github.com/flyteorg/flyte/pull/5550
* Updated map task information to indicate array node is now the default and optional return type by @pryce-turner in https://github.com/flyteorg/flyte/pull/5561
* reverted mockery-v2 on ExecutionContext by @hamersaw in https://github.com/flyteorg/flyte/pull/5562
* Fix issues with helm chart release process by @Sovietaced in https://github.com/flyteorg/flyte/pull/5560
* Add FlyteDirectory to file_types.rst template by @ppiegaze in https://github.com/flyteorg/flyte/pull/5564
* Fully populate Abort task event fields by @va6996 in https://github.com/flyteorg/flyte/pull/5551
* [fix] Add blob typechecker by @ddl-rliu in https://github.com/flyteorg/flyte/pull/5519
* Refactor echo plugin by @pingsutw in https://github.com/flyteorg/flyte/pull/5565
* [Bug] fix ArrayNode state's TaskPhase reset by @pvditt in https://github.com/flyteorg/flyte/pull/5451
* Remove confusing prometheus configuration options in helm charts  by @Sovietaced in https://github.com/flyteorg/flyte/pull/5549
* [Housekeeping] Bump Go version to 1.22 by @lowc1012 in https://github.com/flyteorg/flyte/pull/5032
* Fix typos by @omahs in https://github.com/flyteorg/flyte/pull/5571
* Respect original task definition retry strategy for single task executions by @katrogan in https://github.com/flyteorg/flyte/pull/5577
* Clarify the support for the Java/Scala SDK in the docs by @eapolinario in https://github.com/flyteorg/flyte/pull/5582
* Fix spelling issues by @nnsW3 in https://github.com/flyteorg/flyte/pull/5580
* Give flyte binary cluster role permission to create service accounts by @shreyas44 in https://github.com/flyteorg/flyte/pull/5579
* Simplify single task retry strategy check by @eapolinario in https://github.com/flyteorg/flyte/pull/5584
* Fix failures in `generate_helm` CI check by @eapolinario in https://github.com/flyteorg/flyte/pull/5587
* Update GPU docs by @davidmirror-ops in https://github.com/flyteorg/flyte/pull/5515
* Update azblob 1.1.0 -> 1.4.0 / azcore 1.7.2 -> 1.13.0 by @ddl-ebrown in https://github.com/flyteorg/flyte/pull/5590
* Bump github.com/go-jose/go-jose/v3 from 3.0.0 to 3.0.3 in /flyteadmin by @dependabot in https://github.com/flyteorg/flyte/pull/5591
* add execution mode to ArrayNode proto by @pvditt in https://github.com/flyteorg/flyte/pull/5512
* Fix incorrect YAML for unpartitoned GPU by @davidmirror-ops in https://github.com/flyteorg/flyte/pull/5595
* Another YAML fix by @davidmirror-ops in https://github.com/flyteorg/flyte/pull/5596
* DOC-431 Document pyflyte option --overwrite-cache by @ppiegaze in https://github.com/flyteorg/flyte/pull/5567
* Upgrade docker dependency to address vulnerability by @katrogan in https://github.com/flyteorg/flyte/pull/5614
* Support offloading workflow CRD inputs by @katrogan in https://github.com/flyteorg/flyte/pull/5609
* [flyteadmin] Refactor panic recovery into middleware by @Sovietaced in https://github.com/flyteorg/flyte/pull/5546
* Snowflake agent Doc by @Future-Outlier in https://github.com/flyteorg/flyte/pull/5620
* [flytepropeller][compiler] Error Handling when Type is not found by @Future-Outlier in https://github.com/flyteorg/flyte/pull/5612
* Fix nil pointer when task plugin load returns error by @Sovietaced in https://github.com/flyteorg/flyte/pull/5622
* Log stack trace when refresh cache sync recovers from panic by @Sovietaced in https://github.com/flyteorg/flyte/pull/5623
* [Doc] Fix snowflake agent secret documentation error by @Future-Outlier in https://github.com/flyteorg/flyte/pull/5626
* [Doc] Explain how Agent Secret Works by @Future-Outlier in https://github.com/flyteorg/flyte/pull/5625
* Fix typo in execution manager by @ddl-rliu in https://github.com/flyteorg/flyte/pull/5619
* Amend Admin to use grpc message size by @wild-endeavor in https://github.com/flyteorg/flyte/pull/5628
* [Docs] document the process of setting ttl for a ray cluster by @pingsutw in https://github.com/flyteorg/flyte/pull/5636
* Add CustomHeaderMatcher to pass additional headers by @andrewwdye in https://github.com/flyteorg/flyte/pull/5563
* Turn flyteidl and flytectl releases into manual gh workflows by @eapolinario in https://github.com/flyteorg/flyte/pull/5635
* docs: fix typo by @cratiu222 in https://github.com/flyteorg/flyte/pull/5643
* Use enable_deck=True in docs by @thomasjpfan in https://github.com/flyteorg/flyte/pull/5645
* Fix flyteidl release  checkout all tags by @eapolinario in https://github.com/flyteorg/flyte/pull/5646
* Install pyarrow in sandbox functional tests by @eapolinario in https://github.com/flyteorg/flyte/pull/5647
* docs: add documentation for configuring notifications in GCP by @desihsu in https://github.com/flyteorg/flyte/pull/5545
* Correct "sucessfile" to "successfile" by @shengyu7697 in https://github.com/flyteorg/flyte/pull/5652
* Fix ordering for custom template values in cluster resource controller by @katrogan in https://github.com/flyteorg/flyte/pull/5648
* Don't error when attempting to trigger schedules for inactive projects by @katrogan in https://github.com/flyteorg/flyte/pull/5649

## New Contributors
* @omahs made their first contribution in https://github.com/flyteorg/flyte/pull/5571
* @nnsW3 made their first contribution in https://github.com/flyteorg/flyte/pull/5580
* @shreyas44 made their first contribution in https://github.com/flyteorg/flyte/pull/5579
* @cratiu222 made their first contribution in https://github.com/flyteorg/flyte/pull/5643
* @desihsu made their first contribution in https://github.com/flyteorg/flyte/pull/5545
* @shengyu7697 made their first contribution in https://github.com/flyteorg/flyte/pull/5652

**Full Changelog**: https://github.com/flyteorg/flyte/compare/v1.13.0...v1.13.1-rc0
