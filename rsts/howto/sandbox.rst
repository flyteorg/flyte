.. _howto-sandbox:

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

Ensure ``kubectl`` is installed. Follow `kubectl installation docs <https://kubernetes.io/docs/tasks/tools/install-kubectl/>`_. On Mac::

    brew install kubectl



.. tabs::

    .. tab:: Docker Image

        Refer to :ref:`flyte-tutorials-firstrun`

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

******************************************************************
Deploy Flyte Sandbox environment to a Cloud Kubernetes cluster
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
