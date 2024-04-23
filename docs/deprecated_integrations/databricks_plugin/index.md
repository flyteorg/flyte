# Databricks plugin

```{eval-rst}
.. tags:: Spark, Integration, DistributedComputing, Data, Advanced
```

```{note}

This is a legacy implementation of the Databricks integration. We recommend using the {ref}`Databricks agent <databricks_agent>` instead.

```

Flyte can be integrated with the [Databricks](https://www.databricks.com/) service,
enabling you to submit Spark jobs to the Databricks platform.

## Installation

The Databricks plugin comes bundled with the Spark plugin. To install the Spark plugin, run the following command:

```
pip install flytekitplugins-spark

```

## Flyte deployment configuration

To run the Databricks plugin on a Flyte cluster, you must configure it in your Flyte deployment. For more information, see the
{std:ref}`Databricks plugin setup guide <flyte:deployment-plugin-setup-webapi-databricks>`.

## Example usage

For a usage example, see the {doc}`Databricks plugin example <databricks_plugin_example>` page.

### Run the example on the Flyte cluster

To run the provided example on a Flyte cluster, use the following command:

```
pyflyte run --remote \
  --image ghcr.io/flyteorg/flytecookbook:databricks_plugin-latest \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/databricks_plugin/databricks_plugin/databricks_job.py \
  my_databricks_job
```

Using Spark on Databricks allows comprehensive versioning through a
custom-built Spark container. This container also facilitates the execution of standard Spark tasks.

To use Spark, the image should employ a base image provided by Databricks,
and the workflow code must be copied to `/databricks/driver`.

```{literalinclude} ../../../examples/databricks_plugin/Dockerfile
:language: docker
:emphasize-lines: 1,7-8,20
```


```{toctree}
:maxdepth: -1
:hidden:

databricks_plugin_example
```
