.. _deployment-agent-setup-chatgpt:

ChatGPT agent
=================

This guide provides an overview of how to set up the ChatGPT agent in your Flyte deployment.
Please note that you have to set up the OpenAI API key in the agent server to to run ChatGPT tasks.

Specify agent configuration
----------------------------

.. tabs::

    .. group-tab:: Flyte binary

      Edit the relevant YAML file to specify the agent.

      .. code-block:: bash

        kubectl edit configmap flyte-sandbox-config -n flyte

      .. code-block:: yaml
        :emphasize-lines: 7,11,16

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
              - chatgpt: agent-service

        plugins:
          agent-service:
            supportedTaskTypes:
            - chatgpt

    .. group-tab:: Flyte core

      Create a file named ``values-override.yaml`` and add the following configuration to it:

      .. code-block:: yaml

        configmap:
          enabled_plugins:
            # -- Tasks specific configuration [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#GetConfig)
            tasks:
              # -- Plugins configuration, [structure](https://pkg.go.dev/github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config#TaskPluginConfig)
              task-plugins:
                # -- [Enabled Plugins](https://pkg.go.dev/github.com/flyteorg/flyteplugins/go/tasks/config#Config). Enable sagemaker*, athena if you install the backend
                enabled-plugins:
                  - container
                  - sidecar
                  - k8s-array
                  - agent-service
                default-for-task-types:
                  container: container
                  sidecar: sidecar
                  container_array: k8s-array
                  chatgpt: agent-service
            plugins:
              agent-service:
                supportedTaskTypes:
                - chatgpt

Add the OpenAI API token
-------------------------------

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
          flyte_openai_api_key: <BASE64_ENCODED_OPENAI_API_TOKEN>
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

For ChatGPT agent on the Flyte cluster, see `ChatGPT agent <https://docs.flyte.org/en/latest/flytesnacks/examples/chatgpt_agent/index.html>`_.
