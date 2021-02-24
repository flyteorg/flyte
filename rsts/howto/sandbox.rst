.. _howto_sandbox:

################################
How do I try out/install Flyte?
################################


**********************
What is Flyte Sandbox?
**********************
Flyte can be run using a Kubernetes cluster only. This installs all the dependencies as kubernetes deployments. We call this a Sandbox deployment. Flyte sandbox can be deployed by simply applying a kubernetes YAML.

.. note::

    #. A Sandbox deployment takes over the entire cluster
    #. It needs special cluster roles that will need access to create namespaces, pods etc
    #. The sandbox deployment is not suitable for production environments. For an in-depth overview of how to productionize your flyte deployment, checkout the :ref:`howto_productionize`.


*********************************************************
Deploy Flyte Sandbox environment locally - on your laptop
*********************************************************

Flyte Sandbox deployment can be run locally on your laptop. This will allow only some of the functionality, but is great to get started with.
Refer to :ref:`tutorials-getting-started-run-on-flyte`.


.. _howto-sandbox-dedicated-k8s-cluster:

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

We'll proceed like with :ref:`locally hosted flyte <tutorials-getting-started-run-on-flyte>` with deploying the sandbox
Flyte configuration on your remote cluster.

.. warning::
    The sandbox deployment is not suitable for production environments. For an in-depth overview of how to productionize your flyte deployment, checkout the :ref:`howto_productionize`.

#. The Flyte sandbox can be deployed with a single command ::

    kubectl create -f https://raw.githubusercontent.com/flyteorg/flyte/master/deployment/sandbox/flyte_generated.yaml


#. You can now port-forward (or if you have load-balancer enabled then get an LB) to connect to remote FlyteConsole, as follows::

    kubectl port-forward svc/envoy 30081:80


#. Open console http://localhost:30081/console.

***************************************************************
Deploy Flyte Sandbox environment to a shared kubernetes cluster
***************************************************************

The goal here is to deploy to an existing Kubernetes cluster - within one namespace only. This would allow multiple Flyte clusters to run within one K8s cluster.

.. caution:: coming soon!
