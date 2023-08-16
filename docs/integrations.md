(integrations)=

# Integrations

Flyte is designed to be highly extensible and can be customized in multiple ways.

```{note}
Want to contribute an example? Check out the {doc}`Example Contribution Guide <contribute>`.
```

## Flytekit Plugins

Flytekit plugins are simple plugins that can be implemented purely in python, unit tested locally and allow extending
Flytekit functionality. These plugins can be anything and for comparison can be thought of like
[Airflow Operators](https://airflow.apache.org/docs/apache-airflow/stable/howto/operator/index.html).

```{list-table}
:header-rows: 0
:widths: 20 30

* - {doc}`SQL <auto_examples/sql_plugin/index>`
  - Execute SQL queries as tasks.
* - {doc}`Great Expectations <auto_examples/greatexpectations_plugin/index>`
  - Validate data with `great_expectations`.
* - {doc}`Papermill <auto_examples/papermill_plugin/index>`
  - Execute Jupyter Notebooks with `papermill`.
* - {doc}`Pandera <auto_examples/pandera_plugin/index>`
  - Validate pandas dataframes with `pandera`.
* - {doc}`Modin <auto_examples/modin_plugin/index>`
  - Scale pandas workflows with `modin`.
* - {doc}`Dolt <auto_examples/dolt_plugin/index>`
  - Version your SQL database with `dolt`.
* - {doc}`DBT <auto_examples/dbt_plugin/index>`
  - Run and test your `dbt` pipelines in Flyte.
* - {doc}`WhyLogs <auto_examples/whylogs_plugin/index>`
  - `whylogs`: the open standard for data logging.
* - {doc}`MLFlow <auto_examples/mlflow_plugin/index>`
  - `mlflow`: the open standard for model tracking.
* - {doc}`ONNX <auto_examples/onnx_plugin/index>`
  - Convert ML models to ONNX models seamlessly.
* - {doc}`DuckDB <auto_examples/duckdb_plugin/index>`
  - Run analytical queries using DuckDB.
```

:::{dropdown} {fa}`info-circle` Using flytekit plugins
:animate: fade-in-slide-down

Data is automatically marshalled and unmarshalled in and out of the plugin. Users should mostly implement the
{py:class}`~flytekit.core.base_task.PythonTask` API defined in Flytekit.

Flytekit Plugins are lazily loaded and can be released independently like libraries. We follow a convention to name the
plugin like `flytekitplugins-*`, where * indicates the package to be integrated into Flytekit. For example
`flytekitplugins-papermill` enables users to author Flytekit tasks using [Papermill](https://papermill.readthedocs.io/en/latest/).

You can find the plugins maintained by the core Flyte team [here](https://github.com/flyteorg/flytekit/tree/master/plugins).
:::


## Native Backend Plugins

Native Backend Plugins are the plugins that can be executed without any external service dependencies because the compute is
orchestrated by Flyte itself, within its provisioned Kubernetes clusters.

```{list-table}
:header-rows: 0
:widths: 20 30

* - {doc}`K8s Pods <auto_examples/k8s_pod_plugin/index>`
  - Execute K8s pods for arbitrary workloads.
* - {doc}`K8s Cluster Dask Jobs <auto_examples/k8s_dask_plugin/index>`
  - Run Dask jobs on a K8s Cluster.
* - {doc}`K8s Cluster Spark Jobs <auto_examples/k8s_spark_plugin/index>`
  - Run Spark jobs on a K8s Cluster.
* - {doc}`Kubeflow PyTorch <auto_examples/kfpytorch_plugin/index>`
  - Run distributed PyTorch training jobs using `Kubeflow`.
* - {doc}`Kubeflow TensorFlow <auto_examples/kftensorflow_plugin/index>`
  - Run distributed TensorFlow training jobs using `Kubeflow`.
* - {doc}`MPI Operator <auto_examples/kfmpi_plugin/index>`
  - Run distributed deep learning training jobs using Horovod and MPI.
* - {doc}`Ray Task <auto_examples/ray_plugin/index>`
  - Run Ray jobs on a K8s Cluster.
```

(external_service_backend_plugins)=


## External Service Backend Plugins

As the term suggests, external service backend plugins relies on external services like
[AWS Sagemaker](https://aws.amazon.com/sagemaker),
[Hive](https://docs.qubole.com/en/latest/user-guide/engines/hive/index.html) or
[Snowflake](https://www.snowflake.com/) for handling the workload defined in
the Flyte task that use the respective plugin.

```{list-table}
:header-rows: 0
:widths: 20 30

* - {doc}`AWS Sagemaker: Model Training <auto_examples/sagemaker_training_plugin/index>`
  - Train models with built-in or define your own custom algorithms.
* - {doc}`AWS Sagemaker: Pytorch Training <auto_examples/sagemaker_pytorch_plugin/index>`
  - Train Pytorch models using Sagemaker, with support for distributed training.
* - {doc}`AWS Athena <auto_examples/athena_plugin/index>`
  - Execute queries using AWS Athena
* - {doc}`AWS Batch <auto_examples/aws_batch_plugin/index>`
  - Running tasks and workflows on AWS batch service
* - {doc}`Hive <auto_examples/hive_plugin/index>`
  - Run Hive jobs in your workflows.
* - {doc}`Snowflake <auto_examples/snowflake_plugin/index>`
  - Run Snowflake jobs in your workflows.
* - {doc}`Databricks <auto_examples/databricks_plugin/index>`
  - Run Databricks jobs in your workflows.
* - {doc}`BigQuery <auto_examples/bigquery_plugin/index>`
  - Run BigQuery jobs in your workflows.
```

(enable-backend-plugins)=

::::{dropdown} {fa}`info-circle` Enabling Backend Plugins
:animate: fade-in-slide-down

To enable a backend plugin you have to add the `ID` of the plugin to the enabled plugins list. The `enabled-plugins` is available under the `tasks > task-plugins` section of FlytePropeller's configuration.
The plugin configuration structure is defined [here](https://pkg.go.dev/github.com/flyteorg/flytepropeller@v0.6.1/pkg/controller/nodes/task/config#TaskPluginConfig). An example of the config follows,

:::{rli} https://raw.githubusercontent.com/flyteorg/flyte/master/kustomize/overlays/sandbox/flyte/config/propeller/enabled_plugins.yaml
:language: yaml
:::

**Finding the `ID` of the Backend Plugin**

This is a little tricky since you have to look at the source code of the plugin to figure out the `ID`. In the case of Spark, for example, the value of `ID` is used [here](https://github.com/flyteorg/flyteplugins/blob/v0.5.25/go/tasks/plugins/k8s/spark/spark.go#L424) here, defined as [spark](https://github.com/flyteorg/flyteplugins/blob/v0.5.25/go/tasks/plugins/k8s/spark/spark.go#L41).

**Enabling a Specific Backend Plugin in Your Own Kustomize Generator**

Flyte uses Kustomize to generate the the deployment configuration which can be leveraged to [kustomize your own deployment](https://github.com/flyteorg/flyte/tree/master/kustomize).

::::


## Custom Container Tasks

Because Flyte uses executable docker containers as the smallest unit of compute, you can write custom tasks with the
{py:class}`flytekit.ContainerTask` via the [flytekit](https://github.com/flyteorg/flytekit) SDK.

```{list-table}
:header-rows: 0
:widths: 20 30

* - {doc}`Raw Container Tasks <auto_examples/containerization/raw_container>`
  - Execute arbitrary containers: You can write C++ code, bash scripts and any containerized program.
```

## SDKs for Writing Tasks and Workflows

The {ref}`community <community>` would love to help you with your own ideas of building a new SDK. Currently the available SDKs are:

```{list-table}
:header-rows: 0
:widths: 20 30

* - [flytekit](https://flytekit.readthedocs.io)
  - The Python SDK for Flyte.
* - [flytekit-java](https://github.com/spotify/flytekit-java)
  - The Java/Scala SDK for Flyte.
```

## Flyte Operators

Flyte can be integrated with other orchestrators to help you leverage Flyte's
constructs natively within other orchestration tools.

```{list-table}
:header-rows: 0
:widths: 20 30

* - {doc}`Airflow <auto_examples/airflow_plugin/index>`
  - Trigger Flyte executions from Airflow.
```

```{toctree}
:maxdepth: -1
:caption: Integrations
:hidden:

auto_examples/sql_plugin/index
auto_examples/greatexpectations_plugin/index
auto_examples/papermill_plugin/index
auto_examples/pandera_plugin/index
auto_examples/modin_plugin/index
auto_examples/dolt_plugin/index
auto_examples/dbt_plugin/index
auto_examples/whylogs_plugin/index
auto_examples/mlflow_plugin/index
auto_examples/onnx_plugin/index
auto_examples/duckdb_plugin/index
auto_examples/k8s_pod_plugin/index
auto_examples/k8s_dask_plugin/index
auto_examples/k8s_spark_plugin/index
auto_examples/kfpytorch_plugin/index
auto_examples/kftensorflow_plugin/index
auto_examples/kfmpi_plugin/index
auto_examples/ray_plugin/index
auto_examples/sagemaker_training_plugin/index
auto_examples/sagemaker_pytorch_plugin/index
auto_examples/athena_plugin/index
auto_examples/aws_batch_plugin/index
auto_examples/hive_plugin/index
auto_examples/snowflake_plugin/index
auto_examples/databricks_plugin/index
auto_examples/bigquery_plugin/index
auto_examples/airflow_plugin/index
```
