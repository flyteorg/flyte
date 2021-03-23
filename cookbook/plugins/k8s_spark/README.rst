Executing Spark Jobs natively on K8s Cluster
============================================
Flyte can execute spark jobs natively on a Kubernetes Cluster. Flyte will manage the lifecycle, spin-up and tear down
of a virtual cluster. This leverages open source `Spark On K8s Operator <https://github.com/GoogleCloudPlatform/spark-on-k8s-operator>`_
and can be enabled without signing up for any service. If you want to enable Spark for your Flyte Cluster refer to
std:ref:`plugins-spark-k8s`_. Flytekit makes it possible to write pyspark code natively as a task and the
Spark cluster will be automatically configured using the decorated SparkConf. The examples in this section provide a
hands on tutorial of writing pyspark tasks.

Pre-Requisites / Setup
----------------------

#. Install ``flytekitplugins-spark`` using ``pip`` in your environment that contains ``flytekit >= 0.16.0``

    .. code-block:: bash

        pip install flytekitplugins-spark

#. Build Spark Image correctly as explained below.

#. Enable Spark Plugin for Flyte following std:ref:`plugins-spark-k8s`_.
In-addition, Flyte uses the SparkOperator to run Spark Jobs as well as separate K8s Service Account/Role per namespace. All of these are created as part of the standard Flyte deploy.

#. Ensure you have enough resources on your K8s cluster. Based on the resources required for your spark job (across driver/executors), you might have to tweak resource-quotas for the namespace.

.. _spark-docker-image:

How to build your Dockerfile for Spark on K8s
---------------------------------------------

Using Spark on K8s is extremely easy and provides full versioning using the custom built Spark container. The
built container can also execute regular Spark tasks.
For Spark, the image must contain spark dependencies as well as the correct entrypoint for the Spark driver/executors. This can be achieved by using the `flytekit_install_spark.sh <https://github.com/lyft/flytekit/blob/67b00ef6173c77a940dbe612baa9b76408ef1448/scripts/flytekit_install_spark3.sh>`__ script provided as referenced in the Dockerfile included here.

.. literalinclude:: ../Dockerfile
    :language: dockerfile
    :emphasize-lines: 26-37,47-54
    :linenos:

