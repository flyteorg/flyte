.. _howto-template-only-tasks:

#####################################################################
How do I write a custom task that doesn't depend on the user's image?
#####################################################################

As we'll see throughout this how-to, the answer to the title question also addresses:

#. What is meant by a "flytekit-only plugin"?
#. How do custom task authors optimize a task's Docker image?

**********************
Background
**********************

Process Differences
=====================

Normal function tasks
---------------------

Most tasks that are in the cookbook and other Flyte introductory material are basic Python function tasks. That is, they are created by decorating a Python function with the ``@task`` decorator. Please see the basic Task concept doc for more details. With the decorator in place, the process is

#. At serialization time, a Docker container image is required. The assumption is that this Docker image has the task code.
#. The task is serialized into a ``TaskTemplate``. This template contains instructions to the container on how to reconstitute the task.
#. When Flyte runs the task, the container from step 1. is launched, and the instructions from step 2. recreate a Python object representing the task, using the user code in the container.
#. The task object is run.

The key points here are that
* the task object that gets serialized at compile-time is recreated using the user's code at run time, and
* at platform-run-time, the user-decorated function is executed.

TaskTemplate based tasks
------------------------

The execution process for task template based tasks differ from the above in that
#. At serialization time, the Docker container image is hardcoded into the task definition (by the author of that task type).
#. When serialized into a ``TaskTemplate``, the template should contain all the information needed to run that instance of the task (but not necessarily to reconstitute it).
#. When Flyte runs the task, the container from step 1. is launched. The container should have an executor built into it that knows how to execute the task, purely based on the ``TaskTemplate``.

The two points above differ in that
* the task object that gets serialized at compile-time does not exist at run time.
* at platform-run-time, there is no user function, and the executor is responsible for producing outputs, given the inputs to the task.

Why
===
These tasks are useful because
* Shift the burden of writing the Dockerfile from the user using the task in workflows, to the author of the task type.
* Allow the author to optimize the image that the task runs.
* Make it possible to arbitrarily (mostly) extend Flyte task execution behavior without the need for a backend golang plugin. The caveat is that these tasks still cannot access the K8s cluster, so if you want a custom task type that creates some CRD, you'll still need a backend plugin.

*************************
Using a Task
*************************
Take a look at the example PR where we moved the built-in SQLite3 task from the old style of writing a task to the new one.
From the user's perspective, not much changes. You still just

#. Install whatever Python library contains the task type definition (for SQLite3, it's bundled into flytekit itself, but usually this will not be the case).
#. Import and instantiate the task as you would any other type of non-function based task.

***************************
Writing a Task
***************************
There's three components to writing one of these new tasks.
* A Dockerfile - this is what is run when any user runs your task. It'll likely contain flytekit, Python, and your task extension code.
* Your task extension code, which consists of 1) a class for the Task and 2) a class for the Executor.

Image
=======


Python Library
================

Task
-------


Executor
--------










