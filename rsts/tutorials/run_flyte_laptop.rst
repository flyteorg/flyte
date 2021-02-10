.. _tutorials-getting-started-flyte-laptop:

##################################
Run Your Workflow on a Local Flyte
##################################

This guide will walk you through a quick installation of Flyte on your laptop and then how to register and execute your
workflows against this deployment.

Estimated time to complete: 5 minutes.

************************
Installing Flyte Locally
************************

Prerequisites
=============

Kubernetes and its ``kubectl`` client are the only strict prerequisites to installing Flyte.

Linux
-------
For Linux, you'll need to have Docker set up. For the local Kubernetes cluster itself, we've found that
`KinD <https://kind.sigs.k8s.io/docs/user/quick-start>`__ works better than MicroK8s, but this may change in the future.
Minikube should also work.

.. note::
   The Docker daemon typically runs as root in Linux (though there is a new option for running it rootless -
   we haven't tested that with KinD yet, so it may or may not work). Because of this, you may want to use
   ``sudo /full/path/to/kind`` and prepend sudo to your kubectl commands as well.

Mac OS
---------
For Macs, we recommend `Docker Desktop <https://www.docker.com/products/docker-desktop>`__. Docker Desktop ships with a
Kubernetes cluster, which is the easiest option to use. One can also use KinD.

Flyte Sandbox Deployment
========================

The simplest Flyte deployment is the "sandbox" deployment, which includes everything you need in order to use Flyte.
The Flyte sandbox can be deployed with a single command ::

  kubectl create -f https://raw.githubusercontent.com/flyteorg/flyte/master/deployment/sandbox/flyte_generated.yaml

Once the deployment comes up (this will take a moment or two) you should be able to verify that Flyte is running locally
by visiting `http://localhost:80/console <http://localhost:80/console>`__

For Linux, you'll need to forward the port over before being able to hit that link with your browser

See note above on sudo ::

  sudo kubectl -n flyte port-forward service/contour 80:80

(for Minikube deployment, you need to run minikube tunnel and use the ip that Minikube tunnel outputs)

****************************
Running your Flyte Workflows
****************************

Registration
============

Register a new project (optional)
---------------------------------

Once your Flyte deployment is up running you'll see a few example projects registered in the console. For the sake of this
exercise, let's create a new project you'll use to register your new workflows, but before that, if you have not already, install flytekit ::

  pip install flytekit==0.16.6b  # if you haven't already

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
