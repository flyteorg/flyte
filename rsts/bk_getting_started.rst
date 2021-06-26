.. _gettingstarted:

Getting started
---------------

.. rubric:: Estimated time to complete: 3 minutes.

Prerequisites
***************

Make sure you have `docker installed <https://docs.docker.com/get-docker/>`__ and `git <https://git-scm.com/>`__ installed, then install flytekit:

Steps
*****

1. First install the python Flytekit SDK and clone the ``flytesnacks`` repo:

.. prompt:: bash

  pip install --pre flytekit
  git clone git@github.com:flyteorg/flytesnacks.git flytesnacks
  cd flytesnacks

2. The repo comes with some useful Make targets to make your experimentation workflow easier. Run ``make help`` to get the supported commands.
   Let's start a sandbox cluster:

.. prompt:: bash

  make start

.. tip::
  In case make start throws any error please refer to the troubleshooting guide here `Troubleshoot <https://docs.flyte.org/en/latest/community/troubleshoot.html>`__
    
3. Take a minute to explore Flyte Console through the provided URL.

.. image:: https://github.com/flyteorg/flyte/raw/static-resources/img/first-run-console-2.gif
    :alt: A quick visual tour for launching your first Workflow.

4. Open ``hello_world.py`` in your favorite editor.

.. code-block::

  cookbook/core/flyte_basics/hello_world.py

5. Add ``name: str`` as an argument to both ``my_wf`` and ``say_hello`` functions. Then update the body of ``say_hello`` to consume that argument.

.. tip::

  .. code-block:: python

    @task
    def say_hello(name: str) -> str:
        return f"hello world, {name}"

.. tip::

  .. code-block:: python

    @workflow
    def my_wf(name: str) -> str:
        res = say_hello(name=name)
        return res

6. Update the simple test at the bottom of the file to pass in a name. E.g.

.. tip::

  .. code-block:: python

    print(f"Running my_wf(name='adam') {my_wf(name='adam')}")

7. When you run this file locally, it should output ``hello world, adam``. Run this command in your terminal:

.. prompt:: bash

  python cookbook/core/flyte_basics/hello_world.py

*Congratulations!* You have just run your first workflow. Now, let's run it on the sandbox cluster deployed earlier.

8. Run:

.. prompt:: bash

  REGISTRY=cr.flyte.org/flyteorg make fast_register

.. note::
   If the images are to be re-built, run ``make register`` command.

9. Visit `the console <http://localhost:30081/console/projects/flytesnacks/domains/development/workflows/core.basic.hello_world.my_wf>`__, click launch, and enter your name as the input.

10. Give it a minute and once it's done, check out "Inputs/Outputs" on the top right corner to see your updated greeting.

.. image:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/flytesnacks/tutorial/exercise.gif
    :alt: A quick visual tour for launching a workflow and checking the outputs when they're done.

.. admonition:: Recap

  You have successfully:

  1. Run a flyte sandbox cluster,
  2. Run a flyte workflow locally,
  3. Run a flyte workflow on a cluster.

  .. rubric:: 🎉 Congratulations, you just ran your first Flyte workflow 🎉

Next Steps: User Guide
#######################

To experience the full capabilities of Flyte, take a look at the `User Guide <https://docs.flyte.org/projects/cookbook/en/latest/user_guide.html>`__ 🛫
