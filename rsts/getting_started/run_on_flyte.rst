.. _getting-started-run-on-flyte:

#####################################
Run Your Workflow on a Flyte Backend
#####################################

************************
Installing Flyte Locally
************************

This guide will walk you through:

* A quick installation of Flyte on your device

* How to register and execute your workflows against this deployment. 

(The tips below have an option to install Flyte on a cloud provider as well)

.. rubric:: Estimated time: 1 minute

Prerequisites
=============

1. Follow :ref:`getting-started-firstrun` 

2. Prepare a running docker container

3. Access FlyteConsole on http://localhost:30081/console

.. tip:: To check if your flyte-sandbox container is running you can run ``docker ps`` and it should show image ``ghcr.io/flyteorg/flyte-sandbox`` running

.. tip:: If you prefer using k3d, Minikube, docker for mac, or a hosted Kubernetes cluster like AWS-EKS, GCP-GKE, Azure Kubernetes refer to :ref:`howto-sandbox`. It is recommended that you use a simple Docker based approach when you are first getting started with Flyte.

.. _getting-started-run-flyte-laptop:

****************************
Running Your Flyte Workflows
****************************

Register Your Workflows
=======================
Registration is the process of shipping your code to the Flyte backend. This creates an immutable, versioned record of your code with the FlyteAdmin service.

From within root directory of ``flyteexamples`` you previously created with the :ref:`Write Your First Workflow tutorial <getting-started-first-example>`,
feel free to make any changes and then register: ::

  FLYTE_AWS_ENDPOINT=http://localhost:30084/ FLYTE_AWS_ACCESS_KEY_ID=minio \
  FLYTE_AWS_SECRET_ACCESS_KEY=miniostorage make fast_register


.. tip:: Flyte sandbox uses minio as a substitue for S3/GCS etc. It is port-forwarded in the first command to 30084. If you use S3/GCS or a different port-forward you can drop or change the ``FLYTE_AWS_ENDPOINT`` accordingly.

.. rubric:: It's that simple!

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


Once you have accessed your workflows, 

* Select your workflow
* Click the bright purple "Launch Workflow" button in the upper right
* Update the "name" input argument
* Proceed to launch to trigger an execution

.. note::

    After registration, Flyte Workflows exist in the FlyteAdmin service and can be triggered using:
      - Console
      - Command line
      - Directly invoking the REST API
      - On a schedule


Create a New Project
--------------------
Visit :ref:`howto_new_project`.
