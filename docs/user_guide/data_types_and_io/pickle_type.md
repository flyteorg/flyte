(pickle_type)=

# Pickle type

```{eval-rst}
.. tags:: Basic
```

Flyte enforces type safety by utilizing type information for compiling tasks and workflows,
enabling various features such as static analysis and conditional branching.

However, we also strive to offer flexibility to end-users so they don't have to invest heavily
in understanding their data structures upfront before experiencing the value Flyte has to offer.

Flyte supports the `FlytePickle` transformer, which converts any unrecognized type hint into `FlytePickle`,
enabling the serialization/deserialization of Python values to/from a pickle file.

:::{important}
Pickle can only be used to send objects between the exact same Python version.
For optimal performance, it's advisable to either employ Python types that are supported by Flyte
or register a custom transformer, as using pickle types can result in lower performance.
:::

This example demonstrates how you can utilize custom objects without registering a transformer.

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/pickle_type.py
:caption: data_types_and_io/pickle_type.py
:lines: 1
```

`Superhero` represents a user-defined complex type that can be serialized to a pickle file by Flytekit
and transferred between tasks as both input and output data.

:::{note}
Alternatively, you can {ref}`turn this object into a dataclass <dataclass>` for improved performance.
We have used a simple object here for demonstration purposes.
:::

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/pickle_type.py
:caption: data_types_and_io/pickle_type.py
:lines: 7-26
```

## Batch size

By default, if the list subtype is unrecognized, a single pickle file is generated.
To optimize serialization and deserialization performance for scenarios involving a large number of items
or significant list elements, you can specify a batch size.
This feature allows for the processing of each batch as a separate pickle file.
The following example demonstrates how to set the batch size.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/pickle_type.py
:caption: data_types_and_io/pickle_type.py
:lines: 35-58
```

:::{note}
The `welcome_superheroes` task will generate two pickle files: one containing two superheroes and the other containing one superhero.
:::

You can run the workflows locally as follows:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/data_types_and_io/data_types_and_io/pickle_type.py
:caption: data_types_and_io/pickle_type.py
:lines: 62-64
```

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/data_types_and_io/
