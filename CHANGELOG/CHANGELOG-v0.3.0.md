# Flyte V0.3.0

## Changes since v0.2.0

### Core Platform
- Improved differentiation of System and User errors with Deeper visibility
- Support for Presto queries (no-op client to be replaced with custom client)
- Support for optional interruptible Task executions - on Spot Like environments
- Better backoff behavior on Kubernetes resource creation failures
- Publish notifications to GCP pubsub
- Improved filtering logic in Service Layer
- Configurable Scheduler for K8s pod executions
- Spark Operator webhook support
- Performance improvements
- [Experimental] Inclusion of an EKS archetype with Terraform files to make it easy to deploy and test in a cloud


### Flytekit (SDK improvements)
- Dynamic Subworkflow and Launch Plan support
- Inclusion of queuing budget and interruptible flag in workflow and task decorators
- Ability to call non-Python based Spark jobs

### Flyte Console (UI)
- Added support for a custom banner to show live status messages
- Tweaks in execution list page

