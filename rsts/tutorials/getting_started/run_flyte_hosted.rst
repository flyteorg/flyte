.. _tutorials-getting-started-flyte-hosted:

##############################################
[Optional] Run Your Workflow on a Hosted Flyte
##############################################

*************
Prerequisites
*************

Ensure you have kubernetes up and running on your choice of cloud provider:

- `AWS EKS <https://aws.amazon.com/eks/>`_ (Amazon)
- `GCP GKE <https://cloud.google.com/kubernetes-engine/>`_ (Google)
- `Azure AKS <https://azure.microsoft.com/en-us/services/kubernetes-service/>`_ (Microsoft)

If you can access your cluster with ``kubectl cluster-info``, you're ready to deploy Flyte.

**********
Deployment
**********

We'll proceed like with :ref:`locally hosted flyte <tutorials-getting-started-flyte-laptop>` with deploying the sandbox
Flyte configuration on your remote cluster.

.. warning::
    The sandbox deployment is not suitable for production environments.

For an in-depth overview of how to productionize your flyte deployment, checkout the :ref:`how to <howto_productionize>`

The Flyte sandbox can be deployed with a single command ::

  kubectl create -f https://raw.githubusercontent.com/lyft/flyte/master/deployment/sandbox/flyte_generated.yaml


********************************
Running Flyte Workflows Remotely
********************************

Prerequisites
=============

In the case your remote deployment is not publicly available you'll need to enable port-forwarding to easily access your
workflows and seamlessly register them ::

  kubectl -n flyte port-forward svc/envoy 8001:80

And visit the `console <localhost:8001/console>`__.

Otherwise, verify you can access your console by visiting ``http://<external-ip>/console``

Registration
============

For the sake of this exercise, let's create a new project you'll use to register your new workflows.
(Be sure to update the ``-h`` Flyte host argument if pertinent.) ::

  flyte-cli register-project -i -h localhost:80 -p myflyteproject --name "My Flyte Project" \
    --description "My very first project onboarding onto Flyte"


Once again, commit your changes and register them like so ::

  git add . && git commit -m "Added an example flyte workflow"
  PROJECT=myflyteproject FLYTE_HOST=localhost:80 make register

Replace ``localhost`` with your external IP if you are **not** port-forwarding the service.

Visit the console ``http://localhost/console/projects/myflyteproject/workflows`` substituting the URL with your
external IP if necessary.

Select your workflow, click the bright purple "Launch Workflow" button in the upper right, update the "name" input
argument as you please, proceed to launch and you'll have triggered an execution!
