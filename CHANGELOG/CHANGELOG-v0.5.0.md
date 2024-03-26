# Flyte v0.5.0

## Infrastructure
- Moved CI/CD to Github Actions
- Added end-to-end tests as part of the PR & master merges.
- Enable CI system to run on forks.

## Core Platform
- [Single Task Execution](https://docs.flyte.org/en/latest/user_guide/development_lifecycle/running_tasks.html) to enable registering and launching tasks outside the scope of a workflow to enable faster iteration and a more intuitive development workflow.
- [Run to completion](https://docs.flyte.org/en/latest/protos/docs/core/core.html#ref-flyteidl-core-workflowmetadata-onfailurepolicy) to enable workflows to continue executing even if one or more branches fail.
- Fixed retries for dynamically yielded nodes.
- PreAlpha Support for Raw container with FlyteCoPilot. (docs coming soon). [Sample Notebooks](https://github.com/lyft/flytekit/blob/master/sample-notebooks/raw-container-shell.ipynb). This makes it possible to run workflows with arbitrary containers

## Plugins
- Cleaned up Presto executor to send namespace as user to improve queueing
