.. _deployment-plugin-setup-gcp-bigquery:

Google Bigquery Plugin
======================

This guide provides an overview of setting up BigQuery in your Flyte deployment.
Please note that the BigQuery plugin requires Flyte deployment in the GCP cloud;
it is not compatible with demo/AWS/Azure.

Set up the GCP Flyte cluster
----------------------------

* Ensure you have a functional Flyte cluster running in `GCP <https://docs.flyte.org/en/latest/deployment/gcp/index.html#deployment-gcp>`__.
* Create a service account for BigQuery. For more details, refer to: https://cloud.google.com/bigquery/docs/quickstarts/quickstart-client-libraries.
* Verify that you have the correct kubeconfig and have selected the appropriate Kubernetes context.
* Confirm that you have the correct Flytectl configuration at ``~/.flyte/config.yaml``.

Specify plugin configuration
----------------------------

.. tabs::

  .. group-tab:: Flyte binary

    Edit the relevant YAML file (``eks-starter`` / ``eks-production``) to specify the plugin.

    .. code-block:: yaml
      :emphasize-lines: 7,11

      tasks:
        task-plugins:
          enabled-plugins:
            - container
            - sidecar
            - k8s-array
            - bigquery
          default-for-task-types:
            - container: container
            - container_array: k8s-array
            - bigquery_query_job_task: bigquery

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
                  - bigquery
                default-for-task-types:
                  container: container
                  sidecar: sidecar
                  container_array: k8s-array
                  bigquery_query_job_task: bigquery

Ensure that the propeller has the correct service account for BigQuery.

Upgrade the Flyte Helm release
------------------------------

.. tabs::

  .. group-tab:: Flyte binary

    .. code-block:: bash

      helm upgrade flyte-backend flyteorg/flyte-binary -n flyte --values <YOUR-YAML-FILE>

    Replace ``<YOUR-YAML-FILE>`` with the name of your YAML file.

  .. group-tab:: Flyte core

    .. code-block:: bash
    
      helm upgrade -f values-override.yaml flyte flyte/flyte-core -n flyte
