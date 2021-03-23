.. _getting-started-first-example:

###############################
Write Your First Flyte Workflow
###############################

By the end of this guide you will become familiar with how easy it is to author a Flyte workflow and run it locally.

.. rubric:: Estimated time: <3 minutes


Prerequisites
*************

#. Ensure that you have `git <https://git-scm.com/>`__ installed.

#. Let us make a virutal environment **(recommended)** - and then install flytekit using
    ``--pre`` is used because we are currently using the beta version of flytekit 0.16.0, this introduces a completely new SDK for authoring workflows
    ::

        pip install --pre flytekit


#. Now we will use the ``flytekit-python-template`` repo to create our own git repository called ``flyteexamples`` ::

      git clone git@github.com:flyteorg/flytekit-python-template.git flyteexamples
      cd flyteexamples
      rm -rf .git
      git init


Flyte Tasks and Workflows
*************************

Take a look at the example workflow found in `myapp/workflows/example.py <https://github.com/flyteorg/flytekit-python-template/blob/main/myapp/workflows/example.py>`__

.. rli:: https://raw.githubusercontent.com/flyteorg/flytekit-python-template/main/myapp/workflows/example.py
   :language: python

The most basic Flyte primitive is a :std:doc:`task <generated/flytekit.task>`

Flyte tasks are units of work that can be composed in a :std:doc:`workflow <generated/flytekit.workflow>`

You can call this task:

.. code-block:: python

   greet(name="world")

and iterate it locally before adding it to part of a larger overall workflow.

Similarly, you can call this workflow:

.. code-block:: python

   hello_world(name=...)

and iterate it locally before registering it with Flyte.

.. tip:: Every invocation of a Flyte workflow requires specifying keyword arguments as in the example - ``hello_world(name="name")``. Calling the workflow without the keyword ``name`` will raise an ``AssertionError``.
