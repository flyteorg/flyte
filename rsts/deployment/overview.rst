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

Flyte uses Helm to manage its deployment releases onto a K8s cluster. The chart and templates are located under ``helm``. There is a base ``values.yaml`` file but there are several files that fine tune those settings.

* ``values-eks.yaml`` should be additionally applied for AWS EKS deployments.
* ``values-gcp.yaml`` should be additionally applied for GCP GKE deployments.
* ``values-sandbox.yaml`` should be additionally applied for our sandbox install. See the :ref:`deployment-sandbox` page for more information.

Specific instructions for Helm are covered in both the :ref:`Opta <deployment-aws-opta>` and :ref:`manual <deployment-aws-manual>` AWS setup guides.

.. TODO::
   Additional instructions on the Helm after cleaning it up, how it's laid out, easy guide to know which values to
   change for each component - plugins, sns/sqs, auth, etc.

*********************
Relational Database
*********************

The ``FlyteAdmin`` and ``DataCatalog`` components rely on PostgreSQL to store persistent records. In the sandbox deployment, a containerized version of Postgres is included but for a proper Flyte installation, we recommend one of the cloud provided databases.  For AWS, we recommend their `RDS <https://aws.amazon.com/rds/postgresql/>`__ service, for GCP, `Cloud SQL <https://cloud.google.com/sql/docs/postgres/>`__, and Azure, `PostgreSQL <https://azure.microsoft.com/en-us/services/postgresql/>`__.

*****************************
Production Grade Object Store
*****************************

``FlyteAdmin``, ``FlytePropeller``, and ``DataCatalog`` rely on an Object Store to hold files. The sandbox deployment comes with a containerized Minio, which offers AWS S3 compatibility. We recommend swapping this out for `AWS S3 <https://aws.amazon.com/s3/>`__ or `GCP GCS <https://cloud.google.com/storage/>`__.

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
Flyte relies on cloud-provided pub/sub and schedulers to provide automated periodic execution of your launch plans. In AWS,

* `CloudWatch Events <https://docs.aws.amazon.com/cloudwatch/index.html>`_ are used for the triggering mechanism to accurately serve periodic executions.
* `SNS <https://aws.amazon.com/sns>`_ and `SQS <https://aws.amazon.com/sqs/>`_ are used for handling notifications.


**************
Authentication
**************
Flyte ships with its own authorization server, as well as the ability to use an external authorization server if your external IDP supports it.  See the :ref:`authorization <deployment-cluster-config-auth-setup>` page for detailed configuration.
