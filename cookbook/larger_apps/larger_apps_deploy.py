"""
.. _larger_apps_deploy:

Deploy to the Cloud
-------------------

Prerequisites
^^^^^^^^^^^^^^^^
Make sure you have `Docker <https://docs.docker.com/get-docker/>`__ and Docker
Daemon is running.

.. note::

   Being connected to a VPN may cause problems downloading the image.

Setup and Configuration
^^^^^^^^^^^^^^^^^^^^^^^^

First, install `flytectl <https://docs.flyte.org/projects/flytectl/en/latest/#installation>`__.

Then you can setup a local :ref:`Flyte Sandbox <deployment-sandbox>` cluster or configure ``flytectl`` to use a
pre-provisioned remote Flyte cluster.

.. tip::

   Learn how to deploy to a Flyte Cluster using the :ref:`Deployment Guides <deployment>`.

.. tabs::
   .. group-tab:: Flyte Sandbox

      To start the Flyte Sandbox, run:

      .. prompt:: bash $

         flytectl sandbox start --source .

      .. note::

         The ``'.'`` represents the current directory, which would be ``my_flyte_project`` in this case.

         If you're having trouble with starting the sandbox cluster, refer to :ref:`troubleshoot`.

   .. group-tab:: Remote Flyte Cluster

      Setup flytectl remote cluster config

      .. prompt:: bash $

         flytectl config init --host={FLYTEADMIN_URL}

      where ``FLYTEADMIN_URL`` is your custom url.

Build & Deploy Your Application to the Cluster
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Flyte uses Docker containers to package the workflows and tasks, and sends them to the remote Flyte cluster. Therefore,
there is a ``Dockerfile`` already included in the cloned repo. You can build the Docker container and push the built
image to a registry.

.. tabs::
   .. group-tab:: Flyte Sandbox

      Since ``flyte-sandbox`` runs locally in a Docker container, you do not need to push the Docker image. You can
      combine the build and push step by simply building the image inside the Flyte-sandbox container. This can be done
      using the following command:

      .. prompt:: bash $

         flytectl sandbox exec -- docker build . --tag "my_flyte_project:v1"

      .. tip::
         Why are we not pushing the Docker image? To understand the details, refer to :ref:`deployment-sandbox`.

   .. group-tab:: Remote Flyte Cluster

      If you are using a remote Flyte cluster, then you need to build your container and push it to a registry that
      is accessible by the Flyte Kubernetes cluster.

      .. prompt:: bash $

         docker build . --tag <registry/repo:version>
         docker push <registry/repo:version>

      .. tip:: Recommended

         The ``flytekit-python-template`` ships with a helper script called `docker_build_and_tag.sh
         <https://github.com/flyteorg/flytekit-python-template/blob/main/simple-example/%7B%7Bcookiecutter.project_name%7D%7D/docker_build_and_tag.sh>`__,
         which makes it possible to build and image, tag it correctly and optionally use the git-SHA as the version. We recommend using
         such a script to track versions more effectively and using a CI/CD pipeline to deploy your code.

         .. prompt:: bash $

            ./docker_build_and_tag.sh -r <registry> -a <repo> [-v <version>]

Next, package the workflow using the ``pyflyte`` cli bundled with Flytekit and upload it to the Flyte backend. Note that
the image is the same as the one built in the previous step.

.. tabs::

   .. group-tab:: Flyte Sandbox

      .. prompt:: bash $

         pyflyte --pkgs flyte.workflows package --image "my_flyte_project:v1"

   .. group-tab:: Remote Flyte Cluster

      .. prompt:: bash $

         pyflyte --pkgs flyte.workflows package --image <registry/repo:version>

Upload this package to the Flyte backend. We refer to this as ``registration``. The version here ``v1`` does not have to
match the version used in the commands above, but it is generally recommended to match the versions to make it easier to
track.

.. note::

   Note that we are simply using an existing project ``flytesnacks`` and an existing domain ``development`` to register
   the workflows and tasks. It is possible to create your own project
   and configure domains. Refer to :ref:`control-plane` to understand projects and domains.

.. prompt:: bash $

   flytectl register files --project flytesnacks --domain development --archive flyte-package.tgz --version v1

Finally, visualize the registered workflow in a web browser.

.. prompt:: bash $

   flytectl get workflows --project flytesnacks --domain development flyte.workflows.example.my_wf --version v1 -o doturl

You can also view the workflow as a ``strict digraph`` on the command line.

.. prompt:: bash $

   flytectl get workflows --project flytesnacks --domain development flyte.workflows.example.my_wf --version v1 -o dot

.. _getting-started-execute:

Execute on the Flyte Cluster
^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Launch and monitor from the CLI using Flytectl.
More details can be found `here <https://docs.flyte.org/projects/flytectl/en/stable/gen/flytectl_create_execution.html>`__.

Generate an execution spec file.

.. prompt:: bash $

   flytectl get launchplan --project flytesnacks --domain development flyte.workflows.example.my_wf --latest --execFile exec_spec.yaml

Create an execution using the exec spec file.

.. prompt:: bash $

   flytectl create execution --project flytesnacks --domain development --execFile exec_spec.yaml

You should see an output that looks like:

.. prompt:: text

   execution identifier project:"flytesnacks" domain:"development" name:"<execution_name>"

Monitor the execution by providing the execution name from the ``create execution`` command.

.. prompt:: bash $

   flytectl get execution --project flytesnacks --domain development <execution_name>


**Alternatively, you can FlyteConsole to launch an execution.**

.. tabs::

   .. group-tab:: Flyte Sandbox

      Visit ``http://localhost:30080/console`` on your browser

   .. group-tab:: Remote Flyte Cluster

      Visit ``{FLYTEADMIN_URL}/console`` on your browser

Then use the FlyteConsole to launch an execution:

.. image:: https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/index/getting_started_reg.gif
    :alt: A quick visual tour for launching a workflow and checking the outputs when they're done.


Conclusion
^^^^^^^^^^^

We've successfully packaged your workflow and tasks and pushed them to a Flyte cluster! 🎉
Next, let's learn how to :ref:`iterate on and re-deploy <larger_apps_iterate>` our Flyte app.

"""
