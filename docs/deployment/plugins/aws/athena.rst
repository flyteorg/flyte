.. _deployment-plugin-setup-aws-athena:

Athena Plugin
=============

This guide provides an overview of setting up Athena in your Flyte deployment.

.. note::
  Please note that the Athena plugin requires a Flyte deployment in the AWS cloud; it won't work with demo/GCP/Azure.

Set up the AWS Flyte cluster
----------------------------

1. Ensure you have a functional Flyte cluster up and running in `AWS <https://docs.flyte.org/en/latest/deployment/aws/index.html#deployment-aws>`__
2. Verify that you possess the correct ``kubeconfig`` and have selected the appropriate Kubernetes context
3. Double-check that your ``~/.flyte/config.yaml`` file contains the correct Flytectl configuration

Specify plugin configuration
----------------------------

.. tabs::

  .. group-tab:: Flyte binary

    Edit the relevant YAML file to specify the plugin.

    .. code-block:: yaml
      :emphasize-lines: 7,11

      tasks:
        task-plugins:
          enabled-plugins:
            - container
            - sidecar
            - k8s-array
            - athena
          default-for-task-types:
            - container: container
            - container_array: k8s-array
            - athena: athena

  .. group-tab:: Flyte core

    Create a file named ``values-override.yaml`` and include the following configuration:

    .. code-block:: yaml

        configmap:
          enabled_plugins:
            tasks:
              task-plugins:
                enabled-plugins:
                  - container
                  - sidecar
                  - k8s-array
                  - athena
                default-for-task-types:
                  container: container
                  sidecar: sidecar
                  container_array: k8s-array
                  athena: athena

Ensure that the propeller has the correct service account for Athena.

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
