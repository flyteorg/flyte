# Task caching

Caching allows you to avoid re-running potentially expensive tasks. You can
specify which tasks to cache with the `cache` and `cache_version` arguments:

```{code-cell} ipython3
from typing import List
from flytekit import task, workflow


@task(cache=True, cache_version="1")
def compute_mean(data: List[float]) -> float:
    return sum(data) / len(data)

@workflow
def wf(data: List[float]) -> float:
    return compute_mean(data=data)
```

Caching works both locally and on a Flyte backend.

```{code-cell} ipython3
%timeit -n 1 -r 1 wf(data=[float(x) for x in range(100_000)])
```

```{code-cell} ipython3
%timeit -n 1 -r 1 wf(data=[float(x) for x in range(100_000)])
```

As you can see, the second call to the `wf` workflow takes less time because
Flyte simply hits the cache to obtain the result.

```{note}
For file-like data types like {py:class}`flytekit.types.file.FlyteFile` and offloaded
data types like `pandas.DataFrame` objects, you can provide a hash function that
represents the cache key. Learn more in the {ref}`User Guide <cache-offloaded-objects>`.
```

## Task caching (copied from Union docs)

Flyte provides the ability to cache the output of task executions to make subsequent executions faster.

Task caching is useful when many executions with the same inputs may occur.
For example, consider the following scenarios:

* Running a task periodically on a schedule
* Running the code multiple times when debugging workflows
* Running the commonly shared tasks amongst different workflows, which receive the same inputs

Here's a video with a brief explanation and demo:

<iframe width="560" height="315" src="https://www.youtube.com/embed/WNkThCp-gqo?si=sFATJHv3avFRf6Tn" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

There are three parameters available in the `@task` decorator that pertain to caching:

* `cache`(`bool`): Enables or disables caching of the task.
By default, task caching is disabled.
This is to avoid unintended consequences when caching tasks with side effects.
To enable caching set `cache=True`.
* `cache_version` (`str`): Part of the cache key.
A change to this parameter will invalidate the cache.
This allows the programmer to explicitly indicate when a change has been made to the task that should invalidate any existing cached results.
Note that this is not the only change that will invalidate the cache (see below).
* `cache_serialize` (`bool`): Enables or disables [cache serialization](#cache-serialization).
When enabled for a task, Flyte ensures that a single instance of the task is evaluated before any other instances that would otherwise run concurrently.
This allows the initial instance to cache its result and lets the later instances reuse the resulting cached outputs.
Cache serialization is disabled by default.

Here is an example:

```{code-block} python
@task(cache=True, cache_version="1.0", cache_serialize=True)
def t(n: int) -> int:
    return n * n
```

In the above example, calling `t(n=2)` twice (even if it’s across different executions or different workflows) will only execute the multiplication operation once.
The next time, the output will be made available immediately since it is captured from the previous execution with the same inputs.

### How caching works

When a task (with caching enabled) completes on Union, a **key-value entry** is created in the **caching table**. The **value** of the entry is the output.
The **key** is composed of:

* **Project:** A task run under one project cannot use the cached task execution from another project which would cause inadvertent results between project teams that could result in data corruption.
* **Domain:** To separate test, staging, and production data, task executions are not shared across these environments.
* **Cache version:** When task functionality changes, you can change the `cache_version` of the task.
Flyte will know not to use older cached task executions and create a new cache entry on the subsequent execution.
* **Task signature:** The cache is specific to the task signature associated with the execution.
The signature constitutes the task name, input parameter names/types, and the output parameter name/type.
If the task signature changes, the cache entry is invalidated.
* **Task input values:** A well-formed Flyte task always produces deterministic outputs.
This means that, given a set of input values, every execution should have identical outputs.
When task execution is cached, the input values are part of the cache key.
If you run a task with a new set of inputs a new cache entry is created for the combination of that particular task with those particular inputs.

The result is that within a given project and domain, a cache entry is created for each distinct combination of task name and task input.
If the same task with the same input values is encountered again, the cached output is used instead of running the task again.

However, for a given task name, all its cache entries (for every input) will be removed from the table if either

* The **cache version** changes, or
* The **task signature** changes.

#### Cache version

When a change to code is made that should invalidate the cache for that task, the programmer can explicitly indicate this by incrementing the `cache_version`parameter value in the `@task` decorator.
For example:

```{code-block} python
@task(cache=True, cache_version="1.1", cache_serialize=True)
def t(n: int) -> int:
   return n * n + 1
```

Here the `cache_version` parameter has been bumped from `1.0`to `1.1`, resulting in the removal of the cache entries for that task.
The next time the task is called it will be executed and the result re-cached under an updated key.

Because the `cache_version` parameter must be explicitly changed, not every version (in the sense of git revision) of a task will necessarily invalidate the cache.
A change in Git SHA does not necessarily correlate to a change in the task functionality.
You can refine a task's code without invalidating its outputs as long as you don't change the `cache_version` parameter (or the task signature, see below)

The idea behind this is to decouple syntactic sugar (for example, changed documentation or renamed variables) from changes to logic that can affect the task's result.
When you use Git (or any version control system), you have a new version per code change.
For most tasks in a repository, except for the few you touch per PR, their behavior is probably still the same, so you don't want their cached outputs to be lost.

When a task's behavior does change though, you can bump `cache_version` to invalidate the cache entry and make the system recompute the outputs.

#### Task signature

If the user modifies the task signature by adding, removing, or editing input parameters or output return types, Flyte considers it a change in task functionality and invalidates the cache entries for that task.
During the next execution, Flyte executes the task again and caches the outputs as new values stored under an updated key.

#### Caching when running locally

The description above applies to caching when running a task remotely (that is, your Union cluster).
Caching is also available [when running on a local cluster](../../getting-started/running-in-a-local-cluster.md).

When running locally the caching mechanism is the same except that the cache key does not include **project** or **domain** (since there are none).
The cache key is composed only of **cache version**, **task signature**, and **task input**.
The results of local task executions are stored under `~/.flyte/local-cache/`.

Similar to the remote case, a local cache entry for a task will be invalidated if either the `cache_version` or the task signature is modified.
In addition, the local cache can also be emptied by running

```{code-block} shell
$ pyflyte local-cache clear
```

which essentially removes the contents of the `~/.flyte/local-cache/` directory.

### Cache serialization

Cache serialization means only executing a single instance of a unique cacheable task (determined by the `cache_version` parameter and task signature) at a time.
Using this mechanism, Flyte ensures that during multiple concurrent executions of a task only a single instance is evaluated and all others wait until completion and reuse the resulting cached outputs.

Ensuring serialized evaluation requires a small degree of overhead to coordinate executions using a lightweight artifact reservation system.
Therefore, this should be viewed as an extension to rather than a replacement for non-serialized cacheable tasks.
It is particularly well fit for long-running or otherwise computationally expensive tasks executed in scenarios similar to the following examples:

* Periodically scheduled workflow where a single task evaluation duration may span multiple scheduled executions.
* Running a commonly shared task within different workflows (which receive the same inputs).

#### Enabling cache serialization

Task cache serializing is disabled by default to avoid unexpected behavior for task executions.
To enable set `cache_serialize=True` in the `@task` decorator (this only has an effect if `cache=True` is also set)

The cache key definitions follow the same rules as non-serialized cache tasks.
It is important to understand the implications of the task signature and `cache_version` parameter in defining cached results.

```{code-block} python
@task(cache=True, cache_version="1.0", cache_serialize=True)
def t(n: int) -> int:
    return n * n
```

In the above example calling `t(n=2)` multiple times concurrently (even in different executions or workflows) will only execute the multiplication operation once.
Concurrently evaluated tasks will wait for completion of the first instance before reusing the cached results and subsequent evaluations will instantly reuse existing cache results.

#### How does cache serialization work?

The cache serialization paradigm introduces a new artifact reservation system. Tasks may use this reservation system to acquire an artifact reservation, indicating that they are actively evaluating the task, and release the reservation, once the execution is completed.
Flyte uses a clock-skew algorithm to define reservation timeouts. Therefore, tasks are required to periodically extend the reservation during execution.

The first execution of a serializable cached task will successfully acquire the artifact reservation.
Execution will be performed as usual and upon completion, the results are written to the cache, and reservation is released.
Concurrently executed task instances (those that would otherwise run in parallel with the initial execution) will observe an active reservation, in which case these instances will wait until the next reevaluation and perform another check.
Once the initial execution completes they will reuse the cached results as will any subsequent instances of the same task.

Flyte handles task execution failures using a timeout on the reservation.
If the task currently holding the reservation fails to extend it before it times out, another task may acquire the reservation and begin executing the task.

### Caching of offloaded objects

In some cases, the default behavior displayed by Flyte’s caching feature might not match the user's intuition. For example, this code makes use of pandas dataframes:

```{code-block} python
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
    v = bar(df=df)
```

If run twice with the same inputs, one would expect that `bar` would trigger a cache hit, but it turns out that’s not the case because of how dataframes are represented in Flyte.

However, Flyte provides a new way to control the caching behavior of literals.
This is done via a `typing.Annotated` call on the task signature.
For example, in order to cache the result of calls to `bar`, you can rewrite the code above like this:

```{code-block} python
def hash_pandas_dataframe(df: pandas.DataFrame) -> str:
    return str(pandas.util.hash_pandas_object(df))


@task
def foo_1(a: int, b: str) -> Annotated[pandas.DataFrame, HashMethod(hash_pandas_dataframe)]:
    df = pandas.DataFrame(...)
    ...
    return df


@task(cache=True, cache_version="1.0")
def bar_1(df: pandas.DataFrame) -> int:
    ...


@workflow
def wf_1(a: int, b: str):
    df = foo(a=a, b=b)
    v = bar(df=df)
```

Note how the output of the task `foo` is annotated with an object of type `HashMethod`.
Essentially, it represents a function that produces a hash that is used as part of the cache key calculation in calling the task `bar`.

#### How does caching of offloaded objects work?

Recall how task input values are taken into account to derive a cache key.
This is done by turning the literal representation into a string and using that string as part of the cache key.
In the case of dataframes annotated with `HashMethod` we use the hash as the representation of the Literal.
In other words, the literal hash is used in the cache key.
This feature also works in local execution.

Here’s a complete example of the feature:

```{code-block} python
def hash_pandas_dataframe(df: pandas.DataFrame) -> str:
    return str(pandas.util.hash_pandas_object(df))


@task
def uncached_data_reading_task() -> Annotated[pandas.DataFrame, HashMethod(hash_pandas_dataframe)]:
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
```



<!-- INCORPORATE THE FOLLOWING ABOVE WHERE NECESSARY

`cache`: Boolean that indicates if caching should be enabled

`cache_serialize`: Boolean that indicates if identical (ie. same inputs) instances of this task should be executed in serial when caching is enabled. This means that given multiple concurrent executions over identical inputs, only a single instance executes and the rest wait to reuse the cached results. This parameter does nothing without also setting the cache parameter.

`cache_version`: Cache version to use. Changes to the task signature will automatically trigger a cache miss, but you can always manually update this field as well to force a cache miss. You should also manually bump this version if the function body/business logic has changed, but the signature hasn’t.

`cache_ignore_input_vars`: Input variables that should not be included when calculating hash for cache.

As a result, tasks (or more precisely, their inputs and output) can be cached

They and can be reused in multiple workflows.
Tasks are executed in containers, which allows them to have their own independent set of software dependencies and hardware requirements.
Tasks can be chained together to form workflows, with the output of one task feeding into the input of the next.

Flyte supports memoization of task outputs to ensure that identical
invocations of a task are not executed repeatedly, thereby saving
compute resources and execution time. For example, if you wish to run
the same piece of code multiple times, you can reuse the output instead
of re-computing it. For more information on memoization, refer to the
`Caching Example </flytesnacks/examples/development_lifecycle/task_cache>`{.interpreted-text
role="std:doc"}.The most common

-->
