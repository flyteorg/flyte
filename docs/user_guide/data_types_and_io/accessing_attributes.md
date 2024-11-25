(attribute_access)=

# Accessing attributes in workflows

```{eval-rst}
.. tags:: Basic
```

You can directly access attributes on output promises for lists, dictionaries, dataclasses, and combinations of these types in Flyte.
Note that while this functionality may appear to be the normal behavior of Python, code in `@workflow` functions is not actually Python, but rather a Python-like DSL that is compiled by Flyte.
Consequently, accessing attributes in this manner is, in fact, a specially implemented feature.
This functionality facilitates the direct passing of output attributes within workflows, enhancing the convenience of working with complex data structures.

```{important}
Flytekit version >= v1.14.0 supports Pydantic BaseModel V2, you can do attribute access on Pydantic BaseModel V2 as well.
```

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

To begin, import the required dependencies and define a common task for subsequent use:

```{literalinclude} /examples/data_types_and_io/data_types_and_io/attribute_access.py
:caption: data_types_and_io/attribute_access.py
:lines: 1-9
```

## List
You can access an output list using index notation.

:::{important}
Flyte currently does not support output promise access through list slicing.
:::

```{literalinclude} /examples/data_types_and_io/data_types_and_io/attribute_access.py
:caption: data_types_and_io/attribute_access.py
:lines: 13-22
```

## Dictionary
Access the output dictionary by specifying the key.

```{literalinclude} /examples/data_types_and_io/data_types_and_io/attribute_access.py
:caption: data_types_and_io/attribute_access.py
:lines: 26-34
```

## Data class
Directly access an attribute of a dataclass.

```{literalinclude} /examples/data_types_and_io/data_types_and_io/attribute_access.py
:caption: data_types_and_io/attribute_access.py
:lines: 38-51
```

## Complex type
Combinations of list, dict and dataclass also work effectively.

```{literalinclude} /examples/data_types_and_io/data_types_and_io/attribute_access.py
:caption: data_types_and_io/attribute_access.py
:lines: 55-78
```

You can run all the workflows locally as follows:

```{literalinclude} /examples/data_types_and_io/data_types_and_io/attribute_access.py
:caption: data_types_and_io/attribute_access.py
:lines: 82-86
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
