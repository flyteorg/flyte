Kubeflow TensorFlow
===================

TensorFlow operator is helpful to run distributed TensorFlow training jobs on Flyte. It is a wrapper around Kubeflow's TensorFlow operator.

Installation
------------

To install the Flytekit plugin, run the following command:

.. code-block:: bash

    pip install flytekitplugins-kftensorflow

To enable the plugin, follow the instructions outlined in the :std:ref:`flyte:deployment-plugin-setup-k8s-tensorflow-operator` guide.

GPU to CPU?
-----------

**GPU support has been enabled in the code by default**. If you want to test your code on a CPU, make the following changes:

* You can disable GPU support by replacing ``FROM tensorflow/tensorflow:latest-gpu`` with ``FROM tensorflow/tensorflow:latest`` in the Dockerfile.
* In a non-sandbox environment, if GPU support has to be disabled, modify the ``gpu_request`` and ``gpu_limit`` variables.

Dockerfile
----------

Here's how the Dockerfile has been composed. It uses the TensorFlow-GPU image.

.. literalinclude:: ../../../../../integrations/kubernetes/kftensorflow/Dockerfile
    :emphasize-lines: 1
    :language: docker