.. _getting_started:

Getting Started
---------------

Prerequisites
*************

Kubernetes and its ``kubectl`` client are the only strict prerequisites to installing Flyte.

Kubernetes can be installed on your local machine to run Flyte locally, or in the cloud for a scalable multi-user setup. Some installation options are listed below.

Local:

- `Minikube <https://kubernetes.io/docs/tasks/tools/install-minikube/>`_
- `Docker for Mac <https://blog.docker.com/2018/01/docker-mac-kubernetes/>`_

Cloud Providers:

- `AWS EKS <https://aws.amazon.com/eks/>`_ (Amazon)
- `GCP GKE <https://cloud.google.com/kubernetes-engine/>`_ (Google)
- `Azure AKS <https://azure.microsoft.com/en-us/services/kubernetes-service/>`_ (Microsoft)

Once you have kubernetes set up and can access it with ``kubectl cluster-info``, you're ready to deploy flyte.

Flyte has a few different deployment configurations. We'll start with the easiest, and expand on it to increase scale and reliability.


Sandbox Deployment
******************

The simplest Flyte deployment is the "sandbox" deployment, which includes everything you need in order to use Flyte. The Flyte sandbox can be deployed with a single command ::

  kubectl create -f https://raw.githubusercontent.com/lyft/flyte/master/deployment/sandbox/flyte_generated.yaml

This deployment uses a kubernetes `NodePort <https://kubernetes.io/docs/concepts/services-networking/service/#nodeport>`_ for Flyte ingress.
Once deployed, you can access the Flyte console on any kubernetes node at ``http://{{ any kubernetes node }}:30081/console`` (note that it will take a moment to deploy).

For local deployments, this endpoint is typically http://localhost:30081/console.

(for Minikube deployment, you need to run ``minikube tunnel`` and use the ip that Minikube tunnel outputs)

WARNING:
  The sandbox deployment is not well suited for production use.
  Most importantly, Flyte needs access to an object store, and a PostgreSQL database.
  In the sandbox deployment, the object store and PostgreSQL database are each installed as a single kubernetes pod.
  These pods are sufficient for testing and playground purposes, but they not designed to handle production load.
  Read on to learn how to configure Flyte for production.
