.. _howto-multi-cluster:

#########################################
How Do I use Multiple Kubernetes Clusters
#########################################

Scaling Beyond Kubernetes
-------------------------

As described in the high-level architecture doc, the Flyte Control Plane sends workflows off to the Data Plane for execution.
The Data Plane fulfills these workflows by launching pods in kubernetes.

At some point, your total compute needs could exceed the limits of a single kubernetes cluster.
To address this, you can deploy the Data Plane to several isolated kubernetes clusters.
The Control Plane (FlyteAdmin) can be configured to load-balance workflows across these isolated Data Planes.
This protects you from a failure in a single kubernetes cluster, and increases scalability.

First, you'll need to create additional kubernetes clusters. For this example, we'll assume you have 3 kubernetes clusters, and can access them all with ``kubectl``. We'll call these clusters "cluster1", "cluster2", and "cluster3".

We want to deploy **just** the Data Plane to these clusters. To do this, we'll remove the DataPlane components from the ``flyte`` overlay, and create a new overlay containing **only** the dataplane resources.

Data Plane Deployment
*********************

NOTE:
  With v0.8.0 and the entire setup overhaul, this section is getting revisited. Keep on the lookout for an update soon

To create the "Data Plane only" overlay, lets make a ``dataplane`` subdirectory inside our main deployment directory (my-flyte-deployment). This directory will contain contain only the dataplane resources. ::

  mkdir dataplane

Now, lets copy the ``flyte`` config into the dataplane config ::

  cp flyte/kustomization.yaml dataplane/kustomization.yaml

Since the dataplane resources will live in the new deployment, they are no longer needed in the main ``flyte`` deployment. Remove the Data Plane resources from the flyte deploy by opening ``flyte/kustomization.yaml`` and removing everything in the ``DATA PLANE RESOURCES`` section.

Likewise, the User Plane / Control Plane resources are not needed in the dataplane deployment. Remove these resources from the dataplane deploy by opening ``dataplane/kustomization.yaml`` and removing everything in the ``USER PLANE / CONTROL PLANE RESOURCES`` section.

Now Run ::

  kustomize build dataplane > dataplane_generated.yaml

You will notice that the only the Data Plane resources are included in this file.

You can point your ``kubectl`` context at each of the 3 clusters and deploy the dataplane with ::

  kubectl apply -f dataplane_generated.yaml

User and Control Plane Deployment
*********************************

In order for FlyteAdmin to create "flyteworkflows" on the 3 remote clusters, it will need a secret ``token`` and ``cacert`` to access each cluster.

Once you have deployed the dataplane as described above, you can retrieve the "token" and "cacert" by pointing your ``kubectl`` context each dataplane cluster and executing the following commands.

:token:
  ``kubectl get secrets -n flyte | grep flyteadmin-token | awk '{print $1}' | xargs kubectl get secret -n flyte -o jsonpath='{.data.token}'``

:cacert:
  ``kubectl get secrets -n flyte | grep flyteadmin-token | awk '{print $1}' | xargs kubectl get secret -n flyte -o jsonpath='{.data.ca\.crt}'``

These credentials will need to be included in the Control Plane. Create a new file ``admindeployment/secrets.yaml`` that looks like this ::

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

Include the new ``secrets.yaml`` file in the ``admindeployment`` by opening ``admindeployment/kustomization.yaml`` and add the following line under ``resources:`` to include the secrets in the deploy ::

  - secrets.yaml

Next, we'll need to attach these secrets to the FlyteAdmin pods so that FlyteAdmin can access them. Open ``admindeployment/deployment.yaml`` and add an entry under ``volumes:`` ::

  volumes:
  - name: cluster_credentials
    secret:
      secretName: cluster_credentials

Now look for the container labeled ``flyteadmin``. Add a ``volumeMounts`` to that section. ::

  volumeMounts:
  - name: cluster_credentials
    mountPath: /var/run/credentials

This mounts the credentials inside the FlyteAdmin pods, however, FlyteAdmin needs to be configured to use these credentials. Open the ``admindeployment/configmap.yaml`` file and add a ``clusters`` key to the configmap, with an entry for each cluster ::

  clusters:
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

Now re-run ::

  kustomize build flyte > flyte_generated.yaml

You will notice that the Data Plane resources have been removed from the ``flyte_generated.yaml`` file, and your new configurations have been added.

Deploy the user/control plane to one cluster (you could use one of 3 existing clusters, or an entirely separate cluster). ::

  kubectl apply -f flyte_generated.yaml


FlyteAdmin Remote Cluster Access
*********************************

Some deployments of Flyte may choose to run the control plane separate from the data plane. Flyte Admin is designed to create kubernetes resources in one or more Flyte data plane clusters. For Admin to access remote clusters, it needs credentials to each cluster. In kubernetes, scoped service credentials are created by configuring a “Role” resource in a Kubernetes cluster. When you attach that role to a “ServiceAccount”, Kubernetes generates a bearer token that permits access. We create a flyteadmin `ServiceAccount <https://github.com/lyft/flyte/blob/c0339e7cc4550a9b7eb78d6fb4fc3884d65ea945/artifacts/base/adminserviceaccount/adminserviceaccount.yaml>`_ in each data plane cluster to generate these tokens.

When you first create the Flyte Admin ServiceAccount in a new cluster, a bearer token is generated, and will continue to allow access unless the ServiceAccount is deleted. Once we create the Flyte Admin ServiceAccount on a cluster, we should never delete it. In order to feed the credentials to Flyte Admin, you must retrieve them from your new data plane cluster, and upload them to Admin somehow (within Lyft, we use Confidant for example).

The credentials have two parts (ca cert, bearer token). Find the generated secret via ::

  kubectl get secrets -n flyte | grep flyteadmin-token

Once you have the name of the secret, you can copy the ca cert to your clipboard with ::

  kubectl get secret -n flyte {secret-name} -o jsonpath='{.data.ca\.crt}' | base64 -D | pbcopy

You can copy the bearer token to your clipboard with ::

  kubectl get secret -n flyte {secret-name} -o jsonpath='{.data.token}’ | base64 -D | pbcopy

