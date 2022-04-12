.. _deployment-sandbox:

##################
Sandbox Deployment
##################

**************************
Flyte Deployment - Sandbox
**************************
This guide helps you set up Flyte from scratch, on any kubernetes cluster, It will helpful in trying out flyte. It details step-by-step instructions to go from a bare Kubernetes cluster to a fully functioning Flyte deployment that members of your company can use.

.. note::

  Please make sure you have a working kubernetes cluster with write permission


1. Add Helm repo for Flyte,Postgres & Minio

.. code-block::

 helm repo add flyteorg https://helm.flyte.org

2. Install Flyte dependency helm chart

.. code-block::

 helm install -n flyte flyteorg/flyte-deps flyte-deps --create-namespace

.. note::
	If you are trying it locally with Kind/k3d then you need to port-forward the api port(9000), It is required for [fast registration](https://docs.flyte.org/projects/cookbook/en/latest/auto/deployment/deploying_workflows.html#register-your-workflows-and-tasks). And you can customize your [minio chart](https://github.com/bitnami/charts/tree/master/bitnami/minio) as per your need

	.. code-block::

	kubectl port-forward svc/minio 30084:9000 -n flyte

3. Install Flyte Helm chart

.. code-block::

 helm install flyte flyteorg/flyte-core -n default -f https://raw.githubusercontent.com/flyteorg/flyte/sandbox-deployment/charts/flyte-core/values-sandbox.yaml

