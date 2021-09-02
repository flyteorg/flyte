# Flyte v0.17.0

## Platform
1. Recovery Mode: Executions that fail due to external system failures (e.g. external system being down) can now be rerun in recovery mode ([flytectl --recover docs](https://docs.flyte.org/projects/flytectl/en/latest/gen/flytectl_create_execution.html)). It's also available in the UI:
![](https://i.imgur.com/hYYzkLK.png)


## Flytekit
1. Great Expectations Integration ([docs](https://docs.flyte.org/projects/cookbook/en/latest/auto/integrations/flytekit_plugins/greatexpectations/index.html#great-expectations)).
1. Access to durable blob stores (AWS/GCS/etc) are now pluggable.
1. Local task execution has been updated to also trigger the type engine.
1. Tasks that have cache=True should now be cached when running locally as well ([docs](https://docs.flyte.org/projects/cookbook/en/latest/auto/core/flyte_basics/task_cache.html#how-local-caching-works)).

Please see the [flytekit release](https://github.com/flyteorg/flytekit/releases/tag/v0.22.0) for the full list and more details.

## UI
1. Shiny new Graph UX. The graph rendering has been revamped to be more functional and accessible. More updates are coming for better visualization for nested executions and branches.
![](https://i.imgur.com/HTfuios.png)
1. JSON Validation for json-based types in the UI.
1. Enum support in UI
![](https://i.imgur.com/9bFZlei.png)


## FlyteCtl
1. `flytectl upgrade` to automatically upgrade itself ([docs](https://docs.flyte.org/projects/flytectl/en/latest/gen/flytectl_upgrade.html)).
1. `--dryRun` is available in most commands with server-side-effects to simulate the operations before committing any changes.

And various stabilization [fixes](https://github.com/flyteorg/flyte/milestone/17?closed=1)!
