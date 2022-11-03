.. _administrator-deployment:

###################
Deployment Options
###################

*****************************
Layout of this guide
*****************************
Discussion of the various components you'll use.

********************************
Components of a Flyte Deployment
********************************

This section mostly copied from the existing docs.

Helm
====

Relational Database
===================

The ``FlyteAdmin`` and ``DataCatalog`` components rely on PostgreSQL to store persistent records. In the sandbox deployment, a containerized version of Postgres is included but for a proper Flyte installation, we recommend one of the cloud provided databases.  For AWS, we recommend their `RDS <https://aws.amazon.com/rds/postgresql/>`__ service, for GCP, `Cloud SQL <https://cloud.google.com/sql/docs/postgres/>`__, and Azure, `PostgreSQL <https://azure.microsoft.com/en-us/services/postgresql/>`__.

Production Grade Object Store
=============================

Core Flyte components such as Admin, Propeller, and DataCatalog, as well as user runtime containers rely on an Object Store to hold files. The sandbox deployment comes with a containerized Minio, which offers AWS S3 compatibility. We recommend swapping this out for `AWS S3 <https://aws.amazon.com/s3/>`__ or `GCP GCS <https://cloud.google.com/storage/>`__.

Project Configuration
=============================

As your Flyte user-base evolves, adding new projects is as simple as registering them through the command line ::

   $ flytectl create project --id myflyteproject --name "My Flyte Project" --description "My very first project onboarding onto Flyte"

A cron which runs at the cadence specified in FlyteAdmin configuration ensures that all Kubernetes resources necessary for the new project are created and new workflows can successfully
be registered and executed within it. See :std:ref:`flytectl <flytectl:flytectl_create_project>` for more information.

This project should immediately show up in the Flyte console after refreshing.


.. toctree::
    :maxdepth: 1
    :name: deployment options toc
    :hidden:

    sandbox
    cloud_simple
    cloud_production
    multicluster

