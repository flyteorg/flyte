"""
Task Cache
----------------

Flyte provides the ability to cache the output of task executions in order to make subsequent executions faster. A well-behaved Flyte Task should generate deterministic output given the same inputs and task functionality.
This is useful in situations where a user knows that many executions with the exact same inputs can occur. For example, your task may be periodically run on a schedule, run multiple times when debugging workflows, or 
commonly shared across different workflows but receive the same inputs.
"""

# %%
# For any task in flyte, there is always one required import
from flytekit import task


# %%
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
# In the above example, calling `square(n=2)` twice (even if it's across different executions or different workflows), will only actually execution the multiplication operation once. The second time the output will be made
# available immediately.

# %%
# If, in a subsequent code update, we update the signature of the task to return the original number along with the result, it'll automatically invalidate the cache (even though the cache version remains the same). ::
#
#   :py:func:`flytekit.task`
#   @task(cache=True, cache_version="1.0")
#   def square(n: int) -> (int, int):
#       ...

# %%
# To read more about Task caching and how a unique signature is calculated, please proceed to the `Task Cache documentation <https://flyte.readthedocs.io/en/stable/user/features/task_cache.html#features-task-cache>`__.
