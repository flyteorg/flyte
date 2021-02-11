.. _tutorials-getting-started-flyte-laptop:

##################################
Run Your Workflow on a Local Flyte
##################################

************************
Installing Flyte Locally
************************

This guide will walk you through a quick installation of Flyte on your laptop and then how to register and execute your
workflows against this deployment.

.. rubric:: Estimated time to complete: 5 minutes.

Prerequisites
=============

#. Ensure ``kubectl`` is installed. Follow `kubectl installation docs <https://kubernetes.io/docs/tasks/tools/install-kubectl/>`_. On Mac::

    brew install kubectl

#. If running locally ensure you have docker installed - as explained `here <https://docs.docker.com/get-docker/>`_
#. If you prefer to run the Flyte test cluster on a cloud environment like `AWS EKS <https://aws.amazon.com/eks/>`_, `Google GKE <https://cloud.google.com/kubernetes-engine>`_, then follow the instructions for hosted Flyte below.

Steps
======

.. tabs::

    .. tab:: Using k3d

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


        #. Connect to `FlyteConsole <localhost:30081/console>`_
        #. [Optional] You can delete the cluster once you are done with the tutorial using - ::

            k3d cluster delete flyte


        .. note::

            #. Sometimes Flyteconsole will not open up. This is probably because your docker networking is impacted. One solution is to restart docker and re-do the previous steps.
            #. To debug you can try a simple excercise - run nginx as follows::

                docker run -it --rm -p 8083:80 nginx

               Now connect to `locahost:8083 <localhost:8083>`_. If this does not work, then for sure the networking is impacted, please restart docker daemon.

    .. tab:: Using Docker for Mac

        Coming soon

    .. tab:: Hosted Flyte

        Refer to the :ref:`howto-sandbox-dedicated-k8s-cluster` guide.

        Once you've deployed flyte sandbox to a cloud provider, refer to :ref:`tutorials-run-flyte-laptop` to run your first workflow remotely.


.. _tutorials-run-flyte-laptop:

****************************
Running your Flyte Workflows
****************************

Registration
============

Set-up
------

In a separate process, port-forward the flyte minio service. If you're using a locally deployed Flyte sandbox, simply: ::

  kubectl -n flyte port-forward service/minio 9000


Register your workflows
-----------------------

From within root directory of ``flyteexamples`` you created :ref:`previously <tutorials-getting-started-first-example>`
feel free to make any changes and then register ::

  FLYTE_AWS_ENDPOINT=http://localhost:9000 FLYTE_AWS_ACCESS_KEY_ID=minio  \
    FLYTE_AWS_SECRET_ACCESS_KEY=miniostorage make fast_register

If you're port-forwarding minio somewhere else, substitute the ``FLYTE_AWS_ENDPOINT`` accordingly.

.. rubric:: Boom! It's that simple.

Run your workflows
------------------

Visit the page housing workflows registered for your project (method if you used k3d):
`http://localhost:30081/console/projects/flyteexamples/workflows <http://localhost:30081/console/projects/flyteexamples/workflows>`__
else if you used docker-desktop or something else, then copy paste this URL into the browser and fill in the ``<host:port>``::

    http://<host:port>/console/projects/flyteexamples/workflows


Select your workflow, click the bright purple "Launch Workflow" button in the upper right, update the "name" input
argument as you please, proceed to launch and you'll have triggered an execution!

.. note::

    After registration Flyte Workflows exist in the FlyteAdmin service and can be triggered using the
      - console
      - Command line
      - directly invoking the REST API
      - on a schedule

    More on this later

Optionally you can create a new project
----------------------------------------
Refer to :ref:`howto_new_project`.