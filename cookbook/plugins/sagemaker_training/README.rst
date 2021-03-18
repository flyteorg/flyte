AWS Sagemaker Training
======================
This section provides examples of Flyte Plugins that are designed to work with
AWS Hosted services like Sagemaker, EMR, Athena, Redshift etc


Installation
------------

To use the flytekit aws sagemaker plugin simply run the following:

.. prompt:: bash

   pip install flytekitplugins-awssagemaker==0.16.0


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
- SageMaker plugins are `enabled in flytepropeller's config <https://github.com/lyft/flytepropeller/blob/f9819ab2f4ff817ce5f8b8bb55a837cf0aeaf229/config.yaml#L35-L36>`_
- You have your `AWS role set up correctly for SageMaker <https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-roles.html#sagemaker-roles-amazonsagemakerfullaccess-policy>`_
- `AWS SageMaker k8s operator <https://github.com/aws/amazon-sagemaker-operator-for-k8s>`_ is installed in your k8s cluster

Creating a dockerfile for Sagemaker custom training [Required]
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

.. literalinclude:: ../Dockerfile
    :language: dockerfile
    :emphasize-lines: 22-24
    :linenos:
    :caption: Dockerfile for Sagemaker, similar to base dockerfile, but installs sagemaker-training and set training script
