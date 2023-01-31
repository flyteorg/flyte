.. _deployment-plugin-setup-webapi-snowflake:

Snowflake Plugin Setup
----------------------

This guide gives an overview of how to set up Snowflake in your Flyte deployment.

Add Flyte Chart Repo to Helm
============================

.. code-block::

   helm repo add flyteorg https://flyteorg.github.io/flyte


Setup the Cluster
=================

.. tabs::

   .. tab:: Sandbox

      Start the sandbox cluster
   
      .. prompt:: bash $
   
         flytectl demo start
   
      Generate flytectl config
   
      .. prompt:: bash $
   
         flytectl config init
   
   .. tab:: AWS/GCP

      Follow the :ref:`deployment-deployment-cloud-simple` or
      :ref:`deployment-deployment-multicluster` guide to set up your cluster.
      After following these guides, make sure you have:

      * The correct kubeconfig and selected the correct kubernetes context
      * The correct flytectl config at ``~/.flyte/config.yaml``

Specify Plugin Configuration
============================

Create a file named ``values-override.yaml`` and add the following config to it:

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
              - snowflake
            default-for-task-types:
              container: container
              sidecar: sidecar
              container_array: k8s-array
              snowflake: snowflake

Get an API Token
================

Next, create a trial Snowflake account and follow the docs for creating an API
key. Add the snowflake JWT token to FlytePropeller.

.. note::
   
   Refer to the `Snowflake docs <https://docs.snowflake.com/en/developer-guide/sql-api/guide.html#using-key-pair-authentication>`__
   to understand setting up the Snowflake JWT token.

.. prompt:: bash $

    kubectl edit secret -n flyte flyte-secret-auth

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

Upgrade the Flyte Helm release
==============================

.. prompt:: bash $

   helm upgrade -n flyte -f https://raw.githubusercontent.com/flyteorg/flyte/master/charts/flyte-core/values-sandbox.yaml -f values-override.yaml flyteorg/flyte-core

Register the Snowflake plugin example
=====================================

.. prompt:: bash $

   flytectl register files https://github.com/flyteorg/flytesnacks/releases/download/v0.2.226/snacks-cookbook-external_services-snowflake.tar.gz --archive -p flytesnacks -d development


Launch an execution
===================

.. tabs::

   .. tab:: Flyte Console
   
     * Navigate to Flyte Console's UI (e.g. `sandbox <http://localhost:30081/console>`_) and find the workflow.
     * Click on `Launch` to open up the launch form.
     * Submit the form.
   
   .. tab:: Flytectl
   
      Retrieve an execution form in the form of a yaml file:
   
      .. prompt:: bash $
   
         flytectl get launchplan --config ~/.flyte/flytectl.yaml \
             --project flytesnacks \
             --domain development \
             snowflake.workflows.example.snowflake_wf \
             --latest \
             --execFile exec_spec.yaml
   
      Launch! 🚀
   
      .. prompt:: bash $
   
         flytectl --config ~/.flyte/flytectl.yaml create execution \
             -p flytesnacks \
             -d development \
             --execFile ~/exec_spec.yaml
