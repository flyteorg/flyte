PyTorch plugin usage example
============================

This demo is built on top of this `example`_

#############
Prerequisites
#############

Before running this make sure that
    - pytorch plugin is enabled in flytepropeller's config
    - `Kubeflow pytorch operator`_ is installed in your k8s cluster (you can use `base`_ and configure it in your deploy)
    - [if using GPU] `GPU device plugin`_ is deployed as well

#####
Usage
#####

Build container and push it to the registry

.. code-block::

    cd pytorch
    make docker_build

Create flyte project

.. code-block::

    curl -v -X POST http://127.0.0.1:30081/api/v1/projects -d '{"project": {"id": "pytorch-mnist", "name": "pytorch-mnist"} }'

Register workflow

.. code-block::

    docker run --network host flytesnacks-pytorch:<TAG> pyflyte -p pytorch-mnist -d development -c flyte.config register workflows

Navigate to https://127.0.0.1:30081/console/projects/pytorch-mnist/workflows?domain=development and launch workflow.

#####
Notes
#####

In order to disable GPU usage for the MNIST task just comment out ``per_replica_gpu_limit="1"`` in ``pytorch_task`` decorator

.. code-block::

    @pytorch_task(
        workers_count=2,
        per_replica_cpu_request="500m",
        per_replica_memory_request="4Gi",
        per_replica_memory_limit="8Gi",
        #per_replica_gpu_limit="1",
    )

.. _`example`: https://github.com/kubeflow/pytorch-operator/blob/b7fef224fef1ef0117f6e74961b557270fcf4b04/examples/mnist/mnist.py
.. _`Kubeflow pytorch operator`: https://github.com/kubeflow/pytorch-operator
.. _`base`: https://github.com/lyft/flyte/blob/master/kustomize/base/operators/kfoperators/pytorch/kustomization.yaml
.. _`GPU device plugin`: https://kubernetes.io/docs/tasks/manage-gpus/scheduling-gpus/#deploying-nvidia-gpu-device-plugin