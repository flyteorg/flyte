.. _tutorials-getting-started-first-example:

###########################
Writing Your First Workflow
###########################

By the end of this getting started guide you'll be familiar with how easy it is to author a Flyte workflow and run it locally.

Estimated time to complete: <3 minutes.

The easiest way to author a Flyte Workflow is using the provided python SDK called "FlyteKit".

You can save some effort by cloning the ``flytekit-python-template`` repo, and re-initializing it as a new git repository ::

  pip install flytekit==0.16.0b6
  git clone git@github.com:lyft/flytekit-python-template.git myflyteproject
  cd myflyteproject
  rm -rf .git
  git init

Writing a Task
*****************

The most basic Flyte primitive is a "task". Flyte Tasks are units of work that can be composed in a workflow. The simplest way to write a Flyte task is using the Flyte Python SDK - flytekit.

Start by creating a new file ::


   touch myapp/workflows/first.py


And add the required imports we'll need for this example:


.. code-block:: python

  import flytekit
  from flytekit import task, workflow
  from flytekit.types.file import FlyteFile

From there, we can begin to write our first task.  It should look something like this:

.. code-block:: python

    @task
    def persist_and_greet(name: str) -> str:
        working_dir = flytekit.current_context().working_directory
        output_file = '{}/greeting.txt'.format(working_dir)

        greeting = f"Hello {name}"
        with open(output_file,"a") as output:
            output.write(greeting)

        print(greeting)
        return greeting


Some of the new concepts demonstrated here are:

* Use the :py:func:`flytekit.task` decorator to convert your typed python function to a Flyte task.
* A :py:meth:`flytekit.current_context` is a Flytekit entity that represents the execution context.  It is useful for getting run-time parameters such as the current working dir.


You can call this task

.. code-block:: python

   greet(name="world")

and iterate locally before adding it to part of a larger overall workflow.

Writing a Workflow
*********************
Next you need to call that task from a workflow.  In the same file, add these lines.

.. code-block:: python

   @workflow
   def GreeterWorkflow(name: str="world") -> str:
        greeting = persist_greet(name=name)
        return greeting

This code block creates a workflow, with one task. The workflow itself has an input (the link to an image) that gets passed into the task, and an output, which is the processed image.

You can call this workflow

.. code-block:: python

   GreeterWorkflow(name=...)

iterate locally before moving on to register it with Flyte.

.. note::

   Every invocation of a Flyte workflow requires specifying keyword args.
