# Flyte v0.6.0

## Core Platform
- Catalog and Caching information propagated to FlyteAdmin. This makes it possible to weave all the dependencies using the FlyteAdmin API. this also provides visibility when the Cache was populated,
  cache was hit and cache was disabled.
- More system visibility metrics

## Console
- Better layout of table view
- Ability to launch Blob objects
- Better error message display in execution page
- In multi-cluster mode, Console shows the cluster in which the execution landed
- Many bug fixes.

## Plugins
- Sagemaker Support for Builtin Algorithms with simplified API

## Flytekit
- Support for Spark 3
- Papermill based notebook task support for - Python and Spark (basic input and output types)
- Simplified Sagemaker Interface for leveraging Builtin Algorithms
