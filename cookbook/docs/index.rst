.. _userguide:

##############
User Guide
##############

If this is your first time using Flyte, check out the `Getting Started <https://docs.flyte.org/en/latest/getting_started.html>`_ guide.

The :ref:`User Guide <userguide>` and :ref:`Tutorials <tutorials>` sections of the Flyte documentation covers all of the
key features of Flyte organized by topic. Each of the sections below introduces a topic and discusses how you can use
Flyte to address a specific problem. Code for all of the examples in the user guide be found in the
`flytesnacks repo <https://github.com/flyteorg/flytesnacks>`_.

`Flytesnacks <https://github.com/flyteorg/flytesnacks>`_ comes with a highly customized environment to make running,
documenting and contributing samples easy. If this is your first time running these examples, follow the setup guide
below to get started.

.. _setup_flytesnacks_env:

.. dropdown:: :fa:`info-circle` Setting up your environment to run the examples
   :animate: fade-in-slide-down

   **Prerequisites**

   * Make sure you have `docker <https://docs.docker.com/get-docker/>`_ and `git <https://git-scm.com/>`_ installed.
   * Install :doc:`flytectl <flytectl:index>`. ``flytectl`` is a commandline interface for flyte.

   .. tabbed:: OSX

       .. prompt:: bash

           brew install flyteorg/homebrew-tap/flytectl

       To upgrade, run:

       .. prompt:: bash

           brew upgrade flytectl

   .. tabbed:: Most other platforms

       .. prompt:: bash

           curl -s https://raw.githubusercontent.com/lyft/flytectl/master/install.sh | bash

   **Steps**

   #. Install the python ``Flytekit`` SDK and clone the ``flytesnacks`` repo:

      .. tip::
         It's recommended to create a new python virtual environment to make sure it doesn't interfere with your development
         environment.

      .. prompt:: bash

         git clone --depth 1 git@github.com:flyteorg/flytesnacks.git flytesnacks
         cd flytesnacks
         pip install -r cookbook/core/requirements.txt

   #. Run ``hello_world.py`` locally

      .. prompt:: bash

         python cookbook/core/flyte_basics/hello_world.py

      .. raw:: html

          <details>
          <summary><a>Expected Output</a></summary>

      .. prompt::

         Running my_wf() hello world

      .. raw:: html

          </details>

      .. admonition:: 🎉 **Congratulations** 🎉

         You have just run your first workflow. Now, let's run it on the `sandbox cluster deployed earlier <https://docs.flyte.org/en/latest/getting_started.html>`_.

   #. We've packaged all the required components to run a sandboxed flyte cluster into a single docker image. You can start
      one by running:

      .. prompt:: bash

         flytectl sandbox start --source ${PWD}

      .. tip::
         In case make start throws any error please refer to the troubleshooting guide here `Troubleshoot <https://docs.flyte.org/en/latest/community/troubleshoot.html>`_

      Check status:

      .. prompt:: bash

         flytectl sandbox status

      Teardown:

      .. prompt:: bash

         flytectl sandbox teardown

   #. Take a minute to explore Flyte Console through the provided URL.

      .. figure:: https://github.com/flyteorg/flyte/raw/static-resources/img/first-run-console-2.gif
         :alt: A quick visual tour for launching your first Workflow.

         A quick visual tour for launching your first Workflow.

   #. Register all examples from cookbook/core into the cluster. This step compiles your python code into the intermediate
      flyteIdl language and store them on the control plane running inside the cluster.

      .. prompt:: bash

         REGISTRY=cr.flyte.org/flyteorg make fast_register

      .. note::
         If the images are to be re-built, run ``make register`` command.

   #. Let's launch our first execution from the UI. Visit `the console <http://localhost:30081/console/projects/flytesnacks/domains/development/workflows/core.flyte_basics.hello_world.my_wf>`_, click launch.

   #. Give it a minute and once it's done, check out "Inputs/Outputs" on the top right corner to see your greeting.

      .. figure:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/flytesnacks/tutorial/exercise.gif
         :alt: A quick visual tour for launching a workflow and checking the outputs when they're done.

         A quick visual tour for launching a workflow and checking the outputs when they're done.

   .. admonition:: Recap

     You have successfully:

     #. Run a flyte workflow locally,
     #. Run a flyte sandbox cluster,
     #. Run a flyte workflow on a cluster.

     .. rubric:: 🎉 Congratulations, now you can interactively explore Flyte's features outlined in the :ref:`Table of Contents` 🎉


******************
Table of Contents
******************

.. panels::
   :header: text-center

   .. link-button:: auto/core/flyte_basics/index
      :type: ref
      :text: 🔤 Flyte Basics
      :classes: btn-block stretched-link
   ^^^^^^^^^^^^
   Learn about tasks, workflows, launch plans, caching, and working with files and directories.

   ---

   .. link-button:: auto/core/control_flow/index
      :type: ref
      :text: 🚰 Control Flow
      :classes: btn-block stretched-link
   ^^^^^^^^^^^^
   Implement conditionals, nested and dynamic workflows, map tasks, and even recursion!

   ---

   .. link-button:: auto/core/type_system/index
      :type: ref
      :text: ⌨️ Type System
      :classes: btn-block stretched-link
   ^^^^^^^
   Improve pipeline robustness with Flyte's portable and extensible type system.

   ---

   .. link-button:: auto/core/scheduled_workflows/index
      :type: ref
      :text: ⏱ Scheduled Workflows
      :classes: btn-block stretched-link
   ^^^^^^^^^^^^
   Learn about scheduled workflows.

   ---

   .. link-button:: auto/testing/index
      :type: ref
      :text: ⚗️ Testing
      :classes: btn-block stretched-link
   ^^^^^^^
   Test tasks and workflows with Flyte's testing utilities.

   ---

   .. link-button:: auto/core/containerization/index
      :type: ref
      :text: 📦  Containerization
      :classes: btn-block stretched-link
   ^^^^^^^^^^^^^^^^
   Easily manage the complexity of configuring the containers that run Flyte tasks.

   ---

   .. link-button:: auto/deployment/index
      :type: ref
      :text: 🚢  Configuring Production Features
      :classes: btn-block stretched-link
   ^^^^^^^^^^
   Ship and configure your machine learning pipelines on a production Flyte installation.

   ---

   .. link-button:: auto/remote_access/index
      :type: ref
      :text: 🎮 Remote Access
      :classes: btn-block stretched-link
   ^^^^^^^^^^
   Register, inspect, and monitor tasks and workflows on a Flyte backend.

   ---

   .. link-button:: integrations
      :type: ref
      :text: 🔌  Integrations
      :classes: btn-block stretched-link
   ^^^^^^^^^^^^
   Leverage a rich ecosystem of plugins from compute infrastructure to jupyter notebooks.

   ---

   .. link-button:: auto/core/extend_flyte/index
      :type: ref
      :text: 🏗 Extending Flyte
      :classes: btn-block stretched-link
   ^^^^^^^^^^^^^^^
   Define custom plugins that aren't currently supported in the Flyte ecosystem.

.. toctree::
   :maxdepth: 1
   :hidden:

   |plane| Getting Started <https://docs.flyte.org/en/latest/getting_started.html>
   |book-reader| User Guide <index>
   |chalkboard| Tutorials <tutorials>
   |project-diagram| Concepts <https://docs.flyte.org/en/latest/concepts/basics.html>
   |rocket| Deployment <https://docs.flyte.org/en/latest/deployment/index.html>
   |book| API Reference <https://docs.flyte.org/en/latest/reference/index.html>
   |hands-helping| Community <https://docs.flyte.org/en/latest/community/index.html>

.. toctree::
   :maxdepth: -1
   :caption: User Guide
   :hidden:

   Introduction <self>
   Basics <auto/core/flyte_basics/index>
   Control Flow <auto/core/control_flow/index>
   Type System <auto/core/type_system/index>
   Testing <auto/testing/index>
   Containerization <auto/core/containerization/index>
   Remote Access <auto/remote_access/index>
   Configuring Production Features <auto/deployment/index>
   Scheduling Workflows <auto/core/scheduled_workflows/index>
   integrations
   Extending flyte <auto/core/extend_flyte/index>
   contribute

.. toctree::
   :maxdepth: -1
   :caption: Tutorials
   :hidden:

   Introduction <tutorials>
   ml_training
   feature_engineering
