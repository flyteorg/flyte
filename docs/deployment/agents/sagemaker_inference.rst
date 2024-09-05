.. _deployment-agent-setup-sagemaker-inference:

SageMaker Inference Agent
=========================

This guide provides an overview of how to set up the SageMaker inference agent in your Flyte deployment.

Specify agent configuration
---------------------------

.. tabs::

    .. group-tab:: Flyte binary

      Edit the relevant YAML file to specify the agent.

      .. code-block:: bash

        kubectl edit configmap flyte-sandbox-config -n flyte

      .. code-block:: yaml
        :emphasize-lines: 7,11-12

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
              - boto: agent-service
              - sagemaker-endpoint: agent-service

    .. group-tab:: Flyte core

      Create a file named ``values-override.yaml`` and add the following configuration to it:

      .. code-block:: yaml
        :emphasize-lines: 9,14-15

        configmap:
          enabled_plugins:
            tasks:
              task-plugins:
                enabled-plugins:
                  - container
                  - sidecar
                  - k8s-array
                  - agent-service
                default-for-task-types:
                  container: container
                  sidecar: sidecar
                  container_array: k8s-array
                  boto: agent-service
                  sagemaker-endpoint: agent-service

AWS credentials
---------------

When running the code locally, you can set AWS credentials as
`environment variables <https://boto3.amazonaws.com/v1/documentation/api/latest/guide/credentials.html#environment-variables>`__.
When running on a production AWS cluster, the IAM role is used by default. Ensure that it has the `AmazonSageMakerFullAccess` policy attached.

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

You can refer to the documentation `here <https://docs.flyte.org/en/latest/flytesnacks/examples/sagemaker_inference_agent/index.html>`__.
