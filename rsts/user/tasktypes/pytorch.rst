.. _pytorch-task-type:

PyTorch Task
============

PyTorch Task Type allows users to run distributed PyTorch training jobs on the Kubernetes cluster via `PyTorch Operator`_.

#######
 Setup
#######

In order to build image that is to be eventually submitted to Kubernetes, you'll need to make sure it includes following:
- pytorch and its dependencies (GPU support, distributed communication backend libs and etc.)
- flytekit with pytorch extra (``pip install flytekit[pytroch]``)
- user defined flyte workflows and its dependencies

You might want to leverage official `Dockerfile`_ or `prebuilt images`_.

#######
Usage
#######

Use pytorch_task_ decorator for configuring job execution resources. Here you can specify number of worker replicas (in addition to single master) and resource `requests and limits`_ on per replica basis.

.. code-block:: python
   :caption: PyTorch task type decorator example

    @inputs(int_list=[Types.Integer])
    @outputs(result=Types.Integer)
    @pytorch_task(
         workers_count=2,
         per_replica_cpu_request="500m",
         per_replica_memory_request="4Gi",
         per_replica_memory_limit="8Gi",
         per_replica_gpu_limit="1",
    )
    def my_pytorch_job(wf_params, int_list, result):
        pass

Note that if you request GPU resources, toleration `flyte/gpu=dedicated:NoSchedule` is added to pod spec automatically. So you can use respective taint_ to make GPU-enabled nodes available exclusively for flyte-originated GPU-oriented tasks.

.. _`PyTorch Operator`: https://github.com/kubeflow/pytorch-operator
.. _Dockerfile: https://github.com/pytorch/pytorch/blob/master/docker/pytorch/Dockerfile
.. _`prebuilt images`: https://hub.docker.com/r/pytorch/pytorch/tags
.. _pytorch_task: https://lyft.github.io/flyte/flytekit/flytekit.sdk.html#flytekit.sdk.tasks.pytorch_task
.. _`requests and limits`: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#requests-and-limits
.. _taint: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/
