.. _deployment-agent-setup-bigquery:

Google BigQuery Agent
======================

This guide provides an overview of setting up BigQuery agent in your Flyte deployment.
Please note that the BigQuery agent requires Flyte deployment in the GCP cloud;
it is not compatible with demo/AWS/Azure.

Set up the GCP Flyte cluster
----------------------------

* Ensure you have a functional Flyte cluster running in `GCP <https://docs.flyte.org/en/latest/deployment/gcp/index.html#deployment-gcp>`__.
* Create a service account for BigQuery. For more details, refer to: https://cloud.google.com/bigquery/docs/quickstarts/quickstart-client-libraries.
* Verify that you have the correct kubeconfig and have selected the appropriate Kubernetes context.
* Confirm that you have the correct Flytectl configuration at ``~/.flyte/config.yaml``.

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
              - bigquery_query_job_task: agent-service
        
        plugins:
          agent-service:
            supportedTaskTypes:
            - bigquery_query_job_task

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
                  bigquery_query_job_task: agent-service
            plugins:
              agent-service:
                supportedTaskTypes:
                - bigquery_query_job_task

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

For BigQuery plugin on the Flyte cluster, please refer to `BigQuery Plugin Example <https://docs.flyte.org/en/latest/flytesnacks/examples/bigquery_plugin/bigquery.html>`_
