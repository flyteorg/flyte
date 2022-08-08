.. _integrations:

############
Integrations
############

Flyte is designed to be highly extensible and can be customized in multiple ways.

****************
Flytekit Plugins
****************

Flytekit plugins are simple plugins that can be implemented purely in python, unit tested locally and allow extending
Flytekit functionality. These plugins can be anything and for comparison can be thought of like
`Airflow Operators <https://airflow.apache.org/docs/apache-airflow/stable/howto/operator/index.html>`__.


.. panels::
   :header: text-center
   :column: col-lg-12 p-2

   .. link-button:: auto/integrations/flytekit_plugins/sql/index
      :type: ref
      :text: SQL
      :classes: btn-block stretched-link
   ^^^^^^^^^^^^
   Execute SQL queries as tasks.

   ---

   .. link-button:: auto/integrations/flytekit_plugins/greatexpectations/index
      :type: ref
      :text: Great Expectations
      :classes: btn-block stretched-link
   ^^^^^^^^^^^^
   Validate data with ``great_expectations``.

   ---

   .. link-button:: auto/integrations/flytekit_plugins/papermilltasks/index
      :type: ref
      :text: Papermill
      :classes: btn-block stretched-link
   ^^^^^^^^^^^^
   Execute Jupyter Notebooks with ``papermill``.

   ---

   .. link-button:: auto/integrations/flytekit_plugins/pandera_examples/index
      :type: ref
      :text: Pandera
      :classes: btn-block stretched-link
   ^^^^^^^^^^^^
   Validate pandas dataframes with ``pandera``.

   ---

   .. link-button:: auto/integrations/flytekit_plugins/modin_examples/index
      :type: ref
      :text: Modin
      :classes: btn-block stretched-link
   ^^^^^^^^^^^^
   Scale pandas workflows with ``modin``.

   ---
   .. link-button:: auto/integrations/flytekit_plugins/dolt/index
      :type: ref
      :text: Dolt
      :classes: btn-block stretched-link
   ^^^^^^^^^^^^
   Version your SQL database with ``dolt``.

   ---
   .. link-button:: auto/integrations/flytekit_plugins/whylogs_examples/index
      :type: ref
      :text: whylogs
      :classes: btn-block stretched-link
   ^^^^^^^^^^^^
   ``whylogs``, the open standard for data logging.


.. dropdown:: :fa:`info-circle` Using flytekit plugins
   :animate: fade-in-slide-down

   Data is automatically marshalled and unmarshalled in and out of the plugin. Users should mostly implement the
   :py:class:`~flytekit.core.base_task.PythonTask` API defined in Flytekit.

   Flytekit Plugins are lazily loaded and can be released independently like libraries. We follow a convention to name the
   plugin like ``flytekitplugins-*``, where * indicates the package to be integrated into Flytekit. For example
   ``flytekitplugins-papermill`` enables users to author Flytekit tasks using `Papermill <https://papermill.readthedocs.io/en/latest/>`__.

   You can find the plugins maintained by the core Flyte team `here <https://github.com/flyteorg/flytekit/tree/master/plugins>`__.


**********************
Native Backend Plugins
**********************

Native Backend Plugins are the plugins that can be executed without any external service dependencies because the compute is
orchestrated by Flyte itself, within its provisioned Kubernetes clusters.

.. panels::
    :header: text-center
    :column: col-lg-12 p-2

    .. link-button:: auto/integrations/kubernetes/pod/index
       :type: ref
       :text: K8s Pods
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Execute K8s pods for arbitrary workloads.

    ---

    .. link-button:: auto/integrations/kubernetes/k8s_spark/index
       :type: ref
       :text: K8s Cluster Spark Jobs
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Run Spark jobs on a K8s Cluster.

    ---

    .. link-button:: auto/integrations/kubernetes/kfpytorch/index
       :type: ref
       :text: Kubeflow PyTorch
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Run distributed PyTorch training jobs using ``Kubeflow``.

    ---

    .. link-button:: auto/integrations/kubernetes/kftensorflow/index
       :type: ref
       :text: Kubeflow TensorFlow
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Run distributed TensorFlow training jobs using ``Kubeflow``.

    ---

    .. link-button:: auto/integrations/kubernetes/kfmpi/index
       :type: ref
       :text: MPI Operator
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Run distributed deep learning training jobs using Horovod and MPI.

.. _external_service_backend_plugins:

********************************
External Service Backend Plugins
********************************

As the term suggests, external service backend plugins relies on external services like
`AWS Sagemaker <https://aws.amazon.com/sagemaker/>`__,
`Hive <https://docs.qubole.com/en/latest/user-guide/engines/hive/index.html>`__ or `Snowflake <https://www.snowflake.com/>`__ for handling the workload defined in
the Flyte task that use the respective plugin.

.. panels::
    :header: text-center
    :column: col-lg-12 p-2

    .. link-button:: auto/integrations/aws/sagemaker_training/index
       :type: ref
       :text: AWS Sagemaker: Model Training
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Train models with built-in or define your own custom algorithms.

    ---

    .. link-button:: auto/integrations/aws/sagemaker_pytorch/index
       :type: ref
       :text: AWS Sagemaker: Pytorch Training
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Train Pytorch models using Sagemaker, with support for distributed training.

    ---

    .. link-button:: auto/integrations/aws/athena/index
       :type: ref
       :text: AWS Athena
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Execute queries using AWS Athena

    ---

    .. link-button:: auto/integrations/aws/batch/index
       :type: ref
       :text: AWS Batch
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Running tasks and workflows on AWS batch service

    ---

    .. link-button:: auto/integrations/external_services/hive/index
       :type: ref
       :text: Hive
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Run Hive jobs in your workflows.

    ---

    .. link-button:: auto/integrations/external_services/snowflake/index
       :type: ref
       :text: Snowflake
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Run Snowflake jobs in your workflows.

    ---

    .. link-button:: auto/integrations/gcp/bigquery/index
       :type: ref
       :text: BigQuery
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Run BigQuery jobs in your workflows.


.. _enable-backend-plugins:

.. dropdown:: :fa:`info-circle` Enabling Backend Plugins
   :animate: fade-in-slide-down

   To enable a backend plugin you have to add the ``ID`` of the plugin to the enabled plugins list. The ``enabled-plugins`` is available under the ``tasks > task-plugins`` section of FlytePropeller's configuration.
   The `plugin configuration structure is defined here <https://pkg.go.dev/github.com/flyteorg/flytepropeller@v0.6.1/pkg/controller/nodes/task/config#TaskPluginConfig>`_. An example of the config follows,

   .. rli:: https://raw.githubusercontent.com/flyteorg/flyte/master/kustomize/overlays/sandbox/flyte/config/propeller/enabled_plugins.yaml
       :language: yaml

   Finding the ``ID`` of the Backend Plugin
   ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
   This is a little tricky since you have to look at the source code of the plugin to figure out the ``ID``. In the case of Spark, for example, the value of ``ID`` is `used <https://github.com/flyteorg/flyteplugins/blob/v0.5.25/go/tasks/plugins/k8s/spark/spark.go#L424>`_ here, defined as `spark <https://github.com/flyteorg/flyteplugins/blob/v0.5.25/go/tasks/plugins/k8s/spark/spark.go#L41>`_.

   Enabling a Specific Backend Plugin in Your Own Kustomize Generator
   ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
   Flyte uses Kustomize to generate the the deployment configuration which can be leveraged to `kustomize your own deployment <https://github.com/flyteorg/flyte/tree/master/kustomize>`_.

**********************
Custom Container Tasks
**********************

Because Flyte uses executable docker containers as the smallest unit of compute, you can write custom tasks with the
:py:class:`flytekit.ContainerTask` via the `flytekit <https://github.com/flyteorg/flytekit>`_ SDK.

.. panels::
    :header: text-center
    :column: col-lg-12 p-2

    .. link-button:: raw_container
       :type: ref
       :text: Raw Container Tasks
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Execute arbitrary containers: You can write c++ code, bash scripts and any containerized program.


************************************
SDKs for Writing Tasks and Workflows
************************************

The :ref:`community <community>` would love to help you with your own ideas of building a new SDK. Currently the available SDKs are:

.. panels::
    :header: text-center
    :column: col-lg-12 p-2

    .. link-button:: https://flytekit.readthedocs.io
       :type: url
       :text: Flytekit
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    The Python SDK for Flyte.

    ---

    .. link-button:: https://github.com/spotify/flytekit-java
       :type: url
       :text: Flytekit-java
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    The Java/Scala SDK for Flyte.

**********************
Flyte Airflow Provider
**********************

The Flyte Airflow Provider is helpful to call Flyte tasks/workflows from within Airflow.

.. panels::
    :header: text-center
    :column: col-lg-12 p-2

    .. link-button:: auto/integrations/external_services/airflow/index
       :type: ref
       :text: Airflow
       :classes: btn-block stretched-link
    ^^^^^^^^^^^^
    Trigger Flyte executions from Airflow.

.. TODO: write docs on Helm https://github.com/flyteorg/flyte/issues/299
