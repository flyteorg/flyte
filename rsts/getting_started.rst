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

.. note::
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
  - If the ``make start`` command isn't proceeding by any chance, check the pods' statuses -- run the command ``docker exec flyte-sandbox kubectl get po -A``.
  - If you think a pod's crashing by any chance, describe the pod by running the command ``docker exec flyte-sandbox kubectl describe po <pod-name> -n flyte``. This gives a detailed overview of the pod's status.
  - If Kubernetes reports a disk pressure issue:
  
    - Check the memory stats of the docker container using the command ``docker exec flyte-sandbox df -h``.
    - Prune the images and volumes. 
    - Given there's less than 10% free disk space, Kubernetes by default throws the disk pressure error.

1. Take a minute to explore Flyte Console through the provided URL.

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
