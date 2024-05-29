.. _deployment-agent-setup-snowflake:

Snowflake agent
=================

This guide provides an overview of how to set up the Snowflake agent in your Flyte deployment.

1. Set up the key pair authentication in Snowflake. For more details, see the `Snowflake key-pair authentication and key-pair rotation guide <https://docs.snowflake.com/en/user-guide/key-pair-auth>`__.
2. Create a secret with the group "snowflake" and the key "private_key". For more details, see `"Using Secrets in a Task" <https://https://docs.flyte.org/en/latest/flytesnacks/examples/productionizing/use_secrets.html>`__.

.. code-block:: bash

   kubectl create secret generic snowflake-private-key --namespace=flytesnacks-development --from-file=your_private_key_above

Specify agent configuration
----------------------------

.. tabs::

    .. group-tab:: Flyte binary

      Edit the relevant YAML file to specify the agent.

      .. code-block:: bash

        kubectl edit configmap flyte-sandbox-config -n flyte

      .. code-block:: yaml
        :emphasize-lines: 7,11,16

        tasks:
          task-plugins:
            enabled-plugins:
              - container
              - sidecar
              - k8s-array
              - agent-service
            default-for-task-types:
              - container: container
              - container_array: k8s-array
              - snowflake: agent-service

        plugins:
          agent-service:
            supportedTaskTypes:
            - snowflake

    .. group-tab:: Flyte core

      Create a file named ``values-override.yaml`` and add the following configuration to it.

      .. code-block:: yaml

        configmap:
          enabled_plugins:
            # -- Tasks specific configuration [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#GetConfig)
            tasks:
              # -- Plugins configuration, [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#TaskPluginConfig)
              task-plugins:
                # -- [Enabled Plugins](https://pkg.go.dev/github.com/flyteorg/flyteplugins/go/tasks/config#Config). Enable sagemaker*, athena if you install the backend
                enabled-plugins:
                  - container
                  - sidecar
                  - k8s-array
                  - agent-service
                default-for-task-types:
                  container: container
                  sidecar: sidecar
                  container_array: k8s-array
                  snowflake: agent-service
            plugins:
              agent-service:
                supportedTaskTypes:
                - snowflake

Ensure that the propeller has the correct service account for BigQuery.

Upgrade the Flyte Helm release
------------------------------

.. tabs::

  .. group-tab:: Flyte binary

    .. code-block:: bash

      helm upgrade <RELEASE_NAME> flyteorg/flyte-binary -n <YOUR_NAMESPACE> --values <YOUR_YAML_FILE>

    Replace ``<RELEASE_NAME>`` with the name of your release (e.g., ``flyte-backend``),
    ``<YOUR_NAMESPACE>`` with the name of your namespace (e.g., ``flyte``),
    and ``<YOUR_YAML_FILE>`` with the name of your YAML file.

  .. group-tab:: Flyte core

    .. code-block:: bash

      helm upgrade <RELEASE_NAME> flyte/flyte-core -n <YOUR_NAMESPACE> --values values-override.yaml

    Replace ``<RELEASE_NAME>`` with the name of your release (e.g., ``flyte``)

    and ``<YOUR_NAMESPACE>`` with the name of your namespace (e.g., ``flyte``).

For Snowflake agent on the Flyte cluster, see `Snowflake agent <https://docs.flyte.org/en/latest/flytesnacks/examples/snowflake_agent/index.html>`_.
