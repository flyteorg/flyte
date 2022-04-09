.. _deployment-plugin-setup-webapi-snowflake:

Snowflake Plugin Setup
----------------------

This guide gives an overview of how to set up Snowflake in your Flyte deployment.

1. Add Flyte chart repo to Helm

.. code-block::

 helm repo add flyteorg https://flyteorg.github.io/flyte

2. Setup the cluster

  * Make sure you have up and running flyte cluster in `AWS <https://docs.flyte.org/en/latest/deployment/aws/index.html#deployment-aws>`__ / `GCP <https://docs.flyte.org/en/latest/deployment/gcp/index.html#deployment-gcp>`__ /  `GCP <https://docs.flyte.org/en/latest/deployment/gcp/index.html#deployment-gcp>`__
  * Make sure you have correct kubeconfig and selected the correct kubernetes context
  * make sure you have the correct flytectl config at ~/.flyte/config.yaml


3. Create a file named ``values-override.yaml`` and add the following config to it:

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

6. Upgrade the Flyte Helm release.

.. tabbed:: Sandbox

    .. code-block:: bash

       helm upgrade -n flyte -f values-override.yaml flyteorg/flyte --kubeconfig=~/.flyte/k3s/k3s.yaml

.. tabbed:: AWS/GCP

    .. code-block:: bash

        helm upgrade -n flyte -f values-override.yaml flyteorg/flyte-core


7. Register the Snowflake plugin example.

.. code-block:: bash

  flytectl register files https://github.com/flyteorg/flytesnacks/releases/download/v0.2.226/snacks-cookbook-external_services-snowflake.tar.gz --archive -p flytesnacks -d development


8.  Launch an execution

.. tabbed:: Flyte Console

  * Navigate to Flyte Console's UI (e.g. `sandbox <http://localhost:30081/console>`_) and find the workflow.
  * Click on `Launch` to open up the launch form.
  * Submit the form.

.. tabbed:: Flytectl

  * Retrieve an execution form in the form of a yaml file:

    .. code-block:: bash

       flytectl get launchplan --config ~/.flyte/flytectl.yaml --project flytesnacks --domain development snowflake.workflows.example.snowflake_wf  --latest --execFile exec_spec.yaml --config ~/.flyte/flytectl.yaml

  * Launch! 🚀

    .. code-block:: bash

       flytectl --config ~/.flyte/flytectl.yaml create execution -p <project> -d <domain> --execFile ~/exec_spec.yaml
