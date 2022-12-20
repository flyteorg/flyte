.. _mnist-classifier-training:

MNIST Classification With PyTorch and W&B
-----------------------------------------

.. tags:: MachineLearning, GPU, Advanced

PyTorch
=======

`Pytorch <https://pytorch.org/>`__ is a machine learning framework that accelerates the path from research prototyping
to production deployment. You can build *Tensors* and *Dynamic neural networks* in Python with strong GPU acceleration
using PyTorch.

In a nutshell, it is a Python package that provides two high-level features:

- Tensor computation (like NumPy) with strong GPU acceleration
- Deep neural networks built on a tape-based autograd system

Flyte directly has no unique understanding of PyTorch. As per Flyte, PyTorch is just a Python library.
However, when merged with Flyte, the combo helps utilize and bootstrap the infrastructure for PyTorch and ensures that things work well!
Additionally, it also offers other benefits of using tasks and workflows -- checkpointing, separation of concerns, and auto-memoization.

Model Development
=================

Some basics of model development are outlined in the following video, in addition to:

- Bias Variance trade-off
- Model families
- Data parallelism
- Model parallelism, and
- PyTorch parallelism

..  youtube:: FuMtJOMh5uQ

Specify GPU Requirement
=======================

One of the necessary directives applicable when working on deep learning models is explicitly requesting one or more GPUs.
This can be done by giving a simple directive to the task declaration as follows:

.. code-block:: python

    from flytekit import Resources, task

    @task(requests=Resources(gpu="1"), limits=Resources(gpu="1"))
    def my_deep_learning_task():
        ...

.. tip::
    It is recommended to use the same ``requests`` and ``limits`` for a GPU as automatic GPU scaling is not supported.

Moreover, to utilize the power of a GPU, ensure that your Flyte backend has GPU nodes provisioned.

Distributed Data-Parallel Training
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Flyte also supports distributed training for PyTorch models, but this is not native. This is achieved using one of the optional plugins, such as:

- Natively on Kubernetes using :ref:`kf-pytorch-op`
- On AWS using :ref:`aws-sagemaker` training

*Other distributed training plugins are coming soon -- MPIOperator, Google Vertex AI, etc. You can add your favorite services, too!*

Weights & Biases Integration
============================

`Weights & Biases <https://wandb.ai/site>`__, or simply, ``wandb`` helps build better models faster with experiment tracking, dataset versioning, and model management.

We'll use ``wandb`` alongside PyTorch to track our ML experiment and its concerned model parameters.

.. note::
    Before running the example, create a ``wandb`` account and log in to access the API.
    If you're running the code locally, run the command ``wandb login``.
    If it's a remote cluster, you have to include the API key in the Dockerfile.

.. _pytorch-dockerfile:

PyTorch Dockerfile for Deployment
=================================

It is essential to build the Dockerfile with GPU support to use a GPU within PyTorch.
The example in this section uses a simple ``nvidia-supplied GPU Docker image`` as the base, and the rest of the construction is similar to the other Dockerfiles.

.. literalinclude:: ../../../../../case_studies/ml_training/mnist_classifier/Dockerfile
    :language: docker

.. note::
    Run your code in the ``ml_training`` directory, both locally and within the sandbox.
