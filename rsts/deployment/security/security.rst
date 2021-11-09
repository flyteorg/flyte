.. _security-overview:

###################
Security Overview
###################

Here we cover security aspects of running your flyte deployments. In the current state we will be covering the user
which is used for running the flyte services and we will go through why we do this and not run them as root user

#################
Why non-root user
#################
Up until now all the containers that are used for bringing up the flyte deployment used the default root user to bring
up the services. We use docker for container packaging and by default its containers run as root. This gives full
permissions on the system but may not be suitable for production deployments where a security breach could comprise your
application deployments.
Its considered to be a best practise for security because running in constrained permission environment will prevent any
malicious code from utilizing the full permissions of the host.`Ref <https://kubernetes.io/blog/2018/07/18/11-ways-not-to-get-hacked/#8-run-containers-as-a-non-root-user>`__
Also in certain container platforms like `OpenShift <https://engineering.bitnami.com/articles/running-non-root-containers-on-openshift.html>`__ running non-root containers is mandatory.


*******
Changes
*******
New user group and user have been added to the Docker files for all the flyte components
`Flyteadmin <https://github.com/flyteorg/flyteadmin/blob/master/Dockerfile>`__
`Flytepropeller <https://github.com/flyteorg/flytepropeller/blob/master/Dockerfile>`__
`Datacatalog <https://github.com/flyteorg/datacatalog/blob/master/Dockerfile>`__
`Flyteconsole <https://github.com/flyteorg/flyteconsole/blob/master/Dockerfile>`__

And Dockerfile uses `USER command <https://docs.docker.com/engine/reference/builder/#user>`__ which helps in setting user
and group which will be used for running the container.

Additionally the orchestration yaml files for the flyte components define the overriden security context with the created
user and group to run them. Following shows the overriden security context added for flyteadmin
`Flyteadmin <https://github.com/flyteorg/flyte/blob/master/charts/flyte/templates/admin/deployment.yaml>`__


************
Why override
************
There are certain init-containers that still require root permissions and hence we required to override the security
context for these.
For eg : in case of `Flyteadmin <https://github.com/flyteorg/flyte/blob/master/charts/flyte/templates/admin/deployment.yaml>`__
the init container of check-db-ready is not able to resolve the host for the checks and fails. This is mostly due to no read
permissions on etc/hosts file. Only the check-db-ready container is run using the root user and which we will plan to fix aswell.


Flyte uses Helm to manage its deployment releases onto a K8s cluster. The chart and templates are located under the `helm folder <https://github.com/flyteorg/flyte/tree/master/charts>`__. There is a base ``values.yaml`` file but there are several files that fine tune those settings.

* ``values-eks.yaml`` should be additionally applied for AWS EKS deployments.
* ``values-gcp.yaml`` should be additionally applied for GCP GKE deployments.
* ``values-sandbox.yaml`` should be additionally applied for our sandbox install. See the :ref:`deployment-sandbox` page for more information.

Specific instructions for Helm on AWS are covered in both the :ref:`automated Opta <deployment-aws-opta>` and :ref:`manual <deployment-aws-manual>` AWS setup guides.

*********************
Relational Database
*********************

The ``FlyteAdmin`` and ``DataCatalog`` components rely on PostgreSQL to store persistent records. In the sandbox deployment, a containerized version of Postgres is included but for a proper Flyte installation, we recommend one of the cloud provided databases.  For AWS, we recommend their `RDS <https://aws.amazon.com/rds/postgresql/>`__ service, for GCP, `Cloud SQL <https://cloud.google.com/sql/docs/postgres/>`__, and Azure, `PostgreSQL <https://azure.microsoft.com/en-us/services/postgresql/>`__.

*****************************
Production Grade Object Store
*****************************

Core Flyte components such as Admin, Propeller, and DataCatalog, as well as user runtime containers rely on an Object Store to hold files. The sandbox deployment comes with a containerized Minio, which offers AWS S3 compatibility. We recommend swapping this out for `AWS S3 <https://aws.amazon.com/s3/>`__ or `GCP GCS <https://cloud.google.com/storage/>`__.

*********************
Project Configuration
*********************
As your Flyte user-base evolves, adding new projects is as simple as registering them through the command line ::

   $ flytectl create project --id myflyteproject --name "My Flyte Project" --description "My very first project onboarding onto Flyte"

A cron which runs at the cadence specified in Flyte Admin configuration ensures that all Kubernetes resources necessary for the new project are created and new workflows can successfully
be registered and executed within it. See :std:ref:`flytectl <flytectl:flytectl_create_project>` for more information.

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

