.. _getting-started-first-example:

###############################
Write Your First Flyte Workflow
###############################

By the end of this guide you will learn how to author a Flyte workflow and run it locally.

.. rubric:: Estimated time: <3 minutes

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
