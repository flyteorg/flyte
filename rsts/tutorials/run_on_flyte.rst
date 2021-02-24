.. _tutorials-getting-started-run-on-flyte:

######################################
Run Your Workflow on a Flyte Backend
######################################

************************
Installing Flyte Locally
************************

This guide will walk you through a quick installation of Flyte on your laptop and then how to register and execute your
workflows against this deployment. (The tabs below have an option to install Flyte on a cloud provider as well)

.. rubric:: Estimated time to complete: 1 minutes.

Prerequisites
=============

Make Sure you have followed :ref:`flyte-tutorials-firstrun` and have a running docker container and can access FlyteConsole on http://localhost:30081/console.

.. tip:: To check if your flyte-sandbox container is running you can run ``docker ps`` and it should show image ``ghcr.io/flyteorg/flyte-sandbox `` running

.. tip:: If you prefer using k3d, Minikube, docker for mac, or a hosted Kubernetes cluster like AWS-EKS, GCP-GKE, Azure Kubernetes refer to :ref:`howto-sandbox`. It is recommended that you use a simple Docker based approach when you are first getting started with Flyte.

.. _tutorials-run-flyte-laptop:

****************************
Running your Flyte Workflows
****************************

Registration
============
Registration is the process of shipping your code to Flyte backend. This creates an immutable, versioned record of your code with FlyteAdmin service.

Register your workflows
-----------------------

From within root directory of ``flyteexamples`` you created :ref:`previously <tutorials-getting-started-first-example>`
feel free to make any changes and then register ::

  FLYTE_AWS_ENDPOINT=http://localhost:30084/ FLYTE_AWS_ACCESS_KEY_ID=minio \
  FLYTE_AWS_SECRET_ACCESS_KEY=miniostorage make fast_register


.. tip:: Flyte sandbox uses minio as a substitue for S3/GCS etc, in the first command we port-forwarded it to 30084. If you use s3/gcs or a different port-forward you can drop or change ``FLYTE_AWS_ENDPOINT`` accordingly.

.. rubric:: Boom! It's that simple.

Run your workflows
------------------

Visit the page housing workflows registered for your project (method if you used k3d):
`http://localhost:30081/console/projects/flyteexamples/workflows <http://localhost:30081/console/projects/flyteexamples/workflows>`__
else if you used docker-desktop or something else, then copy paste this URL into the browser and fill in the ``<host:port>``::

    http://<host:port>/console/projects/flyteexamples/workflows


Select your workflow, click the bright purple "Launch Workflow" button in the upper right, update the "name" input
argument as you please, proceed to launch and you'll have triggered an execution!

.. note::

    After registration Flyte Workflows exist in the FlyteAdmin service and can be triggered using the
      - console
      - Command line
      - directly invoking the REST API
      - on a schedule

    More on this later

Optionally you can create a new project
----------------------------------------
Refer to :ref:`howto_new_project`.