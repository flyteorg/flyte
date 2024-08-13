# Cache serializing

```{eval-rst}
.. tags:: Intermediate
```

Serializing means only executing a single instance of a unique cacheable task (determined by the cache_version parameter and task signature) at a time. Using this mechanism, Flyte ensures that during multiple concurrent executions of a task only a single instance is evaluated and all others wait until completion and reuse the resulting cached outputs.

Ensuring serialized evaluation requires a small degree of overhead to coordinate executions using a lightweight artifact reservation system. Therefore, this should be viewed as an extension to rather than a replacement for non-serialized cacheable tasks. It is particularly well fit for long running or otherwise computationally expensive tasks executed in scenarios similar to the following examples:

- Periodically scheduled workflow where a single task evaluation duration may span multiple scheduled executions.
- Running a commonly shared task within different workflows (which receive the same inputs).

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

For any {py:func}`flytekit.task` in Flyte, there is always one required import, which is:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/development_lifecycle/development_lifecycle/task_cache_serialize.py
:caption: development_lifecycle/task_cache_serialize.py
:lines: 1
```

Task cache serializing is disabled by default to avoid unexpected behavior for task executions. To enable use the `cache_serialize` parameter.
`cache_serialize` is a switch to enable or disable serialization of the task
This operation is only useful for cacheable tasks, where one may reuse output from a previous execution. Flyte requires implicitly enabling the `cache` parameter on all cache serializable tasks.
Cache key definitions follow the same rules as non-serialized cache tasks. It is important to understand the implications of the task signature and `cache_version` parameter in defining cached results.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/development_lifecycle/development_lifecycle/task_cache_serialize.py
:caption: development_lifecycle/task_cache_serialize.py
:pyobject: square
```

In the above example calling `square(n=2)` multiple times concurrently (even in different executions or workflows) will only execute the multiplication operation once.
Concurrently evaluated tasks will wait for completion of the first instance before reusing the cached results and subsequent evaluations will instantly reuse existing cache results.

## How does serializing caches work?

The cache serialize paradigm introduces a new artifact reservation system. Tasks may use this reservation system to acquire an artifact reservation, indicating that they are actively evaluating the task, and release the reservation, once the execution is completed. Flyte uses a clock-skew algorithm to define reservation timeouts. Therefore, tasks are required to periodically extend the reservation during execution.

The first execution of a serializable cached task will successfully acquire the artifact reservation. Execution will be performed as usual and upon completion, the results are written to the cache and reservation is released. Concurrently executed task instances (i.e. in parallel with the initial execution) will observe an active reservation, in which case the execution will wait until the next reevaluation and perform another check. Once the initial execution completes it will reuse the cached results. Subsequently executed task instances (i.e. after an execution has already completed successfully) will immediately reuse the existing cached results.

Flyte handles task execution failures using a timeout on the reservation. If the task currently holding the reservation fails to extend it before it times out, another task may acquire the reservation and begin executing the task.

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/development_lifecycle/
