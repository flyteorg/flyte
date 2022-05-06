# v1.0.1 Release ChangeLog

## System
1. [Bugfix](https://github.com/flyteorg/flyte/issues/2424) Addresses the regression to allow empty default complex values for workflow parameters
1. [Bugfix](https://github.com/flyteorg/flyte/issues/2448) for flyteadmin not respecting fallback substitution for execution spec parameters (e.g. labels, annotations, security context, raw output data config)
1. [Bugfix](https://github.com/flyteorg/flyteadmin/pull/416) Bump stow version to fix issue with GoogleAccessId

## Flytekit
1. [Bugfix](https://github.com/flyteorg/flyte/issues/2374) Task chaining without the use of the `create_node` utility function
1. [Bugfix](https://github.com/flyteorg/flytekit/pull/983/) Improved validation of image config objects
