.. _getting-started:

###############
Getting Started
###############

This quick start gives an overview of how to get Flyte up and running on your local machine.

Requirements
^^^^^^^^^^^^

- Install `Docker <https://docs.docker.com/get-docker/>`__
- Ensure Docker Daemon is running

Installation
^^^^^^^^^^^^

Install `Flytekit <https://pypi.org/project/flytekit/>`__, Flyte's Python SDK.

.. prompt:: bash

  pip install flytekit


.. dropdown:: :fa:`info-circle` Please read on if you're running python 3.10 on a Mac M1
    :title: text-muted
    :animate: fade-in-slide-down

    Before proceeding, install ``grpcio`` by building it locally via:

    .. prompt:: bash

        pip install --no-binary :all: grpcio --ignore-installed

    Visit https://github.com/flyteorg/flyte/issues/2486 for more details.


Example: Computing Descriptive Statistics
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Let's create a simple Flyte :py:func:`~flytekit.workflow` that involves two steps:

1. Generate a dataset of ``numbers`` drawn from a normal distribution.
2. Compute the mean and standard deviation of the ``numbers`` data.

Create a Workflow
""""""""""""""""""

Copy the following code to a file named ``example.py``.

.. code-block:: python

    import typing
    import pandas as pd
    import numpy as np

    from flytekit import task, workflow

    @task
    def generate_normal_df(n:int, mean: float, sigma: float) -> pd.DataFrame:
        return pd.DataFrame({"numbers": np.random.normal(mean, sigma,size=n)})

    @task
    def compute_stats(df: pd.DataFrame) -> typing.Tuple[float, float]:
        return float(df["numbers"].mean()), float(df["numbers"].std())

    @workflow
    def wf(n: int = 200, mean: float = 0.0, sigma: float = 1.0) -> typing.Tuple[float, float]:
        return compute_stats(df=generate_normal_df(n=n, mean=mean, sigma=sigma))


.. dropdown:: :fa:`info-circle` This looks like Python, but what do ``@task`` and ``@workflow`` do?
    :title: text-muted
    :animate: fade-in-slide-down

    Flyte ``@task`` and ``@workflow`` decorators are designed to work seamlessly with your code-base, provided
    that the *decorated function is at the top-level scope of the module*.

    This means that you can invoke tasks and workflows as regular Python methods and even import and use them in
    other Python modules or scripts.

    .. note::

       A :py:func:`~flytekit.task` is a pure Python function, while a :py:func:`~flytekit.workflow` is actually a
       `DSL <https://en.wikipedia.org/wiki/Domain-specific_language>`__ that only support a subset of Python's semantics.
       Some key things to learn here are:

       - In workflows, you can't use non-deterministic operations like ``rand.random``, ``time.now()``, etc.
       - Within workflows, the outputs of tasks are promises under the hood, so you can't access and operate on them
         like typical Python function outputs. *You can only pass them into other tasks/workflows.*
       - Tasks can only be invoked with keyword arguments, not positional arguments.

       You can read more about tasks :doc:`here <cookbook:auto/core/flyte_basics/task>` and workflows
       :doc:`here <cookbook:auto/core/flyte_basics/basic_workflow>`.


Running Flyte Workflows
^^^^^^^^^^^^^^^^^^^^^^^

You can run the workflow in ``example.py`` on a local Python environment or a Flyte cluster.

Executing Workflows Locally
""""""""""""""""""""""""""""

Run your workflow locally using ``pyflyte``, the CLI that ships with ``flytekit``.

.. prompt:: bash $

  pyflyte run example.py wf --n 500 --mean 42 --sigma 2

.. dropdown:: :fa:`info-circle` Why use ``pyflyte run`` rather than ``python example.py``?
    :title: text-muted
    :animate: fade-in-slide-down

    ``pyflyte run`` enables you to execute a specific workflow in your Python script using the syntax
    ``pyflyte run <path/to/script.py> <workflow_function_name>``.

    Keyword arguments can be supplied to ``pyflyte run`` by passing in options in the format ``--kwarg value``, and in
    the case of ``snake_case_arg`` argument names, you can pass in options in the form of ``--snake-case-arg value``.

    .. note::
       If you want to run a workflow with ``python example.py``, you would have to write a ``main`` module
       conditional at the end of the script to actually run the workflow:

       .. code-block:: python

          if __name__ == "__main__":
              wf(n=100, mean=1.0, sigma=2.0)

       This becomes even more verbose if you want to pass in arguments:

       .. code-block:: python

          if __name__ == "__main__":
              from argparse import ArgumentParser

              parser = ArgumentParser()
              parser.add_argument("--n", type=int)
              ...  # add the other options

              args = parser.parse_args()
              wf(n=args.n, mean=args.mean, sigma=args.sigma)

Creating a Demo Flyte Cluster
"""""""""""""""""""""""""""""""

To start a local demo cluster, install :std:ref:`flytectl`, which is the command-line interface for Flyte.

.. tabbed:: OSX

  .. prompt:: bash $

    brew install flyteorg/homebrew-tap/flytectl

.. tabbed:: Other Operating systems

  .. prompt:: bash $

    curl -L https://raw.githubusercontent.com/flyteorg/flytectl/HEAD/install.sh | bash


Start a Flyte demonstration environment on your local machine via:

.. prompt:: bash $

  flytectl demo start

.. div:: shadow p-3 mb-8 rounded

   **Expected Output:**

   .. code-block::

      👨‍💻 Flyte is ready! Flyte UI is available at http://localhost:30080/console 🚀 🚀 🎉

.. note::

   Make sure to export the ``KUBECONFIG`` and ``FLYTECTL_CONFIG`` environment variables in your shell, replacing
   ``<username>`` with your actual username.

.. dropdown:: :fa:`info-circle` What is a flyte demo environment?
    :title: text-muted
    :animate: fade-in-slide-down

    ``flytectl`` ships with a limited testing environment that can run on your local machine. It's not a substitute for the production environment,
    but it's great for trying out the platform and checking out some of its capabilities.

    However, most :doc:`integrations <cookbook:integrations>` are not directly installed in this environment, and it's not a great
    way to test the platform's performance.

Executing Workflows on a Flyte Cluster
"""""""""""""""""""""""""""""""""""""""

Run the workflow on Flyte cluster via:

.. prompt:: bash $

  pyflyte run --remote example.py wf --n 500 --mean 42 --sigma 2

.. div:: shadow p-3 mb-8 rounded

   **Expected Output:** A URL to the workflow execution on your demo Flyte cluster:

   .. code-block::

      Go to http://localhost:30080/console/projects/flytesnacks/domains/development/executions/<execution_name> to see execution in the console.

   where ``<execution_name>`` is a unique identifier for the workflow execution.

Unlike the previous ``pyflyte run`` invocation, passing the ``--remote`` flag will trigger the execution on the configured backend.

.. dropdown:: :fa:`info-circle` How to handle custom dependencies? Meet the ``--image`` flag!
    :title: text-muted
    :animate: fade-in-slide-down

    * Consistent dependency management is a challenge with python projects, so Flyte uses `Docker containers <https://www.docker.com/resources/what-container/>`__ to manage dependencies for your project.
    * ``pyflyte run --remote`` uses a default image bundled with flytekit, which contains numpy, pandas, and flytekit and matches your current python (major, minor) version.
    * If you want to use a custom image, create a Dockerfile, build the Docker image, and push it to a registry that is accessible to your cluster.

      .. prompt :: bash $

        docker build . --tag <registry/repo:version>
        docker push <registry/repo:version>

    * And, use the ``--image`` flag and provide the fully qualified image name of your image to the ``pyflyte run`` command.

      .. prompt :: bash $

        pyflyte run --image <registry/repo:version> --remote example.py wf --n 500 --mean 42 --sigma 2

    * If you want to build an image with your Flyte project's code built-in, refer to the :doc:`Deploying Workflows Guide <cookbook:auto/deployment/deploying_workflows>`.


Inspect the Results
^^^^^^^^^^^^^^^^^^^
Navigate to the URL produced as the result of running ``pyflyte run``. This will take you to FlyteConsole, the web UI
used to manage Flyte entities such as tasks, workflows, and executions.

.. image:: https://github.com/flyteorg/static-resources/raw/main/flyte/getting_started/getting_started_console.gif

.. note::

   There are a few features about FlyteConsole worth noting in this video:

   - The default execution view shows the list of tasks executing in sequential order.
   - The right-hand panel shows metadata about the task execution, including logs, inputs, outputs, and task metadata.
   - The *Graph* view shows the execution graph of the workflow, providing visual information about the topology of the graph and the state of each node as the workflow progresses.
   - On completion, you can inspect the outputs of each task, and ultimately, the overarching workflow.

Recap
^^^^^
🎉  **Congratulations!  In this getting started guide, you:**

1. 📜 Created a Flyte script, which computes descriptive statistics over some generated data.
2. 🛥 Created a demo Flyte cluster on your local system.
3. 👟 Ran a workflow locally and on a demo Flyte cluster.

What's Next?
^^^^^^^^^^^^
This guide demonstrated how you can quickly iterate on self-contained scripts using ``pyflyte run``.

- To learn about Flyte's features such as caching, conditionals, specifying resource requirements, and scheduling
  workflows, take a look at the `User Guide <https://docs.flyte.org/projects/cookbook/en/latest/user_guide.html>`__.
- To learn about how to organize, package, and register workflows for larger projects, see the guide for
  :ref:`Building Large Apps <cookbook:larger_apps>`.
