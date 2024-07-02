(attribute_access)=

# Accessing attributes

```{eval-rst}
.. tags:: Basic
```

You can directly access attributes on output promises for lists, dicts, dataclasses and combinations of these types in Flyte. This functionality facilitates the direct passing of output attributes within workflows,
enhancing the convenience of working with complex data structures.

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

To begin, import the required dependencies and define a common task for subsequent use:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/attribute_access.py
:caption: data_types_and_io/attribute_access.py
:lines: 1-10
```

## List
You can access an output list using index notation.

:::{important}
Flyte currently does not support output promise access through list slicing.
:::

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/attribute_access.py
:caption: data_types_and_io/attribute_access.py
:lines: 14-23
```

## Dictionary
Access the output dictionary by specifying the key.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/attribute_access.py
:caption: data_types_and_io/attribute_access.py
:lines: 27-35
```

## Data class
Directly access an attribute of a dataclass.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/attribute_access.py
:caption: data_types_and_io/attribute_access.py
:lines: 39-53
```

## Complex type
Combinations of list, dict and dataclass also work effectively.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/attribute_access.py
:caption: data_types_and_io/attribute_access.py
:lines: 57-80
```

You can run all the workflows locally as follows:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/attribute_access.py
:caption: data_types_and_io/attribute_access.py
:lines: 84-88
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

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/data_types_and_io/
