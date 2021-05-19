##################################
Using Multiple Kubernetes Clusters
##################################

Scaling Beyond Kubernetes
-------------------------

.. tip::
  As described in the `Architecture Overview <https://docs.flyte.org/en/latest/concepts/architecture.html>`_, the Flyte “Control Plane “sends workflows off to the “Data Plane “for execution. The data-plane fulfills these workflows by launching pods in Kubernetes.

Often, the total compute needs could exceed the limits of a single Kubernetes cluster. 
To address this, you can deploy the data-plane to several isolated Kubernetes clusters.
The control-plane (FlyteAdmin) can be configured to load-balance workflows across these isolated data-planes, which protects you from failure in a single Kubernetes cluster, increasing scalability.

To achieve this, first, you have to create additional Kubernetes clusters. 
For now, let’s assume you have three Kubernetes clusters and that you can access them all with “kubectl “. Let’s call these clusters "cluster1", "cluster2", and "cluster3".

Next, deploy **just** the data-planes to these clusters. To do this, remove the data-plane components from the “flyte “overlay, and create a new overlay containing **only** the data-plane resources.

Data Plane Deployment
*********************

.. NOTE::
  With v0.8.0 and the entire setup overhaul, this section will get updated soon!

To create the “data-plane only” overlay, create a “data-plane “subdirectory inside the main deployment directory (“my-flyte-deployment “). This directory will only contain the data-plane resources. ::

  mkdir dataplane

Now, copy the ``flyte`` config to the data-plane config. ::

  cp flyte/kustomization.yaml dataplane/kustomization.yaml

Since the data-plane resources will live in the new deployment, they are no longer needed in the main “flyte “deployment. Remove the data-plane resources from the flyte deployment by opening “flyte/kustomization.yaml “and removing everything in the “DATA PLANE RESOURCES “section.

Likewise, the user-plane/control-plane resources are not needed in the data-plane deployment. Remove these resources from the data-plane deployment by opening “dataplane/kustomization.yaml “and removing everything in the “USER PLANE/CONTROL PLANE RESOURCES “section. ::

  kustomize build dataplane > dataplane_generated.yaml

You will notice that only the data-plane resources are included in this file.

You can point your “kubectl “context at each of the three clusters and deploy the data-plane using the following command: ::

  kubectl apply -f dataplane_generated.yaml

User and Control Plane Deployment
*********************************

For FlyteAdmin to create “flyteworkflows” on the three remote clusters, it will need a secret “token “and “cacert “to access each cluster.

Once you have deployed the data-plane as described above, you can retrieve the “token” and “cacert” by pointing your “kubectl “context to each data-plane cluster and executing the following commands:

:token:
  ``kubectl get secrets -n flyte | grep flyteadmin-token | awk '{print $1}' | xargs kubectl get secret -n flyte -o jsonpath='{.data.token}'``

:cacert:
  ``kubectl get secrets -n flyte | grep flyteadmin-token | awk '{print $1}' | xargs kubectl get secret -n flyte -o jsonpath='{.data.ca\.crt}'``

Now, these credentials need to be included in the control-plane. Create a new file named ``admindeployment/secrets.yaml`` that looks like: ::

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

Include the new “secrets.yaml “file in the “admindeployment “by opening “admindeployment/kustomization.yaml “and add the following line under “resources: “to include the secrets in the deploy. ::

  - secrets.yaml

Next, you need to attach these secrets to the FlyteAdmin pods so that FlyteAdmin can access them. Open ``admindeployment/deployment.yaml`` and add an entry under ``volumes``. ::

  volumes:
  - name: cluster_credentials
    secret:
      secretName: cluster_credentials

Look for the container labeled ``flyteadmin``. Add a ``volumeMounts`` to that section. ::

  volumeMounts:
  - name: cluster_credentials
    mountPath: /var/run/credentials

This mounts the credentials inside the FlyteAdmin pods. 

However, FlyteAdmin needs to be configured to use these credentials. Open the ``admindeployment/configmap.yaml`` file and add a ``clusters`` key to the configmap, with an entry for each cluster. ::

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

Now re-run the following command to emit a YAML stream. ::

  kustomize build flyte > flyte_generated.yaml

You will notice that the data-plane resources have been removed from the “flyte_generated.yaml “file, and your new configurations have been added.

Deploy the user-plane/control-plane to one cluster (you can use one of the three existing clusters or an entirely separate cluster). ::

  kubectl apply -f flyte_generated.yaml

FlyteAdmin Remote Cluster Access
*********************************

Some deployments of Flyte may choose to run the control-plane separate from the data-plane. FlyteAdmin is designed to create Kubernetes resources in one or more Flyte data-plane clusters. 
For the admin to access remote clusters, it needs credentials to each cluster. 
In Kubernetes, scoped service credentials are created by configuring a “Role” resource in a Kubernetes cluster. 
When you attach that role to a “ServiceAccount”, Kubernetes generates a bearer token that permits access. We create a FlyteAdmin `ServiceAccount <https://github.com/flyteorg/flyte/blob/c0339e7cc4550a9b7eb78d6fb4fc3884d65ea945/artifacts/base/adminserviceaccount/adminserviceaccount.yaml>`_ in each data-plane cluster to generate these tokens.

When you first create the FlyteAdmin ServiceAccount in a new cluster, a bearer token is generated and will continue to allow access unless the “ServiceAccount “is deleted. Once we create the Flyte Admin ServiceAccount on a cluster, we should never delete it. To feed the credentials to FlyteAdmin, you must retrieve them from your new data-plane cluster, and upload them to admin somehow (within Lyft, we use Confidant, for example).

The credentials have two parts (“ca cert “and “bearer token “). Find the generated secret via, ::

  kubectl get secrets -n flyte | grep flyteadmin-token

Once you have the name of the secret, you can copy the ``ca cert`` to your clipboard using the following command: ::

  kubectl get secret -n flyte {secret-name} -o jsonpath='{.data.ca\.crt}' | base64 -D | pbcopy

You can copy the bearer token to your clipboard using the following command: ::

  kubectl get secret -n flyte {secret-name} -o jsonpath='{.data.token}’ | base64 -D | pbcopy
