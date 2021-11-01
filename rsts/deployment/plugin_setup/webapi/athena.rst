.. _deployment-plugin-setup-webapi-athena:

Athena Plugin Setup
---------------------

This guide gives an overview of how to set up the Athena in your Flyte deployment.

1. First, clone the Flytesnacks repo. This is where we have the example.

   .. code-block:: bash

      git clone https://github.com/flyteorg/flytesnacks.git

2. Start the Flyte sandbox for testing.

   .. code-block:: bash

      // NOTE: MPI plugin is only available in v0.18.0+ flyte release.
      flytectl sandbox start --source=./flytesnacks
      flytectl config init

3. Setup Your Google account for athena.

   .. code-block:: bash

      export KUBECONFIG=$KUBECONFIG:~/.kube/config:~/.flyte/k3s/k3s.yaml
      //TODO: Steps

4. Create a file named ``values-athena.yaml`` and add the following config to it:

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
              - athena
            default-for-task-types:
              container: container
              sidecar: sidecar
              container_array: k8s-array
              athena: athena

5. Upgrade the Flyte Helm release.

   .. code-block:: bash

      helm upgrade -n flyte -f values-athena.yaml flyteorg/flyte --kubeconfig=~/.flyte/k3s/k3s.yaml

6. (Optional) Build & Serialize the Athena plugin example.

   .. code-block:: bash

      cd flytesnacks
      flytectl sandbox exec -- make -C cookbook/integrations/aws/athena serialize

7. Register the MPI plugin example.

   .. code-block:: bash

      flytectl register files https://github.com/flyteorg/flytesnacks/releases/download/v0.2.226/snacks-cookbook-integrations-aws-athena.tar.gz --archive -p flytesnacks -d development

8. Lastly, fetch the launch plan, create and monitor the execution.

   .. code-block:: bash

      flytectl get launchplan --project flytesnacks --domain development  athena.athena.full_hive_demo_wf  --latest --execFile exec_spec.yaml
      flytectl create execution --project flytesnacks --domain development --execFile exec_spec.yaml
      flytectl get execution --project flytesnacks --domain development <execution_id>
