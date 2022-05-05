.. _deployment-sandbox:

###################
Sandbox Overview
###################

.. warning::
    Don't use sandbox deployment for production environments. For an in-depth overview of how to productionize your Flyte deployment, checkout the :ref:`deployment` guide.


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
    #. To productionize your Flyte deployment, check the rest of the :ref:`deployment` guides.


.. image:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flyte/deployment/sandbox/flyte_sandbox_single_k8s_cluster.png
   :alt: Architecture of Sandbox deployment of Flyte. Single K8s cluster


.. _deploy-sandbox-local:

Deploy Flyte Sandbox environment laptop/workstation/single machine
==================================================================

Ensure ``kubectl`` is installed. Follow `kubectl installation docs <https://kubernetes.io/docs/tasks/tools/install-kubectl/>`__. On Mac::

    brew install kubectl

.. tabbed:: Docker Image

    Recommend using ``flytectl sandbox`` as describe in :ref:`getting-started`

    .. prompt:: bash $

          docker run --rm --privileged -p 30081:30081 -p 30084:30084 -p 30088:30088 cr.flyte.org/flyteorg/flyte-sandbox

.. tabbed:: k3d

    .. TODO::

      Allow installing latest version of k3d once this `issue is fixed <https://github.com/rancher/k3d/issues/380>`__

    #. Install k3d Using ``curl``::

        curl -s https://raw.githubusercontent.com/rancher/k3d/main/install.sh | TAG=v4.2.0 bash

       Or Using ``wget`` ::

        wget -q -O - https://raw.githubusercontent.com/rancher/k3d/main/install.sh | TAG=v4.2.0 bash


    #. Start a new K3s cluster called Flyte ::

        k3d cluster create flyte -p 30081:30081 --no-lb  --k3s-server-arg '–no-deploy=traefik' --k3s-server-arg '–no-deploy=servicelb'


    #. Ensure the context is set to the new cluster::

        kubectl config set-context flyte

    #. Install Flyte::

        kubectl create ns flyte
        kubectl create -f  https://raw.githubusercontent.com/flyteorg/flyte/master/deployment/sandbox/flyte_helm_generated.yaml


    #. Connect to `FlyteConsole <localhost:30081/console>`__
    #. [Optional] You can delete the cluster once you are done with the tutorial using - ::

        k3d cluster delete flyte

    .. note::

        #. Sometimes Flyteconsole will not open up. This is probably because your docker networking is impacted. One solution is to restart docker and repeat the previous steps.
        #. To debug you can try a simple exercise - run nginx as follows::

            docker run -it --rm -p 8083:80 nginx

           Now connect to `locahost:8083 <localhost:8083>`__. If this does not work, then the networking is most probably impacted. Please restart docker daemon.

.. tabbed:: Docker-Mac + K8s

    .. TODO::

        These instructions currently still rely on the old kustomize setup, and will be moved over to the Helm chart soon.

    #. Install Docker for mac with Kubernetes as explained `here <https://www.docker.com/blog/docker-mac-kubernetes/>`_
    #. Make sure Kubernetes is started and once started make sure your kubectx is set to the `docker-desktop` cluster, typically ::

            kubectl config set-context docker-desktop

    #. Install Flyte::

        kubectl create -f https://raw.githubusercontent.com/flyteorg/flyte/master/deployment/sandbox/flyte_generated.yaml

    #. Connect to `FlyteConsole <localhost/console>`__

.. tabbed::  Using Minikube (Not recommended)

    .. TODO::

        These instructions currently still rely on the old kustomize setup, and will be moved over to the Helm chart soon.

    #. Install `Minikube <https://kubernetes.io/docs/tasks/tools/install-minikube/>`_

    #. Install Flyte::

        kubectl create -f https://raw.githubusercontent.com/flyteorg/flyte/master/deployment/sandbox/flyte_generated.yaml

    .. note::

        - Minikube runs in a Virtual Machine on your host
        - So if you try to access the flyte console on localhost, that will not work, because the Virtual Machine has a different IP address.
        - Flyte runs within Kubernetes (minikube), so to access FlyteConsole, you cannot just use https://localhost:30081/console. You need to use the IP address of the minikube VM instead of the localhost
        - Refer to https://kubernetes.io/docs/tutorials/hello-minikube/ to understand how to run a sample app on kubernetes using minikube and Katacoda. To  register workflows, tasks, etc. or use the CLI to query Flyte services, you have to use the IP address.
        - If you are building an image locally and want to execute on Minikube hosted Flyte environment, please push the image to docker registry running on the Minikube VM.
        - Another alternative is to change the docker host, to build the docker image on the Minikube hosted docker daemon. https://minikube.sigs.k8s.io/docs/handbook/pushing/ provides more detailed information about this process. Flyte can only run images that are accessible to Kubernetes. To make an image accessible, you could either push it to a remote registry or to a registry that is available to Kubernetes. In the minikube case, this registry is the one that is running on the VM.


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


#. The Flyte sandbox can be deployed via a helm chart. From the root dir of the flyte repo run ::

    helm repo add flyte https://flyteorg.github.io/flyte
    helm install -n flyte -f values.yaml --create-namespace flyte flyte/flyte

#. For customizations instructions, see ``/charts/flyte/README.md`` in the flyte repo.

#. You can now port-forward (or if you have load-balancer enabled then get an LB) to connect to remote FlyteConsole, as follows ::

    kubectl port-forward --address 0.0.0.0 svc/flyte-contour-envoy 30081:80 -n flyte

#. Open the console http://localhost:30081/console.

#. In order to interact with your Flyte instance using ``flytectl``, initialise your configuration to point to this host ::

    flytectl config init --host='localhost:30081' --insecure

#. Open the minio console http://localhost:30088. Your minio username is `minio` and password is `miniostorage`.

#. Open the Kubernetes dashboard http://localhost:30082.

#. You can port-forward to connect postgres using ::

    kubectl port-forward --address 0.0.0.0 svc/postgres 5432:5432 -n flyte

#. Now use these credentials for postgres

   .. code-block::

      dbname: flyteadmin
      host: 127.0.0.1
      port: 5432
      username: postgres


