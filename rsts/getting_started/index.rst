.. _getting_started:

################
Getting Started
################

Prerequisites
^^^^^^^^^^^^^^^^
Make sure you have `Docker <https://docs.docker.com/get-docker/>`__ and Docker Daemon is running.

Install flytectl
^^^^^^^^^^^^^^^^^

#. Install :std:ref:`flytectl`. ``Flytectl`` is a commandline interface for Flyte.

   .. tabbed:: OSX

     .. prompt:: bash $

        brew install flyteorg/homebrew-tap/flytectl

     *Upgrade* existing installation using the following command:

     .. prompt:: bash $

        flytectl upgrade

   .. tabbed:: Other Operating systems

     .. prompt:: bash $

         curl -sL https://ctl.flyte.org/install | sudo bash -s -- -b /usr/local/bin # You can change path from /usr/local/bin to any file system path
         export PATH=$(pwd)/bin:$PATH # Only required if user used different path then /usr/local/bin

     *Upgrade* existing installation using the following command:

     .. prompt:: bash $

        flytectl upgrade

   **Test** if Flytectl is installed correctly (your Flytectl version should be >= 0.1.34.) using the following command:

   .. prompt:: bash $

      flytectl version

#. TODO: Start sandbox (single-binary) in the current directory


Install flytekit
^^^^^^^^^^^^^^^^

Install Flyte's python SDK — `Flytekit <https://pypi.org/project/flytekit/>`__.

.. prompt:: bash

   pip install flytekit


Run your first workflow
^^^^^^^^^^^^^^^^^^^^^^^

#. Copy the following code to a file named ``example.py``

.. code-block:: python

   from flytekit import task, workflow

   @task
   def say_hello(msg: str) -> str:
       return f"hello, {msg}!"

   @workflow
   def wf(msg: str) -> str:
       return say_hello(msg=msg)

#. Run the workflow via ``pyflyte``:

.. prompt:: bash $

   pyflyte run example.py:wf --msg world --remote

#. Navigate to the url produced as the result of running ``pyflyte``, this should take you to Flyte Console, the web UI used to manage Flyte entities.

   TODO: gif containing flyte console running ``example.py``
