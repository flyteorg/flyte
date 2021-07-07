# Flyte v0.15.0

## Platform
- Enum type support throughout the system
- Stabilization improvements for conditionals
- Support AWS Secrets Manager as a source for secrets injector
- Support Map tasks over Pod tasks
- Support max parallelism to limit how many nodes can be allowed to run in parallel.
- Add Athena flytekit plugin and examples
- Add BigQuery plugin

## Flytekit
 - Support Schema of Dataclasses
 - Support node resource overrides

Please see the [flytekit release](https://github.com/flyteorg/flytekit/releases/tag/v0.20.0) for the full list and more details.

## flytectl
 - `flytectl sandbox start` to start a sandbox cluster locally.
 - `flytectl get workflow .... -o dot` to visualize a workflow graph locally.
 - Add Bash completion support
