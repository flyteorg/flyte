"""
Chain Flyte Tasks
-----------------

Data passing between tasks need not always happen through parameters.
Flyte provides a mechanism to chain tasks using the ``>>`` operator and the ``create_node`` function.
You may want to call this function to specify dependencies between tasks that don't consume or produce outputs.

In this example, let's enforce an order for ``read()`` to happen after ``write()``.
"""

import pandas as pd

# %%
# First, we import the necessary dependencies.
from flytekit import task, workflow
from flytekit.core.node_creation import create_node

DATABASE = "https://raw.githubusercontent.com/mwaskom/seaborn-data/master/iris.csv"
# %%
# We define a ``read()`` task to read from the file.


@task
def read() -> pd.DataFrame:
    data = pd.read_csv(DATABASE)
    return data


# %%
# We define a ``write()`` task to write to the file. Let's assume we are populating the CSV file.
@task
def write():
    # dummy code
    df = pd.DataFrame(  # noqa : F841
        data={
            "sepal_length": [5.3],
            "sepal_width": [3.8],
            "petal_length": [0.1],
            "petal_width": [0.3],
            "species": ["setosa"],
        }
    )
    # we write the data to a database
    # pd.to_csv("...")


# %%
# We want to enforce an order here: ``write()`` followed by ``read()``.
# Since no data-passing happens between the tasks, we use ``>>`` operator on the nodes.
@workflow
def chain_tasks_wf() -> pd.DataFrame:
    write_node = create_node(write)
    read_node = create_node(read)

    write_node >> read_node

    return read_node.o0


# %%
# .. note::
#   To send arguments while creating a node, use the following syntax:
#
#   .. code-block:: python
#
#       create_node(task_name, parameter1=argument1, parameter2=argument2, ...)

# %%
# Finally, we can run the workflow locally.
if __name__ == "__main__":
    print(f"Running {__file__} main...")
    print(f"Running chain_tasks_wf()... {chain_tasks_wf()}")
