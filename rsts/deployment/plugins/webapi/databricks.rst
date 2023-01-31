.. _deployment-plugin-setup-webapi-databricks:

Databricks Plugin Setup
-----------------------

This guide gives an overview of how to set up Databricks in your Flyte deployment.

Add Flyte chart repo to Helm

.. prompt:: bash $

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

..  TODO: move this entrypoint.py script to an official Flyte repo

Upload an `entrypoint.py <https://gist.github.com/pingsutw/482e7f0134414dac437500344bac5134>`__
to dbfs or s3. Spark driver node run this file to override the default command
in the dbx job.


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
            - databricks
          default-for-task-types:
            container: container
            sidecar: sidecar
            container_array: k8s-array
            spark: databricks
  databricks:
    enabled: True
    plugin_config:
      plugins:
        databricks:
          entrypointFile: dbfs:///FileStore/tables/entrypoint.py
          databricksInstance: dbc-a53b7a3c-614c

Get an API Token
================

Create a `Databricks account <https://www.databricks.com/>`__ and follow the
docs for creating an `access token <https://docs.databricks.com/dev-tools/auth.html#databricks-personal-access-tokens>`__.

Then, create a `Instance Profile <https://docs.databricks.com/administration-guide/cloud-configurations/aws/instance-profiles.html>`_
for the Spark cluster, it allows the spark job to access your data in the s3
bucket.

Add the Databricks access token to FlytePropeller:

.. code-block:: bash

   kubectl edit secret -n flyte flyte-secret-auth

The configuration should look as follows:

.. code-block:: yaml

    apiVersion: v1
    data:
      FLYTE_DATABRICKS_API_TOKEN: <ACCESS_TOKEN>
      client_secret: Zm9vYmFy
    kind: Secret
    metadata:
      annotations:
        meta.helm.sh/release-name: flyte
        meta.helm.sh/release-namespace: flyte
    ...

Where you need to replace ``<ACCESS_TOKEN>`` with your access token.

Upgrade the Flyte Helm release
==============================

.. code-block:: bash

    helm upgrade -n flyte -f https://raw.githubusercontent.com/flyteorg/flyte/master/charts/flyte-core/values-sandbox.yaml -f values-override.yaml flyteorg/flyte-core
