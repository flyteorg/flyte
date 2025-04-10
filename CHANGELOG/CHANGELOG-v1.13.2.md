# Flyte 1.13.2 Release Notes

## What's Changed
* Enable echo plugin by default by @pingsutw in https://github.com/flyteorg/flyte/pull/5679
* Do not emit execution id label by default in single binary by @eapolinario in https://github.com/flyteorg/flyte/pull/5704
* Using new offloaded metadata literal message for literal offloading by @pmahindrakar-oss in https://github.com/flyteorg/flyte/pull/5705
* Improve error message for MismatchingTypes by @pingsutw in https://github.com/flyteorg/flyte/pull/5639
* [Docs] Echo Task by @Future-Outlier in https://github.com/flyteorg/flyte/pull/5707
* Improve execution name readability by @wayner0628 in https://github.com/flyteorg/flyte/pull/5637
* Configure imagePullPolicy to be Always pull on flyte sandbox environment by @pmahindrakar-oss in https://github.com/flyteorg/flyte/pull/5709
* Set IsDefault to False for echo plugin by @pingsutw in https://github.com/flyteorg/flyte/pull/5713
* Move default execution name generation to flyteadmin by @wayner0628 in https://github.com/flyteorg/flyte/pull/5714
* Update helm/docs per changes in supported task discovery by @Sovietaced in https://github.com/flyteorg/flyte/pull/5694
* [flyteagent] Add Logging for Agent Supported Task Types by @Future-Outlier in https://github.com/flyteorg/flyte/pull/5718
* extend pod customization to include init containers by @samhita-alla in https://github.com/flyteorg/flyte/pull/5685
* Update "Try Serverless" language in Quickstart guide by @neverett in https://github.com/flyteorg/flyte/pull/5698
* Refactor flyteadmin to pass proto structs as pointers by @Sovietaced in https://github.com/flyteorg/flyte/pull/5717
* fix: Use deterministic execution names in scheduler by @pingsutw in https://github.com/flyteorg/flyte/pull/5724
* [flyteagent] Enable `connector-service` in every configuration  by @Future-Outlier in https://github.com/flyteorg/flyte/pull/5730
* [Docs][flyteagent] Remove `connector-service` config in flyte agent documentation by @Future-Outlier in https://github.com/flyteorg/flyte/pull/5731
* Fix flytectl returning the oldest workflow when using --latest flag by @RRap0so in https://github.com/flyteorg/flyte/pull/5716
* Remove explicit go toolchain versions by @ddl-ebrown in https://github.com/flyteorg/flyte/pull/5723
* Add listing api to stow storage by @bgedik in https://github.com/flyteorg/flyte/pull/5741
* Use latest upload/download-artifact action version by @Sovietaced in https://github.com/flyteorg/flyte/pull/5743
* Introduced SMTP notification by @robert-ulbrich-mercedes-benz in https://github.com/flyteorg/flyte/pull/5535
* Added literal offloading for array node map tasks by @pmahindrakar-oss in https://github.com/flyteorg/flyte/pull/5697
* pytorch object.inv moved by @wild-endeavor in https://github.com/flyteorg/flyte/pull/5755
* [RFC] Binary IDL With MessagePack Bytes by @Future-Outlier in https://github.com/flyteorg/flyte/pull/5742
* Rename literal offloading flags and enable offloading in local config by @eapolinario in https://github.com/flyteorg/flyte/pull/5754
* Add semver check for dev and beta version by @pmahindrakar-oss in https://github.com/flyteorg/flyte/pull/5757
* Add comments for semver regex by @pmahindrakar-oss in https://github.com/flyteorg/flyte/pull/5758
* Enable caching in flyte-core helmchart by @cpaulik in https://github.com/flyteorg/flyte/pull/5739
* [Flyte][1][IDL] Binary IDL With MessagePack by @Future-Outlier in https://github.com/flyteorg/flyte/pull/5751
* [Flyte][2][Literal Type For Scalar] Binary IDL With MessagePack by @Future-Outlier in https://github.com/flyteorg/flyte/pull/5761
* Backoff on etcd errors by @EngHabu in https://github.com/flyteorg/flyte/pull/5710
* [Docs][flyteagent] Update Databricks Agent Setup to V2 by @Future-Outlier in https://github.com/flyteorg/flyte/pull/5766
* [Docs][flyteagent] Improve Agent Secret Setup by @Future-Outlier in https://github.com/flyteorg/flyte/pull/5765
* [RFC] Offloaded Raw Literals by @wild-endeavor in https://github.com/flyteorg/flyte/pull/5103
* Use pluggable clock for auto refresh cache and make unit tests fast by @Sovietaced in https://github.com/flyteorg/flyte/pull/5767
* [Docs] Flyte Deck example by @Future-Outlier in https://github.com/flyteorg/flyte/pull/5777
* Update ImageSpec documentation by @pingsutw in https://github.com/flyteorg/flyte/pull/5748
* Revert "Improve execution name readability" by @pingsutw in https://github.com/flyteorg/flyte/pull/5740
* [Docs] Flyte Deck example V2 by @Future-Outlier in https://github.com/flyteorg/flyte/pull/5781
* [flytepropeller][flyteadmin] Compiler unknown literal type error handling by @Future-Outlier in https://github.com/flyteorg/flyte/pull/5651
* [Temporary fix] Pin flytekit version for single binary builds by @pmahindrakar-oss in https://github.com/flyteorg/flyte/pull/5798
* Fix propeller crash when inferring literal type for an offloaded literal by @pmahindrakar-oss in https://github.com/flyteorg/flyte/pull/5771

## New Contributors
* @wayner0628 made their first contribution in https://github.com/flyteorg/flyte/pull/5637
* @robert-ulbrich-mercedes-benz made their first contribution in https://github.com/flyteorg/flyte/pull/5535
* @cpaulik made their first contribution in https://github.com/flyteorg/flyte/pull/5739

**Full Changelog**: https://github.com/flyteorg/flyte/compare/v1.13.1...v1.13.2
