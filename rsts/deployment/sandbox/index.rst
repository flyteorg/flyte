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

 helm install -n flyte flyte-deps flyteorg/flyte-deps --create-namespace --set webhook.enabled=false,minio.service.type=LoadBalancer,contour.enabled=true,contour.envoy.service.type=LoadBalancer

.. note::
	If you are trying it locally with Kind/k3d then you need to port-forward the api port(9000), It is required for [fast registration](https://docs.flyte.org/projects/cookbook/en/latest/auto/deployment/deploying_workflows.html#register-your-workflows-and-tasks). And you can customize your [minio chart](https://github.com/bitnami/charts/tree/master/bitnami/minio) as per your need

	.. code-block::

	kubectl port-forward svc/minio 30084:9000 -n flyte

3. Install Flyte Helm chart

.. code-block::

 helm install flyte flyteorg/flyte-core -n flyte -f https://raw.githubusercontent.com/flyteorg/flyte/master/charts/flyte-core/values-sandbox.yaml


4. You can now port-forward (or if you have load-balancer enabled then get an LB) to connect to remote FlyteConsole, as follows ::

    kubectl port-forward --address 0.0.0.0 svc/flyte-deps-contour-envoy 30081:80 -n flyte

5. Open the console http://localhost:30081/console.

6. In order to interact with your Flyte instance using ``flytectl``, initialise your configuration to point to this host ::

    flytectl config init --host='localhost:30081' --insecure

7. Open the minio console http://localhost:30088. Your minio username is `minio` and password is `miniostorage`.

8. You can port-forward to connect postgres using ::

    kubectl port-forward --address 0.0.0.0 svc/flyte-deps-kubernetes-dashboard 5432:5432 -n flyte

9. Now use these credentials for postgres

   .. code-block::

      dbname: flyteadmin
      host: 127.0.0.1
      port: 5432
      username: postgres
