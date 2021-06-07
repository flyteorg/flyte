"""
Dynamic Workflows
-----------------

A workflow is typically static where the directed acyclic graph's (DAG) structure is known at compile-time. However, scenarios exist where a run-time parameter (e.g. the output of an earlier task) determines the full DAG structure.

In such cases, dynamic workflows can be used. Here's a code example that counts the common characters between any two strings.

Inputs: s1 = "Pear", s2 = "Earth"

Output: 3

"""

# %%
# Let's first import all the required libraries.
import typing

from flytekit import dynamic, task, workflow

# %%
# Next, we write a task that returns the index of a character (A-Z/a-z is equivalent to 0 to 25).
@task
def return_index(character: str) -> int:
    """
    Computes the character index (which needs to fit into the 26 characters list)"""
    if character.islower():
        return ord(character) - ord("a")
    else:
        return ord(character) - ord("A")


# %%
# We now write a task that prepares the 26-character list by populating the frequency of every character.
@task
def update_list(freq_list: typing.List[int], list_index: int) -> typing.List[int]:
    """
    Notes the frequency of characters"""
    freq_list[list_index] += 1
    return freq_list


# %%
# Next we find the number of common characters between the two strings.
@task
def derive_count(freq1: typing.List[int], freq2: typing.List[int]) -> int:
    """
    Derives the number of common characters"""
    count = 0
    for i in range(26):
        count += min(freq1[i], freq2[i])
    return count


# %%
# In this step, we perform the following:
#
# #. Initialize the empty 26-character list to be sent to the ``update_list`` task
# #. Loop through every character of the first string (s1) and populate the frequency list
# #. Loop through every character of the second string (s2) and populate the frequency list
# #. Derive the number of common characters by comparing the two frequency lists
#
# The looping is dependent on the number of characters of both the strings which aren't known until the run time. If the ``@task`` decorator is used to encapsulate the calls mentioned above, the compilation will fail very early on due to the absence of the literal values.
# Therefore, ``@dynamic`` decorator has to be used.
#
# Dynamic workflow is effectively both a task and a workflow. The key thing to note is that the ``body of tasks is run at run time and the
# body of workflows is run at compile (aka registration) time``. Essentially, this is what a dynamic workflow leverages -- it’s a workflow that is compiled at run time (the best of both worlds)!
#
# At execution (run) time, Flytekit runs the compilation step, and produces
# a ``WorkflowTemplate`` (from the dynamic workflow), which Flytekit then passes back to Flyte Propeller for further running, exactly how sub-workflows are handled.
#
# .. note:: For iterating over a list, the dynamic pattern is not always the most efficient method. `Map tasks <https://github.com/flyteorg/flytekit/blob/8528268a29a07fe7e9ce9f7f08fea68c41b6a60b/flytekit/core/map_task.py/>`_ might be more efficient in certain cases, keeping in mind they only work for Python tasks (tasks decorated with the @task decorator, not sql/spark/etc). 
#
# We now define the dynamic workflow encapsulating the above mentioned points.
@dynamic
def count_characters(s1: str, s2: str) -> int:
    """
    Calls the required tasks and returns the final result"""

    # s1 and s2 are accessible

    # initiliazing an empty list consisting of 26 empty slots corresponding to every alphabet (lower and upper case)
    freq1 = [0] * 26
    freq2 = [0] * 26

    # looping through the string s1
    for i in range(len(s1)):

        # index and freq1 are not accesible as they are promises
        index = return_index(character=s1[i])
        freq1 = update_list(freq_list=freq1, list_index=index)

    # looping through the string s2
    for i in range(len(s2)):

        # index and freq2 are not accesible as they are promises
        index = return_index(character=s2[i])
        freq2 = update_list(freq_list=freq2, list_index=index)

    # counting the common characters
    return derive_count(freq1=freq1, freq2=freq2)


# %%
# When tasks are called within any workflow, they return Promise objects. Likewise, in a dynamic workflow, the tasks' outputs are Promise objects that cannot be directly accessed (they shall be fulfilled by Flyte later).
# Because of this fact, operations on the ``index`` variable like ``index + 1`` are not valid.
# To manage this problem, the values need to be passed to the other tasks to unwrap them.
#
# .. note:: The local execution will work when a ``@dynamic`` decorator is used because Flytekit treats it like a ``task`` that will run with the Python native inputs. 
# Therefore, there are no Promise objects locally within the function decorated with ``@dynamic`` as it is treated as a ``task``.

# %%
# Finally, we define a workflow that calls the dynamic workflow.
@workflow
def wf(s1: str, s2: str) -> int:
    """
    Calls the dynamic workflow and returns the result"""

    # sending two strings to the workflow
    return count_characters(s1=s1, s2=s2)


if __name__ == "__main__":
    print(wf(s1="Pear", s2="Earth"))
