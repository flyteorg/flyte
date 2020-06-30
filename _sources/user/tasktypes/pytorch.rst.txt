.. _pytorch-task-type:

PyTorch Task
============

PyTorch Task Type allows users to run distributed PyTorch training jobs on the Kubernetes cluster via `PyTorch Operator`_.

#####
Setup
#####

In order to build image that is to be eventually submitted to Kubernetes, you'll need to make sure it includes following:
  - pytorch and its dependencies (GPU support, distributed communication backend libs and etc.)
  - flytekit with pytorch extra (``pip install flytekit[pytorch]``)
  - user defined flyte workflows and its dependencies

You might want to leverage official `Dockerfile`_ or `prebuilt images`_.

Also make sure that your flyte installation is compliant with these requirements:
  - pytorch plugin is enabled in flytepropeller's config
  - `Kubeflow pytorch operator`_ is installed in your k8s cluster (you can use `base`_ and configure it in your deploy)
  - [if using GPU] `GPU device plugin`_ is deployed as well


#####
Usage
#####

Use pytorch_task_ decorator for configuring job execution resources. Here you can specify number of worker replicas (in addition to single master) and resource `requests and limits`_ on per replica basis.

.. code-block:: python
   :caption: PyTorch task example (an excerpt from `flytesnacks`_)

   @inputs(
       batch_size=Types.Integer,
       test_batch_size=Types.Integer,
       epochs=Types.Integer,
       learning_rate=Types.Float,
       sgd_momentum=Types.Float,
       seed=Types.Integer,
       log_interval=Types.Integer,
       dir=Types.String)
   @outputs(epoch_accuracies=[Types.Float], model_state=Types.Blob)
   @pytorch_task(
       workers_count=2,
       per_replica_cpu_request="500m",
       per_replica_memory_request="4Gi",
       per_replica_memory_limit="8Gi",
       per_replica_gpu_limit="1",
   )
   def mnist_pytorch_job(workflow_params, batch_size, test_batch_size, epochs, learning_rate, sgd_momentum, seed, log_interval, dir, epoch_accuracies, model_state):
       backend_type = dist.Backend.GLOO

       torch.manual_seed(seed)

       device = torch.device('cuda' if torch.cuda.is_available else 'cpu')

       if should_distribute():
           dist.init_process_group(backend=backend_type)

       kwargs = {'num_workers': 1, 'pin_memory': True} if torch.cuda.is_available else {}
       train_loader = torch.utils.data.DataLoader(
           datasets.MNIST('../data', train=True, download=True,
                          transform=transforms.Compose([
                              transforms.ToTensor(),
                              transforms.Normalize((0.1307,), (0.3081,))
                          ])),
           batch_size=batch_size, shuffle=True, **kwargs)
       test_loader = torch.utils.data.DataLoader(
           datasets.MNIST('../data', train=False, transform=transforms.Compose([
               transforms.ToTensor(),
               transforms.Normalize((0.1307,), (0.3081,))
           ])),
           batch_size=test_batch_size, shuffle=False, **kwargs)

       model = Net().to(device)

       if is_distributed():
           Distributor = nn.parallel.DistributedDataParallel if torch.cuda.is_available \
               else nn.parallel.DistributedDataParallelCPU
           model = Distributor(model)

       optimizer = optim.SGD(model.parameters(), lr=learning_rate, momentum=sgd_momentum)

       accuracies = [epoch_step(model, device, train_loader, test_loader, optimizer, epoch, writer, log_interval) for epoch in range(1, epochs + 1)]

       model_file = "mnist_cnn.pt"
       torch.save(model.state_dict(), model_file)

       model_state.set(model_file)
       epoch_accuracies.set(accuracies)

   def should_distribute():
       return dist.is_available() and WORLD_SIZE > 1


   def is_distributed():
       return dist.is_available() and dist.is_initialized()


Note that if you request GPU resources, toleration like, `flyte/gpu=dedicated:NoSchedule` (configured in the common flyteplugins configuration) is added to pod spec automatically. So you can use respective taint_ to make GPU-enabled nodes available exclusively for flyte-originated GPU-oriented tasks.

.. _`PyTorch Operator`: https://github.com/kubeflow/pytorch-operator
.. _Dockerfile: https://github.com/pytorch/pytorch/blob/master/docker/pytorch/Dockerfile
.. _`prebuilt images`: https://hub.docker.com/r/pytorch/pytorch/tags
.. _pytorch_task: https://lyft.github.io/flyte/flytekit/flytekit.sdk.html#flytekit.sdk.tasks.pytorch_task
.. _`requests and limits`: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#requests-and-limits
.. _taint: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/
.. _`Kubeflow pytorch operator`: https://github.com/kubeflow/pytorch-operator
.. _`base`: https://github.com/lyft/flyte/blob/master/kustomize/base/operators/kfoperators/pytorch/kustomization.yaml
.. _`GPU device plugin`: https://kubernetes.io/docs/tasks/manage-gpus/scheduling-gpus/#deploying-nvidia-gpu-device-plugin
.. _`flytesnacks`: https://github.com/lyft/flytesnacks/blob/761426a2a41809c339a5444d111dfc637434f015/pytorch/workflows/mnist.py#L1