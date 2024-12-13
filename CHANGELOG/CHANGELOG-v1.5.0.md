# Flyte 1.5 release

## Platform

We're laying the foundation for an improved experience to help performance investigations. Stay tuned for more details!

We can now submit Ray jobs to separate clusters (other than the one flytepropeller is running). Thanks to https://github.com/flyteorg/flyteplugins/pull/321.

Several bug fixes, including:
- [Fix fast-cache bug on first node event](https://github.com/flyteorg/flyteadmin/pull/483)
- [Split flyte-binary services into http and grpc in helm charts](https://github.com/flyteorg/flyte/pull/3518)

### Database Migrations
One of the improvements planned requires us to clean up our database migrations. We have done so in this release so you should see a series of new migrations.
These should have zero impact if you are otherwise up-to-date on migrations (which is why they are all labeled `noop`) but please be aware that it will add a minute or so to the
init container/command that runs the migrations in the default Helm charts. Notably, because these should be a no-op, they also do not come with any rollback commands.
If you experience any issues, please let us know.

## Flytekit

Python 3.11 is now officially supported.

### Revamped Data subsystem
The data persistence layer was completely revamped. We now rely exclusively on [fsspec](https://filesystem-spec.readthedocs.io/en/latest/) to handle IO.

Most users will benefit from a more performant IO subsystem, in other words,
no change is needed in user code.


The data persistence layer has undergone a thorough overhaul. We now exclusively utilize [fsspec](https://filesystem-spec.readthedocs.io/en/latest/) for managing input and output operations.

For the majority of users, the improved IO subsystem provides enhanced performance, meaning that no modifications are required in their existing code.

This change opened the door for flytekit to rely on fsspec streaming capabilities. For example, let's say we want to stream a file, now we're able to do:

```
@task
def copy_file(ff: FlyteFile) -> FlyteFile:
    new_file = FlyteFile.new_remote_file(ff.remote_path)
    with ff.open("r", cache_type="simplecache", cache_options={}) as r:
        with new_file.open("w") as w:
            w.write(r.read())
    return new_file
```

This feature is marked as experimental. We'd love feedback on the API!

### Limited support for partial tasks
We can use [functools.partial](https://docs.python.org/3/library/functools.html#functools.partial) to "freeze"
some task arguments. Let's take a look at an example where we partially fix the parameter for a task:

```
@task
def t1(a: int, b: str) -> str:
    return f"{a} -> {b}"

t1_fixed_b = functools.partial(t1, b="hello")

@workflow
def wf(a: int) -> str:
    return t1_fixed_b(a=a)
```

Notice how calls to `t1_fixed_b` do not need to specify the `b` parameter.

This also works for [Map Tasks](https://docs.flyte.org/en/latest/user_guide/advanced_composition/map_tasks.html) in a limited capacity. For example:

```
from flytekit import task, workflow, partial, map_task

@task
def t1(x: int, y: float) -> float:
    return x + y

@workflow
def wf(y: List[float]):
   partial_t1 = partial(t1, x=5)
   return map_task(partial_t1)(y=y)
```

We are currently seeking feedback on this feature, and as a result, it is labeled as experimental for now.

Also worth mentioning that fixing parameters of type list is not currently supported. For example, if we try to register this workflow:

```
from functools import partial
from typing import List
from flytekit import task, workflow, map_task

@task
def t(a: int, xs: List[int]) -> str:
    return f"{a} {xs}"

@workflow
def wf():
    partial_t = partial(t, xs=[1, 2, 3])
    map_task(partial_t)(a=[1, 2])
```

We're going to see this error:

```
❯ pyflyte run workflows/example.py wf
Failed with Unknown Exception <class 'ValueError'> Reason: Map tasks do not support partial tasks with lists as inputs.
Map tasks do not support partial tasks with lists as inputs.
```

## Flyteconsole

Multiple bug fixes around [waiting for external inputs](https://docs.flyte.org/en/latest/user_guide/advanced_composition/waiting_for_external_inputs.html).
Better support for dataclasses in the launch form.
