.. _deployment-plugin-setup-pytorch-operator:

PyTorch plugin Setup
------------------------

.. _pytorch-operator:

####################################
Install PyTorch Operator
####################################

Clone Flytesnacks

.. code-block:: bash

   git clone https://github.com/flyteorg/flytesnacks.git

Start the sandbox for testing

.. code-block:: bash

   flytectl sandbox start --source=./flytesnacks

Install Pytorch Operator

.. code-block:: bash

   helm repo add bitnami https://charts.bitnami.com/bitnami --kubeconfig=~/.flyte/k3s/k3s.yaml
   helm install my-release bitnami/pytorch


Create a file values-pytorch.yaml and add the below values

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

Upgrade flyte helm release

.. code-block:: bash

   helm upgrade -n flyte -f values-pytorch.yaml flyteorg/flyte --kubeconfig=~/.flyte/k3s/k3s.yaml

Build & Serialize Pytorch plugin example

.. code-block:: bash

   cd flytesnacks
   flytectl sandbox exec -- make -C cookbook/integrations/kubernetes/kfpytorch serialize

Register Pytorch plugin example

.. code-block:: bash

   flytectl register files cookbook/integrations/kubernetes/kfpytorch/_pb_output/* -p flytesnacks -d development


Create executions

.. code-block:: bash

   flytectl get launchplan --project flytesnacks --domain development kfpytorch.pytorch_mnist.pytorch_training_wf  --latest --execFile exec_spec.yaml
   flytectl create execution --project flytesnacks --domain development --execFile exec_spec.yaml
