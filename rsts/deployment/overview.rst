.. _deployment-overview:

###################
Deployment Overview
###################

Up until now the Flyte backend you've been working with likely has been accessible only on ``localhost`` and likely ran
entirely in one Docker container.  In order to handle production load, and make use of all the additional features
Flyte offers, we'll need to replace, add, and configure some components. This page describes at a high-level what a
production ready deployment might look like.

*******************
Usage of Helm
*******************


*************************
Relational Database
*************************

The ``FlyteAdmin`` and ``DataCatalog`` components rely on PostgreSQL to store persistent records.

In this section, we'll modify the Flyte deploy to use a remote PostgreSQL database instead.

First, you'll need to set up a reliable PostgreSQL database. The easiest way achieve this is to use a cloud provider like AWS `RDS <https://aws.amazon.com/rds/postgresql/>`__, GCP `Cloud SQL <https://cloud.google.com/sql/docs/postgres/>`__, or Azure `PostgreSQL <https://azure.microsoft.com/en-us/services/postgresql/>`__ to manage the PostgreSQL database for you. Create one and make note of the username, password, endpoint, and port.

Next, remove old sandbox database by opening up the ``flyte/kustomization.yaml`` file and deleting database component. ::

  - github.com/flyteorg/flyte/kustomize/dependencies/database

With this line removed, you can re-run ``kustomize build flyte > flyte_generated.yaml`` and see that the the postgres deployment has been removed from the ``flyte_generated.yaml`` file.


*****************************
Production Grade Object Store
*****************************

``FlyteAdmin``, ``FlytePropeller``, and ``DataCatalog`` components rely on an Object Store to hold files.

In this section, we'll modify the Flyte deploy to use `AWS S3 <https://aws.amazon.com/s3/>`__ for object storage.
The process for other cloud providers like `GCP GCS <https://cloud.google.com/storage/>`__ should be similar.

To start, `create an s3 bucket <https://docs.aws.amazon.com/AmazonS3/latest/gsg/CreatingABucket.html>`__.

Next, remove the old sandbox object store by opening up the ``flyte/kustomization.yaml`` file and deleting the storage line. ::

  - github.com/flyteorg/flyte/kustomize/dependencies/storage

With this line gone, you can re-run ``kustomize build flyte > flyte_generated.yaml`` and see that the sandbox object store has been removed from the ``flyte_generated.yaml`` file.

Next, open the configs ``admindeployment/flyteadmin_config.yaml``, ``propeller/config.yaml``, ``datacatalog/datacatalog_config.yaml`` and look for the ``storage`` configuration.

*******************************
Dynamically Configured Projects
*******************************

As your Flyte user-base evolves, adding new projects is as simple as registering them through the cli ::

  flyte-cli register-project -h {{ your-flyte-admin-host.com }} -p myflyteproject --name "My Flyte Project" \
    --description "My very first project onboarding onto Flyte"

A cron which runs at the cadence specified in flyteadmin config will ensure that all the kubernetes resources necessary for the new project are created and new workflows can successfully
be registered and executed under the new project.

This project should immediately show up in the Flyte console after refreshing.

*******************************
Cloud Based Pub/Sub Integration
*******************************


**************
Authentication
**************




