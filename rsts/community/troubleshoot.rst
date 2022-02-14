.. _troubleshoot:

Troubleshooting Guide
---------------------

.. admonition:: Why have we crafted this guide?

    Let go of overthinking; peep into this page.

We've been working diligently to help users sort out issues. 

Here are a couple of techniques we believe would help you jump out of the pandora box quickly! 


Debug yourself
^^^^^^^^^^^^^^^

- Sandbox is a Docker container that runs Kubernetes and Flyte in it. So you can simply ``exec`` into it;

.. prompt:: bash $

 docker ps

.. code-block::

 CONTAINER ID   IMAGE                                      COMMAND                  CREATED         STATUS         PORTS                                                                                                           NAMES
 d3ab7e4cb17c   cr.flyte.org/flyteorg/flyte-sandbox:dind   "tini flyte-entrypoi…"   7 minutes ago   Up 7 minutes   127.0.0.1:30081-30082->30081-30082/tcp, 127.0.0.1:30084->30084/tcp, 2375-2376/tcp, 127.0.0.1:30086->30086/tcp   flyte-sandbox

.. prompt:: bash $

 docker exec -it <imageid> bash

- and run: ::

    kubectl get pods -n flyte

- You can check on the pending pods and perform a detailed check as to why a pod is failing::

    kubectl describe po <pod-name> -n flyte 

- Also, you can simply export this variable to use local kubectl::

    export KUBECONFIG=$HOME/.flyte/k3s/k3s.yaml


- Another useful way to debug Docker is::

    docker system df

- If you have trouble running ``flytectl sandbox start``, it could be because your Docker daemon is constrained on disk, memory, or CPU. If the process hangs at ``Waiting for Flyte to become ready...`` for a while; OR ends with a message ``Timed out while waiting for the datacatalog rollout to be created``,
    - Reclaim disk space using the following command: ::

        docker system prune [OPTIONS]

    - Increase mem/CPU available for Docker.

.. note::
    Why Docker? Refer to :ref:`deployment-sandbox`.


Troubles with ``flyte sandbox`` log viewing
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

- When testing locally using the ``flyte sandbox`` command, one way to view the logs is using the ``Kubernetes Logs (User)`` option on the Flyte Console. 
- This takes you to the Kubernetes dashboard which requires a login.

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

   There is a ``skip`` button that takes you straight to the logs without logging in.

Troubles with FlyteCTL command within proxy setting
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

- FlyteCTL uses GRPC APIs of FlyteAdmin to administer Flyte resources and in the case of proxy settings, it uses an additional ``CONNECT`` handshake at the GRPC layer to perform the same. Additional info is available `here <https://github.com/grpc/grpc-go/blob/master/Documentation/proxy.md>`__.

- On the Windows environment, it has been noticed that the ``NO_PROXY`` variable doesn't work to bypass the proxy settings. `This <https://github.com/grpc/grpc/issues/9989>`__ GRPC issue provides additional details, though it doesn't seem to have been tested on Windows yet. To get around this issue, unset both the ``HTTP_PROXY`` and ``HTTPS_PROXY`` variables.

Troubles with FlyteCTL commands with Cloudflare DNS
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

- FlyteCTL throws permission error with Cloudflare DNS endpoint
- Cloudflare instance by default proxies the requests and would filter out GRPC.
- To fix this: 
    - Enable grpc in the network tab; OR
    - Turn off the proxy.

Troubles with FlyteCTL commands with auth enabled
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

- FlyteCTL commands use OpenID connect if auth is enabled in the Flyte environment
- It opens an ``HTTP`` server port on localhost:53593. It has a callback endpoint for the OpenID connect server to call into for the response
    - If the callback server call fails, please check if FlyteCTL failed to run the server
    - Verify if you have an entry for localhost in your ``/etc/hosts`` file
    - It could also mean that the callback took longer and the FlyteCTL deadline expired on the wait which defaults to 15 secs


I NEED HELP!
^^^^^^^^^^^^^
Our `Slack <http://flyte-org.slack.com/>`__ community is always available and ready to help!
