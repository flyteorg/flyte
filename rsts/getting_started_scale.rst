.. _gettingstarted_scale:

Getting Started: Scale
----------------------

.. raw:: html
  
    <p style="color: #808080; font-weight: 500; font-size: 20px; padding-top: 10px;">A step-by-step guide to building, deploying, and iterating on Flyte tasks and workflows</p>

.. div:: getting-started-panels

   .. panels::
      :body: text-justify
      :container: container-xs
      :column: col-lg-4 col-md-4 col-sm-6 col-xs-12 p-2

      ---
      .. link-button:: gettingstarted_implement
         :type: ref
         :text: 1. Implement
         :classes: btn-outline-primary btn-block stretched-link
      ---
      .. link-button:: gettingstarted_scale
         :type: ref
         :text: 2. Scale
         :classes: btn-primary btn-block stretched-link
      ---
      .. link-button:: gettingstarted_iterate
         :type: ref
         :text: 3. Iterate
         :classes: btn-block stretched-link


Scale by Deploying Your Workflows into the Cloud
=================================================

.. _getting-started-build-deploy:

Prerequisites
^^^^^^^^^^^^^^^^
Make sure you have `Docker <https://docs.docker.com/get-docker/>`__ and Docker Daemon is running. Also some of our users noted, that if connected to VPN may cause problems downloading the image.

Install flytectl
^^^^^^^^^^^^^^^^^

#. Install :std:ref:`flytectl`. ``Flytectl`` is a commandline interface for Flyte.

   .. tabbed:: OSX

     .. prompt:: bash $

        brew install flyteorg/homebrew-tap/flytectl

     *Upgrade* existing installation using the following command:

     .. prompt:: bash $

        flytectl upgrade

   .. tabbed:: Other Operating systems

     .. prompt:: bash $

         curl -s https://ctl.flyte.org/install | sudo bash -s -- -b /usr/local/bin # You can change path from /usr/local/bin to any file system path
         export PATH=$(pwd)/bin:$PATH # Only required if user used different path then /usr/local/bin

     *Upgrade* existing installation using the following command:

     .. prompt:: bash $

        flytectl upgrade

   **Test** if Flytectl is installed correctly (your Flytectl version should be >= 0.1.34.) using the following command:

   .. prompt:: bash $

      flytectl version

#. Flyte can be deployed locally using a single Docker container — we refer to this as the ``flyte-sandbox`` environment. You can also run this getting started against a hosted or pre-provisioned environment. Refer to :ref:`deployment` section to learn how to deploy a flyte cluster.

   .. tabbed:: Start a new sandbox cluster

     .. tip:: Want to dive under the hood into flyte-sandbox, refer to :ref:`deployment-sandbox`.

     Here '.' represents current directory and assuming you have changed into ``myflyteapp`` — the git-cloned directory you created.

     .. prompt:: bash $

        flytectl sandbox start --source .

     Setup flytectl sandbox config

     .. prompt:: bash $

        flytectl config init

     **NOTE** if having trouble with starting the sandbox refer to :ref:`troubleshoot`.

   .. tabbed:: Connect to an existing Flyte cluster

     .. prompt:: bash $

         flytectl config init --host={FLYTEADMIN_URL} --storage


Build & Deploy Your Application to the Cluster
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
#. Flyte uses Docker containers to package the workflows and tasks and sends them to the remote Flyte cluster. Thus, there is a ``Dockerfile`` already included in the cloned repo. You can build the Docker container and push the built image to a registry.

   .. tabbed:: Flyte Sandbox

     Since ``flyte-sandbox`` runs locally in a Docker container, you need not push the Docker image. You can combine the build and push step by simply building the image inside the Flyte-sandbox container. This can be done using the following command:

     .. prompt:: bash $

         flytectl sandbox exec -- docker build . --tag "myapp:v1"

     .. tip::
      #. Why are we not pushing the Docker image? Want to understand the details — refer to :ref:`deployment-sandbox`
      #. *Recommended:* Use the bundled `./docker_build_and_tag.sh`. It will automatically build the local Dockerfile, name it and tag it with the current git-SHA. This helps in achieving GitOps style workflows.

   .. tabbed:: Remote Flyte Cluster

     If you are using a remote Flyte cluster, then you need to build your container and push it to a registry that is accessible by the Flyte Kubernetes cluster.

     .. prompt:: bash $

         docker build . --tag <registry/repo:version>
         docker push <registry/repo:version>

     **OR** ``flytekit-python-template`` ships with a helper `docker build script <https://github.com/flyteorg/flytekit-python-template/blob/main/docker_build_and_tag.sh>`__ which make it possible to build and image, tag it correctly and optionally use the git-SHA as the version.
     We recommend using such a script to track versions more effectively and use a CI/CD pipeline to deploy your code.

     .. prompt:: bash $

         ./docker_build_and_tag.sh -r <registry> -a <repo> [-v <version>]

#. Next, package the workflow using the ``pyflyte`` cli bundled with Flytekit and upload it to the Flyte backend. Note that the image is the same as the one built in the previous step.

   .. prompt:: bash (venv)$

      pyflyte --pkgs myapp.workflows package --image <registry/repo:version>

#. Upload this package to the Flyte backend. We refer to this as ``registration``. The version here ``v1`` does not have to match the version
   used in the commands above. It's generally recommended to match the versions to make it easier to track.

   .. note::

      Note that we are simply using an existing project ``flytesnacks`` and an existing domain ``development`` to register the workflows and tasks. It is possible to create your own project
      and configure domains. Refer to :ref:`control-plane` to understand projects and domains.

   .. prompt:: bash $

      flytectl register files --project flytesnacks --domain development --archive flyte-package.tgz --version v1

#. Finally, visualize the registered workflow.

   .. prompt:: bash $

      flytectl get workflows --project flytesnacks --domain development myapp.workflows.example.my_wf --version v1 -o doturl


.. _getting-started-execute:

Execute on Flyte Cluster
^^^^^^^^^^^^^^^^^^^^^^^^
Use FlyteConsole to launch an execution and keep tabs on the window! 

.. image:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/flytesnacks/tutorial/getting_started_reg.gif
    :alt: A quick visual tour for launching a workflow and checking the outputs when they're done.

**Alternatively, you can execute using the command line.** 

Launch and monitor from CLI using Flytectl.
More details can be found `here <https://docs.flyte.org/projects/flytectl/en/stable/gen/flytectl_create_execution.html>`__.

#. Generate an execution spec file.

   .. prompt:: bash $

      flytectl get launchplan --project flytesnacks --domain development myapp.workflows.example.my_wf --latest --execFile exec_spec.yaml

#. Create an execution using the exec spec file.

   .. prompt:: bash $

      flytectl create execution --project flytesnacks --domain development --execFile exec_spec.yaml

#. Monitor the execution by providing the execution name from the ``create execution`` command.

   .. prompt:: bash $

      flytectl get execution --project flytesnacks --domain development <execname>


.. admonition:: Recap

  .. rubric:: 🎉  You have successfully packaged your workflow and tasks and pushed them to a Flyte cluster. Let's learn how to iterate.

.. toctree::
   :maxdepth: -1
   :caption: Getting Started
   :hidden:

   Iterate "fast"er<getting_started_iterate>
   User Guide <https://docs.flyte.org/projects/cookbook/en/latest/user_guide.html>