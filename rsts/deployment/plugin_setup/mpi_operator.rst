.. _deployment-plugin-setup-mpi-operator:

MPI Operator Setup
------------------------

.. _mpi-operator:

##############
Install MPI Operator
##############

Start the sandbox for testing

.. code-block:: bash

   flytectl sandbox start

Clone MPI Operator

.. code-block:: bash

   git clone https://github.com/kubeflow/mpi-operator.git && cd mpi-operator/manifests

Install MPI Operator

.. code-block:: bash

   kustomize build overlays/kubeflow | kubectl apply --kubeconfig=~/.flyte/k3s/k3s.yaml -f -

Create a file values-mpi.yaml and add the below values

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
            default-for-task-types:
              container: container
              sidecar: sidecar
              container_array: k8s-array
              mpi: mpi

Upgrade flyte helm release

.. code-block:: bash

   helm upgrade -n flyte -f values-mpi.yaml flyteorg/flyte --kubeconfig=~/.flyte/k3s/k3s.yaml