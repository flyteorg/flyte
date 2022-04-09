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

2. Install Minio & Postgres deployment in kubernetes cluster

.. code-block::

 kubectl create ns flyte
 kubectl apply -f https://raw.githubusercontent.com/flyteorg/flyte/master/deployment/sandbox/flyte_helm_generated.yaml -n flyte

.. note::
	If you are trying it locally with Kind/k3d then you need to port-forward the api port(9000), It is required for [fast registration](https://docs.flyte.org/projects/cookbook/en/latest/auto/deployment/deploying_workflows.html#register-your-workflows-and-tasks). And you can customize your [minio chart](https://github.com/bitnami/charts/tree/master/bitnami/minio) as per your need

	.. code-block::

	kubectl port-forward svc/minio 30084:9000 -n flyte

4. Install Flyte Helm chart

.. code-block::

 helm install flyte flyteorg/flyte-core -n default -f https://raw.githubusercontent.com/flyteorg/flyte/sandbox-deployment/charts/flyte-core/values-sandbox.yaml

5. Create a values-override.yaml file

.. code-block::

   # -- Domain configuration for Flyte project. This enables the specified number of domains across all projects in Flyte.
  replicaCount: 1
  contour:
	# -- Default resources requests and limits for Contour
	resources:
	  # -- Requests are the minimum set of resources needed for this pod
	  requests:
		cpu: 10m
		memory: 50Mi
	  # -- Limits are the maximum set of resources needed for this pod
	  limits:
		cpu: 100m
		memory: 100Mi
  envoy:
	service:
	  type: NodePort
	  ports:
		http: 80
	  nodePorts:
		http: 30081
	# -- Default resources requests and limits for Envoy
	resources:
	  # -- Requests are the minimum set of resources needed for this pod
	  requests:
		cpu: 10m
		memory: 50Mi
	  # -- Limits are the maximum set of resources needed for this pod
	  limits:
		cpu: 100m
		memory: 100Mi

6. Install contour helm chart for ingress setup
.. code-block::

	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm install contour bitnami/contour -n flyte -f values-override.yaml

6. Install Kubernetes Dashboard for logs for flyte cluster

.. code-block::

	helm repo add dashboard https://kubernetes.github.io/dashboard/
	helm install kubernetes-dashboard dashboard/kubernetes-dashboard -n flyte \
	--set service.type=LoadBalancer,rbac.clusterReadOnlyRole=true,service.externalPort=30082,protocolHttp=true,extraArgs={enable-skip-login,enable-insecure-login,disable-settings-authorizer}

.. note::
	If you are trying it locally with Kind/k3d then you need to port-forward the api port(30082), Flyte will automatically provide you Kubernetes Dashboard URL

	.. code-block::

	kubectl port-forward svc/kubernetes-dashboard 30082:30082 -n flyte
