# 0.18.1 Release ChangeLog

[Closed Issues](https://github.com/flyteorg/flyte/issues?q=is%3Aissue+milestone%3A0.18.1+is%3Aclosed)

## UX

### FlyteKit
See the [flytekit 0.24.0 release](https://github.com/flyteorg/flytekit/releases/tag/v0.24.0) for the full list of changes. Here are some of the highlights:

1. We've added the following models to the top-level of the flytekit package. The classes will not move from their original locations, but importing from the top level will always be safer as we promise not to break these.
    ```python
   - Annotations, AuthRole, Labels
   - WorkflowExecutionPhase
   - Literal, LiteralType, Scalar, BlobType, Blob, BlobMetadata
   ```

   Instead of `from flytekit.models.common import Labels, Annotations`, please now do
   ```python
   from flytekit import Labels, Annotations
   ```
2. Support for python pickle. Starting in this release, flytekit is going to pickle inputs and outputs for types which it doesn't have a specific transformer for. This brings a lot of more freedom in porting over code to flyte's model, since it won't force users to write a type transformer in order to use existing code. Keep in mind that all the caveats around pickling code apply in this case.
4. We added a cookiecutter template, simplifying the Getting started docs and also unlocking the path to cookiecutter templates for specific use-cases, e.g. pytorch-enabled samples, etc.
5. Faster installation in Apple M1 Macs. We're now requiring pyarrow 6.0, which contains prebuilt wheels for the M1.

### FlyteConsole
1. Added new UI for Workflow details including execution bar chart.
2. Added new bar chart user-selected filter for workflow executions.
3. Added new launch form controls ("Advanced Options")
4. Minor bug fixes

## System
1. Performance improvements for executions.
    1. Smaller workflow CRDs,
    2. Better handling of partial failures in large fanout scenarios,
3. All flyte containers now run as non-root users. [Docs](https://docs.flyte.org/en/latest/deployment/security/security.html) (Thanks @frsann)
4. Stability and bug fixes

## Documentation 
1. Plugin Setup [Docs](https://docs.flyte.org/en/latest/deployment/plugin_setup/index.html) (e.g. MPI, Tensorflow, Spark Operators, AWS Batch & Athena, Snowflake and Google BigQuery) 
