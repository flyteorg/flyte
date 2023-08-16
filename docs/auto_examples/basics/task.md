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
    jupytext_version: 1.14.7
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

+++ {"lines_to_next_cell": 0}

(basics_of_tasks)=

# Tasks

```{eval-rst}
.. tags:: Basic
```

Task is a fundamental building block and an extension point of Flyte, which encapsulates the users' code. They possess the following properties:

1. Versioned (usually tied to the `git sha`)
2. Strong interfaces (specified inputs and outputs)
3. Declarative
4. Independently executable
5. Unit testable

A task in Flytekit can be of two types:

1. A task that has a Python function associated with it. The execution of the task is equivalent to the execution of this function.
2. A task that doesn't have a Python function, e.g., an SQL query or any portable task like Sagemaker prebuilt algorithms, or a service that invokes an API.

Flyte provides multiple plugins for tasks, which can be a backend plugin as well ([Athena](https://github.com/flyteorg/flytekit/blob/master/plugins/flytekit-aws-athena/flytekitplugins/athena/task.py)).

In this example, you will learn how to write and execute a `Python function task`. Other types of tasks will be covered in the later sections.

+++ {"lines_to_next_cell": 0}

For any task in Flyte, there is one necessary import, which is:

```{code-cell}
from flytekit import task
```

```{code-cell}
:lines_to_next_cell: 1

# Importing additional modules.
from sklearn.datasets import load_iris
from sklearn.linear_model import LogisticRegression
from sklearn.model_selection import train_test_split
```

The use of the {py:func}`flytekit.task` decorator is mandatory for a ``PythonFunctionTask``.
A task is essentially a regular Python function, with the exception that all inputs and outputs must be clearly annotated with their types.
These types are standard Python types, which will be further explained in the {ref}`type-system section <flytekit_to_flyte_type_mapping>`.

```{code-cell}
@task
def train_model(hyperparameters: dict, test_size: float, random_state: int) -> LogisticRegression:
    """
    Parameters:
        hyperparameters (dict): A dictionary containing the hyperparameters for the model.
        test_size (float): The proportion of the data to be used for testing.
        random_state (int): The random seed for reproducibility.

    Return:
        LogisticRegression: The trained logistic regression model.
    """
    # Loading the Iris dataset
    iris = load_iris()

    # Splitting the data into train and test sets
    X_train, _, y_train, _ = train_test_split(iris.data, iris.target, test_size=test_size, random_state=random_state)

    # Creating and training the logistic regression model with the given hyperparameters
    clf = LogisticRegression(**hyperparameters)
    clf.fit(X_train, y_train)

    return clf
```

+++ {"lines_to_next_cell": 0}

:::{note}
Flytekit will assign a default name to the output variable like `out0`.
In case of multiple outputs, each output will be numbered in the order
starting with 0, e.g., -> `out0, out1, out2, ...`.
:::

You can execute a Flyte task as any normal function.

```{code-cell}
if __name__ == "__main__":
    print(train_model(hyperparameters={"C": 0.1}, test_size=0.2, random_state=42))
```

## Invoke a Task within a Workflow

The primary way to use Flyte tasks is to invoke them in the context of a workflow.

```{code-cell}
from flytekit import workflow


@workflow
def train_model_wf(
    hyperparameters: dict = {"C": 0.1}, test_size: float = 0.2, random_state: int = 42
) -> LogisticRegression:
    """
    This workflow invokes the train_model task with the given hyperparameters, test size and random state.
    """
    return train_model(hyperparameters=hyperparameters, test_size=test_size, random_state=random_state)
```

```{note}
When invoking the `train_model` task, you need to use keyword arguments to specify the values for the corresponding parameters.
````

## Use `partial` to provide default arguments to tasks

You can use the {py:func}`functools.partial` function to assign default or constant values to the parameters of your tasks.

```{code-cell}
:lines_to_next_cell: 2

import functools


@workflow
def train_model_wf_with_partial(test_size: float = 0.2, random_state: int = 42) -> LogisticRegression:
    partial_task = functools.partial(train_model, hyperparameters={"C": 0.1})
    return partial_task(test_size=test_size, random_state=random_state)
```

In this toy example, we're calling the `square` task twice and returning the result.

+++

(single_task_execution)=

:::{dropdown} Execute a single task *without* a workflow

While workflows are typically composed of multiple tasks with dependencies defined by shared inputs and outputs,
there are cases where it can be beneficial to execute a single task in isolation during the process of developing and iterating on its logic.
Writing a new workflow definition every time for this purpose can be cumbersome, but executing a single task without a workflow provides a convenient way to iterate on task logic easily.

To run a task without a workflow, use the following command:

```{code-block}
pyflyte run task.py train_model --hyperparameters '{"C": 0.1}' --test_size 0.2 --random_state 42
```
:::
