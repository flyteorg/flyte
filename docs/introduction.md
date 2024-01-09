---
jupytext:
  formats: md:myst
  text_representation:
    extension: .md
    format_name: myst

# override the toc-determined page navigation order
next-page: getting_started/quickstart_guide
next-page-title: Quickstart guide
---

(getting_started_index)=

# Introduction to Flyte

Flyte is a workflow orchestrator that unifies machine learning, data engineering, and data analytics stacks for building robust and reliable applications. Flyte features:
* Reproducible, repeatable workflows
* Strongly typed interfaces
* Structured datasets to enable easy conversion of dataframes between types, and column-level type checking
* Easy movement of data between local and cloud storage
* Easy tracking of data lineages
* Built-in data and artifact visualization

For a full list of feature, see the [Flyte features page](https://flyte.org/features).

## Basic Flyte components

Flyte is made up of a user plane, control plane, and data plane.
* The **user plane** contains the elements you need to develop the code that will implement your application's directed acyclic graph (DAG). These elements are FlyteKit and Flytectl. Data scientists and machine learning engineers primarily work in the user plane.
* The **control plane** is part of the Flyte backend that is configured by platform engineers or others tasked with setting up computing infrastructure. It consists of FlyteConsole and FlyteAdmin, which serves as the main Flyte API to process requests from clients in the user plane. The control plane sends workflow execution requests to the data plane for execution, and stores information such as current and past running workflows, and provides that information upon request.
* The **data plane** is another part of the Flyte backend that contains FlytePropeller, the core engine of Flyte that executes workflows. FlytePropeller is designed as a [Kubernetes Controller](https://kubernetes.io/docs/concepts/architecture/controller/). The data plane sends status events back to the control plane so that information can be stored and surfaced to end users.

## Next steps

* To quickly try out Flyte on your machine, follow the {ref}`Quickstart guide <getting_started_quickstart_guide>`.
* To create a Flyte project that can be used to package workflow code for deployment to a Flyte cluster, see {ref}`"Getting started with workflow development" <getting_started_workflow_development>`.
* To set up a Flyte cluster, see the [Deployment documentation](https://docs.flyte.org/en/latest/deployment/index.html).

```{toctree}
:maxdepth: 1
:hidden:

Getting Started <self>
User Guide <userguide>
Tutorials <tutorials>
Concepts <https://docs.flyte.org/en/latest/concepts/basics.html>
Deployment <https://docs.flyte.org/en/latest/deployment/index.html>
API Reference <https://docs.flyte.org/en/latest/reference/index.html>
Community <https://docs.flyte.org/en/latest/community/index.html>
```

```{toctree}
:maxdepth: -1
:caption: Getting Started
:hidden:

Introduction to Flyte <self>
getting_started/quickstart_guide
getting_started/getting_started_with_workflow_development
getting_started/flyte_fundamentals
getting_started/core_use_cases
```

```{toctree}
:maxdepth: -1
:caption: User Guide
:hidden:

📖 User Guide <userguide>
🌳 Environment Setup <environment_setup>
🔤 Basics <auto_examples/basics/index>
⌨️ Data Types and IO <auto_examples/data_types_and_io/index>
🔮 Advanced Composition <auto_examples/advanced_composition/index>
🧩 Customizing Dependencies <auto_examples/customizing_dependencies/index>
🏡 Development Lifecycle <auto_examples/development_lifecycle/index>
⚗️ Testing <auto_examples/testing/index>
🚢 Productionizing <auto_examples/productionizing/index>
🏗 Extending <auto_examples/extending/index>
📝 Contributing <contribute>
```

```{toctree}
:maxdepth: -1
:caption: Tutorials
:hidden:

Tutorials <tutorials>
Model Training <ml_training>
feature_engineering
bioinformatics_examples
flyte_lab
```

```{toctree}
:maxdepth: -1
:caption: Integrations
:hidden:

Integrations <integrations>
auto_examples/airflow_plugin/index
auto_examples/athena_plugin/index
auto_examples/aws_batch_plugin/index
auto_examples/sagemaker_pytorch_plugin/index
auto_examples/sagemaker_training_plugin/index
auto_examples/bigquery_plugin/index
auto_examples/k8s_dask_plugin/index
auto_examples/databricks_plugin/index
auto_examples/dbt_plugin/index
auto_examples/dolt_plugin/index
auto_examples/duckdb_plugin/index
auto_examples/flyin_plugin/index
auto_examples/greatexpectations_plugin/index
auto_examples/hive_plugin/index
auto_examples/k8s_pod_plugin/index
auto_examples/mlflow_plugin/index
auto_examples/mmcloud_plugin/index
auto_examples/modin_plugin/index
auto_examples/kfmpi_plugin/index
auto_examples/onnx_plugin/index
auto_examples/papermill_plugin/index
auto_examples/pandera_plugin/index
auto_examples/kfpytorch_plugin/index
auto_examples/ray_plugin/index
auto_examples/sensor/index
auto_examples/snowflake_plugin/index
auto_examples/k8s_spark_plugin/index
auto_examples/sql_plugin/index
auto_examples/kftensorflow_plugin/index
auto_examples/whylogs_plugin/index
```

```{toctree}
:maxdepth: -1
:caption: Tags
:hidden:

_tags/tagsindex
```
