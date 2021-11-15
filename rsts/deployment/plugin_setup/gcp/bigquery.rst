.. _deployment-plugin-setup-gcp-bigquery:

Google Bigquery Plugin Setup
----------------------------

This guide gives an overview of how to set up BigQuery in your Flyte deployment. BigQuery plugin needs Flyte deployment in GCP cloud; sandbox/AWS/Azure wouldn't work.

1. Create a file named ``values-bigquery.yaml`` and add the following config to it. Please make sure that the propeller has the correct service account for BigQuery.

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
              - bigquery
            default-for-task-types:
              container: container
              sidecar: sidecar
              container_array: k8s-array
              bigquery: bigquery

2. Upgrade the Flyte Helm release.

   .. code-block:: bash

      helm upgrade -n flyte -f values-bigquery.yaml flyteorg/flyte

3. Register the BigQuery plugin example.

   .. code-block:: bash

      TODO: https://github.com/flyteorg/flyte/issues/1776
      flytectl register files https://github.com/flyteorg/flytesnacks/releases/download/v0.2.226/snacks-cookbook-integrations-gcp-bigquery.tar.gz --archive -p flytesnacks -d development

4. Lastly, fetch the launch plan, create and monitor the execution.

   .. code-block:: bash

      flytectl get launchplan --project flytesnacks --domain development  <TODO: https://github.com/flyteorg/flyte/issues/1776>  --latest --execFile exec_spec.yaml
      flytectl create execution --project flytesnacks --domain development --execFile exec_spec.yaml
      flytectl get execution --project flytesnacks --domain development <execution_id>
