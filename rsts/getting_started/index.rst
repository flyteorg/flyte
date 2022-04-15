.. _getting-started:

################
Getting Started
################

Requirements
^^^^^^^^^^^^^
Make sure you have `Docker <https://docs.docker.com/get-docker/>`__ and Docker Daemon is running.

Installation
^^^^^^^^^^^^

Install `Flytekit <https://pypi.org/project/flytekit/>`__, Flyte's python SDK.

.. prompt:: bash

  pip install flytekit


Example Script: Computing Descriptive Statistics
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Let's create a simple Flyte workflow that involves two steps:

1. Generate a dataset of ``numbers`` drawn from a normal distribution.
2. Compute the mean and standard deviation of the ``numbers`` data.

Create Workflow
""""""""""""""""

Copy the following code to a file named ``example.py``

.. code-block:: python

    import typing
    import pandas as pd
    import numpy as np

    from flytekit import task, workflow

    @task
    def generate_normal_df(n:int, mean: float, sigma: float) -> pd.DataFrame:
        return pd.DataFrame({
            "numbers": np.random.normal(mean, sigma,size=n),
        })

    @task
    def compute_stats(df: pd.DataFrame) -> typing.Tuple[float, float]:
        return float(df["numbers"].mean()), float(df["numbers"].std())

    @workflow
    def wf(n: int = 200, mean: float = 0.0, sigma: float = 1.0) -> typing.Tuple[float, float]:
        return compute_stats(df=generate_normal_df(n=n, mean=mean, sigma=sigma))


Running Flyte Workflows
^^^^^^^^^^^^^^^^^^^^^^^
You can run the workflow in ``example.py`` on a local python environment or a Flyte cluster. In this guide, you will learn how to run a local demo cluster.

Executing Locally
"""""""""""""""""""

Run your workflow locally using ``pyflyte``, the CLI that ships with ``flytekit``.

.. prompt:: bash $

  pyflyte run example.py:wf --n 500 --mean 42 --sigma 2

.. dropdown:: :fa:`info-circle` Why use ``pyflyte run`` rather than ``python example.py``?
    :animate: fade-in-slide-down

    Flyte ``@task`` and ``@workflow`` decorators are designed to continue to work with your code-base with a restriction that they have to be the outermost decorators. 
    Thus you can invoke your ``tasks`` and ``workflows`` as regular python methods.

    Also, note that Tasks are pure python functions, while Workflows are a DSL and only look like a python function.
    Do not use non-deterministic operations in a workflow - like ``rand.random`` or ``time.now()`` etc. Also, in a workflow,
    you cannot simply access the outputs of tasks and operate on them; you can only pass them to other tasks.

Executing on a Flyte Cluster
"""""""""""""""""""""""""""""""

Install :std:ref:`flytectl`, which is the command-line interface for Flyte. Let's use it to start a local demo Flyte cluster.

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


.. dropdown:: :fa:`info-circle` What is a flyte demo environment?
    :animate: fade-in-slide-down

    Flyte demo environment is a fully included testing environment that can run on your local machine. It is not a substitute for the production environment,
    but it is great to try out the platform and check out some capabilities. Most plugins are not directly installed in this environment, and it is not
    a great way to test the platform's performance.

Then run the same workflow on the Flyte cluster:

.. prompt:: bash $

  pyflyte run --remote example.py:wf --n 500 --mean 42 --sigma 2

.. dropdown:: :fa:`info-circle` What does the ``--remote`` flag do?
    :animate: fade-in-slide-down

   * The only difference between the previous ``local`` and this command is the ``--remote`` flag. This will trigger the execution on the configured backend.
   * Dependency management is a challenge with python projects. Flyte uses containers to manage dependencies for your project.
   * ``pyflyte run --remote`` uses a default image bundled with flytekit, which contains numpy, pandas, and flytekit and matches your current python (major, minor) version.
   * If you want to use a custom image, use the ``--image`` flag.
   * Also, it is possible to use an image with your completely built-in code.  Refer to package & register flow.

Inspect the Results
^^^^^^^^^^^^^^^^^^^^^^
Navigate to the URL produced as the result of running ``pyflyte``. This should take you to Flyte Console; the web UI used to manage Flyte entities such as tasks, workflows, and executions.

Recap
^^^^^^^^

🎉  Congratulations 🎉  To summarize, you have:

1. Created a Flyte script called ``example.py``, which creates some data and computes descriptive statistics over it.
2. Run a workflow (i) locally and (ii) on a demo Flyte cluster.

What's Next?
^^^^^^^^^^^^^^^^

To experience the full power of Flyte, take a look at the `User Guide <https://docs.flyte.org/projects/cookbook/en/latest/user_guide.html>`__.
