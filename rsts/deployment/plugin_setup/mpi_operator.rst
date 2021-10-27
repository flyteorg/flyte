.. _deployment-plugin-setup-mpi-operator:

Kubeflow MPI Operator Plugin Setup
----------------------------------

This guide gives an overview of how to set up the Kubeflow MPI operator in your Flyte deployment.

1. First, clone the Flytesnacks repo. This is where we have the example.

   .. code-block:: bash

      git clone https://github.com/flyteorg/flytesnacks.git

2. Start the Flyte sandbox for testing.

   .. code-block:: bash

      // NOTE: MPI plugin is only available in v0.18.0+ flyte release.
      flytectl sandbox start --source=./flytesnacks

3. Install the MPI Operator.

   .. code-block:: bash

      git clone https://github.com/kubeflow/mpi-operator.git
      kustomize build mpi-operator/manifests/overlays/kubeflow | kubectl apply --kubeconfig=~/.flyte/k3s/k3s.yaml -f -

4. Create a file named ``values-mpi.yaml`` and add the following config to it:

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

5. Upgrade the Flyte Helm release.

   .. code-block:: bash

      helm upgrade -n flyte -f values-mpi.yaml flyteorg/flyte --kubeconfig=~/.flyte/k3s/k3s.yaml

6. Build & Serialize the MPI plugin example(Optional).

   .. code-block:: bash

      cd flytesnacks
      flytectl sandbox exec -- make -C cookbook/integrations/kubernetes/kfmpi serialize

7. Register the MPI plugin example.

   .. code-block:: bash

      flytectl register files https://github.com/flyteorg/flytesnacks/releases/download/v0.2.226/snacks-cookbook-integrations-kubernetes-kfmpi.tar.gz --archive -p flytesnacks -d development

8. Lastly, fetch the launch plan, create and monitor the execution.

   .. code-block:: bash

      flytectl get launchplan --project flytesnacks --domain development kfmpi.mpi_mnist.horovod_training_wf  --latest --execFile exec_spec.yaml
      flytectl create execution --project flytesnacks --domain development --execFile exec_spec.yaml
      flytectl get execution --project flytesnacks --domain development <execution_id>
