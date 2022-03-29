.. _gettingstarted_build:


Implement Your Workflows in Python
----------------------------------

.. note::

    We recommend using an OSX or a Linux machine, as we have not tested this on Windows. If you happen to test it, please let us know.


Prerequisites
^^^^^^^^^^^^^^^^
Make sure you have `Git <https://git-scm.com/>`__, and `Python <https://www.python.org/downloads/>`__ >= 3.7 installed. Also ensure you have pip3 (mostly the case).

Start a new project / repository
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
Install Flyte's Python SDK — `Flytekit <https://pypi.org/project/flytekit/>`__ (recommended using a  `virtual environment <https://docs.python.org/3/library/venv.html>`__), run ``pyflyte init {project_name}``, where ``{project_name}`` is the directory that will be created containing the scaffolding for a flyte-ready project. Feel free to use any name as your project name, however for this guide we're going to use ``myflyteapp`` in the  invocation:

.. prompt:: bash (venv)$

    pip3 install flytekit --upgrade
    pyflyte init myflyteapp
    cd myflyteapp

The ``myflyteapp`` directory comes with a sample workflow, which can be found under ``flyte/workflows/example.py``. The structure below shows the most important files and how a typical Flyte app should be laid out.

.. dropdown:: A typical Flyte app should have these files

   .. code-block:: text

       myflyteapp
       ├── Dockerfile
       ├── docker_build_and_tag.sh
       ├── flyte
       │         ├── __init__.py
       │         └── workflows
       │             ├── __init__.py
       │             └── example.py
       └── requirements.txt

   .. note::

       Two things to note here:

       * You can use `pip-compile` to build your requirements file.
       * The Dockerfile that comes with this is not GPU ready, but is a simple Dockerfile that should work for most of your apps.

Run the Workflow Locally
^^^^^^^^^^^^^^^^^^^^^^^^
The workflow can be run locally, simply by running it as a Python script; note the ``__main__`` entry point at the bottom of ``flyte/workflows/example.py``.

.. prompt:: bash (venv)$

   python flyte/workflows/example.py


Expected output:

.. prompt:: text

  Running my_wf() hello world

.. admonition:: Recap

  .. rubric:: 🎉 Congratulations! You just ran your first Flyte workflow locally, let's take it to the cloud!
