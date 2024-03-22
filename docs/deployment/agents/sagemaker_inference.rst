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
        :emphasize-lines: 7,11-12,16-17

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
        plugins:
          agent-service:
            supportedTaskTypes:
            - boto
            - sagemaker-endpoint

    .. group-tab:: Flyte core

      Create a file named ``values-override.yaml`` and add the following configuration to it:

      .. code-block:: yaml
        :emphasize-lines: 9,14-15,19-20

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
            plugins:
              agent-service:
                supportedTaskTypes:
                - boto
                - sagemaker-endpoint

Add the AWS credentials
-----------------------

1. Install the flyteagent pod using helm:

.. code-block::

  helm repo add flyteorg https://flyteorg.github.io/flyte
  helm install flyteagent flyteorg/flyteagent --namespace flyte

2. Get the base64 value of your AWS credentials:

.. code-block::
  
  echo -n "<AWS_CREDENTIAL>" | base64

3. Edit the flyteagent secret:

.. code-block:: bash
  
  kubectl edit secret flyteagent -n flyte

.. code-block:: yaml
  :emphasize-lines: 3-5

  apiVersion: v1
  data:
    aws-access-key: <BASE64_ENCODED_AWS_ACCESS_KEY>
    aws-secret-access-key: <BASE64_ENCODED_AWS_SECRET_ACCESS_KEY>
    aws-session-token: <BASE64_ENCODED_AWS_SESSION_TOKEN>
  kind: Secret

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
