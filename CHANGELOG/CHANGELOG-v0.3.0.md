# Flyte V0.3.0

## Changes since v0.2.0

### Core Platform
- Inclusion of an EKS archetype with Terraform files to make it easy to deploy and test in a cloud
- Better backoff behavior on Kubernetes resource creation failures


### Flytekit (SDK improvements)
- Dynamic Subworkflow and Launch Plan support
- Inclusion of queuing budget and interruptible flag in workflow and task decorators
- Ability to call non-Python based Spark jobs

### Flyte Console (UI)
- Added support for a custom banner to show live status messages

