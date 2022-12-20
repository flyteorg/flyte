"""
.. _advanced_merge_sort:

Implementing Merge Sort
-----------------------

.. tags:: Intermediate

FlyteIdl (the fundamental building block of the Flyte Language) allows various programming language features:
conditionals, recursion, custom typing, and more.

This tutorial will walk you through writing a simple Distributed Merge Sort algorithm. It'll show usage of conditions
as well as recursion using dynamically generated workflows. Flyte imposes limitation on the depth of the recursion to
avoid mis-use and potentially affecting the overall stability of the system.
"""

import typing
from datetime import datetime
from random import random, seed
from typing import Tuple

from flytekit import conditional, dynamic, task, workflow

# seed random number generator
seed(datetime.now().microsecond)


# %%
# A simple split function that divides a list into two halves.


@task
def split(numbers: typing.List[int]) -> Tuple[typing.List[int], typing.List[int], int]:
    return (
        numbers[0 : int(len(numbers) / 2)],
        numbers[int(len(numbers) / 2) :],
        int(len(numbers) / 2),
    )


# %%
# One sample implementation for merging. In a more real world example, this might merge file streams and only load
# chunks into the memory.
@task
def merge(
    sorted_list1: typing.List[int], sorted_list2: typing.List[int]
) -> typing.List[int]:
    result = []
    while len(sorted_list1) > 0 and len(sorted_list2) > 0:
        # Check if current element of first array is smaller than current element of second array. If yes,
        # store first array element and increment first array index. Otherwise do same with second array
        if sorted_list1[0] < sorted_list2[0]:
            result.append(sorted_list1.pop(0))
        else:
            result.append(sorted_list2.pop(0))

    result.extend(sorted_list1)
    result.extend(sorted_list2)

    return result


# %%
# Generally speaking, the algorithm will recurse through the list, splitting it in half until it reaches a size that we
# know is efficient enough to run locally. At which point it'll just use the python-builtin sorted function.

# %%
# This runs the sorting completely locally. It's faster and more efficient to do so if the entire list fits in memory.
@task
def sort_locally(numbers: typing.List[int]) -> typing.List[int]:
    return sorted(numbers)


# %%
# Let's now define the typical merge sort algorithm. We split, merge-sort each half then finally merge. With the simple
# addition of the `@dynamic` annotation, this function will instead generate a plan of execution (a flyte workflow) with
# 4 different nodes that will all run remotely on potentially different hosts. Flyte takes care of ensuring references
# of data are properly passed around and order of execution is maintained with maximum possible parallelism.
@dynamic
def merge_sort_remotely(
    numbers: typing.List[int], run_local_at_count: int
) -> typing.List[int]:
    split1, split2, new_count = split(numbers=numbers)
    sorted1 = merge_sort(
        numbers=split1, numbers_count=new_count, run_local_at_count=run_local_at_count
    )
    sorted2 = merge_sort(
        numbers=split2, numbers_count=new_count, run_local_at_count=run_local_at_count
    )
    return merge(sorted_list1=sorted1, sorted_list2=sorted2)


# %%
# Putting it all together, this is the workflow that also serves as the entry point of execution. Given an unordered set
# of numbers, their length as well as the size at which to sort locally, it runs a condition on the size. The condition
# should look and flow naturally to a python developer. Binary arithmetic and logical operations on simple types as well
# as logical operations on conditions are supported. This condition checks if the current size of the numbers is below
# the cut-off size to run locally, if so, it runs the sort_locally task. Otherwise it runs the above dynamic workflow
# that recurse down the list.
@workflow
def merge_sort(
    numbers: typing.List[int], numbers_count: int, run_local_at_count: int = 10
) -> typing.List[int]:
    return (
        conditional("terminal_case")
        .if_(numbers_count <= run_local_at_count)
        .then(sort_locally(numbers=numbers))
        .else_()
        .then(
            merge_sort_remotely(numbers=numbers, run_local_at_count=run_local_at_count)
        )
    )


# %%
# A helper function to generate inputs for running the workflow locally.
def generate_inputs(numbers_count: int) -> typing.List[int]:
    generated_list = []
    # generate random numbers between 0-1
    for _ in range(numbers_count):
        value = int(random() * 10000)
        generated_list.append(value)

    return generated_list


# %%
# The entire workflow can be executed locally as follows...
if __name__ == "__main__":
    count = 20
    x = generate_inputs(count)
    print(x)
    print(f"Running Merge Sort Locally...{merge_sort(numbers=x, numbers_count=count)}")
