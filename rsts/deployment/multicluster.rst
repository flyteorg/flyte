.. _multicluster-setup:


##################################
Using Multiple Kubernetes Clusters
##################################

Scaling Beyond Kubernetes
-------------------------

.. tip::
  As described in the `Architecture Overview <https://docs.flyte.org/en/latest/concepts/architecture.html>`_, the Flyte ``Control Plane`` sends workflows off to the ``Data Plane`` for execution. The data plane fulfills these workflows by launching pods in Kubernetes.

Often, the total compute needs could exceed the limits of a single Kubernetes cluster.
To address this, you can deploy the data plane to several isolated Kubernetes clusters.
The control plane (FlyteAdmin) can be configured to load-balance workflows across these isolated data planes, protecting you from failure in a single Kubernetes cluster increasing scalability.

To achieve this, first, you have to create additional Kubernetes clusters.
For now, let's assume you have three Kubernetes clusters and that you can access them all with ``kubectl``.
Let's call these clusters ``cluster1``, ``cluster2``, and ``cluster3``.

Next, deploy **just** the data planes to these clusters.
To do this, remove the data plane components from the **flyte** overlay, and create a new overlay containing **only** the data plane resources.


Data Plane Deployment
*********************

* Add Flyteorg Helm repo

.. code-block::

    helm repo add flytyteorg https://flyteorg.github.io/flyte
    helm repo update
    # Get flyte-core helm chart
    helm fetch --untar --untardir . flyteorg/flyte-core
    cd flyte-core

* Install Flyte data plane Helm chart

.. tabbed:: AWS

    .. code-block::

        helm upgrade flyte -n flyte flyteorg/flyte-core values.yaml -f values-gcp.yaml -f values-dataplane.yaml --create-namespace flyte --install

.. tabbed:: GCP

    .. code-block::

        helm upgrade flyte -n flyte flyteorg/flyte-core values.yaml -f values-aws.yaml -f values-dataplane.yaml --create-namespace flyte --install


User and Control Plane Deployment
*********************************

Some Flyte deployments may choose to run the control plane separate from the data plane.
FlyteAdmin is designed to create Kubernetes resources in one or more Flyte data plane clusters.
For the admin to access remote clusters, it needs credentials to each cluster.

In Kubernetes, scoped service credentials are created by configuring a "Role" resource in a Kubernetes cluster.
When you attach the role to a "ServiceAccount", Kubernetes generates a bearer token that permits access.
Hence, create a FlyteAdmin `ServiceAccount <https://github.com/flyteorg/flyte/blob/master/charts/flyte-core/templates/admin/rbac.yaml#L4>`_ in each data plane cluster to generate these tokens.

When you first create the FlyteAdmin ``ServiceAccount`` in a new cluster, a bearer token is generated and will continue to allow access unless the "ServiceAccount" is deleted.
Hence, you should never delete a ``ServiceAccount`` ⚠️.

To feed the credentials to FlyteAdmin, you must retrieve them from your new data plane cluster and upload them to admin (for example, within Lyft, ref:`Confidant <https://github.com/lyft/confidant>`__ is used).

The credentials have two parts ("ca cert" and "bearer token"). Find the generated secret via,::

  kubectl get secrets -n flyte | grep flyteadmin-token

Once you have the name of the secret, you can copy the ``ca cert`` to your clipboard using the following command: ::

  kubectl get secret -n flyte {secret-name} -o jsonpath='{.data.ca\.crt}' | base64 -D | pbcopy

You can copy the bearer token to your clipboard using the following command: ::

  kubectl get secret -n flyte {secret-name} -o jsonpath='{.data.token}' | base64 -D | pbcopy

Now these credentials need to be included in the control plane. 

* Create a new file named ``secrets.yaml`` that looks like: ::

      apiVersion: v1
      kind: Secret
      metadata:
        name: cluster_credentials
        namespace: flyte
      type: Opaque
      data:
        cluster_1_token: {{ cluster 1 token here }}
        cluster_1_cacert: {{ cluster 1 cacert here }}
        cluster_2_token: {{ cluster 2 token here }}
        cluster_2_cacert: {{ cluster 2 cacert here }}
        cluster_3_token: {{ cluster 3 token here }}
        cluster_3_cacert: {{ cluster 3 cacert here }}

* Create cluster credentials secret in the control plane cluster

.. code-block::

    kubectl apply -f secrets.yaml


* Create a file named ``values-override.yaml`` and add the following config to it: ::

      flyteadmin:
        additionalVolumes:
        - name: cluster_credentials
          secret:
            secretName: cluster_credentials
        additionalVolumeMounts:
        - name: cluster_credentials
          mountPath: /var/run/credentials
      configmap:
        labelClusterMap:
          team1:
            - id: cluster_1
              weight: 1
          team2:
            - id: cluster_2
              weight: 0.5
            - id: cluster_3
              weight: 0.5
        clusterConfigs:
        - name: "cluster_1"
          endpoint: {{ your-cluster-1-kubeapi-endpoint.com }}
          enabled: true
          auth:
            type: "file_path"
            tokenPath: "/var/run/credentials/cluster_1_token"
            certPath: "/var/run/credentials/cluster_1_cacert"
        - name: "cluster_2"
          endpoint: {{ your-cluster-2-kubeapi-endpoint.com }}
          auth:
            enabled: true
            type: "file_path"
            tokenPath: "/var/run/credentials/cluster_2_token"
            certPath: "/var/run/credentials/cluster_2_cacert"
        - name: "cluster_3"
          endpoint: {{ your-cluster-3-kubeapi-endpoint.com }}
          enabled: true
          auth:
            type: "file_path"
            tokenPath: "/var/run/credentials/cluster_3_token"
            certPath: "/var/run/credentials/cluster_3_cacert"

* Install Flyte control plane Helm chart

.. tabbed:: AWS

    .. code-block::

        helm upgrade flyte -n flyte flyteorg/flyte-core values.yaml -f values-aws.yaml -f values-controlplane.yaml -f values-override.yaml --create-namespace flyte --install

.. tabbed:: GCP

    .. code-block::

        helm upgrade flyte -n flyte flyteorg/flyte-core values.yaml -f values-gcp.yaml -f values-controlplane.yaml -f values-override.yaml --create-namespace flyte --install

Configure Execution Cluster Labels
**********************************

.. tabbed:: Configure Project & Domain

    * Get execution cluster label of the project and domain

    .. code-block::

        flytectl get execution-cluster-label -p flytesnacks -d development --attrFile ecl.yaml

    * Update the label in `ecl.yaml` ::

        domain: development
        project: flytesnacks
        value: team1

.. tabbed:: Configure Specific Workflow

    * Get execution cluster label of the project and domain

    .. code-block::

        flytectl get execution-cluster-label -p flytesnacks -d development core.control_flow.run_merge_sort.merge_sort --attrFile ecl.yaml

    * Update the label in `ecl.yaml` ::

        domain: development
        project: flytesnacks
        workflow: core.control_flow.run_merge_sort.merge_sort
        value: team1

* Lastly, update the execution cluster label

.. code-block::

    flytectl update execution-cluster-label --attrFile ecl.yaml