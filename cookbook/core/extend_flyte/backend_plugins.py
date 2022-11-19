"""
.. _extend-plugin-flyte-backend:

##########################
Writing Backend Extensions
##########################

.. tags:: Extensibility, Contribute, Intermediate

Now that you have landed here, we can assume that you have exhausted your options of extending and want to extend Flyte in a way that adds new capabilities to the platform.

Let us try to recap, why we should implement a backend plugin,

#. We want to add a new capability to the Flyte Platform, for e.g.
      * ability to talk to a new service like  AWS Sagemaker, Snowflake, Redshift, Athena, BigQuery etc
      * ability to orchestrate a set of containers in a new way like Spark, Flink, Distributed training on Kubernetes (usually using a Kubernetes operator)
      * use a new container orchestration engine like AWS Batch/ECS, Hashicorp' Nomad
      * use a completely new runtime like AWS Lambda, KNative etc
#. For the case of talking to a new service like in 1.a - this can be done using flytekit extensions and usually is the better way to get started. But, once matured most of these extensions are better to be migrated to the backend. For the rest of the cases, it is possible to extend flytekit to achieve these scenarios, but this is less desirable, because of the associated overhead of first launching a container that launches these jobs downstream.
#. You want to retain the capability to updating the plugin implementation and roll out new changes, fixes without affecting the users code, or requiring them to update versions of their plugins etc
#. You want the same plugin to be accessible across multiple language SDK's - build a universal plugin

Basics
=======
We will try to understand the components of a backend plugin using an example plugin - :ref:`plugins-spark-k8s` A Flyte backend extension consists of 3 parts

Interface specification
------------------------
Usually Flyte extensions need information that is not covered by a :std:ref:`Flyte TaskTemplate <protos/docs/core/core:tasktemplate>`. The TaskTemplate consists of a
the interface, task_type identifier, some metadata and other fields. An important field to note here is - :std:ref:`custom <protos/docs/core/core:tasktemplate>`. The custom field is essentially an unstructured JSON.
This makes it possible to extend a task-template beyond the default supported targets -- :std:ref:`container <protos/docs/core/core:tasktemplate>` (WIP, sql etc).

The motivation of the Custom field, is to marshal a JSON structure that specifies information beyond what a regular TaskTemplate can capture. The actual structure of the JSON is known only to the implemented backend-plugin and the SDK components. The core Flyte platform, does not understand of look into the specifics of this structure.
It is highly recommended to use an interface definition language like Protobuf, OpenAPISpec etc to declare specify the structure of the JSON. From here, on we refer to this as the ``Plugin Specification``.

For Spark we decided to use Protobuf to specify the plugin as can be seen `here <https://github.com/flyteorg/flyteidl/blob/master/protos/flyteidl/plugins/spark.proto>`__. Note it is not necessary to have the Plugin structure specified in FlyteIDL, we do it for simplicity, ease of maintenance alongwith the core platform and because of existing tooling to generate code for protobuf.

Flytekit Plugin implementation
--------------------------------
Now that you have a specification, we have to implement a method to generate this new TaskTemplate, with the special custom field. Also, this is where the UX design comes into play. You want to write the best possible interface in the SDK that users are delighted to use. The end goal is to create the TaskTemplate with the Custom field populated with the actual JSON structure.
We will currently refer to Flytekit - python as an example for extending and implementing the SDK. (For java refer to other docs).
The SDK task should be implemented as an extension of :py:class:`flytekit.extend.PythonTask`, or more commonly :py:class:`flytekit.PythonFunctionTask`.
In the case of Spark, we extend the :py:class:`flytekit.PythonFunctionTask`, as shown `here <https://github.com/flyteorg/flytekit/blob/master/plugins/flytekit-spark/flytekitplugins/spark/task.py#L77-L123>`__.

The SparkTask is implemented as a regular flytekit plugin, with one exception, the ``Custom`` field is now actually the ``SparkJob`` protocol buffer. Flytekit base classes when serializing a task, will automatically invoke the `get_custom method <https://github.com/flyteorg/flytekit/blob/c02075d472b5587d199630bcfc7f9937673c6a0e/flytekit/core/base_task.py#L255>`_.


FlytePropeller backend Plugin
------------------------------
The backend plugin is where the actual logic of the execution is implemented. The backend plugin uses Flyte - PluginMachinery interface to implement a plugin which can be one of the following supported types

#. A `Kubernetes operator Plugin <https://pkg.go.dev/github.com/lyft/flyteplugins@v0.5.26/go/tasks/pluginmachinery/k8s#Plugin>`_. The following demo shows 2 examples of K8s backend plugins: Flytekit Athena & Spark, and Flyte K8s Pod & Spark:

.. youtube:: oK2RGQuP94k

#. A Web API plugin - `Async <https://pkg.go.dev/github.com/lyft/flyteplugins@v0.5.26/go/tasks/pluginmachinery/webapi#AsyncPlugin>`_ or `Sync <https://pkg.go.dev/github.com/lyft/flyteplugins@v0.5.26/go/tasks/pluginmachinery/webapi#SyncPlugin>`_.
#. Or if none of the above fits then - a `Core Plugin <https://pkg.go.dev/github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core#Plugin>`_

.. TODO: write content for Kubernetes operator Plugin
.. TODO: write content for Web API plugin
.. TODO: write content for Catch-all - Core Plugin

"""
