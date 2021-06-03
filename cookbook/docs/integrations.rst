############
Integrations
############

Flyte is designed to be highly extensible. Flyte can be extended in the following ways:

#. Flytekit-only plugins: Plugins similar to executing a Python function in a container.
#. Flyte backend global plugins: Plugins that are independent of the SDK, enable backend capabilities in Flyte, and are global for the entire deployment.
#. Flyte custom container executions: These execute arbitrary containers, where the data is loaded into the container as files and read out of the containers. They can be written in C++ code, bash scripts and any other containerized program.
#. Bring your own SDK: The open source community would love to help you with your own ideas of building a new SDK. Ideas include - ``golang``, ``javascript/nodejs``, etc.

Available SDKs:

#. `Flytekit <https://github.com/lyft/flytekit>`_ is the Python SDK for writing Flyte tasks and workflows and is optimized for Machine Learning pipelines and ETL workloads.
#. `Flytekit-Java <https://github.com/spotify/flytekit-java>`_ is the Java/SCALA SDK optimized for ETL and data processing workloads.

What Are Flytekit-Only [Python] Plugins?
========================================
Flytekit plugins are simple plugins that can be implemented purely in Python, unit tested locally, and allow extending Flytekit functionality. These plugins can be compared to Airflow Operators.
The data is automatically marshalled and unmarshalled into and out of the plugin. Users should implement :py:class:`flytekit.core.base_task.PythonTask` API, defined in Flytekit.

Implementation 
^^^^^^^^^^^^^^^

Flytekit Plugins are simply loaded and can be released independently like libraries. We follow a convention to name the plugin like
``flytekitplugins-*``, where * implies the capability. 

For example ``flytekitplugins-papermill`` enables users to author flytekit tasks using `Papermill <https://papermill.readthedocs.io/en/latest/>`_.

What are Backend Plugins?
=========================

Flyte has a unique capability of adding backend plugins. Backend plugins enable the Flyte platform to add new capabilities. This has several advantages, mainly:

#. Advanced introspection capabilities - ways to improve logging, etc.
#. Service oriented architecture - ability to bugfix, deploy plugins without releasing new libraries and forcing all users to update their libraries.
#. Better management of the system communication - for example in case of aborts, Flyte can guarantee cleanup of the remote resources.
#. Reduced cost overhead, for many plugins which launch jobs on a remote service or cluster, the plugins are essentially just polling. This has a huge compute cost in traditional architectures like Airflow etc. Flyte on the other hand, can run these operations in its own control plane.
#. The potential to create drastically new interfaces that work across multiple languages and platforms.

Flyte backend plugins are more involved and implementation needs writing code in ``Golang`` that gets plugged into the Flyte backend engine. These plugins are statically loaded into the FlytePropeller. The contract for the plugin can be encoded in any serialization format - e.g. JSON, OpenAPI, protobuf. The community in general prefers using protobuf.
Once the backend plugin is implemented, any language SDK can be implemented to provide a specialized interface for the user.

Native Backend Plugins
^^^^^^^^^^^^^^^^^^^^^^^
Native Backend Plugins are plugins that can be executed without any external service dependencies. The compute is orchestrated by Flyte itself, within its provisioned kubernetes clusters. Some examples of native plugins are:

#. Python functions
#. K8s Containerized Spark
#. Array Tasks
#. Pod Tasks
#. K8s native distributed Pytorch training using Kubeflow Pytorch Operator
#. K8s native distributed Tensorflow training using Kubeflow TF Operator

External Service Backend Plugins
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
#. AWS Sagemaker Training
#. AWS Batch
#. Qubole Hive

Enabling Backend Plugins
^^^^^^^^^^^^^^^^^^^^^^^^^
To enable a backend plugin you have to add the ``ID`` of the plugin to the enabled plugins list. The ``enabled-plugins`` is available under the ``tasks > task-plugins`` section of FlytePropeller's configuration.
The `plugin configuration structure is defined here <https://pkg.go.dev/github.com/flyteorg/flytepropeller@v0.6.1/pkg/controller/nodes/task/config#TaskPluginConfig>`_. An example of the config follows,

.. rli:: https://raw.githubusercontent.com/flyteorg/flyte/master/kustomize/overlays/sandbox/flyte/config/propeller/enabled_plugins.yaml
    :language: yaml

Finding the ``ID`` of the Backend Plugin
""""""""""""""""""""""""""""""""""""""""
This is a little tricky since you have to look at the source code of the plugin to figure out the ``ID``. In the case of Spark, for example, the value of ``ID`` is `used <https://github.com/flyteorg/flyteplugins/blob/v0.5.25/go/tasks/plugins/k8s/spark/spark.go#L424>`_ here, defined as `spark <https://github.com/flyteorg/flyteplugins/blob/v0.5.25/go/tasks/plugins/k8s/spark/spark.go#L41>`_.

Enabling a Specific Backend Plugin in Your Own Kustomize Generator
""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""""
Flyte uses Kustomize to generate the the deployment configuration which can be leveraged to `kustomize your own deployment <https://github.com/flyteorg/flyte/tree/master/kustomize>`_.

.. admonition:: Coming Soon!

    We will soon be supporting Helm. Track the `GitHub Issue <https://github.com/flyteorg/flyte/issues/299>`__ to stay tuned!


.. toctree::
    :maxdepth: -1
    :caption: Integrations
    :hidden:
 
    flytekit_plugins
    kubernetes
    aws
    gcp
    external_services
