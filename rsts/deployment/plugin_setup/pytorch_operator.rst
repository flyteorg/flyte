.. _deployment-plugin-setup-pytorch-operator:

PyTorch Operator Setup
----------------------

This guide gives an overview of how to set up the PyTorch operator in your Flyte deployment.

1. First, clone the Flytesnacks repo. This is where we have the example.

   .. code-block:: bash

      git clone https://github.com/flyteorg/flytesnacks.git

2. Start the Flyte sandbox for testing.

   .. code-block:: bash

      flytectl sandbox start --source=./flytesnacks

3. Install the PyTorch Operator.

   .. code-block:: bash

      helm repo add bitnami https://charts.bitnami.com/bitnami --kubeconfig=~/.flyte/k3s/k3s.yaml
      helm install my-release bitnami/pytorch

4. Create a file named ``values-pytorch.yaml`` and add the following config to it:

   .. code-block::

       configmap:
         enabled_plugins:
           # -- Tasks specific configuration [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#GetConfig)
           tasks:
             # -- Plugins configuration, [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#TaskPluginConfig)
             task-plugins:
               # -- [Enabled Plugins](https://pkg.go.dev/github.com/flyteorg/flyteplugins/go/tasks/config#Config). Enable sagemaker*, athena if you install the backend
               # plugins
               enabled-plugins:
                 - container
                 - sidecar
                 - k8s-array
                 - mpi
                 - pytorch
               default-for-task-types:
                 container: container
                 sidecar: sidecar
                 container_array: k8s-array
                 pytorch: pytorch

5. Upgrade the Flyte Helm release.

   .. code-block:: bash

      helm upgrade -n flyte -f values-pytorch.yaml flyteorg/flyte --kubeconfig=~/.flyte/k3s/k3s.yaml

6. Build & Serialize the PyTorch plugin example.

   .. code-block:: bash

      cd flytesnacks
      flytectl sandbox exec -- make -C cookbook/integrations/kubernetes/kfpytorch serialize

7. Register the PyTorch plugin example.

   .. code-block:: bash

      flytectl register files cookbook/integrations/kubernetes/kfpytorch/_pb_output/* -p flytesnacks -d development


8. Lastly, fetch the launch plan, create and monitor the execution.

   .. code-block:: bash

      flytectl get launchplan --project flytesnacks --domain development kfpytorch.pytorch_mnist.pytorch_training_wf  --latest --execFile exec_spec.yaml
      flytectl create execution --project flytesnacks --domain development --execFile exec_spec.yaml
      flytectl get execution --project flytesnacks --domain development <execname>
