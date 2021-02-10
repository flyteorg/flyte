.. _tutorials-getting-started-flyte-laptop:

##################################
Run Your Workflow on a Local Flyte
##################################

************************
Installing Flyte Locally
************************

This guide will walk you through a quick installation of Flyte on your laptop and then how to register and execute your
workflows against this deployment.

Estimated time to complete: 5 minutes.

Prerequisites
=============

#. Ensure ``kubectl`` is installed. Follow docs `here <https://kubernetes.io/docs/tasks/tools/install-kubectl/>`_. On Mac::

    brew install kubectl

#. If running locally ensure you have docker installed - as explained `here <https://docs.docker.com/get-docker/>`_
#. If you prefer to run the Sandbox cluster on a hosted environment like `AWS EKS <https://aws.amazon.com/eks/>`_, `Google GKE <https://cloud.google.com/kubernetes-engine>`_, then jump to :ref:`next section <tutorials-getting-started-flyte-hosted>`_.

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


****************************
Running your Flyte Workflows
****************************

Registration
============

Register a new project (optional)
---------------------------------

Once your Flyte deployment is up running you'll see a few example projects registered in the console. For the sake of this
exercise, let's create a new project you'll use to register your new workflows, but before that, if you have not already, install flytekit ::

  pip install flytekit==0.16.0b6  # if you haven't already

After installing flytekit, you can using ``flyte-cli`` to register a project ::

  flyte-cli register-project -i -h localhost:80 -p myflyteproject --name "My Flyte Project" \
    --description "My very first project onboarding onto Flyte"


If you refresh your `console <http://localhost:80/console>`__ you'll see your new project appear!

Register your workflows
-----------------------

From within root directory of ``myflyteproject`` you created :ref:`previously <tutorials-getting-started-first-example>`
commit any changes and then register them ::

Uploading your workflows to your Flyte deployment requires running the single make target below.
The command will first build a Docker image containing your code changes (later on we'll cover how to avoid building a
new Docker image every time you make code changes)
This invokes a two step process to serialize your Flyte workflow objects into a
`protobuf <https://developers.google.com/protocol-buffers>`__ representation and then makes the network call to upload
these serialized protobufs onto the Flyte platform ::

  git add . && git commit -m "Added an example flyte workflow"
  PROJECT=myflyteproject make register


Boom! It's that simple.

Run your workflows
------------------

Triggering a workflow is super simple. For now, let's do so through the UI (flyte console).

Visit the page housing workflows registered for your project:
`http://localhost/console/projects/myflyteproject/workflows <http://localhost/console/projects/myflyteproject/workflows>`__

Select your workflow, click the bright purple "Launch Workflow" button in the upper right, update the "name" input
argument as you please, proceed to launch and you'll have triggered an execution!

There are ways to trigger executions using the ``flyte-cli`` command line or even the underlying REST API, but for the
purposes of this tutorial we won't get into them quite yet.
