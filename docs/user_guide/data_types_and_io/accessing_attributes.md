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

(attribute_access)=

# Accessing attributes

```{eval-rst}
.. tags:: Basic
```

You can directly access attributes on output promises for lists, dicts, dataclasses and combinations of these types in Flyte.
This functionality facilitates the direct passing of output attributes within workflows,
enhancing the convenience of working with complex data structures.

To begin, import the required dependencies and define a common task for subsequent use.

```{code-cell}
from dataclasses import dataclass

from dataclasses_json import dataclass_json
from flytekit import task, workflow


@task
def print_message(message: str):
    print(message)
    return
```

+++ {"lines_to_next_cell": 0}

## List
You can access an output list using index notation.

:::{important}
Flyte currently does not support output promise access through list slicing.
:::

```{code-cell}
@task
def list_task() -> list[str]:
    return ["apple", "banana"]


@workflow
def list_wf():
    items = list_task()
    first_item = items[0]
    print_message(message=first_item)
```

+++ {"lines_to_next_cell": 0}

## Dictionary
Access the output dictionary by specifying the key.

```{code-cell}
@task
def dict_task() -> dict[str, str]:
    return {"fruit": "banana"}


@workflow
def dict_wf():
    fruit_dict = dict_task()
    print_message(message=fruit_dict["fruit"])
```

+++ {"lines_to_next_cell": 0}

## Data class
Directly access an attribute of a dataclass.

```{code-cell}
@dataclass_json
@dataclass
class Fruit:
    name: str


@task
def dataclass_task() -> Fruit:
    return Fruit(name="banana")


@workflow
def dataclass_wf():
    fruit_instance = dataclass_task()
    print_message(message=fruit_instance.name)
```

+++ {"lines_to_next_cell": 0}

## Complex type
Combinations of list, dict and dataclass also work effectively.

```{code-cell}
@task
def advance_task() -> (dict[str, list[str]], list[dict[str, str]], dict[str, Fruit]):
    return {"fruits": ["banana"]}, [{"fruit": "banana"}], {"fruit": Fruit(name="banana")}


@task
def print_list(fruits: list[str]):
    print(fruits)


@task
def print_dict(fruit_dict: dict[str, str]):
    print(fruit_dict)


@workflow
def advanced_workflow():
    dictionary_list, list_dict, dict_dataclass = advance_task()
    print_message(message=dictionary_list["fruits"][0])
    print_message(message=list_dict[0]["fruit"])
    print_message(message=dict_dataclass["fruit"].name)

    print_list(fruits=dictionary_list["fruits"])
    print_dict(fruit_dict=list_dict[0])
```

+++ {"lines_to_next_cell": 0}

You can run all the workflows locally as follows:

```{code-cell}
:lines_to_next_cell: 2

if __name__ == "__main__":
    list_wf()
    dict_wf()
    dataclass_wf()
    advanced_workflow()
```

## Failure scenario
The following workflow fails because it attempts to access indices and keys that are out of range:

```python
from flytekit import WorkflowFailurePolicy


@task
def failed_task() -> (list[str], dict[str, str], Fruit):
    return ["apple", "banana"], {"fruit": "banana"}, Fruit(name="banana")


@workflow(
    # The workflow remains unaffected if one of the nodes encounters an error, as long as other executable nodes are still available
    failure_policy=WorkflowFailurePolicy.FAIL_AFTER_EXECUTABLE_NODES_COMPLETE
)
def failed_workflow():
    fruits_list, fruit_dict, fruit_instance = failed_task()
    print_message(message=fruits_list[100])  # Accessing an index that doesn't exist
    print_message(message=fruit_dict["fruits"])  # Accessing a non-existent key
    print_message(message=fruit_instance.fruit)  # Accessing a non-existent param
```
