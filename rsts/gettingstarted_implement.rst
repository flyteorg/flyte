.. _gettingstarted_implement:

Getting Started
---------------


.. raw:: html
  
    <p style="color: #808080; font-weight: 500; font-size: 20px; padding-top: 10px;">A step-by-step guide to building, deploying, and iterating on Flyte tasks and workflows</p>

.. panels::
    :body: text-justify
    :container: container-xs
    :column: col-lg-4 col-md-4 col-sm-6 col-xs-12 p-2

    ---
    :body: bg-primary
    ➔ Implement
    ---
    2. Scale
    ---
    3. Iterate

.. caution::

    We recommend using an OSX or a Linux machine, as we have not tested this on Windows. If you happen to test it, please let us know.

1. Implement your workflows in python
======================================

Prerequisites
^^^^^^^^^^^^^^^^
Make sure you have `Docker <https://docs.docker.com/get-docker/>`__ , `Git <https://git-scm.com/>`__, and `Python <https://www.python.org/downloads/>`__ >= 3.7 installed.

Start a new project / repository
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
#. Install Flyte's Python SDK — `Flytekit <https://pypi.org/project/flytekit/>`__ (recommended in a virtual environment) and clone the `flytekit-python-template <https://github.com/flyteorg/flytekit-python-template>`__ repo.

   .. prompt:: bash (venv)$

     pip install flytekit
     git clone https://github.com/flyteorg/flytekit-python-template.git myflyteapp
     cd myflyteapp

#. The repo comes with a sample workflow, which can be found under ``myapp/workflows/example.py``. The structure below shows the most important files and how a typical Flyte app should be laid out.

   .. dropdown:: A typical Flyte app should have these files

       .. code-block:: text

           .
           ├── Dockerfile
           ├── docker_build_and_tag.sh
           ├── myapp
           │         ├── __init__.py
           │         └── workflows
           │             ├── __init__.py
           │             └── example.py
           └── requirements.txt

       .. note::

           Two things to note here:

           * You can use `pip-compile` to build your requirements file. 
           * The Dockerfile that comes with this is not GPU ready, but is a simple Dockerfile that should work for most of your apps.


Run the workflow locally
^^^^^^^^^^^^^^^^^^^^^^^^^
The workflow can be run locally, simply by running it as a Python script — note the ``__main__`` entry point at the `bottom of the file <https://github.com/flyteorg/flytekit-python-template/blob/main/myapp/workflows/example.py#L58>`__.

   .. prompt:: bash (venv)$

       python myapp/workflows/example.py

   .. dropdown:: Expected output

      .. prompt:: text

         Running my_wf() hello world

.. admonition:: Recap

  .. rubric:: 🎉 Congratulations! You just ran your first Flyte workflow locally, let's take it to the cloud!


.. toctree::
   :maxdepth: -1
   :caption: Getting Started
   :hidden:

   Build and Deploy your application<gettingstarted_scale>
   Iterate "fast"er<gettingstarted_iterate>
   User Guide <https://docs.flyte.org/projects/cookbook/en/latest/user_guide.html>