.. _plugins-spark-k8s:

Spark on Kubernetes
====================

Flyte can execute Spark jobs natively on a Kubernetes Cluster, which manages the lifecycle, spin-up, and tear down
of a virtual cluster. It leverages the open-sourced `Spark On K8s Operator <https://github.com/GoogleCloudPlatform/spark-on-k8s-operator>`_
and can be enabled without signing up for any service. This is like running an ``transient spark cluster``, a type of cluster that is spun up for a specific spark job and torn down after it is complete.
These clusters are better for production workloads, but have an extra cost of setup and teardown.
In Flyte/K8s, the cost is ammortized by the fact that pods are faster to create than a machine, but the penalty of downloading docker images may affect the performance. Also remember - starting a pod is not as fast as running a process.


Flytekit makes it possible to write pyspark code natively as a task and the
Spark cluster will be automatically configured using the decorated ``SparkConf``. The examples in this section provide a
hands-on tutorial for writing pyspark tasks.

.. NOTE::

   This has been tested at scale and more than 100k Spark Jobs run through Flyte at Lyft. This still needs a large capacity on Kubernetes and careful configuration.
   We recommend using Multi-cluster mode - :std:ref:`deployment/cluster_config/performance:multi-cluster mode` , and enabling :std:ref:`Resource Quotas <deployment/cluster_config/general:configurable resource types>` for large and extremely frequent Spark Jobs.
   For extremely short running jobs, this is still not a recommended approach, and it might be better to use a pre-spawned cluster. A job can be considered to be ``short`` if the runtime is less than ``2-3`` minutes.
   In this scenario, the cost of pod bring-up out-weighs the cost of execution.

Why Use K8s Spark?
===================
Managing Python dependencies is hard. Flyte makes it easy to version and manage dependencies using containers. The K8s Spark plugin brings all the benefits of containerization
to spark without needing to manage special spark clusters.

Pros:
^^^^^
#. Extremely easy to get started and get complete isolation between workloads
#. Every job runs in isolation and has its own virtual cluster - no more nightmarish dependency management
#. Flyte manages everything for you!

Cons:
^^^^^
#. Short running, bursty jobs are not a great fit - because of the container overhead
#. No interactive spark capabilities available with Flyte K8s spark, which is more suited for running, adhoc and/or scheduled jobs


Step 1: Deploy Spark Plugin in the Flyte Backend
=================================================
Flyte Spark uses the `Spark On K8s Operator <https://github.com/GoogleCloudPlatform/spark-on-k8s-operator>`_ and a custom built `Flyte Spark Plugin <https://pkg.go.dev/github.com/flyteorg/flyteplugins@v0.5.25/go/tasks/plugins/k8s/spark>`_.
The plugin is a backend plugin which has to be enabled in your deployment. To enable a plugin follow the steps in :ref:`enable-backend-plugins`.

.. todo: Update to show Helm deployment

You can optionally configure the Plugin as per the - `backend Config Structure <https://pkg.go.dev/github.com/flyteorg/flyteplugins@v0.5.25/go/tasks/plugins/k8s/spark#Config>`_ and an example Config is defined
`here <https://github.com/flyteorg/flyte/blob/376f61acc37d885d17aa6b4d003db502c4cb6bcf/kustomize/overlays/eks/flyte/config/propeller/plugins/spark.yaml>`_ , which looks like:

.. rli:: https://raw.githubusercontent.com/flyteorg/flyte/376f61acc37d885d17aa6b4d003db502c4cb6bcf/kustomize/overlays/eks/flyte/config/propeller/plugins/spark.yaml
   :language: yaml

Step 1b: Spark service accounts
================================
Spark needs a special service account (with associated role and role bindings) to create executor pods. If you are using IAM for Service accounts or GCP Workload identity, then you need to update the service account to include this too.

You can use ``Flyte Cluster resource manager`` to manage creating the spark service account per namespace. For this, you need to add the cluster resource templates as shown `here <https://github.com/flyteorg/flyte/tree/376f61acc37d885d17aa6b4d003db502c4cb6bcf/kustomize/overlays/eks/flyte/config/clusterresource-templates>`_ . Refer to the ``*spark*.yaml`` files.

Step 2: Environment Setup
==========================

#. Install ``flytekitplugins-spark`` using ``pip`` in your environment that contains ``flytekit >= 0.16.0``.

   .. code-block:: bash

      pip install flytekitplugins-spark

#. Build Spark image correctly as explained in :ref:`spark-docker-image`.

#. Enable Spark plugin for Flyte refering to the :ref:`spark-examples` section. Additionally, Flyte uses the SparkOperator to run Spark Jobs and separate K8s Service Account/Role per namespace, which are created as part of the standard Flyte deployment.

#. Ensure you have enough resources on your K8s cluster. Based on the resources required for your Spark job (across drivers/executors), you might have to tweak resource quotas for the namespace.

.. _spark-docker-image:

How to Build Your Dockerfile for Spark on K8s
---------------------------------------------
Using Spark on K8s is extremely easy and provides full versioning using the custom-built Spark container. The built container can also execute regular Spark tasks.
For Spark, the image must contain Spark dependencies and the correct entry point for the Spark driver/executors. This can be achieved using the `flytekit_install_spark.sh <https://github.com/lyft/flytekit/blob/67b00ef6173c77a940dbe612baa9b76408ef1448/scripts/flytekit_install_spark3.sh>`__ script provided as part of the Dockerfile included here.

.. code-block:: docker
    :linenos:
    :emphasize-lines: 26-38,47-54

    FROM ubuntu:focal
    LABEL org.opencontainers.image.source https://github.com/flyteorg/flytesnacks
    
    WORKDIR /root
    ENV VENV /opt/venv
    ENV LANG C.UTF-8
    ENV LC_ALL C.UTF-8
    ENV PYTHONPATH /root
    ENV DEBIAN_FRONTEND=noninteractive
    
    # Install Python3 and other basics
    RUN apt-get update && apt-get install -y python3.8 python3.8-venv make build-essential libssl-dev python3-pip curl
    
    # Install AWS CLI to run on AWS (for GCS install GSutil). This will be removed
    # in future versions to make it completely portable
    RUN pip3 install awscli
    
    ENV VENV /opt/venv
    # Virtual environment
    RUN python3 -m venv ${VENV}
    ENV PATH="${VENV}/bin:$PATH"
    
    # Install Python dependencies
    COPY kubernetes/k8s_spark/requirements.txt /root
    RUN pip install -r /root/requirements.txt
    
    RUN flytekit_install_spark3.sh
    # Adding Tini support for the spark pods
    RUN wget  https://github.com/krallin/tini/releases/download/v0.18.0/tini && \
        cp tini /sbin/tini && cp tini /usr/bin/tini && \
        chmod a+x /sbin/tini && chmod a+x /usr/bin/tini
    
    # Setup Spark environment
    ENV JAVA_HOME /usr/lib/jvm/java-8-openjdk-amd64
    ENV SPARK_HOME /opt/spark
    ENV SPARK_VERSION 3.0.1
    ENV PYSPARK_PYTHON ${VENV}/bin/python3
    ENV PYSPARK_DRIVER_PYTHON ${VENV}/bin/python3
    
    # Copy the makefile targets to expose on the container. This makes it easier to register.
    COPY in_container.mk /root/Makefile
    COPY kubernetes/k8s_spark/sandbox.config /root
    
    # Copy the actual code
    COPY kubernetes/k8s_spark/ /root/k8s_spark
    
    # This tag is supplied by the build script and will be used to determine the version
    # when registering tasks, workflows, and launch plans
    ARG tag
    ENV FLYTE_INTERNAL_IMAGE $tag
    
    # Copy over the helper script that the SDK relies on
    RUN cp ${VENV}/bin/flytekit_venv /usr/local/bin/
    RUN chmod a+x /usr/local/bin/flytekit_venv
    
    # For spark we want to use the default entrypoint which is part of the
    # distribution, also enable the virtualenv for this image. 
    # Note this relies on the VENV variable we've set in this image.
    ENTRYPOINT ["/usr/local/bin/flytekit_venv", "/opt/entrypoint.sh"]


Step 3: Optionally, Setup visibility
===================================================================

Everytime a spark job is run, users can get a spark application UI link to monitor the Job. And for historical executions they can use the SparkHistory Server to retrieve the archived Spark execution history.
Also Flyte, can create explicit links to the Spark driver logs and individual spark executor logs.

Spark history server and Spark UI links are directly shown in the Flyteconsole and simply depend on configuration.

Setup spark history link in UI
--------------------------------
To get a link to teh spark history UI in FlyteConsole, users need to setup a config variable in the spark section of the Flyteplugins configuration, like so ::

  plugins:
    spark:
      spark-history-server-url: <root-url-forspark-history server>


Checkout the various configuration options available `here <https://github.com/flyteorg/flyteplugins/blob/2e8a22b1b5569d6f24373495fdfec68c5e7d344f/go/tasks/plugins/k8s/spark/config.go>`__

Setup Spark Application UI (more involved)
-------------------------------------------
To get a link for the in-progress spark drivers, spark application UI, you need to configure your kubernetes to have wildcard ingress access -``*.my-domain.net`` and configure the
`Spark On K8s Operator <https://github.com/GoogleCloudPlatform/spark-on-k8s-operator>`_ to create a new ingress route for every application. This can be done as a command-line option to spark-operator called
`ingress-url-format <https://github.com/GoogleCloudPlatform/spark-on-k8s-operator/blob/d38c904a4dd84e849408153cdf4d7a30a7be5a07/main.go#L62>`_


Setup Spark Driver and Executor Logs
-------------------------------------

this can be configured by configuring the ``logs`` configuration for the Spark plugin.

Checkout the configuration `here <https://github.com/flyteorg/flyteplugins/blob/2e8a22b1b5569d6f24373495fdfec68c5e7d344f/go/tasks/plugins/k8s/spark/config.go#L31>`_



.. _spark-examples:

Code Examples
-------------
