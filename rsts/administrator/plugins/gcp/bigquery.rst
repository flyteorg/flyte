.. _deployment-plugin-setup-gcp-bigquery:

Google Bigquery Plugin Setup
----------------------------

This guide gives an overview of how to set up BigQuery in your Flyte deployment. BigQuery plugin needs Flyte deployment in GCP cloud; sandbox/AWS/Azure wouldn't work.

1. Setup the GCP Flyte cluster

.. tabbed:: GCp cluster setup

  * Make sure you have up and running flyte cluster in `GCP <https://docs.flyte.org/en/latest/deployment/gcp/index.html#deployment-gcp>`__
  * Create a service account for BigQuery. More detail: https://cloud.google.com/bigquery/docs/quickstarts/quickstart-client-libraries
  * Make sure you have correct kubeconfig and selected the correct kubernetes context
  * make sure you have the correct FlyteCTL config at ~/.flyte/config.yaml

2. Create a file named ``values-override.yaml`` and add the following config to it. Please make sure that the propeller has the correct service account for BigQuery.

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
              bigquery_query_job_task: bigquery

3. Upgrade the Flyte Helm release.

.. code-block:: bash

  helm upgrade -n flyte -f values-override.yaml flyteorg/flyte-core

4. Register the BigQuery plugin example.

.. code-block:: bash

  # TODO: https://github.com/flyteorg/flyte/issues/1776
  flytectl register files https://github.com/flyteorg/flytesnacks/releases/download/v0.2.226/snacks-cookbook-integrations-gcp-bigquery.tar.gz --archive -p flytesnacks -d development

5.  Launch an execution

.. tabbed:: Flyte Console

  * Navigate to the Flyte Console's UI (e.g. `sandbox <http://localhost:30081/console>`_) and find the relevant workflow
  * Click on `Launch` to open up a launch form
  * Submit the form to launch an execution

.. tabbed:: FlyteCTL

  * Retrieve an execution form in the form of a YAML file:

    .. code-block:: bash

       flytectl get launchplan --config ~/.flyte/flytectl.yaml --project flytesnacks --domain <TODO: https://github.com/flyteorg/flyte/issues/1776>  --latest --execFile exec_spec.yaml

  * Launch! 🚀

    .. code-block:: bash

       flytectl --config ~/.flyte/flytectl.yaml create execution -p <project> -d <domain> --execFile ~/exec_spec.yaml
