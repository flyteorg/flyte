.. _gettingstarted_scale:

Getting Started: Scale
-----------------------

.. raw:: html
  
    <p style="color: #808080; font-weight: 500; font-size: 20px; padding-top: 10px;">A step-by-step guide to building, deploying, and iterating on Flyte tasks and workflows</p>

.. panels::
    :body: text-justify
    :container: container-xs
    :column: col-lg-4 col-md-4 col-sm-6 col-xs-12 p-2

    ---
    .. link-button:: gettingstarted_implement
        :type: ref
        :text: ✔ Implement
        :classes: btn-outline-success btn-block stretched-link
    ---
    .. link-button:: gettingstarted_scale
            :type: ref
            :text: ➔ Scale
            :classes: btn-outline-primary btn-block stretched-link
    ---
    .. link-button:: gettingstarted_iterate
            :type: ref
            :text: 3. Iterate
            :classes: btn-outline-primary btn-block stretched-link


2. Scale: Get your workflows into the cloud
=============================================

.. _getting-started-build-deploy:

Install flytectl
^^^^^^^^^^^^^^^^^

#. install :std:ref:`flytectl`. ``Flytectl`` is a commandline interface for Flyte.

   .. tabs::

      .. tab:: OSX

        .. prompt:: bash $

           brew install flyteorg/homebrew-tap/flytectl

        *Upgrade* existing installation using the following command:

        .. prompt:: bash $

           brew upgrade flytectl

      .. tab:: Other Operating systems

        .. prompt:: bash $

            curl -s https://raw.githubusercontent.com/lyft/flytectl/master/install.sh | bash

   **Test** if Flytectl is installed correctly (your Flytectl version should be > 0.2.0) using the following command:

   .. prompt:: bash $

      flytectl version

#. Flyte can be deployed locally using a single Docker container — we refer to this as the ``flyte-sandbox`` environment. You can also run this getting started against a hosted or pre-provisioned environment. Refer to :ref:`deployment` section to learn how to deploy a flyte cluster.

   .. tabs::

      .. tab:: Start a new sandbox cluster

        .. tip:: Want to dive under the hood into flyte-sandbox, refer to :ref:`deployment-sandbox`.

        Here '.' represents current directory and assuming you have changed into ``myflyteapp`` - the git-cloned directory you created

        .. prompt:: bash $

           flytectl sandbox start --source .

        *NOTE*: Output of the command will contain a recommendation to export an environment variable called FLYTECTL_CONFIG. please export as follows or copy paste

        .. prompt:: bash $

           export FLYTECTL_CONFIG=$HOME/.flyte/config-sandbox.yaml

      .. tab:: Connect to an existing Flyte cluster

        .. prompt:: bash $

            # COMING SOON! flytectl init



Build & Deploy Your Application to the cluster
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
#. Flyte uses Docker containers to package the workflows and tasks and sends them to the remote Flyte cluster. Thus, there is a ``Dockerfile`` already included in the cloned repo. You can build the Docker container and push the built image to a registry.

   .. tabs::

       .. tab:: Flyte Sandbox

           Since ``flyte-sandbox`` runs locally in a Docker container, you do not need to push the Docker image. You can combine the build and push step by simply building the image inside the Flyte-sandbox container. This can be done using the following command:

           .. prompt:: bash $

               flytectl sandbox exec -- docker build . --tag "myapp:v1"

           .. tip::
            #. Why are we not pushing the Docker image? Want to understand the details — refer to :ref:`deployment-sandbox`
            #. *Recommended:* Use the bundled `./docker_build_and_tag.sh`. It will automatically build the local Dockerfile, name it and tag it with the current git-SHA. This helps in achieving GitOps style workflows.

       .. tab:: Remote Flyte Cluster

           If you are using a remote Flyte cluster, then you need to build your container and push it to a registry that is accessible by the Flyte Kubernetes cluster.

           .. prompt:: bash $

               docker build . --tag registry/repo:version
               docker push registry/repo:version

#. Next, package the workflow using the ``pyflyte`` cli bundled with Flytekit and upload it to the Flyte backend. Note that the image is the same as the one built in the previous step.

   .. prompt:: bash (venv)$

      pyflyte --pkgs myapp.workflows package --image myapp:v1

#. Upload this package to the Flyte backend. We refer to this as ``registration``.

   .. prompt:: bash $

      flytectl register files -p flytesnacks -d development --archive flyte-package.tgz  --version v1

#. Finally, visualize the registered workflow.

   .. prompt:: bash $

      flytectl get workflows -p flytesnacks -d development myapp.workflows.example.my_wf --version v1 -o doturl


.. _getting-started-execute:

Execute on Flyte Cluster
^^^^^^^^^^^^^^^^^^^^^^^^
Use FlyteConsole to launch an execution and keep tabs on the window! 

.. image:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/flytesnacks/tutorial/exercise.gif
    :alt: A quick visual tour for launching a workflow and checking the outputs when they're done.

**Alternatively, you can execute using the command line.** 

Launch and monitor from CLI using Flytectl.
More details can be found `here <https://docs.flyte.org/projects/flytectl/en/stable/gen/flytectl_create_execution.html>`__.

#. Generate an execution spec file.

   .. prompt:: bash $

      flytectl get launchplan -p flytesnacks -d development myapp.workflows.example.my_wf --execFile exec_spec.yaml

#. Create an execution using the exec spec file.

   .. prompt:: bash $

      flytectl create execution -p flytesnacks -d development --execFile exec_spec.yaml

#. Monitor the execution by providing the execution id from the ``create execution`` command.

   .. prompt:: bash $

      flytectl get execution -p flytesnacks -d development <execid>


.. admonition:: Recap

  .. rubric:: 🎉  You have successfully packaged your workflow and tasks and pushed them to a Flyte cluster. Let's learn how to iterate?

.. toctree::
   :maxdepth: -1
   :caption: Getting Started
   :hidden:

   Iterate "fast"er<gettingstarted_iterate>
   User Guide <https://docs.flyte.org/projects/cookbook/en/latest/user_guide.html>