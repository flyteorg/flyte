.. _deployment-deployment-multicluster:

##################################
Multiple Kubernetes Cluster Deployment
##################################

.. tags:: Kubernetes, Infrastructure, Advanced

.. note::

    The multicluster deployment described in this section, assumes you have deployed
    the ``flyte-core`` Helm chart, which runs the individual Flyte componentes separately.
    This is needed because in a multicluster setup, the execution engine is
    deployed to multiple K8s clusters; it won't work with the ``flyte-binary``
    Helm chart, since it deploys all Flyte services as one single binary.

Scaling Beyond Kubernetes
-------------------------

.. tip::
   
   As described in the `Architecture Overview <https://docs.flyte.org/en/latest/concepts/architecture.html>`_,
   the Flyte ``Control Plane`` sends workflows off to the ``Data Plane`` for
   execution. The data plane fulfills these workflows by launching pods in
   Kubernetes.


.. image:: https://raw.githubusercontent.com/flyteorg/static-resources/main/common/flyte-multicluster-arch.png

The case for multiple Kubernetes clusters may arise due to security constraints, 
cost effectiveness or a need to scale out computing resources.

To address this, you can deploy Flyte's data plane to multiple Kubernetes clusters.
The control plane (FlyteAdmin) can be configured to load-balance workflows across
these individual data planes, protecting you from failure in a single Kubernetes
cluster, thus increasing scalability.


To deploy *only* the data planes to these clusters, use the `values-dataplane.yaml <https://github.com/flyteorg/flyte/blob/master/charts/flyte-core/values-dataplane.yaml>`__ provided with the Helm chart.

Data Plane Deployment
*********************

This gude assumes that you have three Kubernetes clusters and that you can access
them all with ``kubectl``.

Let's call these clusters ``dataplane1``, ``dataplane2``, and ``dataplane3``.

1. Add the ``flyteorg`` Helm repo:

.. code-block::

    helm repo add flyteorg https://flyteorg.github.io/flyte
    helm repo update
    # Get flyte-core helm chart
    helm fetch --untar --untardir . flyteorg/flyte-core
    cd flyte-core

2. Install Flyte data plane Helm chart:

.. note:: 

   Use here the same ``values-eks`` or ``values-gcp.yaml`` file you used to deploy the controlplane.

.. tabbed:: AWS

    .. code-block::

        helm install flyte-core-data flyteorg/flyte-core -n flyte \ 
        --values values-eks.yaml --values values-dataplane.yaml \ 
        --create-namespace

.. tabbed:: GCP

    .. code-block::

        helm upgrade flyte -n flyte flyteorg/flyte-core -f values.yaml \
            -f values-gcp.yaml \
            -f values-dataplane.yaml \
            --create-namespace flyte --install


User and Control Plane Deployment
*********************************

For ``flyteadmin`` to access and create Kubernetes resources in one or more
Flyte data plane clusters , it needs credentials to each cluster.
Flyte makes use of Kubernetess Service Accounts to enable every data plane cluster to perform
authenticated requests to the Kubernetes API Server.
The default behaviour is that ``flyteadmin`` creates a `ServiceAccount <https://github.com/flyteorg/flyte/blob/master/charts/flyte-core/templates/admin/rbac.yaml#L4>`_
in each data plane cluster. 
In order to verify requests, the API Server expects a `signed bearer token <https://kubernetes.io/docs/reference/access-authn-authz/authentication/#service-account-tokens>`__
attached to the Service Account. As of Kubernetes 1.24 an above, the bearer token has to be generated manually.


1. Use the following manifest to create a long-lived secret for the ``flyteadmin`` Service Account in your dataplane cluster:

   .. prompt:: bash $
   
      kubectl apply -f - <<EOF
      apiVersion: v1
      kind: Secret
      metadata:
        name: dataplane1-token
        namespace: flyte
        annotations:
          kubernetes.io/service-account.name: flyteadmin
      type: kubernetes.io/service-account-token
      EOF
      

2. Create a new file named ``secrets.yaml`` that looks like:

.. code-block:: yaml
   :caption: secrets.yaml

   apiVersion: v1
   kind: Secret
   metadata:
     name: cluster-credentials
     namespace: flyte
   type: Opaque
   stringData:

.. note:: 
  The credentials have two parts (``CA cert`` and ``bearer token``). 

3. Copy the bearer token of the first dataplane cluster's secret to your clipboard using the following command:

.. prompt:: bash $

  kubectl get secret -n flyte cluster-credentials \
      -o jsonpath='{.data.token}' | base64 -D | pbcopy

4. Go to ``secrets.yaml`` and add a new entry under ``stringData`` with the dataplane cluster token:

.. code-block:: yaml
   :caption: secrets.yaml

   apiVersion: v1
   kind: Secret
   metadata:
     name: cluster-credentials
     namespace: flyte
   type: Opaque
   stringData:
     dataplane_1_token: eyJhbGciOiJSUzI1NiIsImtpZCI6IlM0WlhfMm1Yb1U4Z1V4R0t6STZDdkhGTVVvVDBZcDAxbjdVbDc1Y1VxR28ifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJmbHl0ZSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VjcmV0Lm5hbWUiOiJkYXRhcGxhbmUxLXRva2VuIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQubmFtZSI6ImZseXRlYWRtaW4iLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC51aWQiOiJkNTdhNjMwZi00ZTZmLTQzNTgtYjQwOS00M2UyMTlhYjg4NTEiLCJzdWIiOiJzeXN0ZW06c2VydmljZWFjY291bnQ6Zmx5dGU6Zmx5dGVhZG1pbiJ9.Fbn5qJjWP1wyJ08PgZXnrrUdKEhLRYqUzG9Vff1maFO3yBKkv_EBuYc2hjGeW5_ORCrT9qKcFAd3AE_tM3P8AQ-dRoA6K-RcJ2qinxabWmk9RYbtKFr1zujswU6dm-iB7JkjY7yYyBRewbw_m4QRacgG8K11c8bYZ9SZoV86EqGmsNdeCPuv5GiPBiJ0p3hgta4kZ1knCNf8qLBUQVZ-9G5vabYM0lyD6dvGOqlOs1bMzgLeijvpQN471dTLmIZ71anOG2gkuJW_AusnWDF_0rJ3yfISf3dRkhXkLswyq-awgtKbz6ZYjPaJ1eA8dNvSlbDoNrMXOGNlx7p7KhOY-w

5. Obtain the corresponding certificate:

.. prompt:: bash $

  kubectl get secret -n flyte cluster-credentials \
      -o jsonpath='{.data.ca\.crt}' | base64 -D | pbcopy

6. Add another entry on your ``secrets.yaml`` file for the cert, making sure that indentation resembles the following example:

.. code-block:: yaml
   :caption: secrets.yaml

   apiVersion: v1
   kind: Secret
   metadata:
     name: cluster-credentials
     namespace: flyte
   type: Opaque
   stringData:
     dataplane_1_token: eyJhbGciOiJSUzI1NiIsImtpZCI6IlM0WlhfMm1Yb1U4Z1V4R0t6STZDdkhGTVVvVDBZcDAxbjdVbDc1Y1VxR28ifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJmbHl0ZSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VjcmV0Lm5hbWUiOiJkYXRhcGxhbmUxLXRva2VuIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQubmFtZSI6ImZseXRlYWRtaW4iLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC51aWQiOiJkNTdhNjMwZi00ZTZmLTQzNTgtYjQwOS00M2UyMTlhYjg4NTEiLCJzdWIiOiJzeXN0ZW06c2VydmljZWFjY291bnQ6Zmx5dGU6Zmx5dGVhZG1pbiJ9.Fbn5qJjWP1wyJ08PgZXnrrUdKEhLRYqUzG9Vff1maFO3yBKkv_EBuYc2hjGeW5_ORCrT9qKcFAd3AE_tM3P8AQ-dRoA6K-RcJ2qinxabWmk9RYbtKFr1zujswU6dm-iB7JkjY7yYyBRewbw_m4QRacgG8K11c8bYZ9SZoV86EqGmsNdeCPuv5GiPBiJ0p3hgta4kZ1knCNf8qLBUQVZ-9G5vabYM0lyD6dvGOqlOs1bMzgLeijvpQN471dTLmIZ71anOG2gkuJW_AusnWDF_0rJ3yfISf3dRkhXkLswyq-awgtKbz6ZYjPaJ1eA8dNvSlbDoNrMXOGNlx7p7KhOY-w
     dataplane_1_cacert:  |
         -----BEGIN CERTIFICATE-----
         MIIDBTCCAe2gAwIBAgIIQREjtnmWbyYwDQYJKoZIhvcNAQELBQAwFTETMBEGA1UE
         AxMKa3ViZXJuZXRlczAeFw0yMzA5MTIxNzIzMDhaFw0zMzA5MDkxNzIzMDhaMBUx
         EzARBgNVBAMTCmt1YmVybmV0ZXMwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEK
         AoIBAQDn53QComJ6lhauUATrnV7DtDDxreGQDxxDp8HrU0nwvzT5e4ewRJ+6+VKH
         ru6iV8hRSH99XdsbRhb5+HrM9bxwDduTZ4wOsdmI1ghXvbBpOEHQTJFiSoWY82LS
         eyMrlwmo8TU8NUXhN+iE+z+cW/QQUKPnNnDcZYWpWOZYjtdtSoYbvU98/cMrRaNg
         IoDMiC6uWz3aNE9SSodE5IpTQ6VhhmZfU8eGO6+2Nl0l73uVSiKUyaJm/DdyUnp1
         iAx7qMPZw+Bfxa6P8PjrkFTpiccPFsy+9mnmoLfbA07QMx0txMFDb/YGOdBYox7n
         V+yOst26TvfNnl4lW4o7cBzjwEuxAgMBAAGjWTBXMA4GA1UdDwEB/wQEAwICpDAP
         BgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBSPzcH1/5DDrurz+Tu8kWwUZoHlcTAV
         BgNVHREEDjAMggprdWJlcm5ldGVzMA0GCSqGSIb3DQEBCwUAA4IBAQAXcAFgusLs
         PfpqzeQcrmDYnywW067+VLwGn906lpceoJbjxL9NQsHSlluXzS8AqljabbweetKD
         +eYfvSDa+yWHSA0ygS9ddCutMgNtsAm5H8LktKvnhERuZKBDUFYG2HFFlIh5mUak
         5TkaYC3FzBsTUoHg+uBqOPSUKaQhzFsIj4a94oZfpGMF+2Yd7vjTeNjuXPbdpYVK
         2avEma8RucJIhIs5w8pgnclSpNXwyz69HrUJ+FxADot6+YHuirpL31XLFPL/jqX4
         Hde3eDWJs4p6Rr0bOGmolOznGUbLdlBsM1QsHfiipMe7XqrBheNWAQFU+rFeHr8L
         tbjBbrxuMPKV
         -----END CERTIFICATE-----

7. Repeat steps 1-6 for every dataplane cluster in your environment.
8. Connect to your controlplane cluster and create the ``cluster-credentials`` secret:

.. prompt:: bash $

    kubectl apply -f secrets.yaml

9. Create a file named ``values-override.yaml`` and add the following config to it:

.. code-block:: yaml
   :caption: values-override.yaml

   flyteadmin:
     additionalVolumes:
     - name: cluster-credentials
       secret:
         secretName: cluster-credentials
     additionalVolumeMounts:
     - name: cluster-credentials
       mountPath: /var/run/credentials
   configmap:
     clusters:
      labelClusterMap:
        project1: 
        - id: dataplane_1
          weight: 1
        project2:
        - id: dataplane_2
          weight: 0.5
        - id: dataplane_3
          weight: 0.5
      clusterConfigs:
      - name: "dataplane_1"
        endpoint: https://dataplane-1-kubeapi-endpoint.com:443
        enabled: true
        auth:
           type: "file_path"
           tokenPath: "/var/run/credentials/dataplane_1_token"
           certPath: "/var/run/credentials/dataplane_1_cacert"
      - name: "dataplane_2"
        endpoint: https://dataplane-2-kubeapi-endpoint.com:443
        enabled: true
        auth:
            type: "file_path"
            tokenPath: "/var/run/credentials/dataplane_2_token"
            certPath: "/var/run/credentials/dataplane_2_cacert"
      - name: "dataplane_3"
        endpoint: https://dataplane-3-kubeapi-endpoint.com:443
        enabled: true
        auth:
            type: "file_path"
            tokenPath: "/var/run/credentials/dataplane_3_token"
            certPath: "/var/run/credentials/dataplane_3_cacert"

.. note:: 
   
   Typically, you can obtain your Kubernetes endpoint URL using the following command:

   .. prompt:: bash $
      
      kubectl cluster-info

In this configuration, ``team1`` and ``team2`` are just labels that we will use later in the process
to configure the necessary mappings so workflow executions matching those labels, are scheduled
on one or multiple clusters depending on the weight (e.g. ``team1`` on ``dataplane_1``)

10. Update the control plane Helm release:

.. note:: 
   This step will disable ``flytepropeller`` in the control plane cluster, leaving no possibility of running workflows there.

.. tabbed:: AWS

    .. code-block::

        helm upgrade flyte-core flyteorg/flyte-core \ 
        --values values-eks-controlplane.yaml --values values-override.yaml \
        --values values-eks.yaml -n flyte

.. tabbed:: GCP

    .. code-block::

        helm upgrade flyte -n flyte flyteorg/flyte-core values.yaml \
            -f values-gcp.yaml \
            -f values-controlplane.yaml \
            -f values-override.yaml \
            --create-namespace flyte --install

11. Verify that all Pods in the ``flyte`` namespace are ``Running``: 

Example output:

.. prompt:: bash $

   kubectl get pods -n flyte                                                                                                                    ✔ ╱ base  ╱ fthw-controlplane ⎈
   NAME                             READY   STATUS    RESTARTS   AGE
   datacatalog-86f6b9bf64-bp2cj     1/1     Running   0          23h
   datacatalog-86f6b9bf64-fjzcp     1/1     Running   0          23h
   flyteadmin-5bb4c4976d-rdk5l      0/1     Pending   0          23h
   flyteadmin-84f666b6f5-7g65j      1/1     Running   0          23h
   flyteadmin-84f666b6f5-sqfwv      1/1     Running   0          23h
   flyteconsole-cdcb48b56-5qzlb     1/1     Running   0          23h
   flyteconsole-cdcb48b56-zj75l     1/1     Running   0          23h
   flytescheduler-947ccbd6-r8kg5    1/1     Running   0          23h
   syncresources-6d8794bbcb-754wn   1/1     Running   0          23h

12. Verify that your cluster configs landed on the ``flyte-clusterresourcesync-config`` ConfigMap:
   
   .. code-block:: yaml

      clusters.yaml:
      ----
      clusters:
        clusterConfigs:
        - auth:
            certPath: /var/run/credentials/dataplane_1_cacert
            tokenPath: /var/run/credentials/dataplane_1_token
            type: file_path
            enabled: true
            endpoint: https://dataplane-1-kubeapi-endpoint.com:443
            name: dataplane_1
  
        labelClusterMap:
          project1:
          - id: dataplane_1
            weight: 1
    ...


Configure Execution Cluster Labels
**********************************

The next step is to configure project-domain or workflow labels to schedule on a specific
Kubernetes cluster.

.. tabbed:: Configure Project & Domain

   1. Create an ``ecl.yaml`` file with the following contents:
    
   .. code-block:: yaml

      domain: development
      project: project1
      value: project1

   .. note:: 

      Change ``domain`` and ``project`` according to your environment.  The ``value`` file has 
      to match with the entry on ``clusterLabelMap`` that's in your ``flyte-clusterresourcesync-config`` ConfigMap.
      Also, in order to automate the creation of the corresponding ``project-domain`` namespaces in the dataplane, add the following to your ``values-dataplane`` file:
   
      Example:

      .. code-block:: yaml

         initialProjects:
         - project1
    
    
   2. Repeat step 1 for each project-domain mapping you need to configure, creating a YAML file for each one.

   3. Update the  execution cluster label of the project and domain:

      .. prompt:: bash $

         flytectl update execution-cluster-label --attrFile ecl.yaml

      Example output:

      .. prompt:: bash $

         Updated attributes from team1 project and domain development



.. tabbed:: Configure Specific Workflow

    1. Create a ``workflow-ecl.yaml`` file with the following example contents:
    
    .. code-block:: yaml

        domain: development
        project: project1
        workflow: core.control_flow.run_merge_sort.merge_sort
        value: project1

    3. Update execution cluster label of the project and domain

    .. prompt:: bash $

        flytectl update execution-cluster-label \
            -p project1 -d development \
            core.control_flow.run_merge_sort.merge_sort \
            --attrFile workflow-ecl.yaml


Congratulations 🎉! With this, the execution of workflows belonging to a specific
project-domain or a single workflow will be scheduled on the target label
cluster.
