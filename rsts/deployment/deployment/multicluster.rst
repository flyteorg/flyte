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
The control plane (FlyteAdmin) can be configured to submit workflows to
these individual data planes. Additionally, Flyte provides the mechanisms for 
administrators to retain control on the workflow placement logic while enabling
users to reap the benefits using simple abstractions like ``projects`` and ``domains``.

Prerequisites
*************

To make sure that your multicluster deployment is able to scale and process 
requests successfully, the following environment-specific requirements should be met:

.. tabbed:: AWS
   
   1. An IAM Policy that defines the permissions needed for Flyte. A minimum set of permissions include:
   
   .. code-block:: json
      
      "Action": [
          "s3:DeleteObject*",
          "s3:GetObject*",
          "s3:ListBucket",
          "s3:PutObject*"
      ], 
       "Resource": [
                "arn:aws:s3:::<your-S3-bucket>*",
                "arn:aws:s3:::<your-S3-bucket>*/*"
            ],   

   2. At least three IAM Roles configured: one for the controlplane components, another for the dataplane
   and one more for the worker Pods that are bootstraped by Flyte to execute workflow tasks. 

   3. An OIDC Provider associated with each of your EKS clusters. You can use the following command to create and connect the Provider:

   .. prompt:: bash

      eksctl utils associate-iam-oidc-provider --cluster <Name-EKS-Cluster> --approve  

   4. An IAM Trust Relationship that associates each EKS cluster type (controlplane or dataplane) with the Service Account(s) and namespaces 
      where the different elements of the system will run.
  
   Use the steps in this section to complete the requirements indicated above:

   **Control plane role**

   1. Use the following command to simplify the process of both creating a role and configuring an initial Trust Relationship:

   .. prompt:: bash
      
      eksctl create iamserviceaccount --cluster=<controlplane-cluster-name> --name=flyteadmin --role-only --role-name=flyte-controlplane-role --attach-policy-arn <ARN-of-your-IAM-policy> --approve --region <AWS-REGION-CODE> --namespace flyte
      
   2. Go to the **IAM** section in your **AWS Management Console** and select the role that was just created
   3. Go to the **Trust Relationships** tab and **Edit the Trust Policy**
   4. Add the ``datacatalog`` Service Account to the ``sub`` section 
   
   The end result should look similar to the following example:

   .. code-block:: json

                 {
          "Version": "2012-10-17",
          "Statement": [
              {
                  "Effect": "Allow",
                  "Principal": {
                      "Federated": "arn:aws:iam::<ACCOUNT-ID>:oidc-provider/oidc.eks.<REGION>.amazonaws.com/id/<CONTROLPLANE-OIDC-PROVIDER>"
                  },
            "Action": "sts:AssumeRoleWithWebIdentity",
            "Condition": {
                "StringEquals": {
                    "oidc.eks.<REGION>.amazonaws.com/id/<CONTROLPLANE-OIDC-PROVIDER>:aud": "sts.amazonaws.com",
                    "oidc.eks.<REGION>.amazonaws.com/id/<CONTROLPLANE-OIDC-PROVIDER>:sub": [
                        "system:serviceaccount:flyte:flyteadmin",
                        "system:serviceaccount:flyte:datacatalog"
                          ]
                      }
                  }
              }
          ]
      }

   **Data plane role**

   1. Create the role and Trust Relationship:

   .. prompt:: bash
      
      eksctl create iamserviceaccount --cluster=<dataplane1-cluster-name> --name=flytepropeller --role-only --role-name=flyte-dataplane-role --attach-policy-arn <ARN-of-your-IAM-policy> --approve --region <AWS-REGION-CODE> --namespace flyte
      
   2. Verify the Trust Relationship configuration:

   .. prompt:: bash

      aws iam get-role --role-name flyte-dataplane-role --query Role.AssumeRolePolicyDocument
   
   Example output:

   .. code-block:: json

      {
        "Version": "2012-10-17",
        "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Federated": "arn:aws:iam::<ACCOUNT-ID>:oidc-provider/oidc.eks.<REGION>.amazonaws.com/id/<DATAPLANE1-OIDC-PROVIDER>"
            },
            "Action": "sts:AssumeRoleWithWebIdentity",
            "Condition": {
                "StringEquals": {
                    "oidc.eks.us-east-1.amazonaws.com/id/66CBAF563FD1438BC98F1EF39FF8DACD:aud": "sts.amazonaws.com",
                    "oidc.eks.us-east-1.amazonaws.com/id/66CBAF563FD1438BC98F1EF39FF8DACD:sub": "system:serviceaccount:flyte:flytepropeller"
                    }
                }
            }
         ]
      }

   **Workers role**

   1. Create role and initial Trust Relationship:

      .. prompt:: bash
      
      eksctl create iamserviceaccount --cluster=<dataplane1-cluster-name> --name=default --role-only --role-name=flyte-workers-role --attach-policy-arn <ARN-of-your-IAM-policy> --approve --region <AWS-REGION-CODE> --namespace flyte
      
   2. Go to the **IAM** section in your **AWS Management Console** and select the role that was just created
   3. Go to the **Trust Relationships** tab and **Edit the Trust Policy**
   4. By default, every Pod created for Task execution, uses the ``default`` Service Account on their respective namespace. In your cluster, you'll have as many
      namespaces as ``project`` and ``domain`` combinations you may have. Hence, it might be useful to use a ``StringLike`` condition and to set a wildcard for the namespace in the Trust Policy:

      .. code-block:: json

         {
       "Version": "2012-10-17",
       "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Federated": "arn:aws:iam::<ACCOUNT-ID>:oidc-provider/oidc.eks.<REGION>.amazonaws.com/id/<DATAPLANE1-OIDC-PROVIDER>"
            },
            "Action": "sts:AssumeRoleWithWebIdentity",
            "Condition": {
                "StringLike": {
                    "oidc.eks.<REGION>.amazonaws.com/id/<DATAPLANE1-OIDC-PROVIDER>:sub": "system:serviceaccount:*:default",
                    "oidc.eks.<REGION>.amazonaws.com/id/<DATAPLANE1-OIDC-PROVIDER>:aud": "sts.amazonaws.com"
                  }
               }
            }
          ]
       }



Data Plane Deployment
*********************

This guide assumes that you have two Kubernetes clusters and that you can access
them all with ``kubectl``.

Let's call these clusters ``dataplane1`` and ``dataplane2``.

1. Add the ``flyteorg`` Helm repo:

.. prompt:: bash

    helm repo add flyteorg https://flyteorg.github.io/flyte
    helm repo update
    # Get flyte-core helm chart
    helm fetch --untar --untardir . flyteorg/flyte-core
    cd flyte-core

2. Open the ``values-dataplane.yaml`` file and add the following contents:

   .. code-block:: yaml

      configmap:
        admin:
          admin:
            endpoint: <your-Ingress-FQDN>:443 #use here the URL you're using to connect to Flyte
            insecure: false #enables secure communication over SSL. Requires a signed certificate

.. note:: 

   This step is needed so the ``flytepropeller`` instance in the data plane cluster is able to send notifications
   back to the ``flyteadmin`` service in the control plane.

3. Install Flyte data plane Helm chart:

.. note:: 

   Use here the same ``values-eks`` or ``values-gcp.yaml`` file you used to deploy the controlplane.

.. tabbed:: AWS

    .. code-block::

        helm install flyte-core-data flyteorg/flyte-core -n flyte \ 
        --values values-eks.yaml --values values-dataplane.yaml \ 
        --create-namespace

.. tabbed:: GCP

    .. code-block::

        helm install flyte-core-data -n flyte flyteorg/flyte-core  \
            --values values-gcp.yaml \
            --values values-dataplane.yaml \
            --create-namespace flyte 

4. Repeat step 2 and 3 for each dataplane cluster in your environment.

Control Plane Deployment
*********************************

For ``flyteadmin`` to access and create Kubernetes resources in one or more
Flyte data plane clusters , it needs credentials to each cluster.
Flyte makes use of Kubernetess Service Accounts to enable every data plane cluster to perform
authenticated requests to the Kubernetes API Server.
The default behaviour is that ``flyteadmin`` creates a `ServiceAccount <https://github.com/flyteorg/flyte/blob/master/charts/flyte-core/templates/admin/rbac.yaml#L4>`_
in each data plane cluster. 
In order to verify requests, the Kubernetes API Server expects a `signed bearer token <https://kubernetes.io/docs/reference/access-authn-authz/authentication/#service-account-tokens>`__
attached to the Service Account. As of Kubernetes 1.24 and above, the bearer token has to be generated manually.


1. Use the following manifest to create a long-lived bearer token for the ``flyteadmin`` Service Account in your dataplane cluster:

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

  kubectl get secret -n flyte dataplane1-token \
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

  kubectl get secret -n flyte dataplane1-token \
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
     initContainerClusterSyncAdditionalVolumeMounts:
     - name: cluster-credentials
       mountPath: /etc/credentials
   configmap:
     clusters:
      labelClusterMap:
        label1:
        - id: dataplane_1
          weight: 1
        label2:
        - id: dataplane_2
          weight: 1
      clusterConfigs:
      - name: "dataplane_1"
        endpoint: https://<your-dataplane1-kubeapi-endpoint>:443
        enabled: true
        auth:
           type: "file_path"
           tokenPath: "/var/run/credentials/dataplane_1_token"
           certPath: "/var/run/credentials/dataplane_1_cacert"
      - name: "dataplane_2"
        endpoint: https://<your-dataplane2-kubeapi-endpoint>:443
        enabled: true
        auth:
          type: "file_path"
          tokenPath: "/var/run/credentials/dataplane_2_token"
          certPath: "/var/run/credentials/dataplane_2_cacert"

.. note:: 
   
   Typically, you can obtain your Kubernetes API endpoint URL using the following command:

   .. prompt:: bash $
      
      kubectl cluster-info

In this configuration, ``label1`` and ``label2`` are just labels that we will use later in the process
to configure the necessary mappings so workflow executions matching those labels, are scheduled
on one or multiple clusters depending on the weight (e.g. ``label1`` on ``dataplane_1``)

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
            --values values-gcp.yaml \
            --values values-controlplane.yaml \
            --values values-override.yaml

11. Verify that all Pods in the ``flyte`` namespace are ``Running``: 

Example output:

.. prompt:: bash $

   kubectl get pods -n flyte                                                                                                                    ✔ ╱ base  ╱ fthw-controlplane ⎈
   NAME                             READY   STATUS    RESTARTS   AGE
   datacatalog-86f6b9bf64-bp2cj     1/1     Running   0          23h
   datacatalog-86f6b9bf64-fjzcp     1/1     Running   0          23h
   flyteadmin-84f666b6f5-7g65j      1/1     Running   0          23h
   flyteadmin-84f666b6f5-sqfwv      1/1     Running   0          23h
   flyteconsole-cdcb48b56-5qzlb     1/1     Running   0          23h
   flyteconsole-cdcb48b56-zj75l     1/1     Running   0          23h
   flytescheduler-947ccbd6-r8kg5    1/1     Running   0          23h
   syncresources-6d8794bbcb-754wn   1/1     Running   0          23h


Configure Execution Cluster Labels
**********************************

The next step is to configure project-domain or workflow labels to schedule on a specific
Kubernetes cluster.

.. tabbed:: Configure Project & Domain

   1. Create an ``ecl.yaml`` file with the following contents:
    
   .. code-block:: yaml

      domain: development
      project: project1
      value: label1

   .. note:: 

      Change ``domain`` and ``project`` according to your environment.  The ``value`` file has 
      to match with the entry on ``clusterLabelMap`` that's in your ``flyte-clusterresourcesync-config`` ConfigMap.
    
   2. Repeat step 1 for every project-domain mapping you need to configure, creating a YAML file for each one.

   3. Update the  execution cluster label of the project and domain:

      .. prompt:: bash $

         flytectl update execution-cluster-label --attrFile ecl.yaml

      Example output:

      .. prompt:: bash $

         Updated attributes from team1 project and domain development


   4. Execute a workflow indicating project and domain:

      .. prompt:: bash $

         pyflyte run --remote --project team1 --domain development example.py  training_workflow \                                                          ✔ ╱ docs-development-env 
         --hyperparameters '{"C": 0.1}'

.. tabbed:: Configure a Specific Workflow mapping

    1. Create a ``workflow-ecl.yaml`` file with the following example contents:
    
    .. code-block:: yaml

        domain: development
        project: project1
        workflow: example.training_workflow
        value: project1

    2. Update execution cluster label of the project and domain

    .. prompt:: bash $

        flytectl update execution-cluster-label \
            -p project1 -d development \
            example.training_workflow \
            --attrFile workflow-ecl.yaml

    3. Execute a workflow indicationg project and domain:

      .. prompt:: bash $

         pyflyte run --remote --project team1 --domain development example.py  training_workflow \                                                          ✔ ╱ docs-development-env 
         --hyperparameters '{"C": 0.1}'

Congratulations 🎉! With this, the execution of workflows belonging to a specific
project-domain or a single specific workflow will be scheduled on the target label
cluster.
