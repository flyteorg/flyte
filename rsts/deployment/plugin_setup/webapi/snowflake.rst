.. _deployment-plugin-setup-webapi-snowflake:

Snowflake Plugin Setup
----------------------

This guide gives an overview of how to set up the Snowflake in your Flyte deployment.

1. First, clone the Flytesnacks repo. This is where we have the example.

   .. code-block:: bash

      git clone https://github.com/flyteorg/flytesnacks.git

2. Start the Flyte sandbox for testing.

   .. code-block:: bash

      // NOTE: Athena plugin is only available in v0.18.0+ flyte release.
      flytectl sandbox start --source=./flytesnacks
      flytectl config init

3. Setup Your Google account for snowflake.

   .. code-block:: bash

      export KUBECONFIG=$KUBECONFIG:~/.kube/config:~/.flyte/k3s/k3s.yaml
      //TODO: Steps

4. Create a file named ``values-snowflake.yaml`` and add the following config to it:

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
              - snowflake
            default-for-task-types:
              container: container
              sidecar: sidecar
              container_array: k8s-array
              snowflake: snowflake

5. Add snowflake JWT token to Flytepropeller. `here <https://docs.snowflake.com/en/developer-guide/sql-api/guide.html#using-key-pair-authentication>`_ to see more detail to setup snowflake JWT token.

.. code-block:: bash

    kubectl edit secret -n flyte flyte-propeller-auth

Configuration will be like below

.. code-block:: bash

    apiVersion: v1
    data:
      FLYTE_SNOWFLAKE_CLIENT_TOKEN: <JWT_TOKEN>
      client_secret: Zm9vYmFy
    kind: Secret
    metadata:
      annotations:
        meta.helm.sh/release-name: flyte
        meta.helm.sh/release-namespace: flyte
    ...

6. Upgrade the Flyte Helm release.

   .. code-block:: bash

      helm upgrade -n flyte -f values-mpi.yaml flyteorg/flyte --kubeconfig=~/.flyte/k3s/k3s.yaml

7. (Optional) Build & Serialize the Snowflake plugin example.

   .. code-block:: bash

      cd flytesnacks
      flytectl sandbox exec -- make -C cookbook/integrations/external_services/snowflake serialize

8. Register the Snowflake plugin example.

   .. code-block:: bash

      flytectl register files https://github.com/flyteorg/flytesnacks/releases/download/v0.2.226/snacks-cookbook-aws-athena.tar.gz --archive -p flytesnacks -d development

9. Lastly, fetch the launch plan, create and monitor the execution.

   .. code-block:: bash

      flytectl get launchplan --project flytesnacks --domain development snowflake.workflows.example.snowflake_wf  --latest --execFile exec_spec.yaml
      flytectl create execution --project flytesnacks --domain development --execFile exec_spec.yaml
      flytectl get execution --project flytesnacks --domain development <execution_id>
