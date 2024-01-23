.. _deployment-agent-setup-mmcloud:

MMCloud Agent
=================

MemVerge Memory Machine Cloud (MMCloud) empowers users to continuously optimize cloud resources during runtime,
safely execute stateful tasks on spot instances,
and monitor resource usage in real time.
These capabilities make it an excellent fit for long-running batch workloads.

This guide provides an overview of how to set up MMCloud in your Flyte deployment.

Set up MMCloud
--------------

To run a Flyte workflow with Memory Machine Cloud, you will need to deploy Memory Machine Cloud.
Check out the `MMCloud User Guide <https://docs.memverge.com/mmce/current/userguide/olh/index.html>`_ to get started!

By the end of this step, you should have deployed an MMCloud OpCenter.

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
    `flyte-core helm chart <https://github.com/flyteorg/flyte/tree/master/charts/flyte-core>`__, please ensure:

    * You have the correct kubeconfig and have selected the correct Kubernetes context.
    * You have configured the correct flytectl settings in ``~/.flyte/config.yaml``.

.. note::

  Add the Flyte chart repo to Helm if you're installing via the Helm charts.

  .. code-block:: bash

    helm repo add flyteorg https://flyteorg.github.io/flyte

Specify agent configuration
----------------------------

Enable the MMCloud agent by adding the following config to the relevant YAML file(s):

.. code-block:: yaml

  tasks:
    task-plugins:
      enabled-plugins:
        - agent-service
      default-for-task-types:
        - mmcloud_task: agent-service

.. code-block:: yaml

  plugins:
    agent-service:
      agents:
        mmcloud-agent:
          endpoint: <AGENT_ENDPOINT>
          insecure: true
      supportedTaskTypes:
      - mmcloud_task
      agentForTaskTypes:
      - mmcloud_task: mmcloud-agent

Substitute ``<AGENT_ENDPOINT>`` with the endpoint of your MMCloud agent.

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

For MMCloud plugin on the Flyte cluster, please refer to `Memory Machine Cloud Plugin Example <https://docs.flyte.org/en/latest/flytesnacks/examples/mmcloud_plugin/index.html>`_
