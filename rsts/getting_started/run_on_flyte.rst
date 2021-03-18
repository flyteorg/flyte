.. _getting-started-run-on-flyte:

#####################################
Run Your Workflow on a Flyte Backend
#####################################

************************
Installing Flyte Locally
************************

This guide will walk you through:
* A quick installation of Flyte on your laptop 
* How to register and execute your workflows against this deployment.
(The tabs below have an option to install Flyte on a cloud provider as well)

.. rubric:: Estimated time to complete: 1 minute

Prerequisites
=============

1. Follow :ref:`getting-started-firstrun` 
2. Prepare a running docker container 
3. Access FlyteConsole on http://localhost:30081/console.

.. tip:: To check if your flyte-sandbox container is running, run ``docker ps`` and check for image ``ghcr.io/flyteorg/flyte-sandbox`` running

.. tip:: If you prefer using k3d, Minikube, docker for mac, or a hosted Kubernetes cluster like AWS-EKS, GCP-GKE, or Azure Kubernetes, refer to :ref:`howto-sandbox`. It is recommended that you use a simple Docker based approach when you first get started with Flyte.

.. _getting-started-run-flyte-laptop:

****************************
Running Your Flyte Workflows
****************************

Register Your Workflows
=======================
Registration is the process of shipping your code to the Flyte backend. This creates an immutable, versioned record of your code with the FlyteAdmin service.

From within the root directory of ``flyteexamples`` that you previously created in the :ref:`<getting-started-first-example>`
feel free to make any changes and then register ::

  FLYTE_AWS_ENDPOINT=http://localhost:30084/ FLYTE_AWS_ACCESS_KEY_ID=minio \
  FLYTE_AWS_SECRET_ACCESS_KEY=miniostorage make fast_register


.. tip:: Flyte sandbox uses minio as a substitue for S3/GCS, etc. It is port-forwarded in the first command to 30084. If you use S3/GCS or a different port-forward you can drop or change the ``FLYTE_AWS_ENDPOINT`` accordingly.

.. rubric:: Boom! It's that simple.

Run Your Workflows
==================

K3d
---

Visit the page housing workflows registered to your project at:
`http://localhost:30081/console/projects/flyteexamples/workflows <http://localhost:30081/console/projects/flyteexamples/workflows>`__

Docker-desktop or other
-----------------------

Copy and paste this URL into the browser and fill in the ``<host:port>``::

    http://<host:port>/console/projects/flyteexamples/workflows


Once you have accessed your workflows:
1. Select your workflow
2. Click the bright purple "Launch Workflow" button in the upper right
3. Update the "name" input argument
4. Proceed to launch to trigger an execution

.. note::

    After registration, Flyte Workflows exist in the FlyteAdmin service and can be triggered using:
      - Console
      - Command line
      - Directly invoking the REST API
      - On a schedule

    

Create a new project
--------------------
Visit :ref:`howto_new_project`.