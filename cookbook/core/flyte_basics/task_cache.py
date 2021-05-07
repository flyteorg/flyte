"""
Task Cache
-----------

Flyte provides the ability to cache the output of task executions in order to make subsequent executions faster. A well-behaved Flyte Task should generate deterministic output given the same inputs and task functionality.
This is useful in situations where a user knows that many executions with the exact same inputs can occur. For example, your task may be periodically run on a schedule, run multiple times when debugging workflows, or 
commonly shared across different workflows but receive the same inputs.
"""

# %%
# For any task in flyte, there is always one required import
from flytekit import task


# %%
# Default Behavior
# ##################
#
# Task caching is disabled by default. This is to avoid unintended consequences of caching tasks with side-effects. To enable caching and control its behavior, use the `cache` and `cache_version` parameters when constructing
# a task. `cache` controls whether caching is enabled or disabled overall and `cache_version` controls which version of the cache is used. Bumping this version is akin to invalidating the cache, the next execution of that task
# with the same input/interface will actually run, and the results will be cached so that subsequent executions are just replayed from the cache.
#
# :py:func:`flytekit.task`
@task(cache=True, cache_version="1.0")
def square(n: int) -> int:
    """
     Parameters:
        n (int): name of the parameter for the task will be derived from the name of the input variable
               the type will be automatically deduced to be Types.Integer

    Return:
        int: The label for the output will be automatically assigned and type will be deduced from the annotation

    """
    return n * n


# %%
# In the above example, calling `square(n=2)` twice (even if it's across # different executions or different workflows), will only actually execute the multiplication operation once. The second time the output will be made
# available immediately - (captured from the previous execution with the same inputs)

# %%
# If, in a subsequent code update, we update the signature of the task to return the original number along with the result, it'll automatically invalidate the cache (even though the cache version remains the same). ::
#
#   :py:func:`flytekit.task`
#   @task(cache=True, cache_version="1.0")
#   def square(n: int) -> (int, int):
#       ...

# %%
# Enable Caching For a Task - SDK
# ###############################

# In order to enable your task to be cached, mark ``cache=True`` below:

.. code-block:: python

   @task(cache=True, cache_version='1.0.0')
   def hash_string_task(original: str) -> str:
      ...

# %%
# A task execution is cached based on the **Project, Domain, Cache_version, Task Signature and Inputs** associated with the execution of the task.

# - *Project:* A task run under one project cannot use the cached task execution from another project. This could cause inadvertent results between project teams that could cause data corruption.
# - *Domain:* For separation of test, staging, and production data, task executions are not shared across these environments.
# - *Cache_version:* When task functionality changes, you can change the cache_version of the task. Flyte will know not to use older cached task executions and create a new cache entry on the next execution.
# - *Task Signature:*: The cache is specific to the task signature that is associated with the execution. The signature is made up of task name, input parameter names/types and also the output parameter name/types.
# - *Task Input Values*: A well-formed Flyte Task always produces deterministic outputs. This means given a set of input values, every execution should produce identical outputs. When a task execution is cached, the input values are part of the cache key.

# Notice that task executions can be cached across different versions of the task. This is because a change in SHA does not neccessarily mean that it correlates to a change in task functionality.

# Flyte provides several ways to break the old task execution cache, and cache new output:

# - ``cache_version``: this field indicates that the task functionality has changed. Flyte users can manually update this version and Flyte will cache the next execution instead of relying on the old cache.
# - Task signature: If a Flyte user changes the task interface in any way (such as by adding, removing, or editing inputs/outputs), Flyte will treat that as a task functionality change. On the next execution, Flyte will run the task and store the outputs as new cached values.
