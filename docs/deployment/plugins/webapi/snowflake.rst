.. _deployment-plugin-setup-webapi-snowflake:

Snowflake Plugin
================

This guide provides an overview of how to set up Snowflake in your Flyte deployment.

Spin up a cluster
-----------------

.. tabs::

  .. group-tab:: Flyte binary
      
    You can spin up a demo cluster using the following command:
   
    .. code-block:: bash
  
      flytectl demo start

    Or install Flyte using the :ref:`flyte-binary helm chart <deployment-deployment-cloud-simple>`.

  .. group-tab:: Flyte core

    If you've installed Flyte using the 
    `flyte-core helm chart <https://github.com/flyteorg/flyte/tree/master/charts/flyte-core>`__,
    please ensure:

    * You have the correct kubeconfig and have selected the correct Kubernetes context.
    * You have configured the correct flytectl settings in ``~/.flyte/config.yaml``.

.. note::

  Add the Flyte chart repo to Helm if you're installing via the Helm charts.

  .. code-block:: bash

    helm repo add flyteorg https://flyteorg.github.io/flyte

Specify plugin configuration
----------------------------

.. tabs::

  .. group-tab:: Flyte binary

    .. tabs::
         
      .. group-tab:: Demo cluster

        Enable the Snowflake plugin on the demo cluster by adding the following block to ``~/.flyte/sandbox/config.yaml``:

        .. code-block:: yaml

          tasks:
            task-plugins:
              default-for-task-types:
                container: container
                container_array: k8s-array
                sidecar: sidecar
                snowflake: snowflake
              enabled-plugins:
                - container
                - k8s-array
                - sidecar
                - snowflake

      .. group-tab:: Helm chart

        Edit the relevant YAML file to specify the plugin.

        .. code-block:: yaml
          :emphasize-lines: 7,11

          tasks:
            task-plugins:
              enabled-plugins:
                - container
                - sidecar
                - k8s-array
                - snowflake
              default-for-task-types:
                - container: container
                - container_array: k8s-array
                - snowflake: snowflake

  .. group-tab:: Flyte core
    
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

Obtain and add the Snowflake JWT token
--------------------------------------

Create a Snowflake account, and follow the `Snowflake docs 
<https://docs.snowflake.com/en/developer-guide/sql-api/authenticating#using-key-pair-authentication>`__
to generate a JWT token.
Then, add the Snowflake JWT token to FlytePropeller.

.. tabs::

  .. group-tab:: Flyte binary

    .. tabs::

      .. group-tab:: Demo cluster

        Add the JWT token as an environment variable to the ``flyte-sandbox`` deployment.

        .. code-block:: bash

          kubectl edit deploy flyte-sandbox -n flyte

        Update the ``env`` configuration:

        .. code-block:: yaml
          :emphasize-lines: 12-13

          env:
          - name: POD_NAME
            valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.name
          - name: POD_NAMESPACE
            valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
          - name: FLYTE_SECRET_FLYTE_SNOWFLAKE_CLIENT_TOKEN
            value: <JWT_TOKEN>
          image: flyte-binary:sandbox
          ...

      .. group-tab:: Helm chart

        Create a secret as follows (or add to it if it already exists from other plugins):

        .. code-block:: bash

          cat <<EOF | kubectl apply -f -
          apiVersion: v1
          kind: Secret
          metadata:
            name: flyte-binary-external-services
            namespace: flyte
          type: Opaque
          stringData:
            FLYTE_SNOWFLAKE_CLIENT_TOKEN: <JWT_TOKEN>
          EOF
        
        Reference the newly created secret in 
        ``.Values.configuration.inlineSecretRef``
        in your YAML file as follows:

        .. code-block:: yaml
          :emphasize-lines: 2

          configuration:
            inlineSecretRef: flyte-binary-external-services

    Replace ``<JWT_TOKEN>`` with your JWT token.

  .. group-tab:: Flyte core

    Add the JWT token as a secret to ``flyte-secret-auth``.

    .. code-block:: bash

      kubectl edit secret -n flyte flyte-secret-auth

    .. code-block:: yaml
      :emphasize-lines: 3

      apiVersion: v1
      data:
        FLYTE_SNOWFLAKE_CLIENT_TOKEN: <JWT_TOKEN>
        client_secret: Zm9vYmFy
      kind: Secret
      ...

    Replace ``<JWT_TOKEN>`` with your JWT token.

Upgrade the deployment
----------------------

.. tabs::

  .. group-tab:: Flyte binary

    .. tabs::

      .. group-tab:: Demo cluster

        .. code-block:: bash

          kubectl rollout restart deployment flyte-sandbox -n flyte

      .. group-tab:: Helm chart

        .. code-block:: bash

          helm upgrade <RELEASE_NAME> flyteorg/flyte-binary -n <YOUR_NAMESPACE> --values <YOUR_YAML_FILE>

        Replace ``<RELEASE_NAME>`` with the name of your release (e.g., ``flyte-backend``),
        ``<YOUR_NAMESPACE>`` with the name of your namespace (e.g., ``flyte``),
        and ``<YOUR_YAML_FILE>`` with the name of your YAML file.

  .. group-tab:: Flyte core

    .. code-block::

      helm upgrade <RELEASE_NAME> flyte/flyte-core -n <YOUR_NAMESPACE> --values values-override.yaml

    Replace ``<RELEASE_NAME>`` with the name of your release (e.g., ``flyte``)
    and ``<YOUR_NAMESPACE>`` with the name of your namespace (e.g., ``flyte``).

Wait for the upgrade to complete. You can check the status of the deployment pods by running the following command:

.. code-block::

  kubectl get pods -n flyte

  For Snowflake plugin on the Flyte cluster, please refer to `Snowflake Plugin Example <https://docs.flyte.org/en/latest/flytesnacks/examples/snowflake_plugin/snowflake.html>`_
