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

# Enum type

```{eval-rst}
.. tags:: Basic
```

At times, you might need to limit the acceptable values for inputs or outputs to a predefined set.
This common requirement is usually met by using Enum types in programming languages.

You can create a Python Enum type and utilize it as an input or output for a task.
Flytekit will automatically convert it and constrain the inputs and outputs to the predefined set of values.

:::{important}
Currently, only string values are supported as valid enum values.
Flyte assumes the first value in the list as the default, and Enum types cannot be optional.
Therefore, when defining enums, it's important to design them with the first value as a valid default.
:::

To begin, import the dependencies.

```{code-cell}
from enum import Enum

from flytekit import task, workflow
```

+++ {"lines_to_next_cell": 0}

We define an enum and a simple coffee maker workflow that accepts an order and brews coffee ☕️ accordingly.
The assumption is that the coffee maker only understands enum inputs.

```{code-cell}
class Coffee(Enum):
    ESPRESSO = "espresso"
    AMERICANO = "americano"
    LATTE = "latte"
    CAPPUCCINO = "cappucccino"


@task
def take_order(coffee: str) -> Coffee:
    return Coffee(coffee)


@task
def prep_order(coffee_enum: Coffee) -> str:
    return f"Preparing {coffee_enum.value} ..."


@workflow
def coffee_maker(coffee: str) -> str:
    coffee_enum = take_order(coffee=coffee)
    return prep_order(coffee_enum=coffee_enum)
```

+++ {"lines_to_next_cell": 0}

The workflow can also accept an enum value.

```{code-cell}
@workflow
def coffee_maker_enum(coffee_enum: Coffee) -> str:
    return prep_order(coffee_enum=coffee_enum)
```

+++ {"lines_to_next_cell": 0}

You can send a string to the `coffee_maker_enum` workflow during its execution, like this:
```
pyflyte run \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/master/examples/data_types_and_io/data_types_and_io/enum_type.py \
  coffee_maker_enum --coffee_enum="latte"
```

You can run the workflows locally.

```{code-cell}
if __name__ == "__main__":
    print(coffee_maker(coffee="latte"))
    print(coffee_maker_enum(coffee_enum=Coffee.LATTE))
```
