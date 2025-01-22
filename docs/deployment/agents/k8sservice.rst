.. _deployment-agent-setup-k8sservice:

Kubernetes (K8s) Service Agent
==============================

The Kubernetes (K8s) Data Service Agent enables machine learning (ML) users to efficiently handle non-training tasks—such as data loading, caching, and processing—concurrently with training jobs in Kubernetes clusters.
This capability is particularly valuable in deep learning applications, such as those in Graph Neural Networks (GNNs).

This guide offers a comprehensive overview of setting up the K8s Data Service Agent within your Flyte deployment.

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

Enable the K8s service agent by adding the following config to the relevant YAML file(s):

.. code-block:: yaml

  tasks:
    task-plugins:
      enabled-plugins:
        - agent-service
      default-for-task-types:
        - dataservicetask: agent-service

.. code-block:: yaml

  plugins:
    agent-service:
      agents:
        k8sservice-agent:
          endpoint: <AGENT_ENDPOINT>
          insecure: true
      agentForTaskTypes:
      - dataservicetask: k8sservice-agent
      - sensor: k8sservice-agent

Substitute ``<AGENT_ENDPOINT>`` with the endpoint of your MMCloud agent.


Setup the RBAC
----------------------

The K8s Data Service Agent will create a StatefulSet and expose the Service endpoint for the StatefulSet pods.
RBAC needs to be set up to allow the K8s Data Service Agent to perform CRUD operations on the StatefulSet and Service.

The role `flyte-flyteagent-role` set up:
.. code-block:: yaml

  # Example of the role/binding set up for the data service to create/update/delete resources in the sandbox flyte namespace
  apiVersion: rbac.authorization.k8s.io/v1
  kind: Role
  metadata:
    name: flyte-flyteagent-role
    namespace: flyte
    labels:
      app.kubernetes.io/name: flyteagent
      app.kubernetes.io/instance: flyte
  rules:
  - apiGroups:
        - apps
    resources:
      - statefulsets
      - statefulsets/status
      - statefulsets/scale
      - statefulsets/finalizers
    verbs:
      - get
      - list
      - watch
      - create
      - update
      - delete
      - patch
  - apiGroups:
    - ""
    resources:
    - pods
    - configmaps
    - serviceaccounts
    - secrets
    - pods/exec
    - pods/log
    - pods/status
    - services
    verbs:
    - '*'

  
The binding `flyte-flyteagent-rolebinding` for the role `flyte-flyteagent-role`
  .. code-block:: yaml
    # Example of the role/binding set up for the data service to create/update/delete resources in the sandbox flyte namespace
    apiVersion: rbac.authorization.k8s.io/v1
    kind: RoleBinding
    metadata:
      name: flyte-flyteagent-rolebinding
      namespace: flyte
      labels:
        app.kubernetes.io/name: flyteagent
        app.kubernetes.io/instance: flyte
    roleRef:
      apiGroup: rbac.authorization.k8s.io
      kind: Role
      name: flyte-flyteagent-role
    subjects:
    - kind: ServiceAccount
      name: flyteagent
      namespace: flyte

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
