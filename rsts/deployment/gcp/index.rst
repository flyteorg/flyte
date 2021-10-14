.. _deployment-gcp:

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

Initialize Gcloud
===================
Authorize Gcloud sdk to access GCP using your credentials, setup the config for the existing project,
and optionally set the default compute zone. `Init <https://cloud.google.com/sdk/gcloud/reference/init>`__

.. code-block:: bash

   gcloud init


Create Organization
===================
Use the following docs to understand the organization creation process in Google cloud
`Organization Management <https://cloud.google.com/resource-manager/docs/creating-managing-organization>`__

Get the organization id to be used for creating the project. The billing should be linked with the organization so
that all projects under the org use the same billing account.

.. code-block:: bash

   gcloud organizations list

Sample output
.. code-block::

   DISPLAY_NAME                 ID    DIRECTORY_CUSTOMER_ID
   example-org        123456789999                C02ewszsz

<ORG-ID> = ID column value for the organization

.. code-block:: bash

   export ORG_ID=<id>

Create GCE Project
==================
.. code-block:: bash

  export PROJECT-ID=<my-project>
  gcloud projects create $PROJECT-ID --organization $ORG_ID

Of course you can also use an existing project if your account has appropriate permissions to create the required resources.

Set project <my-project> as the default in gcloud or use gcloud init to set this default:

.. code-block:: bash

  gcloud config set project ${PROJECT-ID}

We assume that <my-project> has been set as default for all gcloud commands below.

Permissions
===========

Configure `workload identity <https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity>`__ for Flyte namespace service accounts.
This creates the GSA's which would be mapped to the KSA(kubernetes service account) through annotations and used for authorizing the pods access to the Google cloud services.

* Create GSA for flyteadmin

.. code-block:: bash

  gcloud iam service-accounts create gsa-flyteadmin

* Create GSA for datacatalog

.. code-block:: bash

  gcloud iam service-accounts create gsa-datacatalog

* Create GSA for flytepropeller

.. code-block:: bash

  gcloud iam service-accounts create gsa-flytepropeller


* Create GSA for cluster resource manager

.. code-block:: bash

  gcloud iam service-accounts create flyte-clusterresources


* Create a new customized role enveloping `storage.buckets.get` with name StorageBucketGet which will be used while adding roles.

* Add IAM policy binding for flyteadmin GSA. It requires permissions to get the bucket, create/delete/update objects in the bucket, connecting to cloud sql and also workload identity user role.

.. code-block:: bash

  gcloud iam service-accounts add-iam-policy-binding \
    --role "roles/iam.workloadIdentityUser" \
    --role "projects/${PROJECT-ID}/roles/StorageBucketGet" \
    --role "roles/cloudsql.client" \
    --role "roles/storage.objectAdmin" \
    --member "serviceAccount:${PROJECT-ID}.svc.id.goog[flyte/flyteadmin]" \
    gsa-flyteadmin@${PROJECT-ID}.iam.gserviceaccount.com

* Add IAM policy binding for datacatalog GSA. It requires permissions to get the bucket, create/delete/update objects in the bucket, connecting to cloud sql and also workload identity user role.

.. code-block:: bash

  gcloud iam service-accounts add-iam-policy-binding \
    --role "roles/iam.workloadIdentityUser" \
    --role "projects/${PROJECT-ID}/roles/StorageBucketGet" \
    --role "roles/cloudsql.client" \
    --role "roles/storage.objectAdmin" \
    --member "serviceAccount:${PROJECT-ID}.svc.id.goog[flyte/datacatalog]" \
    gsa-datacatalog@${PROJECT-ID}.iam.gserviceaccount.com

* Add IAM policy binding for flytepropeller GSA. It requires permissions to get the bucket, create/delete/update objects in the bucket, create/update/delete kubernetes objects in the cluster and also workload identity user role.

.. code-block:: bash

  gcloud iam service-accounts add-iam-policy-binding \
    --role roles/iam.workloadIdentityUser \
    --role "projects/${PROJECT-ID}/roles/StorageBucketGet" \
    --role "roles/container.developer" \
    --role "roles/storage.objectAdmin" \
    --member "serviceAccount:${PROJECT-ID}.svc.id.goog[flyte/flytepropeller]" \
    gsa-flytepropeller@${PROJECT-ID}.iam.gserviceaccount.com

* Add IAM policy binding for cluster resource manager GSA. It requires permissions to get the bucket, create/delete/update objects in the bucket and also workload identity user role.

.. code-block:: bash

  gcloud iam service-accounts add-iam-policy-binding \
    --role roles/iam.workloadIdentityUser \
    --role "projects/${PROJECT-ID}/roles/StorageBucketGet" \
    --role "roles/storage.objectAdmin" \
    --member "serviceAccount:${PROJECT-ID}.svc.id.goog[flyte/flytepropeller]" \
    flyte-clusterresources@${PROJECT-ID}.iam.gserviceaccount.com


* Allow the Kubernetes service account to impersonate the Google service account by creating an IAM policy binding between the two. This binding allows the Kubernetes Service account to act as the Google service account

   * flyteadmin

   .. code-block:: bash

    gcloud iam service-accounts add-iam-policy-binding \
    --role roles/iam.workloadIdentityUser \
    --member "serviceAccount:flyte-gcp.svc.id.goog[flyte/flyteadmin]” \
    gsa-flyteadmin@flyte-gcp.iam.gserviceaccount.com

   * flytepropeller

   .. code-block:: bash

    gcloud iam service-accounts add-iam-policy-binding \
    --role roles/iam.workloadIdentityUser \
    --member "serviceAccount:flyte-gcp.svc.id.goog[flyte/flytepropeller]” \
    gsa-flytepropeller@flyte-gcp.iam.gserviceaccount.com

   * datacatalog

   .. code-block:: bash

    gcloud iam service-accounts add-iam-policy-binding \
    --role roles/iam.workloadIdentityUser \
    --member "serviceAccount:flyte-gcp.svc.id.goog[flyte/datacatalog]” \
    gsa-datacatalog@flyte-gcp.iam.gserviceaccount.com

   * flyte-clusterresources
   Do the following for all project-domain combination of namespaces. Following example shows for flytesnacks-development namespace

   .. code-block:: bash

    gcloud iam service-accounts add-iam-policy-binding \
    --role roles/iam.workloadIdentityUser \
    --member "serviceAccount:flyte-gcp.svc.id.goog[flytesnacks-development/default]” \
    flyte-clusterresources@flyte-gcp.iam.gserviceaccount.com


Create GKE Cluster
==================
Create GKE cluster with VPC-native networking and workload identity enabled.
Browse to the gcloud console and Kubernetes Engine tab to start creating the k8s cluster.

Ensure that VPC native traffic routing is enabled under Security enable Workload identity and use project default pool
which would be `${PROJECT-ID}.svc.id.goog`

The recommended way is to create it from the console. This is to make sure the options of VPC-native networking and Workload identity are enabled correctly.
The command can be updated once right options are available to do the above steps.

.. code-block:: bash

  gcloud container clusters create <my-flyte-cluster> \
    --workload-pool=${PROJECT-ID}.svc.id.goog
    --region us-west1 \
    --num-nodes 6

Create GKE context
==================
Initialize your kubecontext to point to GKE cluster using the following command:

.. code-block:: bash

  gcloud container clusters get-credentials <my-flyte-cluster>

Verify by creating a test namespace

.. code-block:: bash

   kubectl create ns test

Create Cloud SQL Database
=========================
Next, create a relational `Cloud SQL for PostgreSQL <https://cloud.google.com/sql/docs/postgres/introduction>`__ database. This database will be used by both the primary control plane service (Flyte Admin) and the Flyte memoization service (Data Catalog).
Follow this `link <https://console.cloud.google.com/sql/choose-instance-engine>`__ to create the cloud sql instance.

* Select PostgreSQL
* Provide an Instance ID
* Provide password for the instance <DB_INSTANCE_PASSWD>
* Use PostgresSQL13 or higher
* Select the Zone based on your availability requirements.
* Select customize your instance and enable Private IP in Connections tab. This is required for the private communication between the GKE apps and cloud SQL instance. Follow the steps to create the private connection (default).
* Create the SQL instance
* After creation of the instance get the private IP of the database <CLOUD-SQL-IP>
* Create flyteadmin database and flyteadmin user account on that instance with <DBPASSWORD>
* Verify the connectivity to the DB from GKE cluster
   * Create a testdb namespace

   .. code-block:: bash

      kubectl create ns test

   * Verify the connectivity using a postgres client

   .. code-block:: bash

      kubectl run pgsql-postgresql-client --rm --tty -i --restart='Never' --namespace testdb --image docker.io/bitnami/postgresql:11.7.0-debian-10-r9 --env="PGPASSWORD=<DBPASSWORD>" --command -- psql testdb --host <CLOUD-SQL-IP> -U flyteadmin -d flyteadmin -p 5432

The recommended way is to create it from the console.. This is to make sure the private IP connectivity works correctly to cloud sql instance.
The command can be updated once right options are available to do the above steps

.. code-block:: bash

  gcloud sql instances create <my-flyte-db> \
    --database-version=POSTGRES_13 \
    --cpu=1 \
    --memory=3840MB \
    --region=us-west1


SSL Certificate
===============
In order to use SSL (which we need to use gRPC clients), we next need to create an SSL certificate. We'll use `Google-managed SSL certificates <https://cloud.google.com/kubernetes-engine/docs/how-to/managed-certs>`__

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

An alternative is to use the certificate manager:

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


Create GCS Bucket
=================
Create <BUCKETNAME> with uniform access

.. code-block:: bash

  gsutil mb -b on -l us-west1 gs://<BUCKETNAME>/

Add access permission for the following principals
* gsa-flytepropeller@${PROJECT-ID}.iam.gserviceaccount.com
* gsa-datacatalog@${PROJECT-ID}.iam.gserviceaccount.com
* gsa-flyteadmin@f${PROJECT-ID}.iam.gserviceaccount.com
* gsa-flyte-clusterresources@${PROJECT-ID}.iam.gserviceaccount.com

Time for Helm
=============

Installing Flyte
-----------------
#. Clone the Flyte repo

.. code-block:: bash

   git clone https://github.com/flyteorg/flyte

#. Update values
   <RELEASE-NAME> to be used as prefix for ssl certificate secretName
   <PROJECT-ID> of your GCP project
   <CLOUD-SQL-IP> private IP of cloud sql instance
   <DBPASSWORD> of the flyteadmin user created for the cloud sql instance
   <BUCKETNAME> of the GCS bucket created
   <HOSTNAME> DNS name of the Flyte deployment

#. Update helm dependencies

.. code-block:: bash

   helm dep update


#. Install Flyte

.. code-block:: bash

   cd helm
   helm install -n flyte -f values-gcp.yaml --create-namespace flyte .


#. Verify all the pods have come up correctly

.. code-block:: bash

   kubectl get pods -n flyte


# Get the ingress IP to be used for updating the zone and getting the name server records for DNS

.. code-block:: bash

  kubectl get ingress -n flyte

Uninstalling Flyte
------------------

.. code-block:: bash

   helm uninstall -n flyte flyte

Upgrading Flyte
---------------

.. code-block:: bash

  helm upgrade -n flyte -f values-gcp.yaml --create-namespace flyte .

Connecting to Flyte
===================

Flyte can be accessed using the UI console or your terminal.

* First, find the Flyte endpoint created by the GKE ingress controller.

.. code-block:: bash

   $ kubectl -n flyte get ingress

Sample O/P

.. code-block:: bash

   NAME         CLASS    HOSTS              ADDRESS     PORTS   AGE
   flyte        <none>   <HOSTNAME>   34.136.165.92   80, 443   18m
   flyte-grpc   <none>   <HOSTNAME>   34.136.165.92   80, 443   18m


* Connecting to flytectl CLI

Add :<FLYTE-ENDPOINT>  to ~/.flyte/config.yaml eg ;

.. code-block:: yaml

    admin:
     # For GRPC endpoints you might want to use dns:///flyte.myexample.com
     endpoint: dns:///<FLYTE-ENDPOINT>
     insecure: true
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


Running workflows
=================

* Docker file changes
Make sure the Dockerfile contains gcloud-sdk installation steps which is needed by flyte to upload the results

.. code-block:: bash

   # Install gcloud for GCP
   RUN apt-get install curl --assume-yes

   RUN curl -sSL https://sdk.cloud.google.com | bash
   ENV PATH $PATH:/root/google-cloud-sdk/bin


* Serializing workflows
For running the flytecookbook examples on GCP make sure you have right registry during serialization.
Following example shows if you are using GCP container registry and us-central zone with project name flyte-gcp and repo name flyterep

.. code-block:: bash

   REGISTRY=us-central1-docker.pkg.dev/flyte-gcp/flyterepo make serialize

* Uploading the image to registry
Following example shows uploading cookbook core examples to gcp container registry. This step must be performed before performing registration of the workflows in flyte

.. code-block:: bash
   docker push us-central1-docker.pkg.dev/flyte-gcp/flyterepo/flytecookbook:core-2bd81805629e41faeaa25039a6e6abe847446356

* Registering workflows
Register workflows by pointing to the output folder for the serialization and providing version to use for the workflow through flytectl

.. code-block:: bash

   flytectl register file  /Users/<user-name>/flytesnacks/cookbook/core/_pb_output/*   -d development  -p flytesnacks --version v1

* Generating exec spec file for workflow
Following example generates exec spec file for the latest version of core.flyte_basics.lp.go_greet workflow part of flytecookbook examples

.. code-block:: bash

   flytectl  get launchplan -p flytesnacks -d development core.flyte_basics.lp.go_greet --latest --execFile lp.yaml

* Modify  exec spec file of the  workflow for inputs
Modify the exec spec file lp.yaml and modify the inputs for the workflow

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

* Create execution using the exec spec file

.. code-block:: bash

   flytectl create execution -p flytesnacks -d development --execFile lp.yaml

Sample O/P


.. code-block:: bash

   execution identifier project:"flytesnacks" domain:"development" name:"f12c787de18304f4cbe7"

* Get the execution details

.. code-block:: bash

    flytectl get executions  -p flytesnacks -d development f12c787de18304f4cbe7



Troubleshooting
===============

* If any pod is not coming up, then describe the pod and check which container or init-containers had an error.

.. code-block:: bash

   kubectl describe pod/<pod-instance> -n flyte

Then check the logs for the container which failed.
eg: to check for <init-container> init container do this.

.. code-block:: bash

   kubectl logs -f <pod-instance> <init-container> -n flyte


* Increasing log level for flytectl
===================================

  Change your logger config to this:

  .. code-block:: yaml

     logger:
     show-source: true
     level: 6

* In case you have a new ingress IP for your Flyte deployment, you would need to flush DNS cache using `this <https://developers.google.com/speed/public-dns/cache>`__
