# Flyte v1.8.0 Release

## Flytekit
* Support configuring pip index url with the image spec by @yini7777 in https://github.com/flyteorg/flytekit/pull/1692
* Improve error message for pyflyte-fast-execute by @pingsutw in https://github.com/flyteorg/flytekit/pull/1686
* Feat: Allow disabling rich tracebacks via env var by @fg91 in https://github.com/flyteorg/flytekit/pull/1695
* Fix task type version in pytorch and mpi task plugin  by @yubofredwang in https://github.com/flyteorg/flytekit/pull/1690
* get_transformer returns pickle transformer if type is unsupported  by @pingsutw in https://github.com/flyteorg/flytekit/pull/1694
* Add Cuda to ImageSpec by @pingsutw in https://github.com/flyteorg/flytekit/pull/1688
* Feat: Warn when doing local torch elastic training with nnodes > 1 by @fg91 in https://github.com/flyteorg/flytekit/pull/1697
* Fix config of user facing execution parameters in spawning elastic tasks by @fg91 in https://github.com/flyteorg/flytekit/pull/1677
* Union | optional return types supported by @kumare3 in https://github.com/flyteorg/flytekit/pull/1703
* Fail registration if output isn't Optional when using map tasks with min_success_ratio < 1 by @pingsutw in https://github.com/flyteorg/flytekit/pull/1709
* Add bigquery project and location to metadata by @pingsutw in https://github.com/flyteorg/flytekit/pull/1702
* Generate decks at local execution by @pingsutw in https://github.com/flyteorg/flytekit/pull/1699
* Improve task type hint by @pingsutw in https://github.com/flyteorg/flytekit/pull/1711
* Detect upstream nodes from container nested promises by @wild-endeavor in https://github.com/flyteorg/flytekit/pull/1707
* Add Iterator Transformer by @pingsutw in https://github.com/flyteorg/flytekit/pull/1710
* Always a default image and streamline image handling serialize vs package by @wild-endeavor in https://github.com/flyteorg/flytekit/pull/1610
* Remove ensure literal by @wild-endeavor in https://github.com/flyteorg/flytekit/pull/1706
* Fix a bug in gx integration by @XinEDprob in https://github.com/flyteorg/flytekit/pull/1675
* Update pythonbuild.yml: add hugging face plugin to CI by @cosmicBboy in https://github.com/flyteorg/flytekit/pull/1684
* add imagespec cache by @RichhLi in https://github.com/flyteorg/flytekit/pull/1717
* Add requirements to imageSpec by @pingsutw in https://github.com/flyteorg/flytekit/pull/1698
* Memory reduction change by @wild-endeavor in https://github.com/flyteorg/flytekit/pull/1716
* use getattr to access task_def.disable_deck in entrypoint by @cosmicBboy in https://github.com/flyteorg/flytekit/pull/1724
* Add a couple tests by @wild-endeavor in https://github.com/flyteorg/flytekit/pull/1722
* Csvtransform by @ChungYujoyce in https://github.com/flyteorg/flytekit/pull/1671

## Flyteadmin
* Add a user-specifiable root to upload link request by @wild-endeavor in https://github.com/flyteorg/flyteadmin/pull/577
* Bump propeller version by @pingsutw in https://github.com/flyteorg/flyteadmin/pull/580
* Upgrade go 1.19 in Dockerfile by @hamersaw in https://github.com/flyteorg/flyteadmin/pull/581

## Flytepropeller
* Update flyteplugins to v1.0.67 by @bstadlbauer in https://github.com/flyteorg/flytepropeller/pull/575
* Correctly validating error code on aborting terminal launchplans by @hamersaw in https://github.com/flyteorg/flytepropeller/pull/574
* Use GetExecutionData instead by @EngHabu in https://github.com/flyteorg/flytepropeller/pull/573
* Propagating environment variables through launchplans by @hamersaw in https://github.com/flyteorg/flytepropeller/pull/576
* upgrade dockerfile go to 1.19 by @hamersaw in https://github.com/flyteorg/flytepropeller/pull/580
* Bump flyteplugins version by @pingsutw in https://github.com/flyteorg/flytepropeller/pull/581
* Support for cross-project secrets for GCP by @jeevb in https://github.com/flyteorg/flytepropeller/pull/582
* Added IsFailurePermanent flag on DynamicTaskStatus by @hamersaw in https://github.com/flyteorg/flytepropeller/pull/567

## Flyteplugins
* Don't add master replica log link when doing elastic pytorch training by @fg91 in https://github.com/flyteorg/flyteplugins/pull/356
* [Bigquery] Add support for impersonation of GSA bound to task's KSA by @jeevb in https://github.com/flyteorg/flyteplugins/pull/355
* Fix initial dask job state by @bstadlbauer in https://github.com/flyteorg/flyteplugins/pull/357
* Feat: Add pod start and finish time in RFC3339 time format to logging link templating variables #minor by @fg91 in https://github.com/flyteorg/flyteplugins/pull/360
* fix v1 pytorch job plugin with elastic policy by @yubofredwang in https://github.com/flyteorg/flyteplugins/pull/359
* Allow using pod start time in kubeflow plugin log links by @fg91 in https://github.com/flyteorg/flyteplugins/pull/362
* Pass location to the BigQuery request by @pingsutw in https://github.com/flyteorg/flyteplugins/pull/365
* Fix map task cache misses by @bstadlbauer in https://github.com/flyteorg/flyteplugins/pull/363
* Switch to official dask operator by @bstadlbauer in https://github.com/flyteorg/flyteplugins/pull/366
* Fix duplicate env vars in container by @hamersaw in https://github.com/flyteorg/flyteplugins/pull/358
* Set scheduler restart policy to Always by @bstadlbauer in https://github.com/flyteorg/flyteplugins/pull/367

## Flyteconsole
* Fix launch plan icon https://github.com/flyteorg/flyteconsole/pull/777
