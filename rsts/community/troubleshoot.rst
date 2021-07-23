.. _troubleshoot:

Troubleshooting Guide
---------------------

.. admonition:: Why have we crafted this guide?

    Let go of overthinking; peep into this page.

We've been working diligently to help users sort out issues. 

Here are a couple of techniques we believe would help you jump out of the pandora box quickly! 

Troubles with ``flytectl sandbox start``
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

- The process hangs at ``Waiting for Flyte to become ready...`` for a while
- OR ends with a message ``Timed out while waiting for the datacatalog rollout to be created``

**Potential causes**

- your docker daemon is constrained on disk, memory or CPU potentially. Why docker? refer to :ref:`deployment-sandbox`.
- a simple solution reclaim disk from docker - `docs <https://docs.docker.com/engine/reference/commandline/system_prune/>`__ ::

   docker system prune [OPTIONS]

- Another simple solution, increase mem / cpu available for docker

**Debug yourself**

- sandbox is a docker container that runs kubernetes and flyte in it. So you can simple ``exec`` into it

.. prompt:: bash $

 docker ps

.. code-block::

 CONTAINER ID   IMAGE                                      COMMAND                  CREATED         STATUS         PORTS                                                                                                           NAMES
 d3ab7e4cb17c   cr.flyte.org/flyteorg/flyte-sandbox:dind   "tini flyte-entrypoi…"   7 minutes ago   Up 7 minutes   127.0.0.1:30081-30082->30081-30082/tcp, 127.0.0.1:30084->30084/tcp, 2375-2376/tcp, 127.0.0.1:30086->30086/tcp   flyte-sandbox

.. prompt:: bash $

 docker exec -it <imageid> bash

Inside the container you can run::

 kubectl get pods -n flyte

you can check on the Pending pods and do a detail check as to why the scheduler is failing. This could also affect your workflows.

Also you can simply export this variable to use local kubectl::

 export KUBECONFIG=$HOME/.flyte/k3s/k3s.yaml


Another useful way to debug docker is::

 docker system df


Troubles with ``flyte sandbox`` log viewing
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

- When testing locally using flyte-sandbox, one way to view the logs is by the link ``Kubernetes Logs (User)`` in the FlyteConsole. 
- This might take you to the kubernetes dashboard, and require a login.

::

     kind: Deployment
     apiVersion: apps/v1
     metadata:
       name: kubernetes-dashboard
       namespace: kubernetes-dashboard
     spec:
       template:
         spec:
           containers:
             - name: kubernetes-dashboard
               args:
                 - --namespace=kubernetes-dashboard
                 - --enable-insecure-login
                 - --enable-skip-login
                 - --disable-settings-authorizer

.. note::

   There is a ``skip`` button, which will take you straight to the logs without needing to log in.


Troubles with ``make start`` command in flytesnacks
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

- ``make start`` usually gets completed within five minutes (could take longer if you aren't in the United States).
- If ``make start`` results in a timeout issue:
   .. code-block:: bash

     Starting Flyte sandbox
     Waiting for Flyte to become ready...
     Error from server (NotFound): deployments.apps "datacatalog" not found
     Error from server (NotFound): deployments.apps "flyteadmin" not found
     Error from server (NotFound): deployments.apps "flyteconsole" not found
     Error from server (NotFound): deployments.apps "flytepropeller" not found
     Timed out while waiting for the Flyte deployment to start

   You can run ``make teardown`` followed by the ``make start`` command.

- If the ``make start`` command isn't proceeding by any chance, check the pods' statuses by run this command

  ::

   docker exec flyte-sandbox kubectl get po -A
- If you think a pod's crashing or getting evicted by any chance, describe the pod by running the command which gives detailed overview of pod's status

  ::

   docker exec flyte-sandbox kubectl describe po <pod-name> -n flyte

- If Kubernetes reports a disk pressure issue: (node.kubernetes.io/disk-pressure)

  - Check the memory stats of the docker container using the command ``docker exec flyte-sandbox df -h``.
  - Prune the images and volumes.
  - Given there's less than 10% free disk space, Kubernetes, by default, throws the disk pressure error.

.. NOTE::

      More coming soon. Stay tuned 👀


I NEED HELP!
^^^^^^^^^^^^^
The community is always available and ready to help `Slack <http://flyte-org.slack.com/>`__.
