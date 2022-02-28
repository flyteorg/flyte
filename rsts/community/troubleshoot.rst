.. _troubleshoot:

Troubleshooting Guide
---------------------

.. admonition:: Why did we craft this guide?

    To help streamline your onboarding experience as much as possible, and sort out common issues.

Here are a couple of techniques we believe could help get you up and running in no time! 

Troubles With ``flytectl sandbox start``
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

- The process hangs at ``Waiting for Flyte to become ready...`` for a while; OR 
- It ends with a message ``Timed out while waiting for the datacatalog rollout to be created``.

How Do I Debug?
"""""""""""""""

- Sandbox is a Docker container that runs Kubernetes and Flyte in it. So you can simply ``exec`` into it;

.. prompt:: bash $

 docker ps

.. code-block::

 CONTAINER ID   IMAGE                                      COMMAND                  CREATED         STATUS         PORTS                                                                                                           NAMES
 d3ab7e4cb17c   cr.flyte.org/flyteorg/flyte-sandbox:dind   "tini flyte-entrypoi…"   7 minutes ago   Up 7 minutes   127.0.0.1:30081-30082->30081-30082/tcp, 127.0.0.1:30084->30084/tcp, 2375-2376/tcp, 127.0.0.1:30086->30086/tcp   flyte-sandbox

.. prompt:: bash $

 docker exec -it <imageid> bash

and run: ::

    kubectl get pods -n flyte

You can check on the pending pods and perform a detailed check as to why a pod is failing by running: ::

    kubectl describe po <pod-name> -n flyte 

- Also, you can use this command to simply export this variable to use local kubectl::

    export KUBECONFIG=$HOME/.flyte/k3s/k3s.yaml

- If you would like to reclaim disk space, run: ::

    docker system prune [OPTIONS]

- Increase mem/CPU available for Docker.


Troubles With ``flyte sandbox`` Log Viewing
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

- When testing locally using the ``flyte sandbox`` command, one way to view the logs is using the ``Kubernetes Logs (User)`` option on the FlyteConsole. 
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

Troubles With Flytectl Commands Within Proxy Settings
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

- Flytectl uses gRPC APIs of FlyteAdmin to administer Flyte resources and in the case of proxy settings, it uses an additional ``CONNECT`` handshake at the gRPC layer to perform the same. Additional info is available on this `gRPC proxy documentation <https://github.com/grpc/grpc-go/blob/master/Documentation/proxy.md>`__ page.

- In the Windows environment, it has been noticed that the ``NO_PROXY`` variable doesn't work to bypass the proxy settings. This `GRPC issue <https://github.com/grpc/grpc/issues/9989>`__ provides additional details, though it doesn't seem to have been tested on Windows yet. To bypass this issue, unset both ``HTTP_PROXY`` and ``HTTPS_PROXY`` variables.

Troubles With Flytectl Commands With Cloudflare DNS
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

- Flytectl produces permission errors with Cloudflare DNS endpoints
- Cloudflare instance proxies by default the requests and filters out gRPC.
- **To fix this**: 
    - Enable gRPC in the network tab; or
    - Turn off the proxy.

Troubles With Flytectl Commands With Auth Enabled
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

- Flytectl commands use OpenID connect if auth is enabled in the Flyte environment
- It opens an ``HTTP`` server port on localhost:53593. It has a callback endpoint for the OpenID connect server to call into for the response.
    - If the callback server call fails, please check if Flytectl failed to run the server.
    - Verify that you have an entry for localhost in your ``/etc/hosts`` file.
    - It could also mean that the callback took longer than the default 15 secs, and the Flytectl wait deadline expired. 


I Still Need Help!
^^^^^^^^^^^^^^^^^^
Our `Slack <https://slack.flyte.org/>`__ community is always available and ready to help!
