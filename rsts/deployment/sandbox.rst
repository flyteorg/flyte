.. _deployment-sandbox:

###################
Sandbox overview
###################

.. warning::
    The sandbox deployment is not suitable for production environments. For an in-depth overview of how to productionize your flyte deployment, checkout the :ref:`deployment` guide.


**********************
What is Flyte Sandbox?
**********************
Flyte Sandbox is a fully standalone minimal environment for running Flyte - given a kubernetes cluster. To simplify this further :std:ref:`flytectl_sandbox` provides a simplified way of running ``flyte-sandbox`` as a single docker container running locally.
The follow section explains how you can use each of these modes and provides more information. We **recommend** to run flyte-sandbox using flytectl locally on your workstations or on a single cloud instance for trying out flyte or testing new ideas. Flyte Sandbox is not a complete representation of Flyte,
many features are intentionally removed from this environment to ensure that the startup times and runtime footprints are low.

*******************************************
FlyteSandbox as a single docker container
*******************************************

Using :std:ref:`flytectl_sandbox` one can start a local sandbox environment for Flyte. This is mini-replica of an entire Flyte deployment, without the scalability and reduced extensions. The idea for this environment originated from the desire of the core team to make it extremely simple for users of Flyte to
try out Flyte and get a feel for the user experience, without having to understand kubernetes and dabble with configuration etc. Flyte single container sandbox is also used by Flyte to run continuous integration tests and used by the `flytesnacks - UserGuide playground environment`. flyte-sandbox can mostly be run
in any environment that support docker containers and an ubuntu docker base image.

Architecture & Reasons why we built it?
========================================
In the single container environment, a mini kubernetes cluster is installed within a container using the excellent `k3s <https://k3s.io/>`__ platform. K3s uses an in-container docker daemon (run using `docker-in-docker configuration <https://www.docker.com/blog/docker-can-now-run-within-docker/>`__) to orchestrate user containers.

When users use ``flytectl sandbox start --source <dir>``, the source ``<dir>`` is mounted within the outer docker container and hence it is possible to build a docker image using the inner docker-daemon. In a typical Flyte installation, one needs to build a docker container for the tasks and push to a repository from which K8s can pull the containers.
But, this is not possible with flyte-sandbox in docker environment, because it does not ship with a docker registry. Users are free to use an external registry and this should as expected, as long as the inner k3s cluster has permissions to pull containers from the external registry. But, to reduce the friction of procuring a new registry, procuring permissions to access it and then pushing to this registry,
we recommend using the ``flytectl sandbox exec -- ...`` mode to trigger a docker build for your code (which is mounted into the sandbox environment) using the docker-in-docker daemon. Since K3s just uses the same docker daemon, it is possible to re-use the container that is built internally. This greatly simplifies the users interaction and ability to try out Flyte, while sacrificing on step in the real world workflow - ``docker push``.

The illustration below shows the archtecture of flyte-sandbox in a single container. It is identical to a Flyte sandbox cluster, but the fact that we have built one docker container, with kubernetes and flyte already installed.

.. image:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/core/flyte_sandbox_single_container.png
   :alt: Architecture of flyte-sandbox single container


Uses
=====
* Try out Flyte locally using a single docker command or using flytectl sandbox
* Run regular integration tests for Flyte
* provide snapshot environment for various Flyte versions, to identify regressions

***************************************************************
Deploying your own Flytesandbox environment to a K8s cluster
***************************************************************

This installs all the dependencies as kubernetes deployments. We call this a Sandbox deployment. Flyte sandbox deployment can be deployed using the default helm template.


.. note::

    #. A Sandbox deployment takes over the entire cluster
    #. It needs special cluster roles that will need access to create namespaces, pods etc
    #. The sandbox deployment is not suitable for production environments. For an in-depth overview of how to productionize your flyte deployment, checkout the :ref:`howto_productionize`.


.. image:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/core/flyte_sandbox_single_k8s_cluster.png
   :alt: Architecture of Sandbox deployment of Flyte. Single K8s cluster


Deploy Flyte Sandbox environment locally - on your laptop
=========================================================

Ensure ``kubectl`` is installed. Follow `kubectl installation docs <https://kubernetes.io/docs/tasks/tools/install-kubectl/>`_. On Mac::

    brew install kubectl



.. tabs::

    .. tab:: Docker Image

        Refer to :ref:`getting-started-firstrun`

    .. tab:: k3d

        #. Install k3d Using ``curl``::

            curl -s https://raw.githubusercontent.com/rancher/k3d/main/install.sh | bash

           Or Using ``wget`` ::

            wget -q -O - https://raw.githubusercontent.com/rancher/k3d/main/install.sh | bash

        #. Start a new K3s cluster called flyte::

            k3d cluster create -p "30081:30081" --no-lb --k3s-server-arg '--no-deploy=traefik' --k3s-server-arg '--no-deploy=servicelb' flyte

        #. Ensure the context is set to the new cluster::

            kubectl config set-context flyte

        #. Install Flyte::

            kubectl create -f https://raw.githubusercontent.com/flyteorg/flyte/master/deployment/sandbox/flyte_generated.yaml


        #. Connect to `FlyteConsole <localhost:30081/console>`__
        #. [Optional] You can delete the cluster once you are done with the tutorial using - ::

            k3d cluster delete flyte


        .. note::

            #. Sometimes Flyteconsole will not open up. This is probably because your docker networking is impacted. One solution is to restart docker and re-do the previous steps.
            #. To debug you can try a simple excercise - run nginx as follows::

                docker run -it --rm -p 8083:80 nginx

               Now connect to `locahost:8083 <localhost:8083>`__. If this does not work, then for sure the networking is impacted, please restart docker daemon.

    .. tab:: Docker-Mac + K8s

        #. `Install Docker for mac with Kubernetes as explained here <https://www.docker.com/blog/docker-mac-kubernetes/>`_
        #. Make sure Kubernetes is started and once started make sure your kubectx is set to the `docker-desktop` cluster, typically ::

                kubectl config set-context docker-desktop

        #. Install Flyte::

            kubectl create -f https://raw.githubusercontent.com/flyteorg/flyte/master/deployment/sandbox/flyte_generated.yaml


        #. Connect to `FlyteConsole <localhost/console>`__

    .. tab::  Using Minikube (Not recommended)

        #. Install `Minikube <https://kubernetes.io/docs/tasks/tools/install-minikube/>`_

        #. Install Flyte::

            kubectl create -f https://raw.githubusercontent.com/flyteorg/flyte/master/deployment/sandbox/flyte_generated.yaml


        .. note::

            - Minikube runs in a Virtual Machine on your host
            - So if you try to access the flyte console on localhost, that will not work, because the Virtual Machine has a different IP address.
            - Flyte runs within Kubernetes (minikube), thus to access FlyteConsole, you cannot just use https://localhost:30081/console, you need to use the IP address of the minikube VM instead of localhost
            - Refer to https://kubernetes.io/docs/tutorials/hello-minikube/ to understand how to access a
                also to register workflows, tasks etc or use the CLI to query Flyte service, you have to use the IP address.
            - If you are building an image locally and want to execute on Minikube hosted Flyte environment, please push the image to docker registry running on the Minikube VM.
            - Another alternative is to change the docker host, to build the docker image on the Minikube hosted docker daemon. https://minikube.sigs.k8s.io/docs/handbook/pushing/ provides more detailed information about this process. As a TL;DR, Flyte can only run images that are accessible to Kubernetes. To make an image accessible, you could either push it to a remote registry or to a regisry that is available to Kuberentes. In case on minikube this registry is the one thats running on the VM.


.. _howto-sandbox-dedicated-k8s-cluster:

Deploy Flyte Sandbox environment to a Cloud Kubernetes cluster
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

We'll proceed like with :ref:`locally hosted flyte <getting-started-run-on-flyte>` with deploying the sandbox
Flyte configuration on your remote cluster.

#. The Flyte sandbox can be deployed with a single command ::

    kubectl create -f https://raw.githubusercontent.com/flyteorg/flyte/master/deployment/sandbox/flyte_generated.yaml


#. You can now port-forward (or if you have load-balancer enabled then get an LB) to connect to remote FlyteConsole, as follows::

    kubectl port-forward svc/envoy 30081:80


#. Open console http://localhost:30081/console.