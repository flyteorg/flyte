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

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

To begin, import the dependencies:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/enum_type.py
:caption: data_types_and_io/enum_type.py
:lines: 1-3
```

We define an enum and a simple coffee maker workflow that accepts an order and brews coffee ☕️ accordingly.
The assumption is that the coffee maker only understands enum inputs:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/enum_type.py
:caption: data_types_and_io/enum_type.py
:lines: 9-35
```

The workflow can also accept an enum value:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/enum_type.py
:caption: data_types_and_io/enum_type.py
:pyobject: coffee_maker_enum
```

You can send a string to the `coffee_maker_enum` workflow during its execution, like this:
```
pyflyte run \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/enum_type.py \
  coffee_maker_enum --coffee_enum="latte"
```

You can run the workflows locally:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/enum_type.py
:caption: data_types_and_io/enum_type.py
:lines: 44-46
```

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/data_types_and_io/
