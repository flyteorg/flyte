.. _troubleshoot:

Troubleshooting Guide
---------------------

.. admonition:: Why have we crafted this guide?

    Let go of overthinking; peep into this page.

We've been working diligently to help users sort out issues. 

Here are a couple of techniques we believe would help you jump out of the pandora box quickly! 

Troubles with ``flytectl sandbox start``
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

- The process hangs at ``Waiting for Flyte to become ready...`` for a while; OR
- The process ends with a message ``Timed out while waiting for the datacatalog rollout to be created``

**Potential causes**

- Your Docker daemon is constrained on disk, memory or CPU potentially. Why Docker? refer to :ref:`deployment-sandbox`.
- A simple solution reclaim disk from Docker - `docs <https://docs.docker.com/engine/reference/commandline/system_prune/>`__ ::

   docker system prune [OPTIONS]

- Another simple solution is to increase mem/CPU available for Docker.

**Debug yourself**

- Sandbox is a Docker container that runs Kubernetes and Flyte in it. So you can simply use ``exec`` .

.. prompt:: bash $

 docker ps

.. code-block::

 CONTAINER ID   IMAGE                                      COMMAND                  CREATED         STATUS         PORTS                                                                                                           NAMES
 d3ab7e4cb17c   cr.flyte.org/flyteorg/flyte-sandbox:dind   "tini flyte-entrypoi…"   7 minutes ago   Up 7 minutes   127.0.0.1:30081-30082->30081-30082/tcp, 127.0.0.1:30084->30084/tcp, 2375-2376/tcp, 127.0.0.1:30086->30086/tcp   flyte-sandbox

.. prompt:: bash $

 docker exec -it <imageid> bash

Inside the container you can run::

 kubectl get pods -n flyte

You can check on the Pending pods and perform a detailed check as to why the scheduler is failing. This could also affect your workflows.

Also you can simply export this variable to use local kubectl::

 export KUBECONFIG=$HOME/.flyte/k3s/k3s.yaml


Another useful way to debug Docker is::

 docker system df


Troubles with ``flyte sandbox`` log viewing
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

- When testing locally using ``flyte sandbox`` command, one way to view the logs is using ``Kubernetes Logs (User)`` in the FlyteConsole. 
- This would take you to the Kubernetes dashboard, and require a login.

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

   There is a ``skip`` button, which will take you straight to the logs without logging in.


Troubles with ``make start`` command in Flytesnacks
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

- The ``make start`` command usually completes execution within five minutes (could take longer if you aren't in the United States).
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

- If the ``make start`` command isn't proceeding by any chance, check the pods' statuses by running this command

  ::

   docker exec flyte-sandbox kubectl get po -A
- If you think a pod has crashed or evicted by any chance, describe the pod by running the command which gives the detailed overview of the pod's status

  ::

   docker exec flyte-sandbox kubectl describe po <pod-name> -n flyte

- If Kubernetes reports a disk pressure issue: (node.kubernetes.io/disk-pressure)

  - Check the memory stats of the Docker container using the command 
  ``docker exec flyte-sandbox df -h``.
  - Prune the images and volumes.
  - Given there is less than 10% free disk space, Kubernetes, by default, throws the disk pressure error.


Troubles with FlyteCTL command within proxy setting
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

- FlyteCTL uses GRPC APIs of the FlyteAdmin to administer Flyte resources and in case of proxy settings, it uses an additional CONNECT handshake at GRPC layer to perfrom the same. Additonal info is available `here <https://github.com/grpc/grpc-go/blob/master/Documentation/proxy.md>`__

- On Windows environment it has been noticed that ``NO_PROXY`` variable doesn't work to bypass the proxy settings. `This <https://github.com/grpc/grpc/issues/9989>`__ GRPC issue provides additional details though it doesn't seem to have been tested on Windows yet. Inorder to get around this issue, unset both the ``HTTP_PROXY`` and ``HTTPS_PROXY`` variables.

Troubles with ``flytectl`` command with Cloudfare DNS
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

- FlyteCTL throws permission error with Cloudfare DNS endpoint
- Cloudflare instance by default would proxy the requests and would filter out GRPC.
- You have two options: 
    - Either enable grpc in the network tab; OR
    - Turn off the proxy.

Troubles with ``flytectl`` command with auth enabled
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

- The ``flytectl`` command uses OpenID connect if auth is enabled in the Flyte environment.
- It opens an ``HTTP`` server port on localhost:53593. It has a callback endpoint for the OpendID connect server to call into for the response.
- If the callback server call fails, please check if ``flytectl`` failed to run the server.
- Verify if you have an entry for localhost in your ``/etc/hosts`` file.
- It could also mean that the call back took longer and FlyteCTL deadline expired on the wait which defaults to 15 secs.

.. NOTE::

      More coming soon. Stay tuned 👀


What if none of the above methods solved my issue?
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
Our `Slack <http://flyte-org.slack.com/>`__ community is always available and ready to help!
