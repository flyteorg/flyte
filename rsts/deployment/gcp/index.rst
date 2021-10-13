.. _deployment-gcp:

###############
GCP (GKE) Setup
###############

************************************************
Flyte Deployment - Manual GCE/GKE Deployment
************************************************
This guide helps you set up Flyte from scratch, on GCE, without using an automated approach. It details step-by-step how to go from a bare GCE account, to a fully functioning Flyte deployment that members of your company can use.

Prerequisites
=============
* Access to `GCE console <https://console.cloud.google.com/>`__
* A domain name for the Flyte installation like flyte.example.org that allows you to set a DNS A record.

Before you begin, please ensure that you have the following tools installed.

* ``gcloud``
* Helm
* ``kubectl``

Initialize Gcloud
===================
Authorize Gcloud sdk to access GCP using your credentials and also additionally setup the config for existing project
and optionally set the default compute zone. `Init <https://cloud.google.com/sdk/gcloud/reference/init>`__

.. code-block::

   gcloud init


Create Organization
===================
Use the following docs to understand the organization creation process in google cloud
`Organization Management <https://cloud.google.com/resource-manager/docs/creating-managing-organization>`__

Get the organization id to be used for creating the project. The billing should be linked with the organization so
that all projects under the org use the same billing account.

.. code-block::

   gcloud organizations list

Sample output
.. code-block::

   DISPLAY_NAME                 ID    DIRECTORY_CUSTOMER_ID
   example-org        123456789999                C02ewszsz

<ORG-ID> = ID column value for the organization

.. code-block::

   export ORG_ID=<id>

Create GCE Project
==================
.. code-block::

  export GCP_PROJECT=<my-project>
  gcloud projects create $GCP_PROJECT --organization $ORG_ID

Of course you can also use an existing project if your account has appropriate permissions to create the required resources.

Set project <my-project> as default in gcloud or use gcloud init to set this default:

.. code-block::

  gcloud config set project ${GCP_PROJECT}

We assume that for the <my-project> has been set as default for all gcloud commands below.

Permissions
===========

Configure `workload identity <https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity>`__ for flyte namespace service accounts.
This creates the GSA's which would be used as mapped to the KSA(kubernetes service account) through annotations and would be used for authorizing the pods access to the google cloud services.

* Create GSA for flyteadmin

.. code-block::

  gcloud iam service-accounts create gsa-flyteadmin

* Create GSA for datacatalog

.. code-block::

  gcloud iam service-accounts create gsa-datacatalog

* Create GSA for flytepropeller

.. code-block::

  gcloud iam service-accounts create gsa-flytepropeller


* Create GSA for cluster resource manager

.. code-block::

  gcloud iam service-accounts create flyte-clusterresources


* Create a new customized role enveloping `storage.buckets.get` with name StorageBucketGet which will be used while adding roles

* Add IAM policy binding for flyteadmin GSA. It requires permissions to get bucket, create/delete/update objects in the bucket, connecting to cloud sql and also workload identity user role.

.. code-block::

  gcloud iam service-accounts add-iam-policy-binding \
    --role "roles/iam.workloadIdentityUser" \
    --role "projects/${GCP_PROJECT}/roles/StorageBucketGet" \
    --role "roles/cloudsql.client" \
    --role "roles/storage.objectAdmin" \
    --member "serviceAccount:${GCP_PROJECT}.svc.id.goog[flyte/flyteadmin]" \
    gsa-flyteadmin@${GCP_PROJECT}.iam.gserviceaccount.com

* Add IAM policy binding for datacatalog GSA.It requires permissions to get bucket, create/delete/update objects in the bucket, connecting to cloud sql and also workload identity user role.

.. code-block::

  gcloud iam service-accounts add-iam-policy-binding \
    --role "roles/iam.workloadIdentityUser" \
    --role "projects/${GCP_PROJECT}/roles/StorageBucketGet" \
    --role "roles/cloudsql.client" \
    --role "roles/storage.objectAdmin" \
    --member "serviceAccount:${GCP_PROJECT}.svc.id.goog[flyte/datacatalog]" \
    gsa-datacatalog@${GCP_PROJECT}.iam.gserviceaccount.com

* Add IAM policy binding for flytepropeller GSA.It requires permissions to get bucket, create/delete/update objects in the bucket, create/update/delete kubernetes objects in the cluster and also workload identity user role.

.. code-block::

  gcloud iam service-accounts add-iam-policy-binding \
    --role roles/iam.workloadIdentityUser \
    --role "projects/${GCP_PROJECT}/roles/StorageBucketGet" \
    --role "roles/container.developer" \
    --role "roles/storage.objectAdmin" \
    --member "serviceAccount:${GCP_PROJECT}.svc.id.goog[flyte/flytepropeller]" \
    gsa-flytepropeller@${GCP_PROJECT}.iam.gserviceaccount.com

* Add IAM policy binding for cluster resource manager GSA.It requires permissions to get bucket, create/delete/update objects in the bucket and also workload identity user role.

.. code-block::

  gcloud iam service-accounts add-iam-policy-binding \
    --role roles/iam.workloadIdentityUser \
    --role "projects/${GCP_PROJECT}/roles/StorageBucketGet" \
    --role "roles/storage.objectAdmin" \
    --member "serviceAccount:${GCP_PROJECT}.svc.id.goog[flyte/flytepropeller]" \
    flyte-clusterresources@${GCP_PROJECT}.iam.gserviceaccount.com


Create GKE Cluster
==================
Creates GKE cluster with VPC-native networking and workload identity enabled.
Browse to the gcloud console and Kubernetes Engine tab to start creating the k8s cluster.

Ensure that VPC native traffic routing is enabled and under Security enable Workload identity and use project default pool
which would be `${GCP_PROJECT}.svc.id.goog`

Recommended way is to create it from the console.

.. code-block::

  gcloud container clusters create <my-flyte-cluster> \
    --workload-pool=${GCP_PROJECT}.svc.id.goog
    --region us-west1 \
    --num-nodes 6

Create GKE context
==================
Initialize your kubecontext to point to GKE cluster using the following command.

.. code-block::

  gcloud container clusters get-credentials <my-flyte-cluster>

Verify by creating a test namespace

.. code-block::

   kubectl create ns test

Create Cloud SQL Database
=========================
Next create a relational `Cloud SQL for PostgreSQL <https://cloud.google.com/sql/docs/postgres/introduction>`__ database. This database will be used by both the primary control plane service (Flyte Admin) and the Flyte memoization service (Data Catalog).
Follow this `link <https://console.cloud.google.com/sql/choose-instance-engine>`__ to create the cloud sql instance

* Select PostgreSQL
* Provide an Instance ID
* Provide password for the instance <DB_INSTANCE_PASSWD>
* Use PostgresSQL13 or higher
* Select the Zone based on your availability requirements.
* Select customize your instance and enable Private IP in Connections tab. This is required for the private communication between the GKE apps and cloud SQL instance. Follow the steps to create the private connection (default)
* Create the SQL instance
* After creation of the instance get the private IP of the database <PRIVATE_IP>
* Create flyteadmin database and flyteadmin user account on that instance with <DB_PASSWORD>
* Verify the connectivity to the DB from GKE cluster
   * Create a testdb namespace

   .. code-block::

      kubectl create ns test
   * Verify the connectivity using a postgres client

   .. code-block::

      kubectl run pgsql-postgresql-client --rm --tty -i --restart='Never' --namespace testdb --image docker.io/bitnami/postgresql:11.7.0-debian-10-r9 --env="PGPASSWORD=<DB_INSTANCE_PASSWD>" --command -- psql testdb --host <PRIVATE_IP> -U postgres -d flyteadmin -p 5432

Recommended way is to create it from the console.

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

Alternative is to use certificate manager

* Install the cert manager

.. code-block::

  helm install cert-manager --namespace flyte --version v0.12.0 jetstack/cert-manager

* Create cert issuer

.. code-block:: 

   apiVersion: cert-manager.io/v1alpha2
   kind: Issuer
   metadata:
     name: letsencrypt-production
   spec:
     acme:
       server: https://acme-v02.api.letsencrypt.org/directory
       email: pmahindrakar@union.ai
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


* Get the ingress IP to be used for updating the zone and getting the name server records for DNS

.. code-block:: bash

  kubectl get ingress -n flyte

Create GCS Bucket
=================
Create <BUCKET-NAME> with uniform access

.. code-block:: bash

  gsutil mb -b on -l us-west1 gs://my-flyte-bucket/

Add access permission for the following principals
* gsa-flytepropeller@${GCP_PROJECT}.iam.gserviceaccount.com
* gsa-datacatalog@${GCP_PROJECT}.iam.gserviceaccount.com
* gsa-flyteadmin@f${GCP_PROJECT}.iam.gserviceaccount.com
* gsa-flyte-clusterresources@${GCP_PROJECT}.iam.gserviceaccount.com

Time for Helm
=============

Installing Flyte
-----------------
#. Clone the Flyte repo

.. code-block:: bash

   git clone https://github.com/flyteorg/flyte

#. Update values

TODO

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

Flyte can be accessed using the UI console or your terminal

* First, find the Flyte endpoint created by the GKE ingress controller.

.. code-block:: bash

   $ kubectl -n flyte get ingress

   NAME         CLASS    HOSTS   ADDRESS                                                       PORTS   AGE
   flyte        <none>   *       k8s-flyte-8699360f2e-1590325550.us-east-2.elb.amazonaws.com   80      3m50s
   flyte-grpc   <none>   *       k8s-flyte-8699360f2e-1590325550.us-east-2.elb.amazonaws.com   80      3m49s

<FLYTE-ENDPOINT> = Value in ADDRESS column and both will be the same as the same port is used for both GRPC and HTTP.


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
* Ignore the certificate error if using a self signed cert

Troubleshooting
===============


* If flyteadmin pod is not coming up, then describe the pod and check which of the container or init-containers had an error.

.. code-block:: bash

   kubectl describe pod/<flyteadmin-pod-instance> -n flyte

Then check the logs for the container which failed.
eg: to check for run-migrations init container do this.

.. code-block:: bash

   kubectl logs -f <flyteadmin-pod-instance> run-migrations -n flyte


* Increasing log level for flytectl
  Change your logger config to this
  .. code-block::

     logger:
     show-source: true
     level: 6

