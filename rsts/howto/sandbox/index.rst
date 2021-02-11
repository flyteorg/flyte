.. _howto_sandbox:

########################
How do I try out Flyte?
########################

**********************
What is Flyte Sandbox?
**********************

*********************************************************
Deploy Flyte Sandbox environment locally - on your laptop
*********************************************************


******************************************************************
Deploy Flyte Sandbox environment to a dedicated kubernetes cluster
******************************************************************

Cluster Requirements
====================

Ensure you have kubernetes up and running on your choice of cloud provider:

- `AWS EKS <https://aws.amazon.com/eks/>`_ (Amazon)
- `GCP GKE <https://cloud.google.com/kubernetes-engine/>`_ (Google)
- `Azure AKS <https://azure.microsoft.com/en-us/services/kubernetes-service/>`_ (Microsoft)

If you can access your cluster with ``kubectl cluster-info``, you're ready to deploy Flyte.


Deployment
==========

We'll proceed like with :ref:`locally hosted flyte <tutorials-getting-started-flyte-laptop>` with deploying the sandbox
Flyte configuration on your remote cluster.

.. warning::
    The sandbox deployment is not suitable for production environments.

For an in-depth overview of how to productionize your flyte deployment, checkout the :ref:`how to <howto_productionize>`

The Flyte sandbox can be deployed with a single command ::

  kubectl create -f https://raw.githubusercontent.com/lyft/flyte/master/deployment/sandbox/flyte_generated.yaml


***************************************************************
Deploy Flyte Sandbox environment to a shared kubernetes cluster
***************************************************************
