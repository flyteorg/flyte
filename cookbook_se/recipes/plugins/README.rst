############################################
Flyte Plugins
############################################
Flyte is designed to be highly extensible. Flyte can be extended in multiple ways

#. Flytekit only plugins: Plugins that are like executing a python function in a container
#. Flyte backend global plugins: Plugins that are independent of the SDK and enable backend capabilities in Flyte and are global for the entire deployment
#. Flyte custom container executions: Execute arbitrary containers - data is loaded into the container as files and read out of the containers. One can write c++ code, bash scripts and any containerized program
#. Bring your own SDK: the community would love to help you with your own ideas of building a new SDK. Ideas include - ``golang``, ``javascript/nodejs`` etc

Available SDKs:

#. `Flytekit <https://github.com/lyft/flytekit>`_ is the Python SDK for writing Flyte tasks and workflows and is optimized for Machine Learning pipelines and ETL workloads
#. `Flytekit-Java <https://github.com/spotify/flytekit-java>`_ is the Java/SCALA SDK optimized for ETL and data processing workloads

What are Flytekit [python] only plugins?
===========================================
Flytekit plugins are simple plugins that can be implemented purely in python, unit tested locally and allow extending Flytekit functionality. These plugins can be anything and for comparison can be thought of like Airflow Operators.
Data is automatically marshalled and unmarshalled into and out of the plugin and mostly users should implement :pyclass:`flytekit.annotated.python_function_task.PythonTask` API, defined in flytekit.
This tutorial will walk you through an example of how a plugin can be written.

Flytekit Plugins are lazily loaded and can be released independently like libraries. We follow a convention to name the plugin like
``flytekitplugins-*``, where * implies the capability. For example ``flytekitplugins-papermill`` enables users to author flytekit tasks using `Papermill <https://papermill.readthedocs.io/en/latest/>`_

Examples of flytekit only plugins:

#. Papermill implementation `flytekitplugins-papermill <https://github.com/lyft/flytekit/tree/master/plugins/papermill>`_
#. SQLite3 implementation `SQLite3 Queries <https://github.com/lyft/flytekit/blob/master/flytekit/extras/sqlite3/task.py>`_

What are Backend Plugins?
=========================
Flyte backend plugins are more involved and implementation needs writing code in ``Golang`` that gets plugged into the Flyte backend engine. These plugins are statically loaded into the FlytePropeller. The contract for the plugin can be encoded in any serialization format - e.g. JSON, OpenAPI, protobuf. The community in general prefers using protobuf.
Once the backend plugin is implemented, any language SDK can be implemented to provide a specialized interface for the user.

Examples

#. `Sagemaker <https://github.com/lyft/flytekit/tree/master/plugins/awssagemaker>`_
#. `K8s Spark <https://github.com/lyft/flytekit/tree/master/plugins/spark>`_

Native Backend Plugins
^^^^^^^^^^^^^^^^^^^^^^^
Native Backend Plugins are plugins that can be executed without any external service
dependencies. The compute is orchestrated by Flyte itself, within its
provisioned kubernetes clusters. Some examples of native plugins are

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