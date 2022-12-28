.. _administrator-deployment-sandbox:

#########################
Flyte Sandbox Deployment
#########################

*******************************
What is a sandbox deployment?
*******************************

The Flyte demo deployment is a fully standalone, sandboxed environment for running Flyte. Flyte requires a 

provides a simplified way of running ``flyte-sandbox`` as a single Docker container running locally.

The follow section explains how you can use each of these modes and provides more information. We **recommend** running the sandbox using flytectl locally on your workstation.
Flyte Sandbox is not a complete representation of Flyte, many features are intentionally removed from this environment to ensure that the startup times and runtime footprints are low.

.. warning::
    The sandbox deployment is not suitable for production environments. For an in-depth overview of how to productionize your Flyte deployment, checkout the :ref:`administrator` guide.

*******************************************
Flyte Sandbox as a Single Docker Container
*******************************************

:std:ref:`flytectl_sandbox` starts a local sandbox environment for Flyte. This is mini-replica of an entire Flyte deployment, without the scalability and with minimal extensions. The idea for this environment originated from the desire of the core team to make it extremely simple for users of Flyte to
try out the platform and get a feel for the user experience, without having to understand Kubernetes or dabble with configuration etc. The Flyte single container sandbox is also used by the team to run continuous integration tests and used by the `flytesnacks - UserGuide playground environment`. The sandbox can be run
in most any environment that supports Docker containers and an Ubuntu docker base image.

Architecture and Reasons Why We Built It
========================================
Within the single container environment, a mini Kubernetes cluster is installed using the excellent `k3s <https://k3s.io/>`__ platform. K3s uses an in-container Docker daemon (run using `docker-in-docker configuration <https://www.docker.com/blog/docker-can-now-run-within-docker/>`__) to orchestrate user containers.

In a typical Flyte installation, one needs to build Docker containers for tasks and push them to a repository from which K8s can pull.

Users are free to use an external registry of course, as long as the inner k3s cluster has permissions to pull from it.

The illustration below shows the architecture of flyte-sandbox in a single container. It is identical to a Flyte sandbox cluster, except that we have built one docker container, with Kubernetes and Flyte already installed.

.. image:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/deployment/sandbox/flyte_sandbox_single_container.png
   :alt: Architecture of single container Flyte Sandbox


Use the Flyte Sandbox to:
=========================
* Try out Flyte locally using a single Docker command or using ``flytectl sandbox``
* Run regular integration tests for Flyte
* Provide snapshot environments for various Flyte versions, to identify regressions

********************
Prerequisites
********************

Ensure ``kubectl`` is installed. Follow `kubectl installation docs <https://kubernetes.io/docs/tasks/tools/install-kubectl/>`__. On Mac::

    brew install kubectl



.. image:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/deployment/sandbox/flyte_sandbox_single_k8s_cluster.png
   :alt: Architecture of Sandbox deployment of Flyte. Single K8s cluster


.. _administrator-deployment-sandbox-local:

Deploy Flyte Sandbox on Your Local Machine
==========================================

You can deploy the sandbox environment on your laptop workstation to run locally.


Recommended using ``flytectl sandbox start`` as described in :ref:`getting-started`

.. prompt:: bash $

        docker run --rm --privileged -p 30081:30081 -p 30084:30084 -p 30088:30088 cr.flyte.org/flyteorg/flyte-sandbox
