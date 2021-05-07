.. _howto-template-only-tasks:

#####################################################################
How to Write a Custom Task That Doesn't Depend on the User's Image?
#####################################################################

As we'll see throughout this how-to, the answer to the title question also addresses:

#. What is meant by a "flytekit-only plugin"?
#. How do custom task authors optimize a task's Docker image?

**********************
Background
**********************

Process Differences
=====================

Normal Function Tasks
---------------------

Most tasks that are in the cookbook and other Flyte introductory material are basic Python function tasks. That is, they are created by decorating a Python function with the :py:function:flytekit.task decorator. Please see the basic Task concept doc for more details. With the decorator in place, the process is

#. At serialization time, a Docker container image is required. The assumption is that this Docker image has the task code.
#. The task is serialized into a :std:ref:`api_msg_flyteidl.core.tasktemplate`. This template contains instructions to the container on how to reconstitute the task.
#. When Flyte runs the task, the container from step 1. is launched, and the instructions from step 2. recreate a Python object representing the task, using the user code in the container.
#. The task object is run.

The key points here are that
* the task object that gets serialized at compile-time is recreated using the user's code at run time, and
* at platform-run-time, the user-decorated function is executed.

TaskTemplate Based Tasks
------------------------

The execution process for task template based tasks differs from the above in that
#. At serialization time, the Docker container image is hardcoded into the task definition (by the author of that task type).
#. When serialized into a ``TaskTemplate``, the template should contain all the information needed to run that instance of the task (but not necessarily to reconstitute it).
#. When Flyte runs the task, the container from step 1. is launched. The container should have an executor built into it that knows how to execute the task, purely based on the ``TaskTemplate``.

The two points above differ in that
* the task object that gets serialized at compile-time does not exist at run time.
* at platform-run-time, there is no user function, and the executor is responsible for producing outputs, given the inputs to the task.

Why
===
These tasks are useful because they
* Shift the burden of writing the Dockerfile from the user using the task in workflows, to the author of the task type.
* Allow the author to optimize the image that the task runs.
* Make it possible to arbitrarily (mostly) extend Flyte task execution behavior without the need for a backend golang plugin. The caveat is that these tasks still cannot access the K8s cluster, so if you want a custom task type that creates some CRD, you'll still need a backend plugin.

*************************
Using a Task
*************************
Take a look at the `example PR <https://github.com/flyteorg/flytekit/pull/470>`__ where we moved the built-in SQLite3 task from the old style of writing a task to the new one.
From the user's perspective, not much changes. You still just

#. Install whatever Python library contains the task type definition (for SQLite3, it's bundled into flytekit itself, but usually this will not be the case).
#. Import and instantiate the task as you would any other type of non-function based task.

***************************
Writing a Task
***************************
There are three components to writing one of these new tasks.
* Your task extension code, which consists of 1) a class for the Task and 2) a class for the Executor.
* A Dockerfile - this is what is run when any user runs your task. It'll likely contain flytekit, Python, and your task extension code.

The `aforementioned PR <https://github.com/flyteorg/flytekit/pull/470>`__ where we migrate the SQLite3 task should be used to follow along the below.

Python Library
================

Task
-------
Authors creating new tasks of this type will need to create a subclass of the ``PythonCustomizedContainerTask`` class.

Specifically, you'll need to customize these three arguments to the parent class constructor:

* ``container_image`` This is the container image that will run when the user's invocation of the task is run on a Flyte platform.
* ``executor_type`` This should be the Python class that subclasses the ``ShimTaskExecutor``.
* ``task_type`` All types have a task type. This is just a string which the Flyte engine uses to determine which plugin to use when running a task. Anything that doesn't have an explicit match will default to the container plugin (which is correct in this case). So you can call this anything, just not anything that's already taken by something else (like "spark" or something).

Referring to the SQLite3 example ::

    container_image="ghcr.io/flyteorg/flytekit-py37:v0.18.1",
    executor_type=SQLite3TaskExecutor,
    task_type="sqlite3",

Note that the container is special in this case - because the definition of the Python classes themselves is bundled in Flytekit, we just use the Flytekit image.

Additionally, you will need to override the ``get_custom`` function. Keep in mind that the execution behavior of the task needs to be completely determined by the serialized form of the task (that is, the serialized ``TaskTemplate``). This function is how you can do that, as it's stored and inserted into the `custom field <https://github.com/flyteorg/flyteidl/blob/7302971c064b6061a148f2bee79f673bc8cf30ee/protos/flyteidl/core/tasks.proto#L114>`__ of the template. Keep the total size of the task template reasonably small though.

Executor
--------
The ``ShimTaskExecutor`` is an abstract class that you will need to subclass and override the ``execute_from_model`` function for. This function is where all the business logic for your task should go, and it will be called in both local workflow execution and at platform-run-time execution.

The signature of this execute function is different from the ``execute`` functions of most other tasks since here, all the business logic, the entirety of how the task is run, is determined from the ``TaskTemplate``.

Image
=======
This is the custom image that you supplied in the ``PythonCustomizedContainerTask`` subclass. Out of the box, these tasks will run a command that looks like the following when the container is run by Flyte ::

    pyflyte-execute --inputs s3://inputs.pb --output-prefix s3://outputs --raw-output-data-prefix s3://user-data --resolver flytekit.core.python_customized_container_task.default_task_template_resolver -- {{.taskTemplatePath}} path.to.your.executor.subclass

This means that your Docker image will need to have Python and Flytekit installed. The Python interpreter that is run by the container should be able to find your custom executor class at that ``path.to.your.executor.subclass`` import path.

Feel free to take a look at the Flytekit Dockerfile as well.
