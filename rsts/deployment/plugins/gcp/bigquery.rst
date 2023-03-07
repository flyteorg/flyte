.. _deployment-plugin-setup-gcp-bigquery:

Google Bigquery Plugin Setup
----------------------------

This guide gives an overview of how to set up BigQuery in your Flyte deployment.
BigQuery plugin needs Flyte deployment in GCP cloud; sandbox/AWS/Azure wouldn't
work.

Setup the GCP Flyte cluster
===========================

.. tabs::

   .. tab:: GCP cluster setup

     * Make sure you have up and running flyte cluster in `GCP <https://docs.flyte.org/en/latest/deployment/gcp/index.html#deployment-gcp>`__
     * Create a service account for BigQuery. More detail: https://cloud.google.com/bigquery/docs/quickstarts/quickstart-client-libraries
     * Make sure you have correct kubeconfig and selected the correct kubernetes context
     * make sure you have the correct FlyteCTL config at ~/.flyte/config.yaml

Specify Plugin Configuration
============================

Create a file named ``values-override.yaml`` and add the following config to it.
Please make sure that the propeller has the correct service account for BigQuery.

.. code-block:: yaml

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

Upgrade the Flyte Helm release
==============================

.. prompt:: bash $

  helm upgrade -n flyte -f values-override.yaml flyteorg/flyte-core

Register the BigQuery plugin example
====================================

.. prompt:: bash $

  flytectl register files https://github.com/flyteorg/flytesnacks/releases/download/v0.2.226/snacks-cookbook-integrations-gcp-bigquery.tar.gz --archive -p flytesnacks -d development

Launch an execution
===================

.. tabs::

   .. tab:: Flyte Console

     * Navigate to the Flyte Console's UI (e.g. `sandbox <http://localhost:30081/console>`_) and find the relevant workflow
     * Click on `Launch` to open up a launch form
     * Submit the form to launch an execution
   
   .. tab:: FlyteCTL

     Retrieve an execution form in the form of a YAML file:
   
     .. code-block:: bash
   
        flytectl get launchplan --config ~/.flyte/flytectl.yaml \
            --project flytesnacks \
            --domain development \
            bigquery.bigquery.full_bigquery_wf \
            --latest --execFile exec_spec.yaml
   
     Launch! 🚀
   
     .. code-block:: bash
   
        flytectl --config ~/.flyte/flytectl.yaml create execution \
            -p flytesnacks \
            -d development \
            --execFile ./exec_spec.yaml
