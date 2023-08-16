# Databricks

```{eval-rst}
.. tags:: Spark, Integration, DistributedComputing, Data, Advanced
```

Flyte backend can be connected with Databricks service. Once enabled it can allow you to submit a spark job to Databricks platform.
This section will provide how to use the Databricks Plugin using flytekit python.

## Installation

The flytekit Databricks plugin is bundled into its Spark plugin, so to use, simply run the following:

```{eval-rst}
.. prompt:: bash

    pip install flytekitplugins-spark
```

## How to Build Your Dockerfile for Spark on Databricks

Using Spark on Databricks is extremely easy and provides full versioning using the custom-built Spark container. The built container can also execute regular Spark tasks.
For Spark, the image must use a base image built by Databricks and the workflow code must copy to `/databricks/driver`

```{literalinclude} ../../../examples/databricks_plugin/Dockerfile
:emphasize-lines: 20-32
:language: docker
:linenos: true
```

## Configuring the backend to get Databricks plugin working

1. Make sure to add "databricks" in `tasks.task-plugins.enabled-plugin` in [enabled_plugins.yaml](https://github.com/flyteorg/flyte/blob/master/deployment/sandbox/flyte_generated.yaml#L2296)
2. Add Databricks access token to Flytepropeller. [here](https://docs.databricks.com/administration-guide/access-control/tokens.html#enable-or-disable-token-based-authentication-for-the-workspace) to see more detail to create Databricks access token.

```bash
kubectl edit secret -n flyte flyte-propeller-auth
```

Configuration will be like below

```bash
apiVersion: v1
data:
  FLYTE_DATABRICKS_API_TOKEN: <ACCESS_TOKEN>
kind: Secret
metadata:
  annotations:
    meta.helm.sh/release-name: flyte
    meta.helm.sh/release-namespace: flyte
...
```

```{auto-examples-toc}
databricks_job
```
