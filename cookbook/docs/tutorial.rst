.. _flyte-tutorial:
.. currentmodule:: flyte_tutorial

############################################
Hello World Example
############################################

.. rubric:: Estimated time to complete: 3 minutes.

The goal of this tutorial is to modify an existing example (hello world) to take the name of the person to green as an argument.

Prerequisites
*************

* Ensure that you have `git <https://git-scm.com/>`__ installed.

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

3. Take a minute to explore Flyte Console through the provided URL.

.. image:: https://github.com/flyteorg/flyte/raw/static-resources/img/first-run-console-2.gif
    :alt: A quick visual tour for launching your first Workflow.

4. Open ``hello_world.py`` in your favorite editor.

.. code-block::

  cookbook/core/basic/hello_world.py

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

  python cookbook/core/basic/hello_world.py

*Congratulations!* You have just run your first workflow. Now, let's run it on the sandbox cluster deployed earlier.

8. Run:

.. prompt:: bash

  REGISTRY=ghcr.io/flyteorg make fast_register

9. Visit `the console <http://localhost:30081/console/projects/flytesnacks/domains/development/workflows/core.basic.hello_world.my_wf>`__, click launch, and enter your name as the input.

10. Give it a minute and once it's done, check out "Inputs/Outputs" on the top right corner to see your updated greeting.

.. image:: https://raw.githubusercontent.com/flyteorg/flyte/static-resources/img/flytesnacks/tutorial/exercise.gif
    :alt: A quick visual tour for launching a workflow and checking the outputs when they're done.

.. admonition:: Recap

  You have successfully:

  1. Run a flyte sandbox cluster,
  2. Run a flyte workflow locally,
  3. Run a flyte workflow on a cluster.

Head over to the next section :ref:`flyte-core` to learn more about FlyteKit and how to start leveraging all the functionality flyte has to offer in simple and idiomatic python.
