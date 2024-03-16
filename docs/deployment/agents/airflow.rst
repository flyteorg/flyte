.. _deployment-agent-setup-airflow:

Airflow agent
=================

This guide provides an overview of how to set up the Airflow agent in your Flyte deployment.
Please note that the you don't need an Airflow cluster to run the Airflow tasks, since Flytekit will
automatically compile Airflow tasks to Flyte tasks and execute them on the Flyte cluster.

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
              - airflow: agent-service

        plugins:
          agent-service:
            supportedTaskTypes:
            - airflow

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
                  airflow: agent-service
            plugins:
              agent-service:
                supportedTaskTypes:
                - airflow


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

For Airflow agent on the Flyte cluster, see `Airflow agent <https://docs.flyte.org/en/latest/flytesnacks/examples/airflow_agent/index.html>`_.
