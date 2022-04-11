.. _getting-started:

################
Getting Started
################

Prerequisites
^^^^^^^^^^^^^^^^
Make sure you have `Docker <https://docs.docker.com/get-docker/>`__ and Docker Daemon is running.

Install flytectl
""""""""""""""""

Install :std:ref:`flytectl`. ``Flytectl`` is a commandline interface for Flyte.

   .. tabbed:: OSX

     .. prompt:: bash $

        brew install flyteorg/homebrew-tap/flytectl

   .. tabbed:: Other Operating systems

     .. prompt:: bash $

         curl -sL https://ctl.flyte.org/install | sudo bash -s -- -b /usr/local/bin # You can change path from /usr/local/bin to any file system path
         export PATH=$(pwd)/bin:$PATH # Only required if user used different path then /usr/local/bin

Install flytekit
""""""""""""""""

Install Flyte's python SDK — `Flytekit <https://pypi.org/project/flytekit/>`__.

.. prompt:: bash

   pip install flytekit


Write your first workflow
^^^^^^^^^^^^^^^^^^^^^^^^^

#. Copy the following code to a file named ``example.py``

    .. code-block:: python

       from flytekit import task, workflow

       @task
       def say_hello(msg: str) -> str:
           return f"hello, {msg}!"

       @workflow
       def wf(msg: str) -> str:
           return say_hello(msg=msg)

#. Run your workflow locally using ``pyflyte``:

    .. prompt:: bash $

       pyflyte run example.py:wf --msg world

Because this is just Python code also, you can also run the ``wf()`` function from the module.

Run on a Flyte backend installation
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

#. Start a Flyte sandbox installation on your local machine

    .. prompt:: bash $

       flytectl sandbox start

#. Now run the same workflow on the Flyte backend

    .. prompt:: bash $

       pyflyte run example.py:wf --msg world --remote

#. Navigate to the url produced as the result of running ``pyflyte``, this should take you to Flyte Console, the web UI used to manage Flyte entities.
