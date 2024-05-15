.. _deployment-agent-setup-openai-batch:

OpenAI Batch Agent
==================

This guide provides an overview of how to set up the OpenAI Batch agent in your Flyte deployment.

Specify agent configuration
---------------------------

.. tabs::

    .. group-tab:: Flyte binary

      Edit the relevant YAML file to specify the agent.

      .. code-block:: bash

        kubectl edit configmap flyte-sandbox-config -n flyte

      .. code-block:: yaml
        :emphasize-lines: 7,11,15

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
              - openai-batch: agent-service
        plugins:
          agent-service:
            supportedTaskTypes:
            - openai-batch

    .. group-tab:: Flyte core

      Create a file named ``values-override.yaml`` and add the following configuration to it:

      .. code-block:: yaml
        :emphasize-lines: 9,14,18

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
                  openai-batch: agent-service
            plugins:
              agent-service:
                supportedTaskTypes:
                - openai-batch

Add the OpenAI API token
------------------------

1. Install flyteagent pod using helm:

.. code-block::

  helm repo add flyteorg https://flyteorg.github.io/flyte
  helm install flyteagent flyteorg/flyteagent --namespace flyte

2. Get the base64 value of your OpenAI API token:

.. code-block::

  echo -n "<OPENAI_API_TOKEN>" | base64

3. Edit the flyteagent secret:

    .. code-block:: bash

      kubectl edit secret flyteagent -n flyte

    .. code-block:: yaml
      :emphasize-lines: 3

      apiVersion: v1
      data:
        FLYTE_OPENAI_API_KEY: <BASE64_ENCODED_OPENAI_API_TOKEN>
      kind: Secret
      metadata:
        annotations:
          meta.helm.sh/release-name: flyteagent
          meta.helm.sh/release-namespace: flyte
        creationTimestamp: "2023-10-04T04:09:03Z"
        labels:
          app.kubernetes.io/managed-by: Helm
        name: flyteagent
        namespace: flyte
        resourceVersion: "753"
        uid: 5ac1e1b6-2a4c-4e26-9001-d4ba72c39e54
      type: Opaque


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

You can refer to the `documentation <https://docs.flyte.org/en/latest/flytesnacks/examples/openai_batch_agent/index.html>`__ 
to run the agent on your Flyte cluster.
