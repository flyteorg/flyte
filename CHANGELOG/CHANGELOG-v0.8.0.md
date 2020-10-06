# Flyte v0.8.0

## Core Platform
- Metadata available for every task execution (First one: if the task was executed on interruptible nodes, useful when running on spot instances)
- Support for Cron schedules with offset (only on supported schedulers - e.g. Styx)
- Plugin overrides using Admin (docs coming soon)
- Custom model support in Sagemaker for single node and distributed training

## Console
- Full inputs and outputs from Execution data
- Timestamp rendering always in UTC

## Flytekit
- Dynamic overridable Spark configuration for spark tasks
- New Tensorflow task
- Formally removed python 2.x support. Supports only 3.6+
- Refactor - getting ready for enhanced auto-typed flytekit
