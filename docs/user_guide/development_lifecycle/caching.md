# Caching

```{eval-rst}
.. tags:: Basic
```

Flyte provides the ability to cache the output of task executions to make the subsequent executions faster.

Task caching is useful when a user knows that many executions with the same inputs may occur. For example, consider the following scenarios:

- Running a task periodically on a schedule
- Running the code multiple times when debugging workflows
- Running the commonly shared tasks amongst different workflows, which receive the same inputs

Let's watch a brief explanation of caching and a demo in this video, followed by how task caching can be enabled.

```{eval-rst}
.. youtube:: WNkThCp-gqo

```

There are four parameters and one command-line flag related to caching.

## Parameters

* `cache`(`bool`): Enables or disables caching of the workflow, task, or launch plan.
By default, caching is disabled to avoid unintended consequences when caching executions with side effects.
To enable caching set `cache=True`.
* `cache_version` (`str`): Part of the cache key.
A change to this parameter will invalidate the cache.
This allows you to explicitly indicate when a change has been made to the task that should invalidate any existing cached results.
Note that this is not the only change that will invalidate the cache (see below).
Also, note that you can manually trigger cache invalidation per execution using the [`overwrite-cache` flag](#overwrite-cache-flag).
* `cache_serialize` (`bool`): Enables or disables [cache serialization](./cache_serializing).
When enabled, Flyte ensures that a single instance of the task is run before any other instances that would otherwise run concurrently.
This allows the initial instance to cache its result and lets the later instances reuse the resulting cached outputs.
Cache serialization is disabled by default.
* `cache_ignore_input_vars` (`Tuple[str, ...]`): Input variables that should not be included when calculating hash for cache. By default, no input variables are ignored. This parameter only applies to task serialization.

Task caching parameters can be specified at task definition time within `@task` decorator or at task invocation time using `with_overrides` method.

## Overwrite cache flag

*  `overwrite-cache` (`bool`): Invalidates the cache and forces re-execution of the task.

This flag can be used when launching an execution from [the command line](#overwrite-cache-on-the-command-line), [the UI](#overwrite-cache-in-the-ui), or programmatically [through `FlyteRemote`](#overwrite-cache-programmatically).

### Overwrite cache on the command line

The `overwrite-cache` flag can be used from the command line with the `pyflyte run` command. For example:

```{code-block} shell
$ pyflyte run --remote  --overwrite-cache example.py wf
```

### Overwrite cache in the UI

You can also trigger cache invalidation when launching an execution from the UI, in the launch modal, by checking the **Overwrite cache** checkbox.

### Overwrite cache programmatically

When using `FlyteRemote`, you can use the `overwrite_cache` parameter in the [`flytekit.remote.remote.FlyteRemote.execute`](https://docs.flyte.org/en/latest/api/flytekit/generated/flytekit.remote.remote.FlyteRemote.html#flytekit.remote.remote.FlyteRemote.execute) method.

## Examples

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

Import the necessary libraries:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/development_lifecycle/development_lifecycle/task_cache.py
:caption: development_lifecycle/task_cache.py
:lines: 1-3
```

For any {py:func}`flytekit.task` in Flyte, there is always one required import, which is:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/development_lifecycle/development_lifecycle/task_cache.py
:caption: development_lifecycle/task_cache.py
:lines: 8-10
```

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/development_lifecycle/development_lifecycle/task_cache.py
:caption: development_lifecycle/task_cache.py
:pyobject: square
```

In the above example, calling `square(n=2)` twice (even if it's across different executions or different workflows) will only execute the multiplication operation once.
The next time, the output will be made available immediately since it is captured from the previous execution with the same inputs.

If in a subsequent code update, you update the signature of the task to return the original number along with the result, it'll automatically invalidate the cache (even though the cache version remains the same).

```python
@task(cache=True, cache_version="1.0")
def square(n: int) -> Tuple[int, int]:
    ...
```

:::{note}
If the user changes the task interface in any way (such as adding, removing, or editing inputs/outputs), Flyte treats that as a task functionality change. In the subsequent execution, Flyte runs the task and stores the outputs as newly cached values.
:::

## How does caching work?

Caching is implemented differently depending on the mode the user is running, i.e. whether they are running locally or using remote Flyte.

### How does remote caching work?

The cache keys for remote task execution are composed of **Project**, **Domain**, **Cache Version**, **Task Signature**, and **Inputs** associated with the execution of the task, as per the following definitions:

- **Project:** A task run under one project cannot use the cached task execution from another project which would cause inadvertent results between project teams that could result in data corruption.
- **Domain:** To separate test, staging, and production data, task executions are not shared across these environments.
- **Cache Version:** When task functionality changes, you can change the `cache_version` of the task. Flyte will know not to use older cached task executions and create a new cache entry on the subsequent execution.
- **Task Signature:** The cache is specific to the task signature associated with the execution. The signature constitutes the task name, input parameter names/types, and the output parameter name/type.
- **Task Input Values:** A well-formed Flyte task always produces deterministic outputs. This means, given a set of input values, every execution should have identical outputs. When task execution is cached, the input values are part of the cache key.

The remote cache for a particular task is invalidated in three ways:

1. Modifying the `cache_version`;
2. Updating the task signature.
3. Using the `overwrite_cache` flag.

:::{note}
Task executions can be cached across different versions of the task because a change in SHA does not necessarily mean that it correlates to a change in the task functionality.
:::

### How does local caching work?

The flytekit package uses the [diskcache](https://github.com/grantjenks/python-diskcache) package, specifically [diskcache.Cache](http://www.grantjenks.com/docs/diskcache/tutorial.html#cache), to aid in the memoization of task executions. The results of local task executions are stored under `~/.flyte/local-cache/` and cache keys are composed of **Cache Version**, **Task Signature**, and **Task Input Values**.

Similar to the remote case, a local cache entry for a task will be invalidated if either the `cache_version` or the task signature is modified. In addition, the local cache can also be emptied by running the following command: `pyflyte local-cache clear`, which essentially obliterates the contents of the `~/.flyte/local-cache/` directory.
To disable the local cache, you can set the `local.cache_enabled` config option (e.g. by setting the environment variable `FLYTE_LOCAL_CACHE_ENABLED=False`).

:::{note}
The format used by the store is opaque and not meant to be inspectable.
:::

(cache-offloaded-objects)=

## Caching of non-Flyte offloaded objects

The default behavior displayed by Flyte's memoization feature might not match the user intuition. For example, this code makes use of pandas dataframes:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/development_lifecycle/development_lifecycle/task_cache.py
:caption: development_lifecycle/task_cache.py
:lines: 39-54
```

If run twice with the same inputs, one would expect that `bar` would trigger a cache hit, but it turns out that's not the case because of how dataframes are represented in Flyte.
However, with release 1.2.0, Flyte provides a new way to control memoization behavior of literals. This is done via a `typing.Annotated` call on the task signature.
For example, in order to cache the result of calls to `bar`, you can rewrite the code above like this:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/development_lifecycle/development_lifecycle/task_cache.py
:caption: development_lifecycle/task_cache.py
:lines: 64-85
```

Note how the output of task `foo` is annotated with an object of type `HashMethod`. Essentially, it represents a function that produces a hash that is used as part of the cache key calculation in calling the task `bar`.

### How does caching of offloaded objects work?

Recall how task input values are taken into account to derive a cache key.
This is done by turning the literal representation into a string and using that string as part of the cache key. In the case of dataframes annotated with `HashMethod` we use the hash as the representation of the Literal. In other words, the literal hash is used in the cache key.

This feature also works in local execution.

Here's a complete example of the feature:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/development_lifecycle/development_lifecycle/task_cache.py
:caption: development_lifecycle/task_cache.py
:lines: 97-134
```

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/development_lifecycle/
