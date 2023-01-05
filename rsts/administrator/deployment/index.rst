.. _administrator-deployment:

###################
Deployment Guide
###################
The articles in this section will guide a new Flyte administrator through deploying Flyte. Experience tells us that the most complicated parts of a Flyte deployment are authentication, and ingress, DNS, and SSL support. For this reason, we recommend first deploying Flyte without these, and relying on K8s port forwarding to test. After the base deployment is tested, these additional features can be turned on more seamlessly.

********************************
Components of a Flyte Deployment
********************************
We recommend working with your infrastructure team to set up the cloud service requirements below.

Relational Database
===================
The ``FlyteAdmin`` and ``DataCatalog`` components rely on PostgreSQL to store persistent records. In the sandbox deployment, a containerized version of Postgres is included but for a proper Flyte installation, we recommend one of the cloud provided databases.  For AWS, we recommend their `RDS <https://aws.amazon.com/rds/postgresql/>`__ service, for GCP, `Cloud SQL <https://cloud.google.com/sql/docs/postgres/>`__, and Azure, `PostgreSQL <https://azure.microsoft.com/en-us/services/postgresql/>`__.

Production Grade Object Store
=============================

Core Flyte components such as Admin, Propeller, and DataCatalog, as well as user runtime containers rely on an Object Store to hold files. The sandbox deployment comes with a containerized Minio, which offers AWS S3 compatibility. We recommend swapping this out for `AWS S3 <https://aws.amazon.com/s3/>`__ or `GCP GCS <https://cloud.google.com/storage/>`__.

Helm
====
Flyte uses Helm as the K8s release packaging solution for now, though you may still see some old ``kustomize`` artifacts in the repo. There are Helm charts that correspond with the deployment options below.

Cluster Configuration
=====================
Flyte has the ability to configure K8s clusters to work with it. For example, as your Flyte user-base evolves, adding new projects is as simple as registering them through the command line ::

   $ flytectl create project --id myflyteproject --name "My Flyte Project" --description "My very first project onboarding onto Flyte"

A process runs at a configurable cadence that ensures that all Kubernetes resources necessary for the new project are created and new workflows can successfully
be registered and executed within it. See :std:ref:`flytectl <flytectl:flytectl_create_project>` for more information. This project should immediately show up in the Flyte console after refreshing.

************************
Flyte Deployment Options
************************
There are broadly three different styles of deploying a Flyte backend, with the middle option below being what we recommend for a capable though not massively scalable cluster. If all your ML compute can `fit on one EKS cluster <https://docs.aws.amazon.com/eks/latest/userguide/service-quotas.html>`__ (which as of this writing is north of 13000 nodes), this is likely the option for you.

* Sandboxed
  This uses portable replacements for the blob store and database mentioned above and is not suitable for production use. This is a fantastic experimentation and testing environment for Flyte though and is what is used for our ``flytectl demo`` environment. Please see :ref:`administrator-deployment-sandbox`.
* Unified Flyte on one cluster
  We recommend this for most Flyte deployments because it's the simplest. Flyte is bundled as one executable, requiring only one Deployment for itself. It runs on a single K8s cluster and but still supports the rich ecosystem of extensions and plugins Flyte offers. Refer to the :ref:`administrator-deployment-cloud-simple` page to get started.
* Multi-cluster
  For the largest of deployments, it may be worthwhile or necessary to have multiple K8s clusters. Flyte's control plane (Admin, Console, Data Catalog) is separated from Flyte's execution engine (Propeller), which runs typically once per compute cluster. See :ref:`administrator-deployment-multicluster` for the details.
  
Whatever the style, note that Propeller itself can be sharded as well, though typically that's not required.

.. toctree::
    :maxdepth: 1
    :name: deployment options toc
    :hidden:

    sandbox
    cloud_simple
    cloud_production
    multicluster

