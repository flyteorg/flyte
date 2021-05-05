.. _howto-template-only-tasks:

#####################################################################
How do I write a custom task that doesn't depend on the user's image?
#####################################################################

As we'll see throughout this how-to, the answer to the title question also addresses:

#. What is meant by a "flytekit-only plugin"?
#. How do custom task authors optimize a task's Docker image?

**********************
How normal tasks work
**********************

Most tasks that are in the cookbook and other Flyte introductory material are basic Python function tasks. That is, they are created by decorating a Python function with the ``@task`` decorator. Please see the basic Task concept doc for more details but from here, the process is

#. At serialization time, a Docker container image is required. The assumption is that this Docker image has the task code.
#. The task is serialized into a ``TaskTemplate``. This template contains instructions to the container on how to reconstitute the task.
#. When Flyte runs the task, the container from step 1. is launched, and the instructions from step 2. recreate a Python object representing the task, using the user code in the container.
#. The task object is run.

The key point here is that the task object that gets serialized at compile-time is recreated using the user's code at run time.

**************************
Task Template based tasks
**************************


Using one of these tasks


Writing one of these tasks






