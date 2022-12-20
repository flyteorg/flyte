.. _aws-sagemaker:

AWS Sagemaker Training
======================

.. tags:: Integration, MachineLearning, AWS, Advanced

This section provides examples of Flyte Plugins that are designed to work with
AWS Hosted services like Sagemaker, EMR, Athena, Redshift etc


Installation
------------

To use the flytekit aws sagemaker plugin simply run the following:

.. prompt:: bash

   pip install flytekitplugins-awssagemaker


Builtin Algorithms
------------------
Amazon SageMaker provides several built-in machine learning algorithms that you can use for a variety of problem types. Built-in algorithms are the fastest to get started with, as they are already pre-built and optimized on Sagemaker. To understand how they work and the various options available please refer to Amazon Sagemaker Official Documentation

Flyte Sagemaker plugin intends to greatly simplify using Sagemaker for training. We have tried to distill the API into a meaningful subset that makes it easier for users to adopt and run with Sagemaker. Due to the nature of the Sagemaker built-in algorithms, it is possible to run them completely from a local notebook using Flyte. This is because, Flyte will automatically use a pre-built Image for the given algorithm.

The Algorithm Images are configured in FlytePlugins in the plugin Configuration here. In the default setting, we have configured XGBoost.

.. note::

    Sagemaker Builtin Algorithms do not require explicit docker images to be specified.

Training a custom model
-----------------------
Flyte Sagemaker plugin intends to greatly simplify using Sagemaker for training. We have tried to distill the API into a meaningful subset that makes it easier for users to adopt and run with Sagemaker. Training code that runs on Sagemaker looks almost identical to writing any other task on Flyte. Once a custom job is defined, hyper parameter optimization for pre-built algorithms or custom jobs is identical. Users need to wrap their training tasks into an HPO Task and launch it.

.. note::

    Sagemaker custom algorithms work by building our own Docker images. These images need to be pushed to ECR for Sagemaker to access them. Thus, this examples need to be compiled and pushed to your own AWS ECR docker registry to actually execute on Sagemaker.

When a remote execution is triggered, the Sagemaker API is invoked to launch a job in and users function ``@task`` is invoked and all parameters are passed to it. The returns are automatically captured. The users can provide a special predicate that marks when to capture outputs from Rank-0 task in distributed training.

Prerequisites
--------------

Before following this example, make sure that
- SageMaker plugins are `enabled in flytepropeller's config <https://github.com/flyteorg/flytepropeller/blob/fe1981b1f10e19f0a189509c3826f2d8f5a75da0/propeller-config.yaml#L36-L39>`_
- You have your `AWS role set up correctly for SageMaker <https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-roles.html>`_
- `AWS SageMaker k8s operator <https://github.com/aws/amazon-sagemaker-operator-for-k8s>`_ is installed in your k8s cluster

Creating a Dockerfile for Sagemaker Custom Training [Required]
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. code-block:: docker
    :emphasize-lines: 22-23
    :linenos:

    FROM python:3.8-buster
    LABEL org.opencontainers.image.source https://github.com/flyteorg/flytesnacks

    WORKDIR /root
    ENV LANG C.UTF-8
    ENV LC_ALL C.UTF-8
    ENV PYTHONPATH /root

    # Install the AWS cli separately to prevent issues with boto being written over
    RUN pip install awscli

    # Setup a virtual environment
    ENV VENV /opt/venv
    # Virtual environment
    RUN python3 -m venv ${VENV}
    ENV PATH="${VENV}/bin:$PATH"

    # Install Python dependencies
    COPY aws/sagemaker_training/requirements.txt /root
    RUN pip install -r /root/requirements.txt

    # Setup Sagemaker entrypoints
    ENV SAGEMAKER_PROGRAM /opt/venv/bin/flytekit_sagemaker_runner.py

    # Copy the makefile targets to expose on the container. This makes it easier to register.
    COPY in_container.mk /root/Makefile
    COPY aws/sagemaker_training/sandbox.config /root

    # Copy the actual code
    COPY aws/sagemaker_training/ /root/sagemaker_training

    # This tag is supplied by the build script and will be used to determine the version
    # when registering tasks, workflows, and launch plans
    ARG tag
    ENV FLYTE_INTERNAL_IMAGE $tag

