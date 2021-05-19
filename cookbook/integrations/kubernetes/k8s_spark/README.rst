.. _plugins-spark-k8s:

Executing Spark Jobs Natively on K8s Cluster
============================================
Flyte can execute Spark jobs natively on a Kubernetes Cluster, which manages the lifecycle, spin-up, and tear down
of a virtual cluster. It leverages the open-sourced `Spark On K8s Operator <https://github.com/GoogleCloudPlatform/spark-on-k8s-operator>`_
and can be enabled without signing up for any service. 

Flytekit makes it possible to write pyspark code natively as a task and the
Spark cluster will be automatically configured using the decorated ``SparkConf``. The examples in this section provide a
hands-on tutorial for writing pyspark tasks.

Environment Setup
----------------------

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
    
.. _spark-examples:

Code Examples
-------------
