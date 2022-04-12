.. _getting-started:

################
Getting Started
################

Requirements
^^^^^^^^^^^^^
Make sure you have `Docker <https://docs.docker.com/get-docker/>`__ and Docker Daemon is running.

Installation
^^^^^^^^^^^^

Install Flyte's python SDK — `Flytekit <https://pypi.org/project/flytekit/>`__.

.. prompt:: bash

  pip install flytekit


Example
^^^^^^^^

Create It
""""""""""

Copy the following code to a file named ``example.py``

.. code-block:: python

    import typing
    import pandas as pd
    import numpy as np

    from flytekit import task, workflow


    @task
    def generate_normal_df(n:int, mean: float, sigma: float) -> pd.DataFrame:
        return pd.DataFrame(dict(
          numbers=np.random.normal(mean, sigma,size=n),
        ))

    @task
    def compute_stats(df: pd.DataFrame) -> typing.Tuple[float, float]:
        return df["numbers"].mean(), df["numbers"].std()

    @workflow
    def wf(n: int = 200, mean: float = 0.0, sigma: float = 1.0) -> typing.Tuple[float, float]:
        return compute_stats(df=generate_normal_df(n=n, mean=mean, sigma=sigma))


Run It
^^^^^^^
You can either execute the code in python environment or on a remote cluster. We will show how to run on a demo local cluster.

Locally
""""""""

Run your workflow locally using ``pyflyte``

.. prompt:: bash $

  pyflyte run example.py:wf --n 500 --mean 42 --sigma 2

*This uses the default image bundled with flytekit, which contains numpy, pandas, flytekit, you can optionally pass a different image*

.. tip:: This is just Python code also, you can also simply execute the workflow as a python function -> ``wf()``.

On a Flyte cluster
"""""""""""""""""""

Install :std:ref:`flytectl`. ``Flytectl`` is a commandline interface for Flyte. This is useful to start a local demo cluster of Flyte.

.. tabbed:: OSX

  .. prompt:: bash $

    brew install flyteorg/homebrew-tap/flytectl

.. tabbed:: Other Operating systems

  .. prompt:: bash $

    curl -sL https://ctl.flyte.org/install | sudo bash -s -- -b /usr/local/bin # You can change path from /usr/local/bin to any file system path
    export PATH=$(pwd)/bin:$PATH # Only required if user used different path then /usr/local/bin


* Start a Flyte demonstration environment on your local machine

.. prompt:: bash $

  flytectl demo start

* Now run the same workflow on the Flyte backend

.. prompt:: bash $

  pyflyte run --remote example.py:wf --n 500 --mean 42 --sigma 2

.. note:: The only difference between previous ``local`` and this command is the ``--remote`` flag. This will trigger an execute on the configured backend


Check It
^^^^^^^^^
Navigate to the url produced as the result of running ``pyflyte``, this should take you to Flyte Console, the web UI used to manage Flyte entities.
