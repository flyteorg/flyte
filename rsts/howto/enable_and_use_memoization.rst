.. _howto_enable_use_memoization:

#########################################
How do I enable and use memoization?
#########################################

Flyte provides the ability to cache the output of task executions in order to make subsequent executions faster. A well-behaved Flyte Task should generate deterministic output given the same inputs and task functionality. This is useful in situations where a user knows that many executions with the exact same inputs can occur. For example, your task may be periodically run on a schedule, run multiple times when debugging workflows, or commonly shared across different workflows but receive the same inputs.

Enable Caching For a Task - SDK?
-----------------------------------

In order to enable your task to be cached, mark ``cache=True`` below:

.. code-block:: python

   @task(cache=True, cache_version='1.0.0')
   def hash_string_task(original: str) -> str:
      ...

A task execution is cached based on the **Project, Domain, cache_version, the task signature and inputs** associated with the execution of the task.

- *Project:* A task run under one project cannot use the cached task execution from another project. This could cause inadvertent results between project teams that could cause data corruption.
- *Domain:* For separation of test, staging, and production data, task executions are not shared across these environments.
- *cache_version:* When task functionality changes, you can change the cache_version of the task. Flyte will know not to use older cached task executions and create a new cache entry on the next execution.
- *Task signature:*: The cache is specific to the task signature that is associated with the execution. The signature is made up of task name, input parameter names/types and also the output parameter name/types.
- *Task input values*: A well-formed Flyte Task always produces deterministic outputs. This means given a set of input values, every execution should produce identical outputs. When a task execution is cached, the input values are part of the cache key.

Notice that task executions can be cached across different versions of the task. This is because a change in SHA does not neccessarily mean that it correlates to a change in task functionality.

Flyte provides several ways to break the old task execution cache, and cache new output:

- ``cache_version``: this field indicates that the task functionality has changed. Flyte users can manually update this version and Flyte will cache the next execution instead of relying on the old cache.
- Task signature: If a Flyte user changes the task interface in any way (such as by adding, removing, or editing inputs/outputs), Flyte will treat that as a task functionality change. On the next execution, Flyte will run the task and store the outputs as new cached values.


Enable Caching in your FlytePlatform
--------------------------------------