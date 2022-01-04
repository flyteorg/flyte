.. _plugins-spark-k8s:

Kubernetes Spark Jobs
=====================

Flyte can execute Spark jobs natively on a Kubernetes Cluster, which manages a virtual cluster's lifecycle, spin-up, and tear down.
It leverages the open-sourced `Spark On K8s Operator <https://github.com/GoogleCloudPlatform/spark-on-k8s-operator>`__
and can be enabled without signing up for any service. This is like running a ``transient spark cluster``—a type of cluster spun up for a specific Spark job and torn down after completion.
These clusters are better for production workloads but have an extra cost of setup and teardown.

In Flyte/K8s, the cost is amortized because pods are faster to create than a machine, but the penalty of downloading Docker images may affect the performance.
Also, remember that starting a pod is not as fast as running a process.

Flytekit makes it possible to write PySpark code natively as a task and the
Spark cluster will be automatically configured using the decorated ``SparkConf``. The examples in this section provide a
hands-on tutorial for writing PySpark tasks.

.. NOTE::

   This plugin has been tested at scale, and more than 100k Spark Jobs run through Flyte at Lyft. This still needs a large capacity on Kubernetes and careful configuration.
   We recommend using Multi-cluster mode - :std:ref:`deployment/cluster_config/performance:multi-cluster mode` , and enabling :std:ref:`Resource Quotas <deployment/cluster_config/general:configurable resource types>` for large and extremely frequent Spark Jobs.
   This is not recommended for extremely short-running jobs, and it might be better to use a pre-spawned cluster. A job can be considered ``short`` if the runtime is less than ``2-3`` minutes.
   In this scenario, the cost of pod bring-up outweighs the cost of execution.

Why Use K8s Spark?
------------------

Managing Python dependencies is hard. Flyte makes it easy to version and manage dependencies using containers.
The K8s Spark plugin brings all the benefits of containerization to Spark without needing to manage special Spark clusters.

**Pros:**

#. Extremely easy to get started; get complete isolation between workloads
#. Every job runs in isolation and has its own virtual cluster — no more nightmarish dependency management!
#. Flyte manages everything for you!

**Cons:**

#. Short running, bursty jobs are not a great fit because of the container overhead
#. No interactive Spark capabilities are available with Flyte K8s Spark, which is more suited for running adhoc and scheduled jobs


Step 1: Deploy Spark Plugin in the Flyte Backend
-------------------------------------------------

Flyte Spark uses the `Spark On K8s Operator <https://github.com/GoogleCloudPlatform/spark-on-k8s-operator>`__ and a custom built `Flyte Spark Plugin <https://pkg.go.dev/github.com/flyteorg/flyteplugins@v0.5.25/go/tasks/plugins/k8s/spark>`__.
This is a backend plugin which has to be enabled in your deployment; you can follow the steps mentioned in the :ref:`flyte:deployment-plugin-setup-k8s` section.

.. todo: Update to show Helm deployment

You can optionally configure the plugin as per the `backend config structure <https://pkg.go.dev/github.com/flyteorg/flyteplugins@v0.5.25/go/tasks/plugins/k8s/spark#Config>`__ and an example config is defined
`here <https://github.com/flyteorg/flyte/blob/376f61acc37d885d17aa6b4d003db502c4cb6bcf/kustomize/overlays/eks/flyte/config/propeller/plugins/spark.yaml>`__. This is how it looks:

.. rli:: https://raw.githubusercontent.com/flyteorg/flyte/376f61acc37d885d17aa6b4d003db502c4cb6bcf/kustomize/overlays/eks/flyte/config/propeller/plugins/spark.yaml
   :language: yaml

Spark Service Accounts
^^^^^^^^^^^^^^^^^^^^^^

Spark needs a special service account (with associated role and role bindings) to create executor pods.
If you use IAM for Service accounts or GCP Workload identity, you need to update the service account to include this.

You can use ``Flyte cluster resource manager`` to manage creating the spark service account per namespace.
For this, you need to add the cluster resource templates as shown `here <https://github.com/flyteorg/flyte/tree/376f61acc37d885d17aa6b4d003db502c4cb6bcf/kustomize/overlays/eks/flyte/config/clusterresource-templates>`__ (refer to the ``*spark*.yaml`` files).

Step 2: Environment Setup
-------------------------

#. Install ``flytekitplugins-spark`` using ``pip`` in your environment that contains ``flytekit >= 0.16.0``.

   .. code-block:: bash

      pip install flytekitplugins-spark

#. Build Spark image correctly as explained in :ref:`spark-docker-image`.

#. Enable Spark plugin for Flyte refering to the :ref:`spark-examples` section. Additionally, Flyte uses the SparkOperator to run Spark Jobs and separate K8s Service Account/Role per namespace, which are created as part of the standard Flyte deployment.

#. Ensure you have enough resources on your K8s cluster. Based on the resources required for your Spark job (across drivers/executors), you may have to tweak resource quotas for the namespace.

.. _spark-docker-image:

How to Build Your Dockerfile for Spark on K8s
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Using Spark on K8s is extremely easy and provides full versioning using the custom-built Spark container. The built container can also execute regular Spark tasks.
For Spark, the image must contain Spark dependencies and the correct entry point for the Spark driver/executors.
This can be achieved using the `flytekit_install_spark.sh <https://github.com/lyft/flytekit/blob/67b00ef6173c77a940dbe612baa9b76408ef1448/scripts/flytekit_install_spark3.sh>`__ script provided as part of the Dockerfile included here.

.. literalinclude:: ../../../../../integrations/kubernetes/k8s_spark/Dockerfile
    :linenos:
    :emphasize-lines: 26-38
    :language: docker

Step 3: Optionally, Setup Visibility
-------------------------------------

Every time a Spark job is run, you can get a Spark application UI link to monitor the Job.
And for historical executions, you can use the SparkHistory Server to retrieve the archived Spark execution history.
Also, Flyte can create explicit links to the Spark driver logs and the individual Spark executor logs.

Spark history server and Spark UI links are directly shown in the Flyte Console and simply depend on the configuration.

Setup Spark History Link in UI
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
To get a link to the Spark history UI in Flyte Console, you need to set up a config variable in the Spark section of the Flyteplugins configuration. ::

  plugins:
    spark:
      spark-history-server-url: <root-url-forspark-history server>


Checkout the various configuration options available `here <https://github.com/flyteorg/flyteplugins/blob/2e8a22b1b5569d6f24373495fdfec68c5e7d344f/go/tasks/plugins/k8s/spark/config.go>`__.

Setup Spark Application UI (more involved)
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

To get a link for the in-progress Spark drivers, Spark application UI, you need to configure your Kubernetes to have wildcard ingress access -``*.my-domain.net`` and configure the
`Spark On K8s Operator <https://github.com/GoogleCloudPlatform/spark-on-k8s-operator>`__ to create a new ingress route for every application. This can be done as a command-line option to Spark-operator called
`ingress-url-format <https://github.com/GoogleCloudPlatform/spark-on-k8s-operator/blob/d38c904a4dd84e849408153cdf4d7a30a7be5a07/main.go#L62>`__.

Setup Spark Driver and Executor Logs
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

This can be configured by configuring the ``logs`` configuration for the Spark plugin. Spark Plugin uses the same default log configuration as explained in :ref:`configure-logging`.

SparkPlugin supports separating User (spark user code) vs. System (spark core logs) to enhance visibility into Spark, which is only available if you can route the spark user logs separately from the core logs. 
**Flyte does not automatically separate the logs.** Checkout the configuration structure `here <https://github.com/flyteorg/flyteplugins/blob/2e8a22b1b5569d6f24373495fdfec68c5e7d344f/go/tasks/plugins/k8s/spark/config.go#L31>`__.

- *Mixed*: Get unseparated logs from Spark Driver (both user and system), which follow the same structure as all log plugins. You can get links to the K8s dashboard, or a log aggregator of your choice, as long as it can generate standardized links.
- *User*: Logs from the driver which are separated (if log separation is available)
- *System*: Logs from executors—usually will not return unique links per executors; more like a prefix where all executors logs can be found
- *AllUser*: Logs all user logs across spark-submit, driver, and executor

**Log config example**

.. code-block:: yaml

    plugins:
        spark:
          logs:
            user:
              kubernetes-enabled: true
              kubernetes-url: <the existing k8s url you have in the main logs section>
            mixed:
              cloudwatch-enabled: true
              cloudwatch-template-uri: "https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logStream:group=<LogGroupName>;prefix=var.log.containers.{{.podName}};streamFilter=typeLogStreamPrefix"
            system:
              cloudwatch-enabled: true
              cloudwatch-template-uri: "https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logStream:group=<LogGroupName>;prefix=system_log.var.log.containers.{{.podName}};streamFilter=typeLogStreamPrefix"
            all-user:
              cloudwatch-enabled: true
              cloudwatch-template-uri: "https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logStream:group=<LogGroupName>;prefix=var.log.containers.{{.podName}};streamFilter=typeLogStreamPrefix"


More configuration
^^^^^^^^^^^^^^^^^^
Spark plugin supports further enhanced configuration options; for example, if you want some Spark features to be enabled by default for every Spark application, default Spark configurations are to be applied.
Refer to the `configuration structure <https://github.com/flyteorg/flyteplugins/blob/d76eb152eb36b9a77887985ab0ff3be923261bfb/go/tasks/plugins/k8s/spark/config.go#L24-L29>`__ for more details.

.. _spark-examples:

Code Examples
-------------
