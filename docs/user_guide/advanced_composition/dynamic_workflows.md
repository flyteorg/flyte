(dynamic_workflow)=

# Dynamic workflows

```{eval-rst}
.. tags:: Intermediate
```

A workflow whose directed acyclic graph (DAG) is computed at run-time is a {py:func}`~flytekit.dynamic` workflow.
The tasks in a dynamic workflow are executed at runtime using dynamic inputs.
This type of workflow shares similarities with the {py:func}`~flytekit.workflow`, as it employs a Python-esque DSL
to declare dependencies between the tasks or define new workflows. A key distinction lies in the dynamic workflow being assessed at runtime.
This means that the inputs are initially materialized and forwarded to dynamic workflow, resembling the behavior of a task.
However, the return value from a dynamic workflow is a {py:class}`~flytekit.extend.Promise` object,
which can be materialized by the subsequent tasks.

Think of a dynamic workflow as a combination of a task and a workflow.
It is used to dynamically decide the parameters of a workflow at runtime.
It is both compiled and executed at run-time. You can define a dynamic workflow using the `@dynamic` decorator.

Within the `@dynamic` context, each invocation of a {py:func}`~flytekit.task` or a derivative of
{py:class}`~flytekit.core.base_task.Task` class leads to deferred evaluation using a promise,
rather than the immediate materialization of the actual value. While nesting other `@dynamic` and
`@workflow` constructs within this task is possible, direct interaction with the outputs of a task/workflow is limited,
as they are lazily evaluated. If interaction with the outputs is desired, it is recommended to separate the
logic in a dynamic workflow and create a new task to read and resolve the outputs.

Dynamic workflows become essential when you require:

- Modifying the logic of the code at runtime
- Changing or deciding on feature extraction parameters on-the-go
- Building AutoML pipelines
- Tuning hyperparameters during execution

This example utilizes dynamic workflow to count the common characters between any two strings.

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

To begin, we import the required libraries.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/dynamic_workflow.py
:caption: advanced_composition/dynamic_workflow.py
:lines: 1
```

We define a task that returns the index of a character, where A-Z/a-z is equivalent to 0-25.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/dynamic_workflow.py
:caption: advanced_composition/dynamic_workflow.py
:pyobject: return_index
```

We also create a task that prepares a list of 26 characters by populating the frequency of each character.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/dynamic_workflow.py
:caption: advanced_composition/dynamic_workflow.py
:pyobject: update_list
```

We define a task to calculate the number of common characters between the two strings.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/dynamic_workflow.py
:caption: advanced_composition/dynamic_workflow.py
:pyobject: derive_count
```

We define a dynamic workflow to accomplish the following:

1. Initialize an empty 26-character list to be passed to the `update_list` task
2. Iterate through each character of the first string (`s1`) and populate the frequency list
3. Iterate through each character of the second string (`s2`) and populate the frequency list
4. Determine the number of common characters by comparing the two frequency lists

The looping process is contingent on the number of characters in both strings, which is unknown until runtime.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/dynamic_workflow.py
:caption: advanced_composition/dynamic_workflow.py
:pyobject: count_characters
```

A dynamic workflow is modeled as a task in the backend,
but the body of the function is executed to produce a workflow at run-time.
In both dynamic and static workflows, the output of tasks are promise objects.

Propeller executes the dynamic task within its Kubernetes pod, resulting in a compiled DAG, which is then accessible in the console.
It utilizes the information acquired during the dynamic task's execution to schedule and execute each node within the dynamic task.
Visualization of the dynamic workflow's graph in the UI becomes available only after the dynamic task has completed its execution.

When a dynamic task is executed, it generates the entire workflow as its output, termed the *futures file*.
This nomenclature reflects the anticipation that the workflow is yet to be executed, and all subsequent outputs are considered futures.

:::{note}
Local execution works when a `@dynamic` decorator is used because Flytekit treats it as a task that runs with native Python inputs.
:::

Define a workflow that triggers the dynamic workflow.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/dynamic_workflow.py
:caption: advanced_composition/dynamic_workflow.py
:pyobject: dynamic_wf
```

You can run the workflow locally as follows:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/dynamic_workflow.py
:caption: advanced_composition/dynamic_workflow.py
:lines: 78-79
```

## Why use dynamic workflows?

### Flexibility

Dynamic workflows streamline the process of building pipelines, offering the flexibility to design workflows
according to the unique requirements of your project. This level of adaptability is not achievable with static workflows.

### Lower pressure on etcd

The workflow Custom Resource Definition (CRD) and the states associated with static workflows are stored in etcd,
the Kubernetes database. This database maintains Flyte workflow CRDs as key-value pairs, tracking the status of each node's execution.

However, there is a limitation with etcd — a hard limit on data size, encompassing the workflow and node status sizes.
Consequently, it's crucial to ensure that static workflows don't excessively consume memory.

In contrast, dynamic workflows offload the workflow specification (including node/task definitions and connections) to the blobstore.
Still, the statuses of nodes are stored in the workflow CRD within etcd.

Dynamic workflows help alleviate some of the pressure on etcd storage space, providing a solution to mitigate storage constraints.

## Dynamic workflows vs. map tasks

Dynamic tasks come with overhead for large fan-out tasks as they store metadata for the entire workflow.
In contrast, {ref}`map tasks <map_task>` prove efficient for such extensive fan-out scenarios since they refrain from storing metadata,
resulting in less noticeable overhead.

(advanced_merge_sort)=
## Merge sort

Merge sort is a perfect example to showcase how to seamlessly achieve recursion using dynamic workflows.
Flyte imposes limitations on the depth of recursion to prevent misuse and potential impacts on the overall stability of the system.

```python
from typing import Tuple

from flytekit import conditional, dynamic, task, workflow


@task
def split(numbers: list[int]) -> Tuple[list[int], list[int], int, int]:
    return (
        numbers[0 : int(len(numbers) / 2)],
        numbers[int(len(numbers) / 2) :],
        int(len(numbers) / 2),
        int(len(numbers)) - int(len(numbers) / 2),
    )


@task
def merge(sorted_list1: list[int], sorted_list2: list[int]) -> list[int]:
    result = []
    while len(sorted_list1) > 0 and len(sorted_list2) > 0:
        # Compare the current element of the first array with the current element of the second array.
        # If the element in the first array is smaller, append it to the result and increment the first array index.
        # Otherwise, do the same with the second array.
        if sorted_list1[0] < sorted_list2[0]:
            result.append(sorted_list1.pop(0))
        else:
            result.append(sorted_list2.pop(0))

    # Extend the result with the remaining elements from both arrays
    result.extend(sorted_list1)
    result.extend(sorted_list2)

    return result


@task
def sort_locally(numbers: list[int]) -> list[int]:
    return sorted(numbers)


@dynamic
def merge_sort_remotely(numbers: list[int], run_local_at_count: int) -> list[int]:
    split1, split2, new_count1, new_count2 = split(numbers=numbers)
    sorted1 = merge_sort(numbers=split1, numbers_count=new_count1, run_local_at_count=run_local_at_count)
    sorted2 = merge_sort(numbers=split2, numbers_count=new_count2, run_local_at_count=run_local_at_count)
    return merge(sorted_list1=sorted1, sorted_list2=sorted2)


@workflow
def merge_sort(numbers: list[int], numbers_count: int, run_local_at_count: int = 5) -> list[int]:
    return (
        conditional("terminal_case")
        .if_(numbers_count <= run_local_at_count)
        .then(sort_locally(numbers=numbers))
        .else_()
        .then(merge_sort_remotely(numbers=numbers, run_local_at_count=run_local_at_count))
    )
```

By simply adding the `@dynamic` annotation, the `merge_sort_remotely` function transforms into a plan of execution,
generating a Flyte workflow with four distinct nodes. These nodes run remotely on potentially different hosts,
with Flyte ensuring proper data reference passing and maintaining execution order with maximum possible parallelism.

`@dynamic` is essential in this context because the number of times `merge_sort` needs to be triggered is unknown at compile time.
The dynamic workflow calls a static workflow, which subsequently calls the dynamic workflow again,
creating a recursive and flexible execution structure.

## Run the example on the Flyte cluster

To run the provided workflows on the Flyte cluster, you can use the following commands:

```
pyflyte run --remote \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/dynamic_workflow.py \
  dynamic_wf --s1 "Pear" --s2 "Earth"
```

```
pyflyte run --remote \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/dynamic_workflow.py \
  merge_sort --numbers '[1813, 3105, 3260, 2634, 383, 7037, 3291, 2403, 315, 7164]' --numbers_count 10
```

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/advanced_composition/
