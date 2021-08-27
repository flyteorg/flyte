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
* Access to [GCE console](https://console.cloud.google.com/)
* A domain name for the Flyte installation like flyte.example.org that allows you to set a DNS A record.

Before you begin, please ensure that you have the following tools installed.

* ``gcloud``
* Helm
* ``kubectl``

Create GCE Project
==================
.. code-block::

  gcloud create project <my-project>

Of course you can also use an existing project if your account has appropriate permissions to create the required resources.

Set project <my-project> as default in gcloud:

.. code-block::

  gcloud config set project <my-project>

We assume that for the <my-project> has been set as default for all gcloud commands below.

Permissions
===========

Configure `workload identity <https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity>`__ for flyte namespace service accounts.

.. code-block:: bash

  gcloud iam service-accounts create flyte

  gcloud iam service-accounts add-iam-policy-binding \
    --role roles/iam.workloadIdentityUser \
    --member "serviceAccount:<my-project>.svc.id.goog[flyte/flyteadmin]" \
    flyte@<my-project>.iam.gserviceaccount.com

  gcloud iam service-accounts add-iam-policy-binding \
    --role roles/iam.workloadIdentityUser \
    --member "serviceAccount:<my-project>.svc.id.goog[flyte/datacatalog]" \
    flyte@<my-project>.iam.gserviceaccount.com

  gcloud iam service-accounts add-iam-policy-binding \
    --role roles/iam.workloadIdentityUser \
    --member "serviceAccount:<my-project>.svc.id.goog[flyte/flytepropeller]" \
    flyte@<my-project>.iam.gserviceaccount.com

Create GKE Cluster
==================
.. code-block::

  gcloud container clusters create <my-flyte-cluster> \
    --region us-west1 \
    --num-nodes 1

Create Cloud SQL Database
=========================
Next create a relational `Cloud SQL for PostgreSQL https://cloud.google.com/sql/docs/postgres/introduction` database. This database will be used by both the primary control plane service (Flyte Admin) and the Flyte memoization service (Data Catalog).

.. code-block::
  gcloud sql instances create <myinstance> \
    --database-version=POSTGRES_13 \
    --cpu=1 \
    --memory=3840MB \
    --region=us-
    
TODO get DB password

SSL Certificate
===============
In order to use SSL (which we need to use gRPC clients), we next need to create an SSL certificate. We'll use [Google-managed SSL certificates](
https://cloud.google.com/kubernetes-engine/docs/how-to/managed-certs)

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

Ingress
=======

Create a static IP address.

.. code-block:: bash

  gcloud compute addresses create flyte-example --global
  gcloud compute addresses describe flyte-example --global

TODO

Create GCS Bucket
=================

.. code-block:: bash
  gsutil mb -b on -l us-west1 gs://my-flyte-bucket/

TODO bucket permissions

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
     insecureSkipVerify: true # only required if using a self-signed cert. Caution: not to be used in production
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

