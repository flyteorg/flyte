.. _deployment-plugin-setup-webapi-snowflake:

Snowflake Plugin Setup
----------------------

This guide gives an overview of how to set up Snowflake in your Flyte deployment.

1. First, clone the Flytesnacks repo. This is where we have the example.

   .. code-block:: bash

      git clone https://github.com/flyteorg/flytesnacks.git

2. Start the Flyte sandbox for testing.

   .. code-block:: bash

      flytectl sandbox start --source=./flytesnacks
      flytectl config init

3. Create a file named ``values-snowflake.yaml`` and add the following config to it:

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

4. Create a trial Snowflake account and follow the docs for creating an API key.

5. Add snowflake JWT token to FlytePropeller.

.. note::
        Refer to the `Snowflake docs <https://docs.snowflake.com/en/developer-guide/sql-api/guide.html#using-key-pair-authentication>`__ to understand setting up the Snowflake JWT token.

.. code-block:: bash

    kubectl edit secret -n flyte flyte-propeller-auth

The configuration will look as follows:

.. code-block:: yaml

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

Replace ``<JWT_TOKEN>`` with your JWT token.

6. Upgrade the Flyte Helm release.

   .. code-block:: bash

      helm upgrade -n flyte -f values-snowflake.yaml flyteorg/flyte --kubeconfig=~/.flyte/k3s/k3s.yaml

7. Register the Snowflake plugin example.

   .. code-block:: bash

      flytectl register files https://github.com/flyteorg/flytesnacks/releases/download/v0.2.226/snacks-cookbook-external_services-snowflake.tar.gz --archive -p flytesnacks -d development

8. Lastly, fetch the launch plan, create and monitor the execution.

   .. code-block:: bash

      flytectl get launchplan --project flytesnacks --domain development snowflake.workflows.example.snowflake_wf  --latest --execFile exec_spec.yaml
      flytectl create execution --project flytesnacks --domain development --execFile exec_spec.yaml
      flytectl get execution --project flytesnacks --domain development <execution_id>
