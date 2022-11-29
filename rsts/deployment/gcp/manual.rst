.. _deployment-gcp-manual:

###############
GCP (GKE) Setup
###############

************************************************
Flyte Deployment - Manual GCE/GKE Deployment
************************************************
This guide helps you set up Flyte from scratch, on GCE, without using an automated approach. It details step-by-step instructions to go from a bare GCE account to a fully functioning Flyte deployment that members of your company can use.

Prerequisites
=============
* Access to `GCE console <https://console.cloud.google.com/>`__
* A domain name for the Flyte installation like flyte.example.org that allows you to set a DNS A record.

Before you begin, please ensure that you have the following tools installed:

* `gcloud <https://cloud.google.com/sdk/docs/install>`__
* `Helm <https://helm.sh/docs/intro/install/>`__
* `kubectl <https://kubernetes.io/docs/tasks/tools/>`__
* `flytectl <https://docs.flyte.org/projects/flytectl/en/stable/#install>`__

Initialize Gcloud
===================
Authorize Gcloud sdk to access GCP using your credentials, setup the config for the existing project,
and optionally set the default compute zone. `Init <https://cloud.google.com/sdk/gcloud/reference/init>`__

.. code-block:: bash

   gcloud init


Create an Organization (Optional)
=================================
This step is optional if you already have an organization linked to a billing account.
Use the following docs to understand the organization creation process in Google cloud:
`Organization Management <https://cloud.google.com/resource-manager/docs/creating-managing-organization>`__.

Get the organization ID to be used for creating the project. Billing should be linked with the organization so
that all projects under the org use the same billing account.

.. code-block:: bash

   gcloud organizations list

Sample output
.. code-block::

   DISPLAY_NAME                 ID    DIRECTORY_CUSTOMER_ID
   example-org        123456789999                C02ewszsz

<id> = ID column value for the organization

.. code-block:: bash

   export ORG_ID=<id>

Create a GCE Project
====================
.. code-block:: bash

  export PROJECT_ID=<my-project>
  gcloud projects create $PROJECT_ID --organization $ORG_ID

Of course you can also use an existing project if your account has appropriate permissions to create the required resources.

Set project <my-project> as the default in gcloud or use gcloud init to set this default:

.. code-block:: bash

  gcloud config set project ${PROJECT_ID}

We assume that <my-project> has been set as the default for all gcloud commands below.

Permissions
===========

Configure `workload identity <https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity>`__ for Flyte namespace service accounts.
This creates the GSA's which would be mapped to the KSA (kubernetes service account) through annotations and used for authorizing the pods access to the Google cloud services.

* Create a GSA for flyteadmin

.. code-block:: bash

  gcloud iam service-accounts create gsa-flyteadmin

* Create a GSA for flytescheduler

.. code-block:: bash

  gcloud iam service-accounts create gsa-flytescheduler

* Create a GSA for datacatalog

.. code-block:: bash

  gcloud iam service-accounts create gsa-datacatalog

* Create a GSA for flytepropeller

.. code-block:: bash

  gcloud iam service-accounts create gsa-flytepropeller


* Create a GSA for cluster resource manager

Production

.. code-block:: bash

  gcloud iam service-accounts create gsa-production

Staging

.. code-block:: bash

  gcloud iam service-accounts create gsa-staging

Development

.. code-block:: bash

  gcloud iam service-accounts create gsa-development

* Create a new role DataCatalogRole with following permissions
   * storage.buckets.get
   * storage.objects.create
   * storage.objects.delete
   * storage.objects.update
   * storage.objects.get
* Create a new role FlyteAdminRole with following permissions
   * storage.buckets.get
   * storage.objects.create
   * storage.objects.delete
   * storage.objects.get
   * storage.objects.getIamPolicy
   * storage.objects.update
   * iam.serviceAccounts.signBlob
* Create a new role FlyteSchedulerRole with following permissions
   * storage.buckets.get
   * storage.objects.create
   * storage.objects.delete
   * storage.objects.get
   * storage.objects.getIamPolicy
   * storage.objects.update
* Create a new role FlytePropellerRole with following permissions
   * storage.buckets.get
   * storage.objects.create
   * storage.objects.delete
   * storage.objects.get
   * storage.objects.getIamPolicy
   * storage.objects.update
* Create a new role FlyteWorkflowRole with following permissions
   * storage.buckets.get
   * storage.objects.create
   * storage.objects.delete
   * storage.objects.get
   * storage.objects.list
   * storage.objects.update

Refer the following `role <https://cloud.google.com/iam/docs/understanding-roles>`__ page for more details.

* Add IAM policy binding for flyteadmin GSA using FlyteAdminRole.

.. code-block:: bash

  gcloud projects add-iam-policy-binding ${PROJECT_ID}  --member "serviceAccount:gsa-flyteadmin@${PROJECT_ID}.iam.gserviceaccount.com"    --role "projects/${PROJECT_ID}/roles/FlyteAdminRole"

* Add IAM policy binding for flytescheduler GSA using FlyteSchedulerRole.

.. code-block:: bash

  gcloud projects add-iam-policy-binding ${PROJECT_ID}  --member "serviceAccount:gsa-flytescheduler@${PROJECT_ID}.iam.gserviceaccount.com"    --role "projects/${PROJECT_ID}/roles/FlyteSchedulerRole"

* Add IAM policy binding for datacatalog GSA using DataCatalogRole.

.. code-block:: bash

   gcloud projects add-iam-policy-binding ${PROJECT_ID}  --member "serviceAccount:gsa-datacatalog@${PROJECT_ID}.iam.gserviceaccount.com"    --role "projects/${PROJECT_ID}/roles/DataCatalogRole"

* Add IAM policy binding for flytepropeller GSA using FlytePropellerRole.

.. code-block:: bash

  gcloud projects add-iam-policy-binding ${PROJECT_ID}  --member "serviceAccount:gsa-flytepropeller@${PROJECT_ID}.iam.gserviceaccount.com"    --role "projects/${PROJECT_ID}/roles/FlytePropellerRole"

* Add IAM policy binding for cluster resource manager GSA using FlyteWorkflowRole.

Production

.. code-block:: bash

  gcloud projects add-iam-policy-binding ${PROJECT_ID}  --member "serviceAccount:gsa-production@${PROJECT_ID}.iam.gserviceaccount.com"    --role "projects/${PROJECT_ID}/roles/FlyteWorkflowRole"


Staging

.. code-block:: bash

  gcloud projects add-iam-policy-binding ${PROJECT_ID}  --member "serviceAccount:gsa-staging@${PROJECT_ID}.iam.gserviceaccount.com"    --role "projects/${PROJECT_ID}/roles/FlyteWorkflowRole"

Development

.. code-block:: bash

  gcloud projects add-iam-policy-binding ${PROJECT_ID}  --member "serviceAccount:gsa-development@${PROJECT_ID}.iam.gserviceaccount.com"    --role "projects/${PROJECT_ID}/roles/FlyteWorkflowRole"


* Allow the Kubernetes service account to impersonate the Google service account by creating an IAM policy binding the two. This binding allows the Kubernetes Service account to act as the Google service account.

flyteadmin

.. code-block:: bash

 gcloud iam service-accounts add-iam-policy-binding --role "roles/iam.workloadIdentityUser" --member "serviceAccount:${PROJECT_ID}.svc.id.goog[flyte/flyteadmin]" gsa-flyteadmin@${PROJECT_ID}.iam.gserviceaccount.com


flytepropeller

.. code-block:: bash

 gcloud iam service-accounts add-iam-policy-binding --role "roles/iam.workloadIdentityUser" --member "serviceAccount:${PROJECT_ID}.svc.id.goog[flyte/flytepropeller]" gsa-flytepropeller@${PROJECT_ID}.iam.gserviceaccount.com

datacatalog

.. code-block:: bash

 gcloud iam service-accounts add-iam-policy-binding --role "roles/iam.workloadIdentityUser" --member "serviceAccount:${PROJECT_ID}.svc.id.goog[flyte/datacatalog]" gsa-datacatalog@${PROJECT_ID}.iam.gserviceaccount.com

Cluster Resource Manager

We create binding for production, staging and development domains for the Flyte workflows to use.

Production

.. code-block:: bash

 gcloud iam service-accounts add-iam-policy-binding --role "roles/iam.workloadIdentityUser" --member "serviceAccount:${PROJECT_ID}.svc.id.goog[production/default]" gsa-production@${PROJECT_ID}.iam.gserviceaccount.com

Staging

.. code-block:: bash

 gcloud iam service-accounts add-iam-policy-binding --role "roles/iam.workloadIdentityUser" --member "serviceAccount:${PROJECT_ID}.svc.id.goog[staging/default]" gsa-staging@${PROJECT_ID}.iam.gserviceaccount.com

Development

.. code-block:: bash

 gcloud iam service-accounts add-iam-policy-binding --role "roles/iam.workloadIdentityUser" --member "serviceAccount:${PROJECT_ID}.svc.id.goog[development/default]" gsa-development@${PROJECT_ID}.iam.gserviceaccount.com


Create a GKE Cluster
====================
Create a GKE cluster with VPC-native networking and workload identity enabled.
You can enable the GKE workload identity by adding the below lines to `this Dockerfile <https://github.com/flyteorg/flytekit/blob/master/Dockerfile.py3.8>`__. 

.. code-block:: bash

 FROM ghcr.io/flyteorg/flytekit:py3.8-1.0.3

 # Required for gsutil to work with workload-identity
 RUN echo '[GoogleCompute]\nservice_account = default' > /etc/boto.cfg

Adding the above specified lines (``boto.cfg`` configuration) to the Dockerfile also authenticates standalone ``gsutil``. This way, the pod starts without any hiccups. 

Navigate to the gcloud console and Kubernetes Engine tab to start creating the k8s cluster.

Ensure that VPC native traffic routing is enabled under Security enable Workload identity and use project default pool
which would be `${PROJECT_ID}.svc.id.goog`.

It is recommended to create it from the console. This is to make sure the options of VPC-native networking and Workload identity are enabled correctly.
There are multiple commands needed to achieve this. If you create it through the console, it will take care of creating and configuring the right resources:

.. code-block:: bash

  gcloud container clusters create <my-flyte-cluster> \
    --workload-pool=${PROJECT_ID}.svc.id.goog
    --region us-west1 \
    --num-nodes 6

Create a GPU Node Pool in GKE
=============================
Flyte will use the default node pool created by the command above, but if there is need for special hardware like a GPU, you will have to create a node pool that can create nodes with gpus attached.

The first step is to install `this DaemonSet <https://cloud.google.com/kubernetes-engine/docs/how-to/gpus#installing_drivers>`__ that nodes will use to install gpu drivers when initally provisioning the node instance.

.. code-block:: bash

  kubectl apply -f https://raw.githubusercontent.com/GoogleCloudPlatform/container-engine-accelerators/master/nvidia-driver-installer/cos/daemonset-preloaded.yaml

Next you will create the actual node pool. This can be done via the GKE dashboard.

Go to your cluster and enter the cluster, <my-flyte-cluster>. At the top hit the "ADD NODE POOL" button. Since gpu machines can be expensive, it is wise to set the "Number of nodes (per zone)" field to 1. The node pool can be resized to 0 after the node pool creation but during the creation process there needs to be atleast 1. If you do resize the node pool to 0 after creation, then you should check the "Enable cluster autoscaler" box. 

Next in the sidebar you can configure the instance template for the node pool. Select the GPU machine family and adjust the cpu and RAM according to your needs. Then you can hit create at the bottom to create the node pool. 

After the creation of the node pool, you will notice that by default GKE will taint the node pool with nvidia.com/gpu=present. To configure flyte to be able to schedule pods to this node pool, you will have to follow `this <https://docs.flyte.org/projects/cookbook/en/latest/auto/deployment/configure_use_gpus.html#sphx-glr-auto-deployment-configure-use-gpus-py>`__.


Create the GKE Context
======================
Initialize your kubecontext to point to GKE cluster using the following command:

.. code-block:: bash

  gcloud container clusters get-credentials <my-flyte-cluster>

Verify by creating a test namespace:

.. code-block:: bash

   kubectl create ns test

Create a Cloud SQL Database
===========================
Next, create a relational `Cloud SQL for PostgreSQL <https://cloud.google.com/sql/docs/postgres/introduction>`__ database. This database will be used by both the primary control plane service (FlyteAdmin) and the Flyte memoization service (Data Catalog).
Follow this `link <https://console.cloud.google.com/sql/choose-instance-engine>`__ to create the cloud SQL instance.

* Select PostgreSQL
* Provide an Instance ID
* Provide a password for the instance <DB_INSTANCE_PASSWD>
* Use PostgresSQL13 or higher
* Select the Zone based on your availability requirements.
* Select customize your instance and enable Private IP in the Connections tab. This is required for private communication between the GKE apps and cloud SQL instance. Follow the steps to create the private connection (default).
* Create the SQL instance.
* After creation of the instance get the private IP of the database <CLOUD-SQL-IP>.
* Create flyteadmin database and flyteadmin user accounts on that instance with <DBPASSWORD>.
* Verify connectivity to the DB from the GKE cluster.
   * Create a testdb namespace:

   .. code-block:: bash

      kubectl create ns test

   * Verify the connectivity using a postgres client:

   .. code-block:: bash

      kubectl run pgsql-postgresql-client --rm --tty -i --restart='Never' --namespace testdb --image docker.io/bitnami/postgresql:11.7.0-debian-10-r9 --env="PGPASSWORD=<DBPASSWORD>" --command -- psql testdb --host <CLOUD-SQL-IP> -U flyteadmin -d flyteadmin -p 5432

It is recommended to create it from the console. This is to make sure the private IP connectivity works correctly with the cloud SQL instance.
There are multiple commands needed to achieve this. If you create it through the console, it will take care of creating and configuring the right resources.

.. code-block:: bash

  gcloud sql instances create <my-flyte-db> \
    --database-version=POSTGRES_13 \
    --cpu=1 \
    --memory=3840MB \
    --region=us-west1


SSL Certificate
===============
In order to use SSL (which we need to use gRPC clients), we will need to create an SSL certificate. We use `Google-managed SSL certificates <https://cloud.google.com/kubernetes-engine/docs/how-to/managed-certs>`__.

Save the following certificate resource definition as `flyte-certificate.yaml`:

.. code-block:: yaml

  apiVersion: networking.gke.io/v1
  kind: ManagedCertificate
  metadata:
    name: flyte-certificate
  spec:
    domains:
      - flyte.example.org

Then apply it to your cluster:

.. code-block:: bash

  kubectl apply -f flyte-certificate.yaml

.. note::

  ManagedCertificate will only work with GKE ingress, For other ingress please use cert-manager


For nginx ingress please use the certificate manager:

* Install the cert manager

.. code-block:: bash

  helm install cert-manager --namespace flyte --version v0.12.0 jetstack/cert-manager

* Create cert issuer

.. code-block:: yaml

   apiVersion: cert-manager.io/v1alpha2
   kind: Issuer
   metadata:
     name: letsencrypt-production
   spec:
     acme:
       server: https://acme-v02.api.letsencrypt.org/directory
       email: issue-email-id
       privateKeySecretRef:
         name: letsencrypt-production
       solvers:
       - selector: {}
         http01:
           ingress:
             class: nginx

Ingress
=======

* Add the ingress repo

.. code-block:: bash

  helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx


* Install the nginx-ingress

.. code-block:: bash

  helm install nginx-ingress ingress-nginx/ingress-nginx


Create the GCS Bucket
=====================
Create <BUCKETNAME> with uniform access:

.. code-block:: bash

  gsutil mb -b on -l us-west1 gs://<BUCKETNAME>/

Add access permission for the following principals:

* gsa-flytepropeller@${PROJECT_ID}.iam.gserviceaccount.com
* gsa-datacatalog@${PROJECT_ID}.iam.gserviceaccount.com
* gsa-flyteadmin@f${PROJECT_ID}.iam.gserviceaccount.com
* gsa-flyte-clusterresources@${PROJECT_ID}.iam.gserviceaccount.com

Time for Helm
=============

Installing Flyte
-----------------
1. Add the Flyte helm repo

.. code-block:: bash

   helm repo add flyteorg https://flyteorg.github.io/flyte

2. Download the GCP values file

.. code-block:: bash

   curl https://raw.githubusercontent.com/flyteorg/flyte/master/charts/flyte-core/values-gcp.yaml >values-gcp.yaml

3. Update values

.. code-block::

   <RELEASE-NAME> to be used as prefix for ssl certificate secretName
   <PROJECT_ID> of your GCP project
   <CLOUD-SQL-IP> private IP of cloud sql instance
   <DBPASSWORD> of the flyteadmin user created for the cloud sql instance
   <BUCKETNAME> of the GCS bucket created
   <HOSTNAME> to the flyte FQDN (e.g. flyte.example.org)

4. (Optional) Configure Flyte project and domain

To restrict projects, update Helm values. By default, Flyte creates three projects: Flytesnacks, Flytetester, and Flyteexample.

.. code-block::

   # you can define the number of projects as per your need
   flyteadmin:
    initialProjects:
       - flytesnacks
       - flytetester
       - flyteexamples

To restrict domains, update Helm values again. By default, Flyte creates three domains per project: development, staging and production.

.. code-block::

   # -- Domain configuration for Flyte project. This enables the specified number of domains across all projects in Flyte.
   configmap
     domain:
       domains:
         - id: development
           name: development
         - id: staging
           name: staging
         - id: production
           name: production

   # Update Cluster resource manager only if you are using Flyte resource manager. It will create the required resource in the project-domain namespace.
   cluster_resource_manager:
     enabled: true
     config:
       cluster_resources:
          customData:
            - development:
                - projectQuotaCpu:
                  value: "5"
                - projectQuotaMemory:
                  value: "4000Mi"
                - defaultIamRole:
                  value: "gsa-development@{{ .Values.userSettings.googleProjectId }}.iam.gserviceaccount.com"
            - staging:
                - projectQuotaCpu:
                  value: "2"
                - projectQuotaMemory:
                  value: "3000Mi"
                - defaultIamRole:
                  value: "gsa-staging@{{ .Values.userSettings.googleProjectId }}.iam.gserviceaccount.com"
            - production:
                - projectQuotaCpu:
                  value: "2"
                - projectQuotaMemory:
                  value: "3000Mi"
                - defaultIamRole:
                  value: "gsa-production@{{ .Values.userSettings.googleProjectId }}.iam.gserviceaccount.com"

5. Update helm dependencies

.. code-block:: bash

   helm dep update


6. Install Flyte

.. code-block:: bash

   helm install -n flyte -f values-gcp.yaml --create-namespace flyte flyteorg/flyte-core


7. Verify that all pods have come up correctly

.. code-block:: bash

   kubectl get pods -n flyte


8. Get the ingress IP to update the zone and fetch name server records for DNS

.. code-block:: bash

  kubectl get ingress -n flyte

Uninstalling Flyte
------------------

.. code-block:: bash

   helm uninstall -n flyte flyte

Upgrading Flyte
---------------

.. code-block:: bash

  helm upgrade -n flyte -f values-gcp.yaml --create-namespace flyte flyteorg/flyte-core

Connecting to Flyte
===================

Flyte can be accessed using the UI console or your terminal.

* First, find the Flyte endpoint created by the GKE ingress controller.

.. code-block:: bash

   $ kubectl -n flyte get ingress

Sample O/P

.. code-block:: bash

   NAME         CLASS    HOSTS              ADDRESS     PORTS   AGE
   flyte        <none>   <FLYTE-ENDPOINT>   34.136.165.92   80, 443   18m
   flyte-grpc   <none>   <FLYTE-ENDPOINT>   34.136.165.92   80, 443   18m


* Connecting to flytectl CLI

Add :<FLYTE-ENDPOINT>  to ~/.flyte/config.yaml eg ;

.. code-block:: yaml

    admin:
     # For GRPC endpoints you might want to use dns:///flyte.myexample.com
     endpoint: dns:///<FLYTE-ENDPOINT>
     insecure: false
    logger:
     show-source: true
     level: 0
    storage:
      type: stow
      stow:
        kind: google
        config:
          json: ""
          project_id: myproject # GCP Project ID
          scopes: https://www.googleapis.com/auth/devstorage.read_write
      container: mybucket # GCS Bucket Flyte is configured to use

Accessing Flyte Console (web UI)
================================

* Use the https://<FLYTE-ENDPOINT>/console to get access to flyteconsole UI
* Ignore the certificate error if using a self-signed cert


Running Workflows
=================

* Docker file changes

Make sure the Dockerfile contains gcloud-sdk installation steps which are needed by Flyte to upload the results.

.. code-block:: bash

   # Install gcloud for GCP
   RUN apt-get install curl --assume-yes

   RUN curl -sSL https://sdk.cloud.google.com | bash
   ENV PATH $PATH:/root/google-cloud-sdk/bin


* Serializing workflows

For running flytecookbook examples on GCP, make sure you have the right registry during serialization.
The following example shows if you are using GCP container registry and US-central zone with project name flyte-gcp and repo name flyterep:

.. code-block:: bash

   REGISTRY=us-central1-docker.pkg.dev/flyte-gcp/flyterepo make serialize

* Uploading the image to registry

The following example shows uploading cookbook core examples to GCP container registry. This step must be performed before performing registration of the workflows in Flyte:

.. code-block:: bash

   docker push us-central1-docker.pkg.dev/flyte-gcp/flyterepo/flytecookbook:core-2bd81805629e41faeaa25039a6e6abe847446356

* Registering workflows

Register workflows by pointing to the output folder for the serialization and providing the version to use for the workflow through flytectl:

.. code-block:: bash

   flytectl register file  /Users/<user-name>/flytesnacks/cookbook/core/_pb_output/*   -d development  -p flytesnacks --version v1

* Generating exec spec file for workflow

The following example generates the exec spec file for the latest version of core.flyte_basics.lp.go_greet workflow part of flytecookbook examples:

.. code-block:: bash

   flytectl  get launchplan -p flytesnacks -d development core.flyte_basics.lp.go_greet --latest --execFile lp.yaml

* Modifying exec spec files of the workflow for inputs

Modify the exec spec file lp.yaml and modify the inputs for the workflow:

.. code-block:: yaml

   iamRoleARN: ""
   inputs:
       am: true
       day_of_week: "Sunday"
       number: 5
   kubeServiceAcct: ""
   targetDomain: ""
   targetProject: ""
   version: v1
   workflow: core.flyte_basics.lp.go_greet

* Creating execution using the exec spec file

.. code-block:: bash

   flytectl create execution -p flytesnacks -d development --execFile lp.yaml

Sample O/P


.. code-block:: bash

   execution identifier project:"flytesnacks" domain:"development" name:"f12c787de18304f4cbe7"

* Getting the execution details

.. code-block:: bash

    flytectl get executions  -p flytesnacks -d development f12c787de18304f4cbe7



Troubleshooting
===============

* If any pod is not coming up, then describe the pod and check which container or init-containers had an error.

.. code-block:: bash

   kubectl describe pod/<pod-instance> -n flyte

Then check the logs for the container which failed.
E.g.: to check for <init-container> init container type this:

.. code-block:: bash

   kubectl logs -f <pod-instance> <init-container> -n flyte


* Increasing log level for flytectl

Change your logger config to this:

.. code-block:: yaml

  logger:
  show-source: true
  level: 6

* If you have a new ingress IP for your Flyte deployment, you would need to flush DNS cache using `this <https://developers.google.com/speed/public-dns/cache>`__
* If you need to get access logs for your buckets then follow `this <https://cloud.google.com/storage/docs/access-logs>`__ GCP guide
* If you get the following error:

.. code-block::

   ERROR: Policy modification failed. For a binding with condition, run "gcloud alpha iam policies lint-condition" to identify issues in condition.
   ERROR: (gcloud.iam.service-accounts.add-iam-policy-binding) INVALID_ARGUMENT: Identity Pool does not exist

this means that you haven't enabled workload identity on the cluster. Use the following `docs <https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity>`__.
