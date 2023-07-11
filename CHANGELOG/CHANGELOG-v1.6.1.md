# Flyte v1.6.1 Patch Release

## Changes

### Admin - v1.1.100
* Inject user identifier to ExecutionSpec by @ByronHsu in https://github.com/flyteorg/flyteadmin/pull/549
* Fix flaky test by @eapolinario in https://github.com/flyteorg/flyteadmin/pull/563
* Add oauth http proxy for external server & Extract email from azure claim by @ByronHsu in https://github.com/flyteorg/flyteadmin/pull/553
* Remove single task execution default timeout by @hamersaw in https://github.com/flyteorg/flyteadmin/pull/564
* Revert conditional setting of SecurityContext when launching security context by @wild-endeavor in https://github.com/flyteorg/flyteadmin/pull/566

### Console - v1.8.2
* Export Flytedecks support for TLRO by @james-union in https://github.com/flyteorg/flyteconsole/pull/757
* fix: filter executions by version and name by @ursucarina in https://github.com/flyteorg/flyteconsole/pull/758
* fix: task recent runs should filter by version by @ursucarina in https://github.com/flyteorg/flyteconsole/pull/759
* Bug: Execution Page's back button returns Workflows route from Launch Plan route #patch by @FrankFlitton in https://github.com/flyteorg/flyteconsole/pull/760
* chore: add item when mapped task by @jsonporter in https://github.com/flyteorg/flyteconsole/pull/761
* Feature: Fullview Flyte Deck modal by @FrankFlitton in https://github.com/flyteorg/flyteconsole/pull/764


### Propeller - v1.1.90
* Add grpc plugin to loader.go by @pingsutw in https://github.com/flyteorg/flytepropeller/pull/562

