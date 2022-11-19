"""
Caching
-------

.. tags:: Basic

Flyte provides the ability to cache the output of task executions to make the subsequent executions faster. A well-behaved Flyte task should generate deterministic output given the same inputs and task functionality.

Task caching is useful when a user knows that many executions with the same inputs may occur. For example, consider the following scenarios:

- Running a task periodically on a schedule
- Running the code multiple times when debugging workflows
- Running the commonly shared tasks amongst different workflows, which receive the same inputs

Let's watch a brief explanation of caching and a demo in this video, followed by how task caching can be enabled .

.. youtube:: WNkThCp-gqo

"""

# %%
# Import the necessary libraries.
import time

import pandas

# %%
# For any :py:func:`flytekit.task` in Flyte, there is always one required import, which is:
from flytekit import HashMethod, task, workflow
from flytekit.core.node_creation import create_node
from typing_extensions import Annotated

# %%
# Task caching is disabled by default to avoid unintended consequences of caching tasks with side effects. To enable caching and control its behavior, use the ``cache`` and ``cache_version`` parameters when constructing a task.
# `cache` is a switch to enable or disable the cache, and `cache_version` pertains to the version of the cache.
# ``cache_version`` field indicates that the task functionality has changed.
# Bumping the ``cache_version`` is akin to invalidating the cache.
# You can manually update this version and Flyte caches the next execution instead of relying on the old cache.


@task(cache=True, cache_version="1.0")  # noqa: F841
def square(n: int) -> int:
    """
     Parameters:
        n (int): name of the parameter for the task will be derived from the name of the input variable.
                 The type will be automatically deduced to ``Types.Integer``.

    Return:
        int: The label for the output will be automatically assigned, and the type will be deduced from the annotation.

    """
    return n * n


# %%
# In the above example, calling `square(n=2)` twice (even if it's across different executions or different workflows) will only execute the multiplication operation once.
# The next time, the output will be made available immediately since it is captured from the previous execution with the same inputs.

# %%
# If in a subsequent code update, you update the signature of the task to return the original number along with the result, it'll automatically invalidate the cache (even though the cache version remains the same).
#
# .. code-block:: python
#
#    @task(cache=True, cache_version="1.0")
#    def square(n: int) -> (int, int):
#        ...

# %%
# .. note::
#   If the user changes the task interface in any way (such as adding, removing, or editing inputs/outputs), Flyte treats that as a task functionality change. In the subsequent execution, Flyte runs the task and stores the outputs as newly cached values.
#
# How Does Caching Work?
# ######################
#
# Caching is implemented differently depending on the mode the user is running, i.e. whether they are running locally or using remote Flyte.
#
# How Does Remote Caching Work?
# *****************************
#
# The cache keys for remote task execution are composed of **Project**, **Domain**, **Cache Version**, **Task Signature**, and **Inputs** associated with the execution of the task, as per the following definitions:
#
# - **Project:** A task run under one project cannot use the cached task execution from another project which would cause inadvertent results between project teams that could result in data corruption.
# - **Domain:** To separate test, staging, and production data, task executions are not shared across these environments.
# - **Cache Version:** When task functionality changes, you can change the ``cache_version`` of the task. Flyte will know not to use older cached task executions and create a new cache entry on the subsequent execution.
# - **Task Signature:** The cache is specific to the task signature associated with the execution. The signature constitutes the task name, input parameter names/types, and the output parameter name/type.
# - **Task Input Values:** A well-formed Flyte task always produces deterministic outputs. This means, given a set of input values, every execution should have identical outputs. When task execution is cached, the input values are part of the cache key.
#
# The remote cache for a particular task is invalidated in two ways:
#
# 1. Modifying the ``cache_version``;
# 2. Updating the task signature.
#
# .. note::
#   Task executions can be cached across different versions of the task because a change in SHA does not necessarily mean that it correlates to a change in the task functionality.
#
# How Does Local Caching Work?
# ****************************
#
# The flytekit package uses the `diskcache <https://github.com/grantjenks/python-diskcache>`_ package, specifically `diskcache.Cache <http://www.grantjenks.com/docs/diskcache/tutorial.html#cache>`_, to aid in the memoization of task executions. The results of local task executions are stored under ``~/.flyte/local-cache/`` and cache keys are composed of **Cache Version**, **Task Signature**, and **Task Input Values**.
#
# Similar to the remote case, a local cache entry for a task will be invalidated if either the ``cache_version`` or the task signature is modified. In addition, the local cache can also be emptied by running the following command: ``pyflyte local-cache clear``, which essentially obliterates the contents of the ``~/.flyte/local-cache/`` directory.
#
# .. note::
#   The format used by the store is opaque and not meant to be inspectable.
#
# Caching of Non-flyte Offloaded Objects
# ######################################
#
# The default behavior displayed by Flyte's memoization feature might not match the user intuition. For example, this code makes use of pandas dataframes:


@task
def foo(a: int, b: str) -> pandas.DataFrame:
    df = pandas.DataFrame(...)
    ...
    return df


@task(cache=True, cache_version="1.0")
def bar(df: pandas.DataFrame) -> int:
    ...


@workflow
def wf(a: int, b: str):
    df = foo(a=a, b=b)
    v = bar(df=df)  # noqa: F841


# %%
# If run twice with the same inputs, one would expect that ``bar`` would trigger a cache hit, but it turns out that's not the case because of how dataframes are represented in Flyte.
# However, with release 1.2.0, Flyte provides a new way to control memoization behavior of literals. This is done via a ``typing.Annotated`` call on the task signature.
# For example, in order to cache the result of calls to ``bar``, you can rewrite the code above like this:


@task
def hash_pandas_dataframe_function(df: pandas.DataFrame) -> str:
    return str(pandas.util.hash_pandas_object(df))


@task
def foo_1(  # noqa: F811
    a: int, b: str  # noqa: F821
) -> Annotated[
    pandas.DataFrame, HashMethod(hash_pandas_dataframe_function)  # noqa: F821
]:  # noqa: F821
    df = pandas.DataFrame(...)  # noqa: F821
    ...
    return df


@task(cache=True, cache_version="1.0")  # noqa: F811
def bar_1(df: pandas.DataFrame) -> int:  # noqa: F811
    ...  # noqa: F811


@workflow
def wf_1(a: int, b: str):  # noqa: F811
    df = foo(a=a, b=b)  # noqa: F811
    v = bar(df=df)  # noqa: F841


# %%
# Note how the output of task ``foo`` is annotated with an object of type ``HashMethod``. Essentially, it represents a function that produces a hash that is used as part of the cache key calculation in calling the task ``bar``.
#
# How Does Caching of Offloaded Objects Work?
# *******************************************
#
# Recall how task input values are taken into account to derive a cache key.
# This is done by turning the literal representation into a string and using that string as part of the cache key. In the case of dataframes annotated with ``HashMethod`` we use the hash as the representation of the Literal. In other words, the literal hash is used in the cache key.
#
# This feature also works in local execution.
#
# %%
# Here's a complete example of the feature:


def hash_pandas_dataframe(df: pandas.DataFrame) -> str:
    return str(pandas.util.hash_pandas_object(df))


@task
def uncached_data_reading_task() -> Annotated[
    pandas.DataFrame, HashMethod(hash_pandas_dataframe)
]:
    return pandas.DataFrame({"column_1": [1, 2, 3]})


@task(cache=True, cache_version="1.0")
def cached_data_processing_task(df: pandas.DataFrame) -> pandas.DataFrame:
    time.sleep(1)
    return df * 2


@task
def compare_dataframes(df1: pandas.DataFrame, df2: pandas.DataFrame):
    assert df1.equals(df2)


@workflow
def cached_dataframe_wf():
    raw_data = uncached_data_reading_task()

    # Execute `cached_data_processing_task` twice, but force those
    # two executions to happen serially to demonstrate how the second run
    # hits the cache.
    t1_node = create_node(cached_data_processing_task, df=raw_data)
    t2_node = create_node(cached_data_processing_task, df=raw_data)
    t1_node >> t2_node

    # Confirm that the dataframes actually match
    compare_dataframes(df1=t1_node.o0, df2=t2_node.o0)


if __name__ == "__main__":
    df1 = cached_dataframe_wf()
    print(f"Running cached_dataframe_wf once : {df1}")
