---
jupytext:
  cell_metadata_filter: all
  formats: md:myst
  main_language: python
  notebook_metadata_filter: all
  text_representation:
    extension: .md
    format_name: myst
    format_version: 0.13
    jupytext_version: 1.16.1
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

+++ {"lines_to_next_cell": 0}

(nested_parallelization)=

# Nested Parallelization

```{eval-rst}
.. tags:: Advanced
```

For exceptionally large or complicated workflows where dynamic workflows or map tasks aren't enough, it can be benficial to have multiple levels of parallelization.

This is useful for multiple reasons:
1. Better code organization
2. Better code reuse
3. Better testing
4. Better debugging
5. Better monitoring - each subworkflow can be run independently and monitored independently
6. Better performance and scale - each subworkflow is executed as a separate workflow and thus can land on different flytepropeller workers and shards. This allows for better parallelism and scale.

## Nested Dynamics

This example shows how to break down a large workflow into smaller workflows and then compose them together to form a hierarchy. 

```python
"""
The structure for 6 items and a chunk size of 2, the workflow will be broken down as follows:

multi_wf -> level1 -> level2 -> core_wf -> step1 -> step2
                             -> core_wf -> step1 -> step2
                      level2 -> core_wf -> step1 -> step2
                             -> core_wf -> step1 -> step2
                      level2 -> core_wf -> step1 -> step2
                             -> core_wf -> step1 -> step2
"""

import typing
from flytekit import task, workflow, dynamic, LaunchPlan


@task
def step1(a: int) -> int:
    return a + 1


@task
def step2(a: int) -> int:
    return a + 2


@workflow
def core_wf(a: int) -> int:
    return step2(a=step1(a=a))


core_wf_lp = LaunchPlan.get_or_create(core_wf)


@dynamic
def level2(l: typing.List[int]) -> typing.List[int]:
    return [core_wf_lp(a=a) for a in l]


@task
def reduce(l: typing.List[typing.List[int]]) -> typing.List[int]:
    f = []
    for i in l:
        f.extend(i)
    return f


@dynamic
def level1(l: typing.List[int], chunk: int) -> typing.List[int]:
    v = []
    for i in range(0, len(l), chunk):
        v.append(level2(l=l[i:i + chunk]))
    return reduce(l=v)


@workflow
def multi_wf(l: typing.List[int], chunk: int) -> typing.List[int]:
    return level1(l=l, chunk=chunk)
```

+++ {"lines_to_next_cell": 0}

This shows a top-level workflow which uses 2 levels of dynamic workflows to process a list through some simple addition tasks and then flatten it again. Here is a visual representation of the execution in a Flyte console:

:::{figure} https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/user_guide/nested_parallel_top_level.png?raw=true
:alt: Nested Parallelization UI View
:class: with-shadow
:::

For each node in that top-level we can see 2 sub-workflows being run in parallel. 

:::{figure} https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/user_guide/nested_parallel_inner_dynamic.png?raw=true
:alt: Inner Dynamic
:class: with-shadow
:::

Finally, drilling into each sub-workflow, we'll see both those tasks being executed in series.

:::{figure} https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/user_guide/nested_parallel_subworkflow.png?raw=true
:alt: Sub-Workflow
:class: with-shadow
:::

## Mixed Parallelism

This example is similar to the above but instead of using a dynamic to parallelize a 2-task serial workflow, we're using that workflow to call a map task which processes both inputs in parallel. This workflow has one fewer layers of parallelism so the outputs won't be the same, but it does still demonstrate how you can mix these different approaches to achieving concurrency.

```python
"""
The structure for 6 items and a chunk size of 2, the workflow will be broken down as follows:

multi_wf -> level1 -> level2 -> mappable
                             -> mappable
                      level2 -> mappable
                             -> mappable
                      level2 -> mappable
                             -> mappable
"""
import typing
from flytekit import task, workflow, dynamic, map_task


@task
def mappable(a: int) -> int:
    return a + 2


@workflow
def level2(l: typing.List[int]) -> typing.List[int]:
    return map_task(mappable)(a=l)


@task
def reduce(l: typing.List[typing.List[int]]) -> typing.List[int]:
    f = []
    for i in l:
        f.extend(i)
    return f


@dynamic
def level1(l: typing.List[int], chunk: int) -> typing.List[int]:
    v = []
    for i in range(0, len(l), chunk):
        v.append(level2(l=l[i : i + chunk]))
    return reduce(l=v)


@workflow
def multi_wf(l: typing.List[int], chunk: int) -> typing.List[int]:
    return level1(l=l, chunk=chunk)

```
+++ {"lines_to_next_cell": 0}

While the top-level dynamic will be exactly the same as the previous example, here you can see the inner map task nodes as well as links in the sidebar.

:::{figure} https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/user_guide/nested_parallel_inner_map.png?raw=true
:alt: Inner Map Task
:class: with-shadow
:::

## Design Considerations

You can nest even further if needed, or incorporate map tasks if your inputs are all the same type. The design of your workflow should of course be informed by the actual data you're processing. For example if you have a big library of music that you'd like to extract the lyrics for, the first level can loop through all the albums and the second level can go through each song. 

If you're just churning through an enormous list of the same input, it's typically best to keep the code simple and let the scheduler handle optmizing the execution. Additionally, unless you need the features of a dynamic workflow like mixing and matching inputs and outputs, it's usually most efficient to use a map task. This has the added benefit of keeping the UI clean.

You can also choose to limit the scale of parallel execution at a couple levels. The `max_parallelism` attribute can be applied at the workflow level and will limit the number of parallel tasks being executed, this is set to 25 by default. Within map tasks specifically, you can indicate a `concurrency` argument which will limit the number of mapped tasks within running at any given time. 