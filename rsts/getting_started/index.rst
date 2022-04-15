.. _getting-started:

################
Getting Started
################

Requirements
^^^^^^^^^^^^^
Make sure you have `Docker <https://docs.docker.com/get-docker/>`__ and the Docker Daemon is running.

Installation
^^^^^^^^^^^^

Install `Flytekit <https://pypi.org/project/flytekit/>`__, Flyte's python SDK.

.. prompt:: bash

  pip install flytekit


Example: Computing Descriptive Statistics
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Let's create a simple Flyte Workflow that involves two steps:

1. Generate a dataset of ``numbers`` drawn from a normal distribution.
2. Compute the mean and standard deviation of the ``numbers`` data.

Create a Workflow
""""""""""""""""""

Copy the following code to a file named ``example.py``

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


.. dropdown:: :fa:`info-circle` This looks like Python, but what does ``@task`` and ``@workflow`` do?
    :title: text-muted
    :animate: fade-in-slide-down

    Flyte ``@task`` and ``@workflow`` decorators are designed to work seamlessly with your code-base, provided
    that the *decorated function is at the top-level scope of the module*.
    
    This means that you can invoke your Tasks and Workflows as regular python methods and even import and use them in
    other Python modules or scripts.

    .. note::

       A Task is a pure Python function, while a Workflow is actually a `DSL <https://en.wikipedia.org/wiki/Domain-specific_language>`__
       that only support a subset of Python's semantics. Some key things to learn here are:

       - In Workflows, you can't use non-deterministic operations like ``rand.random`` or ``time.now()`` etc.
       - Within Workflows, the outputs of tasks are promises under the hood, so you can't access and operate on them
         like typical Python function outputs. *You can only pass them into other Tasks/Workflows.*
       - Tasks can only be invoked with keyword arguments, not positional arguments.

       You can read more about Tasks :doc:`here <cookbook:auto/core/flyte_basics/task>` and Workflows
       :doc:`here <cookbook:auto/core/flyte_basics/basic_workflow>`


Running Flyte Workflows
^^^^^^^^^^^^^^^^^^^^^^^

You can run the workflow in ``example.py`` on a local python environment or a Flyte cluster.

Executing Workflows Locally
""""""""""""""""""""""""""""

Run your workflow locally using ``pyflyte``, the CLI that ships with ``flytekit``.

.. prompt:: bash $

  pyflyte run example.py:wf --n 500 --mean 42 --sigma 2

.. dropdown:: :fa:`info-circle` Why use ``pyflyte run`` rather than ``python example.py``?
    :title: text-muted
    :animate: fade-in-slide-down

    ``pyflyte run`` enables you to execute a specific workflow in your python script using the syntax
    ``pyflyte run <path/to/script.py>:<workflow_function_name>``.

    Key-word arguments can be supplied to ``pyflyte run`` by passing in options in the format ``--kwarg value``, and in
    the case of ``snake_case_arg`` argument names, you can pass in options in the form of ``--snake-case-arg value``.

    .. note::
       If you wanted to run a workflow with ``python example.py``, you would have to write a ``main`` module
       conditional at the end of the script to actually run the workflow:

       .. code-block:: python

          if __name__ == "__main__":
              wf(n=100, mean=1, sigma=2.0)

       This becomes even more verbose if you want to pass in arguments:

       .. code-block:: python

          if __name__ == "__main__":
              from argpase import ArgumentParser

              parser = ArgumentParser()
              parser.add_argument("--n", type=int)
              ...  # add the other options

              args = parser.parse_args()
              wf(n=args.n, mean=args.mean, sigma=args.sigma)

Creating a Demo Flyte Cluster
"""""""""""""""""""""""""""""""

To start a local demo cluster, first install :std:ref:`flytectl`, which is the command-line interface for Flyte.

.. tabbed:: OSX

  .. prompt:: bash $

    brew install flyteorg/homebrew-tap/flytectl

.. tabbed:: Other Operating systems

  .. prompt:: bash $

    curl -sL https://ctl.flyte.org/install | sudo bash -s -- -b /usr/local/bin # You can change path from /usr/local/bin to any file system path
    export PATH=$(pwd)/bin:$PATH # Only required if user used different path then /usr/local/bin


Start a Flyte demonstration environment on your local machine:

.. prompt:: bash $

  flytectl demo start

.. div:: shadow p-3 mb-8 bg-white rounded

   **Expected Output:**

   .. code-block::
   
      👨‍💻 Flyte is ready! Flyte UI is available at http://localhost:30080/console 🚀 🚀 🎉
      Add KUBECONFIG and FLYTECTL_CONFIG to your environment variable
      export KUBECONFIG=$KUBECONFIG:/Users/<username>/.kube/config:/Users/<username>/.flyte/k3s/k3s.yaml
      export FLYTECTL_CONFIG=/Users/<username>/.flyte/config-sandbox.yaml

.. note::

   Make sure to export the ``KUBECONFIG`` and ``FLYTECTL_CONFIG`` environment variables in your shell, replacing
   ``<username>`` with your actual username.


.. dropdown:: :fa:`info-circle` What is a flyte demo environment?
    :title: text-muted
    :animate: fade-in-slide-down

    ``flytectl`` ships with a limited testing environment that can run on your local machine. It's not a substitute for the production environment,
    but it's great for trying out the platform and checking out some of its capabilities.

    Most :doc:`integrations <cookbook:integrations>` are not directly installed in this environment, and it's not a great
    way to test the platform's performance.


Executing Workflows on a Flyte Cluster
"""""""""""""""""""""""""""""""""""""""

Then run the same Workflow on the Flyte cluster:

.. prompt:: bash $

  pyflyte run --remote example.py:wf --n 500 --mean 42 --sigma 2

.. div:: shadow p-3 mb-8 bg-white rounded

   **Expected Output:** A URL to the Workflow Execution on your demo Flyte cluster:

   .. code-block::

      Go to http://localhost:30080/console/projects/flytesnacks/domains/development/executions/<execution_name> to see execution in the console.

   Where ``<execution_name>`` is a unique identifier for your Workflow Execution.

|

.. dropdown:: :fa:`info-circle` What does the ``--remote`` flag do?
    :title: text-muted
    :animate: fade-in-slide-down

    Unlike the previous ``pyflyte run`` invocation, passing the ``--remote`` flag will trigger the execution on the configured backend.

    .. note::

       * Consistent dependency management is a challenge with python projects, so Flyte uses `docker containers <https://www.docker.com/resources/what-container/>`__ to manage dependencies for your project.
       * ``pyflyte run --remote`` uses a default image bundled with flytekit, which contains numpy, pandas, and flytekit and matches your current python (major, minor) version.
       * If you want to use a custom image, use the ``--image`` flag and provide the fully qualified image name of your image.
       * If you want to build an image with your Flyte project's code built-in, refer to the :doc:`Deploying Workflows Guide <cookbook:auto/deployment/deploying_workflows>`.

Inspect the Results
^^^^^^^^^^^^^^^^^^^^^^
Navigate to the URL produced as the result of running ``pyflyte run``. This will take you to Flyte Console, the web UI
used to manage Flyte entities such as tasks, workflows, and executions.

.. video:: ../_static/videos/getting_started_console.mp4
   :width: 100%
   :autoplay:

.. note::

   There are a few features about the Flyte console worth noting in this video:

   - The default execution view shows the list of Tasks executing in sequential order
   - The right-hand panel shows metadata about the Task Execution, including logs, inputs, outputs, and Task Metadata.
   - The *Graph* view shows the execution graph of the Workflow, providing visual information about the topology
     of the graph and the state of each node as the Workflow progresses.
   - On completion, you can inspect the outputs of each Task, and ultimately, the overarching Workflow.

Recap
^^^^^^^^

🎉  **Congratulations!  In this getting started guide, you:**

1. 📜 Created a Flyte script, which computes descriptive statistics over some generated data.
2. 🛥 Created a demo Flyte cluster on your local system
3. 👟 Ran a workflow locally and on a demo Flyte cluster.

What's Next?
^^^^^^^^^^^^^^^^

To experience the full power of Flyte, take a look at the `User Guide <https://docs.flyte.org/projects/cookbook/en/latest/user_guide.html>`__.
