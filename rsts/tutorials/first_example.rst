.. _tutorials-getting-started-first-example:

######################################
Write your first Flyte workflow
######################################

By the end of this getting started guide you'll be familiar with how easy it is to author a Flyte workflow and run it locally.

.. rubric:: Estimated time to complete: <3 minutes.


Prerequisites
*************

Ensure that you have `git <https://git-scm.com/>`__ installed.

First install the python Flytekit SDK and clone the ``flytekit-python-template`` repo ::

  pip install flytekit==0.16.0b6
  git clone git@github.com:flyteorg/flytekit-python-template.git flyteexamples
  cd flyteexamples
  rm -rf .git
  git init


Flyte Tasks and Workflows
*************************

Let's take a look at the example workflow found in `myapp/workflows/example.py <https://github.com/flyteorg/flytekit-python-template/blob/main/myapp/workflows/example.py>`__

.. rli:: https://raw.githubusercontent.com/flyteorg/flytekit-python-template/main/myapp/workflows/example.py
   :language: python

The most basic Flyte primitive is a :std:doc:`task <flytekit:tasks>`.
Flyte tasks are units of work that can be composed in a :std:doc:`workflow <flytekit:workflow>`.

You can call this task

.. code-block:: python

   greet(name="world")

and iterate locally before adding it to part of a larger overall workflow.

Similarly, you can call this workflow

.. code-block:: python

   hello_world(name=...)

and iterate locally before moving on to register it with Flyte.

.. tip:: Every invocation of a Flyte workflow requires specifying keyword arguments as in the example - ``hello_world(name="name")``. Calling the workflow without the keyword ``name`` will raise an ``AssertionError``.
