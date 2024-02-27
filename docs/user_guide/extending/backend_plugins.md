---
jupytext:
  cell_metadata_filter: all
  formats: md:myst
  main_language: python
  notebook_metadata_filter: all
  text_representation:
    extension: .md
    format_name: myst
    format_version: 0.13
    jupytext_version: 1.16.1
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

%% [markdown]
(extend-plugin-flyte-backend)=

# Backend plugins

```{eval-rst}
.. tags:: Extensibility, Contribute, Intermediate
```

This guide will take you through the why and how of writing a backend plugin for
Flyte.

To recap, here are a few examples of why you would want to implement a backend plugin:

1. We want to add a new capability to the Flyte Platform, for example we might want to:
   - Talk to a new service like  AWS Sagemaker, Snowflake, Redshift, Athena, BigQuery, etc.
   - Orchestrate a set of containers in a new way like Spark, Flink, Distributed
     training on Kubernetes (usually using a Kubernetes operator).
   - Use a new container orchestration engine like AWS Batch/ECS, Hashicorp' Nomad
   - Use a completely new runtime like AWS Lambda, KNative, etc.
3. You want to retain the capability to update the plugin implementation and roll
   out new changes and fixes without affecting the users code or requiring them to update
   versions of their plugins.
4. You want the same plugin to be accessible across multiple language SDK's.

```{note}
Talking to a new service can be done using flytekit extensions and usually is the better way to get started. But, once matured, most of these extensions are better to be migrated to the backend. For the rest of the cases, it is possible to extend flytekit to achieve these scenarios, but this is less desirable, because of the associated overhead of first launching a container that launches these jobs downstream.
```

## Basics

In this section we'll go through the components of a backend plugin using the {ref}`K8s Spark plugin<plugins-spark-k8s>` as a reference. A Flyte backend extension consists of 3 parts: interface
specification, `flytekit` plugin implementation, and `flytepropeller` plugin implementation.

### Interface specification

Usually Flyte extensions need information that is not covered by a {std:ref}`Flyte TaskTemplate <protos/docs/core/core:tasktemplate>`. The TaskTemplate consists of a
the interface, task_type identifier, some metadata and other fields.

```{note}
An important field to note here is {std:ref}`custom <protos/docs/core/core:tasktemplate>`. The custom field is essentially an unstructured JSON. This makes it possible to extend a task-template beyond the default supported targets {std:ref}`container <protos/docs/core/core:tasktemplate>`.

The motivation of the `custom`` field is to marshal a JSON structure that specifies information beyond what a regular TaskTemplate can capture. The actual structure of the JSON is known only to the implemented backend-plugin and the SDK components. The core Flyte platform, does not understand of look into the specifics of this structure.
```

It is highly recommended to use an interface definition language like Protobuf, OpenAPISpec etc to declare specify the structure of the JSON. From here, on we refer to this as the **Plugin Specification**.

```{note}
For Spark we decided to use Protobuf to specify the plugin as can be seen [here](https://github.com/flyteorg/flyteidl/blob/master/protos/flyteidl/plugins/spark.proto). Note it isn't necessary to have the Plugin structure specified in `flyteidl`, but we do it for simplicity, ease of maintenance alongside the core platform, and convenience leveraging existing tooling to generate code for protobuf.
```

### Flytekit plugin implementation

Now that you have a specification, we have to implement a method to generate this new TaskTemplate, with the special custom field. Also, this is where the UX design comes into play. You want to write the best possible interface in the SDK that users are delighted to use. The end goal is to create the TaskTemplate with the Custom field populated with the actual JSON structure.

We will currently refer to the Python `flytekit` SDK as an example for extending and
implementing the SDK.

The SDK task should be implemented as an extension of {py:class}`flytekit.core.base_task.PythonTask`, or more commonly {py:class}`flytekit.PythonFunctionTask`.
In the case of Spark, we extend the {py:class}`flytekit.PythonFunctionTask`, as shown [here](https://github.com/flyteorg/flytekit/blob/master/plugins/flytekit-spark/flytekitplugins/spark/task.py#L77-L123).

The `SparkTask` is implemented as a regular flytekit plugin, with one exception: the `custom` field is now actually the `SparkJob` protocol buffer. When serializing a task, `flytekit` base classes will automatically invoke the [`get_custom` method](https://github.com/flyteorg/flytekit/blob/c02075d472b5587d199630bcfc7f9937673c6a0e/flytekit/core/base_task.py#L255).

### FlytePropeller backend plugin

The backend plugin is where the actual logic of the execution is implemented. The backend plugin uses the **Flyte PluginMachinery** interface to implement a plugin which can be one of the following supported types:

1. [Kubernetes operator Plugin](https://pkg.go.dev/github.com/lyft/flyteplugins@v0.5.26/go/tasks/pluginmachinery/k8s#Plugin): The demo in the video below shows two examples of K8s backend plugins: flytekit `Athena` & `Spark`, and Flyte K8s `Pod` & `Spark`.

   ```{youtube} oK2RGQuP94k
   ```

2. **A Web API plugin:** [Async](https://pkg.go.dev/github.com/lyft/flyteplugins@v0.5.26/go/tasks/pluginmachinery/webapi#AsyncPlugin) or [Sync](https://pkg.go.dev/github.com/lyft/flyteplugins@v0.5.26/go/tasks/pluginmachinery/webapi#SyncPlugin).
3. [Core Plugin](https://pkg.go.dev/github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core#Plugin): if none of the above fits
