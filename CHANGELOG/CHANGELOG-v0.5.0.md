# Flyte V0.5.0

## Infrastructure
- Moved CI/CD to Github Actions
- Added end-to-end tests as part of the PR & master merges. 
- Enable CI system to run on forks.

## Core Platform
- [Run to completion](https://flyte.readthedocs.io/en/latest/user/features/on_failure_policy.html) to enable workflows to continue executing even if one or more branches fail.
- Fixed retries for dynamically yielded nodes.

## Plugins
- Cleaned up Presto executor to send namespace as user to improve queueing
