.. _deployment-deployment-sandbox:

#########################
Flyte Sandbox Deployment
#########################

*****************
Sandboxed Flyte
*****************

What is a sandbox deployment?
=============================

In addition to a K8s cluster, Flyte requires some external cloud-provided resources such as a database and durable blob store. A sandboxed deployment of Flyte is merely the concept of bundling portable versions of these requirements alongside Flyte to simplify setup. For the blob store requirements, we use Minio, which offers an S3 compatible interface, and for Postgres, we use the stock Postgres Docker image and Helm chart.

.. warning::
    The sandbox deployment is not suitable for production environments. For instructions on how to create a production-ready Flyte deployment, checkout the :ref:`deployment` guide.

*******************************************
Flyte Sandbox as a Single Docker Container
*******************************************
Flyte provides one such sandboxed deployment in the form of a self-contained Docker image. This is mini-replica of an entire Flyte deployment, without the scalability and with minimal extensions. The idea for this environment originated from the desire of the core team to make it extremely simple for users of Flyte to try out the platform and get a feel for the user experience, without having to understand Kubernetes or dabble with configuration etc. The Flyte single container sandbox is also used by the team to run continuous integration tests and used by the `flytesnacks - UserGuide playground environment`. The sandbox can be run in any environment that supports containers.

Requirements
============

kubectl
-------
Ensure ``kubectl`` is installed. Follow `kubectl installation docs <https://kubernetes.io/docs/tasks/tools/install-kubectl/>`__. On Mac::

    brew install kubectl

Docker
------
Ensure that Docker is installed. While Flyte can run any OCI compatible task image, using the default Kubernetes container runtime (cri-o), we on the team typically use Docker. Also the ``flytectl demo`` command does rely on Docker APIs, but as this demo environment is just one self-contained image, you can also run the image directly using another run time.

Within the single container environment, a mini Kubernetes cluster is installed using the excellent `k3s <https://k3s.io/>`__ platform. K3s uses an in-container Docker daemon (run using `docker-in-docker configuration <https://www.docker.com/blog/docker-can-now-run-within-docker/>`__) to orchestrate user containers.  

flytectl
--------
``flytectl`` is a control plane CLI for Flyte. Find installation instructions for it on https://github.com/flyteorg/flytectl.

Running
=======
To launch, call ::

    flytectl demo start

This comes with a Docker registry on ``localhost:30000`` so you can build images outside of the docker-in-docker container by tagging your containers with ``localhost:30000/imgname:tag`` and pushing. The local Postgres installation is also available on port ``30001`` for users who wish to dig deeper into the storage layer.

**************************
Flyte Sandbox on the Cloud
**************************
Sometimes it's also helpful to be able to install a sandboxed environment on a cloud provider. That is, you have access to an EKS or GKE cluster, but provisioning a separate database or blob storage bucket is harder because of a lack of infrastructure support. Instructions for how to do this will be forthcoming.
