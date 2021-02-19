Executing Distributed Pytorch training jobs on K8s
==========================================================
This plugin uses the Kubeflow Pytorch Operator and provides an extremely simplified interface for executing distributed training using various pytorch backends.

Installation
------------

To use the flytekit distributed pytorch plugin simply run the following:

.. prompt:: bash

   pip install flytekitplugins-kfpytorch==0.1.0


How to build your Dockerfile for Pytorch on K8s
-----------------------------------------------

.. note::

    If using CPU for training then special dockerfile is NOT REQUIRED. If GPU or TPUs are required then, the dockerfile differs only in the driver setup. The following dockerfile is enabled for GPU accelerated training using CUDA


.. literalinclude:: ../Dockerfile
    :language: dockerfile
    :emphasize-lines: 1
    :linenos:
    :caption: Dockerfile for distributed pytorch training is identical to the base dockerfile, except for using CUDA base
