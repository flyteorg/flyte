.. _deployment-sandbox:

###################
Sandbox Overview
###################

.. warning::
    The sandbox deployment is not suitable for production environments. For an in-depth overview of how to productionize your Flyte deployment, checkout the :ref:`deployment` guide.


**********************
What is Flyte Sandbox?
**********************
The Flyte Sandbox is a fully standalone minimal environment for running Flyte. Basically, :std:ref:`flytectl_sandbox` provides a simplified way of running ``flyte-sandbox`` as a single Docker container running locally.

The follow section explains how you can use each of these modes and provides more information. We **recommend** running the sandbox using flytectl locally on your workstation or on a single cloud instance to try out Flyte or for testing new ideas. Flyte Sandbox is not a complete representation of Flyte,
many features are intentionally removed from this environment to ensure that the startup times and runtime footprints are low.

*******************************************
Flyte Sandbox as a single Docker container
*******************************************

:std:ref:`flytectl_sandbox` starts a local sandbox environment for Flyte. This is mini-replica of an entire Flyte deployment, without the scalability and with minimal extensions. The idea for this environment originated from the desire of the core team to make it extremely simple for users of Flyte to
try out the platform and get a feel for the user experience, without having to understand Kubernetes or dabble with configuration etc. The Flyte single container sandbox is also used by the team to run continuous integration tests and used by the `flytesnacks - UserGuide playground environment`. The sandbox can be run
in most any environment that supports Docker containers and an Ubuntu docker base image.

Architecture and reasons why we built it
========================================
Within the single container environment, a mini Kubernetes cluster is installed using the excellent `k3s <https://k3s.io/>`__ platform. K3s uses an in-container Docker daemon (run using `docker-in-docker configuration <https://www.docker.com/blog/docker-can-now-run-within-docker/>`__) to orchestrate user containers.

When users call ``flytectl sandbox start --source <dir>``, the source ``<dir>`` is mounted within the sandbox container and hence it is possible to build images for that source code, using the inner Docker daemon. In a typical Flyte installation, one needs to build Docker containers for tasks and push them to a repository from which K8s can pull.

This is not possible with the sandbox's Docker environment however, because it does not ship with a Docker registry. Users are free to use an external registry of course, as long as the inner k3s cluster has permissions to pull from it. To reduce the friction of procuring such a registry, configuring permissions to access it, and having to explicitly push to it,
we recommend using the ``flytectl sandbox exec -- ...`` mode to trigger a Docker build for your code (which is mounted in the sandbox environment) using the docker-in-docker daemon. Since K3s uses the same Docker daemon, it is possible to re-use the images built internally. This greatly simplifies the user's interaction and ability to try out Flyte, at the cost of hiding one part of the real-world iteration cycle, calling ``docker push`` on your task images.

The illustration below shows the architecture of flyte-sandbox in a single container. It is identical to a Flyte sandbox cluster, except that we have built one docker container, with Kubernetes and Flyte already installed.

.. image:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/deployment/sandbox/flyte_sandbox_single_container.png
   :alt: Architecture of single container Flyte Sandbox


Use the Flyte Sandbox to:
=========================
* Try out Flyte locally using a single Docker command or using ``flytectl sandbox``
* Run regular integration tests for Flyte
* Provide snapshot environments for various Flyte versions, to identify regressions

***************************************************************
Deploying your own Flyte Sandbox environment to a K8s cluster
***************************************************************

This installs all the dependencies as Kubernetes deployments. We call this a Sandbox deployment. Flyte sandbox deployment can be deployed using the default Helm chart.

.. note::

    #. A Sandbox deployment takes over the entire cluster
    #. It needs special cluster roles that will need access to create namespaces, pods, etc.
    #. The sandbox deployment is not suitable for production environments. For an in-depth overview of how to productionize your Flyte deployment, checkout the rest of the :ref:`deployment` guides.


.. image:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/deployment/sandbox/flyte_sandbox_single_k8s_cluster.png
   :alt: Architecture of Sandbox deployment of Flyte. Single K8s cluster


.. _deploy-sandbox-local:

Deploy Flyte Sandbox environment on laptop/workstation/single machine
=======================================================================


Ensure ``kubectl`` is installed. Follow `kubectl installation docs <https://kubernetes.io/docs/tasks/tools/install-kubectl/>`__. On Mac::

    brew install kubectl

Recommended using ``flytectl sandbox start`` as described in :ref:`getting-started`

.. prompt:: bash $

        docker run --rm --privileged -p 30081:30081 -p 30084:30084 -p 30088:30088 cr.flyte.org/flyteorg/flyte-sandbox

.. _deployment-sandbox-dedicated-k8s-cluster:

Deploy a Flyte Sandbox environment to a Cloud Kubernetes cluster
==================================================================

Cluster Requirements
---------------------

Ensure you have kubernetes up and running on your choice of cloud provider:

- `AWS EKS <https://aws.amazon.com/eks/>`_ (Amazon)
- `GCP GKE <https://cloud.google.com/kubernetes-engine/>`_ (Google)
- `Azure AKS <https://azure.microsoft.com/en-us/services/kubernetes-service/>`_ (Microsoft)

If you can access your cluster with ``kubectl cluster-info``, you're ready to deploy Flyte.


Deployment
-----------

We'll proceed like with :ref:`locally hosted flyte <deploy-sandbox-local>` with deploying the sandbox
Flyte configuration on your remote cluster.


#. Add Helm repo for flyte ::

    helm repo add flyteorg https://helm.flyte.org

#. Install Flyte dependency helm chart (this will install the minio, Postgres, Kubernetes-dashboard, and contour) ::

    helm install -n flyte flyte-deps flyteorg/flyte-deps --create-namespace --set webhook.enabled=false,minio.service.type=LoadBalancer,contour.enabled=true,contour.envoy.service.type=LoadBalancer,kubernetes-dashboard.service.type=LoadBalancer

#. Install flyte-core chart ::

    helm install flyte flyteorg/flyte-core -n flyte -f https://raw.githubusercontent.com/flyteorg/flyte/master/charts/flyte-core/values-sandbox.yaml --wait

#. Verify Flyte deployment using the following command ::

    kubect get pods -n flyte

    .. note::

        Make sure all pods are in Running condition. If you see anything that's crashing, check them in this order: postgres, minio, flyteadmin, datacatalog, flytepropeller.

#. Get the URL of the ingress service ::

    kubect get ingress -n flyte

#. In order to interact with your Flyte instance using ``flytectl``, initialise your configuration to point to this host ::

    flytectl config init --host='<CONTOUR_URL>' --insecure

#. Open the minio console http://<MINIO_URL>. Your minio username is `minio` and password is `miniostorage`.

#. Open the Kubernetes dashboard http://<K8S_DASHBOARD_URL>.

#. Port-forward to connect Postgres using the following command: ::

    kubectl port-forward --address 0.0.0.0 svc/postgres 5432:5432 -n flyte
    
#. Use the following credentials for Postgres:

   .. code-block::

      dbname: flyteadmin
      host: 127.0.0.1
      port: 5432
      username: postgres


