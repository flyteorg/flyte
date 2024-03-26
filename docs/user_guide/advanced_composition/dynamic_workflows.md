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

To begin, we import the required libraries.

```{code-cell}
from flytekit import dynamic, task, workflow
```

+++ {"lines_to_next_cell": 0}

We define a task that returns the index of a character, where A-Z/a-z is equivalent to 0-25.

```{code-cell}
@task
def return_index(character: str) -> int:
    if character.islower():
        return ord(character) - ord("a")
    else:
        return ord(character) - ord("A")
```

+++ {"lines_to_next_cell": 0}

We also create a task that prepares a list of 26 characters by populating the frequency of each character.

```{code-cell}
@task
def update_list(freq_list: list[int], list_index: int) -> list[int]:
    freq_list[list_index] += 1
    return freq_list
```

+++ {"lines_to_next_cell": 0}

We define a task to calculate the number of common characters between the two strings.

```{code-cell}
@task
def derive_count(freq1: list[int], freq2: list[int]) -> int:
    count = 0
    for i in range(26):
        count += min(freq1[i], freq2[i])
    return count
```

+++ {"lines_to_next_cell": 0}

We define a dynamic workflow to accomplish the following:

1. Initialize an empty 26-character list to be passed to the `update_list` task
2. Iterate through each character of the first string (`s1`) and populate the frequency list
3. Iterate through each character of the second string (`s2`) and populate the frequency list
4. Determine the number of common characters by comparing the two frequency lists

The looping process is contingent on the number of characters in both strings, which is unknown until runtime.

```{code-cell}
@dynamic
def count_characters(s1: str, s2: str) -> int:
    # s1 and s2 should be accessible

    # Initialize empty lists with 26 slots each, corresponding to every alphabet (lower and upper case)
    freq1 = [0] * 26
    freq2 = [0] * 26

    # Loop through characters in s1
    for i in range(len(s1)):
        # Calculate the index for the current character in the alphabet
        index = return_index(character=s1[i])
        # Update the frequency list for s1
        freq1 = update_list(freq_list=freq1, list_index=index)
        # index and freq1 are not accessible as they are promises

    # looping through the string s2
    for i in range(len(s2)):
        # Calculate the index for the current character in the alphabet
        index = return_index(character=s2[i])
        # Update the frequency list for s2
        freq2 = update_list(freq_list=freq2, list_index=index)
        # index and freq2 are not accessible as they are promises

    # Count the common characters between s1 and s2
    return derive_count(freq1=freq1, freq2=freq2)
```

+++ {"lines_to_next_cell": 0}

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

```{code-cell}
@workflow
def dynamic_wf(s1: str, s2: str) -> int:
    return count_characters(s1=s1, s2=s2)
```

+++ {"lines_to_next_cell": 0}

You can run the workflow locally as follows:

```{code-cell}
:lines_to_next_cell: 2

if __name__ == "__main__":
    print(dynamic_wf(s1="Pear", s2="Earth"))
```

+++ {"lines_to_next_cell": 0}

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

```{code-cell}
:lines_to_next_cell: 2

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
  https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/advanced_composition/advanced_composition/dynamic_workflow.py \
  dynamic_wf --s1 "Pear" --s2 "Earth"
```

```
pyflyte run --remote \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/advanced_composition/advanced_composition/dynamic_workflow.py \
  merge_sort --numbers '[1813, 3105, 3260, 2634, 383, 7037, 3291, 2403, 315, 7164]' --numbers_count 10
```
