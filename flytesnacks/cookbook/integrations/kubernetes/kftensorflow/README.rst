Kubeflow TensorFlow
===================

.. tags:: Integration, DistributedComputing, MachineLearning, KubernetesOperator, Advanced

TensorFlow operator is useful to natively run distributed TensorFlow training jobs on Flyte.
It is a wrapper built around `Kubeflow's TensorFlow operator <https://www.kubeflow.org/docs/components/training/tftraining/>`__.

Installation
------------

To install the Kubeflow TensorFlow plugin, run the following command:

.. code-block:: bash

    pip install flytekitplugins-kftensorflow

To enable the plugin in the backend, follow instructions outlined in the :std:ref:`flyte:deployment-plugin-setup-k8s` guide.

Code
----

We will write an example that does distributed training using the Kubeflow TensorFlow operator.
Before that, let's look at the compute setup and Dockerfile.

GPU to CPU
^^^^^^^^^^

**GPU support has been enabled in the code by default**. If you want to test your code on a CPU, incorporate the following changes:

* Replace ``FROM tensorflow/tensorflow:latest-gpu`` with ``FROM tensorflow/tensorflow:latest`` in the Dockerfile
* Remove the ``gpu`` parameter from the ``Resources`` definition in the example

Dockerfile
^^^^^^^^^^

The example uses TensorFlow-GPU image.

.. literalinclude:: ../../../../../integrations/kubernetes/kftensorflow/Dockerfile
    :emphasize-lines: 1-3
    :language: docker
